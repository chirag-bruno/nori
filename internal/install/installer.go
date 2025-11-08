package install

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/chirag-bruno/nori/internal/extract"
	"github.com/chirag-bruno/nori/internal/manifest"
	"github.com/chirag-bruno/nori/internal/platform"
)

// Installer handles package installation
type Installer struct{}

// New creates a new installer
func New() *Installer {
	return &Installer{}
}

// Install installs a package from an extracted directory to the install location
func (i *Installer) Install(ctx context.Context, m *manifest.Manifest, version string, p platform.Platform, extractDir string) (string, error) {
	// Validate version and platform
	if err := manifest.ValidateVersion(m, version, p.String()); err != nil {
		return "", err
	}
	
	// Detect archive root
	rootDir, err := extract.DetectRoot(extractDir)
	if err != nil {
		return "", fmt.Errorf("failed to detect archive root: %w", err)
	}
	
	// Validate that all bins exist
	for _, bin := range m.Bins {
		binPath := filepath.Join(rootDir, bin)
		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			return "", fmt.Errorf("bin %q not found in extracted archive", bin)
		}
	}
	
	// Create install directory
	installPath := platform.InstallPath(m.Name, version, p.String())
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create install directory: %w", err)
	}
	
	// Move contents from rootDir to installPath
	if err := moveContents(rootDir, installPath); err != nil {
		// Cleanup on failure
		os.RemoveAll(installPath)
		return "", fmt.Errorf("failed to move contents: %w", err)
	}
	
	// Set executable bits on bin files (POSIX only)
	if runtime.GOOS != "windows" {
		for _, bin := range m.Bins {
			binPath := filepath.Join(installPath, bin)
			if info, err := os.Stat(binPath); err == nil {
				mode := info.Mode()
				if mode&0111 == 0 {
					os.Chmod(binPath, mode|0111)
				}
			}
		}
	}
	
	return installPath, nil
}

// moveContents moves all contents from src to dst
func moveContents(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		
		if err := os.Rename(srcPath, dstPath); err != nil {
			// If rename fails (cross-device), fall back to copy+remove
			if err := copyRecursive(srcPath, dstPath); err != nil {
				return err
			}
			os.RemoveAll(srcPath)
		}
	}
	
	return nil
}

// copyRecursive copies a file or directory recursively
func copyRecursive(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	if info.IsDir() {
		if err := os.MkdirAll(dst, info.Mode()); err != nil {
			return err
		}
		
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		
		for _, entry := range entries {
			if err := copyRecursive(filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name())); err != nil {
				return err
			}
		}
		return nil
	}
	
	// Copy file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()
	
	_, err = io.Copy(dstFile, srcFile)
	return err
}

