package config

import (
	"os"
	"testing"

	"github.com/chirag-bruno/nori/internal/platform"
)

func TestGetActive(t *testing.T) {
	// Use real config directory but clean up after
	activePath := platform.ActiveConfigPath()
	defer os.Remove(activePath)
	
	// Create active.yaml
	configDir := platform.ConfigDir()
	os.MkdirAll(configDir, 0755)
	os.WriteFile(activePath, []byte(`node: "22.2.0"
python: "3.12.0"
`), 0644)
	
	// Test reading
	version, err := GetActive("node")
	if err != nil {
		t.Fatalf("GetActive() failed: %v", err)
	}
	if version != "22.2.0" {
		t.Errorf("GetActive() = %q, want %q", version, "22.2.0")
	}
	
	version, err = GetActive("python")
	if err != nil {
		t.Fatalf("GetActive() failed: %v", err)
	}
	if version != "3.12.0" {
		t.Errorf("GetActive() = %q, want %q", version, "3.12.0")
	}
	
	// Test non-existent package
	version, err = GetActive("nonexistent")
	if err != nil {
		t.Fatalf("GetActive() should not fail for non-existent package")
	}
	if version != "" {
		t.Errorf("GetActive() for non-existent = %q, want empty", version)
	}
}

func TestSetActive(t *testing.T) {
	// Use real config directory but clean up after
	activePath := platform.ActiveConfigPath()
	defer os.Remove(activePath)
	
	err := SetActive("node", "22.2.0")
	if err != nil {
		t.Fatalf("SetActive() failed: %v", err)
	}
	
	// Verify it was written
	version, err := GetActive("node")
	if err != nil {
		t.Fatalf("GetActive() failed: %v", err)
	}
	if version != "22.2.0" {
		t.Errorf("GetActive() = %q, want %q", version, "22.2.0")
	}
	
	// Update to new version
	err = SetActive("node", "20.5.1")
	if err != nil {
		t.Fatalf("SetActive() failed: %v", err)
	}
	
	version, err = GetActive("node")
	if err != nil {
		t.Fatalf("GetActive() failed: %v", err)
	}
	if version != "20.5.1" {
		t.Errorf("GetActive() = %q, want %q", version, "20.5.1")
	}
}

func TestListActive(t *testing.T) {
	// Use real config directory but clean up after
	activePath := platform.ActiveConfigPath()
	defer os.Remove(activePath)
	
	// Set multiple active versions
	SetActive("node", "22.2.0")
	SetActive("python", "3.12.0")
	
	active, err := ListActive()
	if err != nil {
		t.Fatalf("ListActive() failed: %v", err)
	}
	
	if len(active) != 2 {
		t.Errorf("ListActive() count = %d, want 2", len(active))
	}
	
	if active["node"] != "22.2.0" {
		t.Errorf("ListActive() node = %q, want %q", active["node"], "22.2.0")
	}
	
	if active["python"] != "3.12.0" {
		t.Errorf("ListActive() python = %q, want %q", active["python"], "3.12.0")
	}
}

