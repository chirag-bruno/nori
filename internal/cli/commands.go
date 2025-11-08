package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/chirag-bruno/nori/internal/config"
	"github.com/chirag-bruno/nori/internal/extract"
	"github.com/chirag-bruno/nori/internal/fetch"
	"github.com/chirag-bruno/nori/internal/install"
	"github.com/chirag-bruno/nori/internal/manifest"
	"github.com/chirag-bruno/nori/internal/platform"
	"github.com/chirag-bruno/nori/internal/registry"
	"github.com/chirag-bruno/nori/internal/shims"
	urfavecli "github.com/urfave/cli/v3"
)

var (
	style = lipgloss.NewStyle().Bold(true)
)

// InitCommand handles the `nori init` command
func InitCommand(ctx context.Context, c *urfavecli.Command) error {
	shell := detectShell()
	shimsDir := platform.ShimsDir()

	// Ensure shims directory exists
	if err := os.MkdirAll(shimsDir, 0755); err != nil {
		return fmt.Errorf("failed to create shims directory: %w", err)
	}

	var profilePath string
	var pathLine string
	var added bool
	var err error

	switch shell {
	case "zsh":
		home, _ := os.UserHomeDir()
		profilePath = filepath.Join(home, ".zshrc")
		pathLine = `export PATH="$HOME/.nori/shims:$PATH"`
		added, err = addToProfile(profilePath, pathLine)
	case "bash":
		home, _ := os.UserHomeDir()
		profilePath = filepath.Join(home, ".bashrc")
		pathLine = `export PATH="$HOME/.nori/shims:$PATH"`
		added, err = addToProfile(profilePath, pathLine)
	case "fish":
		home, _ := os.UserHomeDir()
		profilePath = filepath.Join(home, ".config", "fish", "config.fish")
		pathLine = `set -gx PATH $HOME/.nori/shims $PATH`
		added, err = addToProfile(profilePath, pathLine)
	case "powershell":
		profilePath = os.Getenv("PROFILE")
		if profilePath == "" {
			home, _ := os.UserHomeDir()
			profilePath = filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
		}
		pathLine = `$env:PATH = "$HOME\.nori\shims;" + $env:PATH`
		added, err = addToProfile(profilePath, pathLine)
	default:
		fmt.Printf("Unable to detect shell. Please manually add %s to your PATH.\n", shimsDir)
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to update %s profile: %w", shell, err)
	}

	if added {
		fmt.Printf("✓ Added nori shims to PATH in %s\n", profilePath)
		fmt.Printf("\nPlease run: source %s\n", profilePath)
		if shell == "powershell" {
			fmt.Printf("Or restart your PowerShell session.\n")
		}
	} else {
		fmt.Printf("✓ nori shims already in PATH\n")
	}

	fmt.Printf("\nShims directory: %s\n", shimsDir)

	return nil
}

// addToProfile adds a line to a shell profile file if it doesn't already exist
func addToProfile(profilePath, line string) (bool, error) {
	// Read existing profile
	data, err := os.ReadFile(profilePath)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	// Check if line already exists
	content := string(data)
	if strings.Contains(content, line) || strings.Contains(content, ".nori/shims") {
		return false, nil // Already added
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(profilePath), 0755); err != nil {
		return false, err
	}

	// Append line to profile
	var newContent string
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		newContent = content + "\n" + line + "\n"
	} else {
		newContent = content + line + "\n"
	}

	if err := os.WriteFile(profilePath, []byte(newContent), 0644); err != nil {
		return false, err
	}

	return true, nil
}

// UpdateCommand handles the `nori update` command
func UpdateCommand(ctx context.Context, c *urfavecli.Command) error {
	reg := registry.NewFromEnv()

	fmt.Println("Updating registry...")
	if err := reg.Update(ctx); err != nil {
		return fmt.Errorf("failed to update registry: %w", err)
	}

	fmt.Println("Registry updated successfully")
	return nil
}

// SearchCommand handles the `nori search` command
func SearchCommand(ctx context.Context, c *urfavecli.Command) error {
	if c.NArg() == 0 {
		return fmt.Errorf("usage: nori search <query>")
	}

	query := c.Args().Get(0)
	reg := registry.NewFromEnv()

	results, err := reg.Search(ctx, query)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Printf("No packages found matching %q\n", query)
		return nil
	}

	fmt.Printf("Found %d package(s):\n\n", len(results))
	for _, pkg := range results {
		fmt.Printf("  %s - %s\n", style.Render(pkg.Name), pkg.Description)
	}

	return nil
}

// InfoCommand handles the `nori info` command
func InfoCommand(ctx context.Context, c *urfavecli.Command) error {
	if c.NArg() == 0 {
		return fmt.Errorf("usage: nori info <package>")
	}

	pkgName := c.Args().Get(0)
	reg := registry.NewFromEnv()

	m, err := reg.LoadPackage(ctx, pkgName)
	if err != nil {
		return fmt.Errorf("failed to load package: %w", err)
	}

	fmt.Printf("%s: %s\n", style.Render(m.Name), m.Description)
	if m.Homepage != "" {
		fmt.Printf("Homepage: %s\n", m.Homepage)
	}
	if m.License != "" {
		fmt.Printf("License: %s\n", m.License)
	}

	fmt.Printf("\nBinaries: %s\n", strings.Join(m.Bins, ", "))

	fmt.Printf("\nVersions:\n")
	for version := range m.Versions {
		fmt.Printf("  %s\n", version)
	}

	return nil
}

// InstallCommand handles the `nori install` command
func InstallCommand(ctx context.Context, c *urfavecli.Command) error {
	if c.NArg() == 0 {
		return fmt.Errorf("usage: nori install <package>@<version>")
	}

	arg := c.Args().Get(0)
	parts := strings.Split(arg, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format: expected <package>@<version>")
	}

	pkgName, version := parts[0], parts[1]

	reg := registry.NewFromEnv()

	// Load manifest
	m, err := reg.LoadPackage(ctx, pkgName)
	if err != nil {
		return fmt.Errorf("failed to load package: %w", err)
	}

	// Detect platform
	p := platform.Detect()
	platformStr := p.String()

	// Validate version/platform
	if err := manifest.ValidateVersion(m, version, platformStr); err != nil {
		return err
	}

	// Get asset
	asset, err := m.GetAsset(version, platformStr)
	if err != nil {
		return err
	}

	fmt.Printf("Installing %s@%s for %s...\n", pkgName, version, platformStr)

	// Fetch with progress
	fetcher := fetch.New()
	
	// Get content length for progress bar
	var totalSize int64
	req, _ := http.NewRequestWithContext(ctx, "HEAD", asset.URL, nil)
	if resp, err := http.DefaultClient.Do(req); err == nil {
		totalSize = resp.ContentLength
		resp.Body.Close()
	}
	
	downloadBar := NewProgressBar(totalSize, "Downloading")
	data, err := fetcher.FetchWithProgress(ctx, asset.URL, asset.Checksum, downloadBar)
	if err != nil {
		downloadBar.Finish()
		fmt.Fprintf(os.Stderr, "\nError: download failed: %v\n", err)
		return fmt.Errorf("download failed: %w", err)
	}
	downloadBar.Finish()

	// Extract with progress
	extractor := extract.New()
	
	// File count progress (unknown total, will show count)
	extractBar := NewFileProgressBar(0, "Extracting")
	fileCount := 0
	
	extractDir, err := extractor.ExtractWithProgress(data, asset.Type, asset.Checksum, func() {
		fileCount++
		extractBar.SetCurrent(fileCount)
	})
	if err != nil {
		extractBar.Finish()
		fmt.Fprintf(os.Stderr, "\nError: extraction failed: %v\n", err)
		return fmt.Errorf("extraction failed: %w", err)
	}
	extractBar.Finish()
	defer os.RemoveAll(extractDir)

	// Install
	installer := install.New()
	fmt.Println("Installing...")
	installPath, err := installer.Install(ctx, m, version, p, extractDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: installation failed: %v\n", err)
		return fmt.Errorf("installation failed: %w", err)
	}

	// Create shims
	shimsDir := platform.ShimsDir()
	shim := shims.New(shimsDir)
	if err := shim.UpdateShims(pkgName, version, m.Bins, installPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create shims: %v\n", err)
		return fmt.Errorf("failed to create shims: %w", err)
	}

	fmt.Printf("Installed %s@%s to %s\n", pkgName, version, installPath)
	return nil
}

// UseCommand handles the `nori use` command
func UseCommand(ctx context.Context, c *urfavecli.Command) error {
	if c.NArg() == 0 {
		return fmt.Errorf("usage: nori use <package>@<version>")
	}

	arg := c.Args().Get(0)
	parts := strings.Split(arg, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format: expected <package>@<version>")
	}

	pkgName, version := parts[0], parts[1]

	// Load manifest and validate version exists
	reg := registry.NewFromEnv()
	m, err := reg.LoadPackage(ctx, pkgName)
	if err != nil {
		return fmt.Errorf("failed to load package: %w", err)
	}

	// Detect platform and validate version/platform
	p := platform.Detect()
	platformStr := p.String()
	if err := manifest.ValidateVersion(m, version, platformStr); err != nil {
		return fmt.Errorf("version %q does not exist for package %q on platform %q", version, pkgName, platformStr)
	}

	// Verify installation exists
	installPath := platform.InstallPath(pkgName, version, p.String())
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		return fmt.Errorf("package %s@%s is not installed", pkgName, version)
	}

	// Set active
	if err := config.SetActive(pkgName, version); err != nil {
		return fmt.Errorf("failed to set active version: %w", err)
	}

	// Update shims (use manifest we already loaded)

	shimsDir := platform.ShimsDir()
	shim := shims.New(shimsDir)
	if err := shim.UpdateShims(pkgName, version, m.Bins, installPath); err != nil {
		return fmt.Errorf("failed to update shims: %w", err)
	}

	fmt.Printf("Using %s@%s\n", pkgName, version)
	return nil
}

// ListCommand handles the `nori list` command
func ListCommand(ctx context.Context, c *urfavecli.Command) error {
	pkgName := ""
	if c.NArg() > 0 {
		pkgName = c.Args().Get(0)
	}

	p := platform.Detect()
	installsDir := platform.InstallsDir()

	if pkgName != "" {
		// List versions for specific package
		pkgDir := filepath.Join(installsDir, pkgName)
		entries, err := os.ReadDir(pkgDir)
		if os.IsNotExist(err) {
			fmt.Printf("Package %s is not installed\n", pkgName)
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to read installs: %w", err)
		}

		fmt.Printf("Installed versions of %s:\n", pkgName)
		for _, entry := range entries {
			if entry.IsDir() {
				versionDir := filepath.Join(pkgDir, entry.Name())
				platformDir := filepath.Join(versionDir, p.String())
				if _, err := os.Stat(platformDir); err == nil {
					active, _ := config.GetActive(pkgName)
					marker := ""
					if active == entry.Name() {
						marker = " (active)"
					}
					fmt.Printf("  %s%s\n", entry.Name(), marker)
				}
			}
		}
	} else {
		// List all installed packages
		entries, err := os.ReadDir(installsDir)
		if os.IsNotExist(err) {
			fmt.Println("No packages installed")
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to read installs: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				fmt.Printf("  %s\n", entry.Name())
			}
		}
	}

	return nil
}

// WhichCommand handles the `nori which` command
func WhichCommand(ctx context.Context, c *urfavecli.Command) error {
	if c.NArg() == 0 {
		return fmt.Errorf("usage: nori which <binary>")
	}

	binName := c.Args().Get(0)

	// Find which package provides this binary
	reg := registry.NewFromEnv()

	// Load index to find packages
	results, err := reg.Search(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to search registry: %w", err)
	}

	var pkgName string
	for _, pkg := range results {
		m, err := reg.LoadPackage(ctx, pkg.Name)
		if err != nil {
			continue
		}
		for _, bin := range m.Bins {
			if filepath.Base(bin) == binName {
				pkgName = pkg.Name
				break
			}
		}
		if pkgName != "" {
			break
		}
	}

	if pkgName == "" {
		return fmt.Errorf("binary %q not found in any package", binName)
	}

	// Get active version
	version, err := config.GetActive(pkgName)
	if err != nil || version == "" {
		return fmt.Errorf("package %s has no active version", pkgName)
	}

	// Resolve path
	p := platform.Detect()
	installPath := platform.InstallPath(pkgName, version, p.String())

	m, err := reg.LoadPackage(ctx, pkgName)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Find bin path
	var binPath string
	for _, bin := range m.Bins {
		if filepath.Base(bin) == binName {
			binPath = filepath.Join(installPath, bin)
			break
		}
	}

	if binPath == "" {
		return fmt.Errorf("binary %q not found in package %s", binName, pkgName)
	}

	fmt.Println(binPath)
	return nil
}

// detectShell detects the current shell
func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		if runtime.GOOS == "windows" {
			return "powershell"
		}
		return "bash"
	}

	if strings.Contains(shell, "zsh") {
		return "zsh"
	}
	if strings.Contains(shell, "fish") {
		return "fish"
	}
	if strings.Contains(shell, "bash") {
		return "bash"
	}

	return "bash"
}
