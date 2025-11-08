package shims

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/chirag-bruno/nori/internal/platform"
)

func TestCreateShimUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix test on Windows")
	}
	
	tmpDir := t.TempDir()
	shimsDir := filepath.Join(tmpDir, "shims")
	os.MkdirAll(shimsDir, 0755)
	
	// Override shims directory for test
	originalShimsDir := platform.ShimsDir
	defer func() {
		_ = originalShimsDir
	}()
	
	targetPath := filepath.Join(tmpDir, "bin", "test")
	os.MkdirAll(filepath.Dir(targetPath), 0755)
	os.WriteFile(targetPath, []byte("#!/bin/sh\necho test"), 0755)
	
	shim := New(shimsDir)
	err := shim.CreateShim("test", targetPath)
	if err != nil {
		t.Fatalf("CreateShim() failed: %v", err)
	}
	
	shimPath := filepath.Join(shimsDir, "test")
	if _, err := os.Lstat(shimPath); os.IsNotExist(err) {
		t.Errorf("Shim was not created at %q", shimPath)
	}
}

func TestCreateShimWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows test on non-Windows")
	}
	
	tmpDir := t.TempDir()
	shimsDir := filepath.Join(tmpDir, "shims")
	os.MkdirAll(shimsDir, 0755)
	
	targetPath := filepath.Join(tmpDir, "bin", "test.exe")
	os.MkdirAll(filepath.Dir(targetPath), 0755)
	os.WriteFile(targetPath, []byte("test binary"), 0644)
	
	shim := New(shimsDir)
	err := shim.CreateShim("test", targetPath)
	if err != nil {
		t.Fatalf("CreateShim() failed: %v", err)
	}
	
	// Check for .cmd file
	cmdPath := filepath.Join(shimsDir, "test.cmd")
	if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
		t.Errorf("Shim .cmd file was not created at %q", cmdPath)
	}
	
	// Check for .ps1 file
	ps1Path := filepath.Join(shimsDir, "test.ps1")
	if _, err := os.Stat(ps1Path); os.IsNotExist(err) {
		t.Errorf("Shim .ps1 file was not created at %q", ps1Path)
	}
}

func TestUpdateShims(t *testing.T) {
	tmpDir := t.TempDir()
	shimsDir := filepath.Join(tmpDir, "shims")
	os.MkdirAll(shimsDir, 0755)
	
	installRoot := filepath.Join(tmpDir, "installs", "testpkg", "1.0.0", "linux-amd64")
	binDir := filepath.Join(installRoot, "bin")
	os.MkdirAll(binDir, 0755)
	
	testBin := filepath.Join(binDir, "test")
	os.WriteFile(testBin, []byte("#!/bin/sh\necho test"), 0755)
	
	shim := New(shimsDir)
	bins := []string{"bin/test"}
	
	err := shim.UpdateShims("testpkg", "1.0.0", bins, installRoot)
	if err != nil {
		t.Fatalf("UpdateShims() failed: %v", err)
	}
	
	// Verify shim was created
	shimPath := filepath.Join(shimsDir, "test")
	if runtime.GOOS == "windows" {
		shimPath = filepath.Join(shimsDir, "test.cmd")
	}
	
	if _, err := os.Stat(shimPath); os.IsNotExist(err) {
		t.Errorf("Shim was not created at %q", shimPath)
	}
}

