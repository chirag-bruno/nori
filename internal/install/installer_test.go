package install

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/chirag-bruno/nori/internal/manifest"
	"github.com/chirag-bruno/nori/internal/platform"
)

func TestInstall(t *testing.T) {
	// Create a temporary extract directory with test files
	// Simulate an archive with a single top-level directory
	extractDir := t.TempDir()
	
	// Create a package directory (like archives typically have)
	pkgDir := filepath.Join(extractDir, "testpkg-1.0.0")
	binDir := filepath.Join(pkgDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create bin directory: %v", err)
	}
	testBin := filepath.Join(binDir, "test")
	if err := os.WriteFile(testBin, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}
	
	// Verify file exists
	if _, err := os.Stat(testBin); os.IsNotExist(err) {
		t.Fatalf("Test binary was not created at %q", testBin)
	}
	
	// Create manifest with current platform
	p := platform.Detect()
	platformStr := p.String()
	
	m := &manifest.Manifest{
		Schema: 1,
		Name:   "testpkg",
		Bins:   []string{"bin/test"},
		Versions: map[string]manifest.Version{
			"1.0.0": {
				Platforms: map[string]manifest.Asset{
					platformStr: {
						Type:     "tar",
						URL:      "https://example.com/test.tar.gz",
						Checksum: "sha256:abcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab",
					},
				},
			},
		},
	}
	
	installer := New()
	ctx := context.Background()
	
	installPath, err := installer.Install(ctx, m, "1.0.0", p, extractDir)
	if err != nil {
		t.Fatalf("Install() failed: %v", err)
	}
	
	// Verify install path
	expectedPath := platform.InstallPath("testpkg", "1.0.0", p.String())
	if installPath != expectedPath {
		t.Errorf("Install() path = %q, want %q", installPath, expectedPath)
	}
	
	// Verify bin file exists
	binPath := filepath.Join(installPath, "bin", "test")
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Errorf("bin file not found at %q", binPath)
	}
	
	// Verify executable bit is set (on Unix)
	if runtime.GOOS != "windows" {
		info, _ := os.Stat(binPath)
		if info.Mode()&0111 == 0 {
			t.Error("bin file should be executable")
		}
	}
}

func TestInstallMissingBin(t *testing.T) {
	extractDir := t.TempDir()
	
	m := &manifest.Manifest{
		Schema: 1,
		Name:   "testpkg",
		Bins:   []string{"bin/missing"},
		Versions: map[string]manifest.Version{
			"1.0.0": {
				Platforms: map[string]manifest.Asset{
					"linux-amd64": {
						Type:     "tar",
						URL:      "https://example.com/test.tar.gz",
						Checksum: "sha256:abcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab",
					},
				},
			},
		},
	}
	
	installer := New()
	ctx := context.Background()
	p := platform.Detect()
	
	_, err := installer.Install(ctx, m, "1.0.0", p, extractDir)
	if err == nil {
		t.Error("Install() should fail when bin is missing")
	}
}

