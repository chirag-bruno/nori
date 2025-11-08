package shims

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Shims manages shim creation and updates
type Shims struct {
	shimsDir string
}

// New creates a new shims manager
func New(shimsDir string) *Shims {
	return &Shims{
		shimsDir: shimsDir,
	}
}

// CreateShim creates a shim for a binary
func (s *Shims) CreateShim(binName, targetPath string) error {
	// Ensure shims directory exists
	if err := os.MkdirAll(s.shimsDir, 0755); err != nil {
		return fmt.Errorf("failed to create shims directory: %w", err)
	}
	
	if runtime.GOOS == "windows" {
		return s.createWindowsShim(binName, targetPath)
	}
	
	return s.createUnixShim(binName, targetPath)
}

// createUnixShim creates a symlink or wrapper script on Unix
func (s *Shims) createUnixShim(binName, targetPath string) error {
	shimPath := filepath.Join(s.shimsDir, binName)
	
	// Try symlink first
	if err := os.Symlink(targetPath, shimPath); err == nil {
		return nil
	}
	
	// Fallback to wrapper script if symlink fails
	script := fmt.Sprintf(`#!/bin/sh
exec "%s" "$@"
`, targetPath)
	
	return os.WriteFile(shimPath, []byte(script), 0755)
}

// createWindowsShim creates .cmd and .ps1 wrappers on Windows
func (s *Shims) createWindowsShim(binName, targetPath string) error {
	// Create .cmd wrapper
	cmdPath := filepath.Join(s.shimsDir, binName+".cmd")
	cmdScript := fmt.Sprintf(`@echo off
"%s" %%*
`, targetPath)
	if err := os.WriteFile(cmdPath, []byte(cmdScript), 0644); err != nil {
		return fmt.Errorf("failed to create .cmd shim: %w", err)
	}
	
	// Create .ps1 wrapper
	ps1Path := filepath.Join(s.shimsDir, binName+".ps1")
	ps1Script := fmt.Sprintf(`& "%s" $args
`, targetPath)
	if err := os.WriteFile(ps1Path, []byte(ps1Script), 0644); err != nil {
		return fmt.Errorf("failed to create .ps1 shim: %w", err)
	}
	
	return nil
}

// UpdateShims updates shims for a package version
func (s *Shims) UpdateShims(pkg, version string, bins []string, installRoot string) error {
	for _, bin := range bins {
		// Get basename of bin path
		binName := filepath.Base(bin)
		
		// Resolve full target path
		targetPath := filepath.Join(installRoot, bin)
		
		// On Windows, append .exe if not present
		if runtime.GOOS == "windows" {
			if filepath.Ext(targetPath) != ".exe" {
				// Check if .exe version exists
				exePath := targetPath + ".exe"
				if _, err := os.Stat(exePath); err == nil {
					targetPath = exePath
				}
			}
		}
		
		// Verify target exists
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			return fmt.Errorf("target binary %q does not exist", targetPath)
		}
		
		// Create or update shim
		if err := s.CreateShim(binName, targetPath); err != nil {
			return fmt.Errorf("failed to create shim for %q: %w", binName, err)
		}
	}
	
	return nil
}

// RemoveShims removes shims for given binary names
func (s *Shims) RemoveShims(binNames []string) error {
	for _, binName := range binNames {
		shimPath := filepath.Join(s.shimsDir, binName)
		if err := os.Remove(shimPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove shim %q: %w", binName, err)
		}
		
		// On Windows, also remove .cmd and .ps1
		if runtime.GOOS == "windows" {
			os.Remove(shimPath + ".cmd")
			os.Remove(shimPath + ".ps1")
		}
	}
	
	return nil
}

