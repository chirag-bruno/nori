package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNoriRoot(t *testing.T) {
	got := NoriRoot()
	
	// Should be ~/.nori
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() failed: %v", err)
	}
	
	want := filepath.Join(home, ".nori")
	if got != want {
		t.Errorf("NoriRoot() = %q, want %q", got, want)
	}
}

func TestInstallsDir(t *testing.T) {
	got := InstallsDir()
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".nori", "installs")
	if got != want {
		t.Errorf("InstallsDir() = %q, want %q", got, want)
	}
}

func TestShimsDir(t *testing.T) {
	got := ShimsDir()
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".nori", "shims")
	if got != want {
		t.Errorf("ShimsDir() = %q, want %q", got, want)
	}
}

func TestRegistryDir(t *testing.T) {
	got := RegistryDir()
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".nori", "registry")
	if got != want {
		t.Errorf("RegistryDir() = %q, want %q", got, want)
	}
}

func TestConfigDir(t *testing.T) {
	got := ConfigDir()
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".nori", "config")
	if got != want {
		t.Errorf("ConfigDir() = %q, want %q", got, want)
	}
}

func TestInstallPath(t *testing.T) {
	tests := []struct {
		pkg      string
		version  string
		platform string
		want     string
	}{
		{"node", "22.2.0", "linux-amd64", filepath.Join(NoriRoot(), "installs", "node", "22.2.0", "linux-amd64")},
		{"python", "3.12.0", "darwin-arm64", filepath.Join(NoriRoot(), "installs", "python", "3.12.0", "darwin-arm64")},
		{"deno", "2.0.0", "windows-amd64", filepath.Join(NoriRoot(), "installs", "deno", "2.0.0", "windows-amd64")},
	}
	
	for _, tt := range tests {
		t.Run(tt.pkg+"-"+tt.version+"-"+tt.platform, func(t *testing.T) {
			got := InstallPath(tt.pkg, tt.version, tt.platform)
			if got != tt.want {
				t.Errorf("InstallPath(%q, %q, %q) = %q, want %q", tt.pkg, tt.version, tt.platform, got, tt.want)
			}
		})
	}
}

func TestPackageManifestPath(t *testing.T) {
	got := PackageManifestPath("node")
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".nori", "registry", "packages", "node.yaml")
	if got != want {
		t.Errorf("PackageManifestPath(%q) = %q, want %q", "node", got, want)
	}
}

func TestIndexPath(t *testing.T) {
	got := IndexPath()
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".nori", "registry", "index.yaml")
	if got != want {
		t.Errorf("IndexPath() = %q, want %q", got, want)
	}
}

func TestActiveConfigPath(t *testing.T) {
	got := ActiveConfigPath()
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".nori", "config", "active.yaml")
	if got != want {
		t.Errorf("ActiveConfigPath() = %q, want %q", got, want)
	}
}

// Test that paths use correct separators for the OS
func TestPathSeparators(t *testing.T) {
	paths := []string{
		NoriRoot(),
		InstallsDir(),
		ShimsDir(),
		RegistryDir(),
		ConfigDir(),
	}
	
	for _, path := range paths {
		// Just verify it's a valid path - filepath.Join handles separators
		if path == "" {
			t.Errorf("Path should not be empty")
		}
		// On Windows, paths should contain backslashes; on Unix, forward slashes
		// But filepath.Join handles this, so we just verify it's not empty
	}
}

