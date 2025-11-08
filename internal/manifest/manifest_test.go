package manifest

import (
	"testing"
)

func TestManifestUnmarshal(t *testing.T) {
	yamlData := `
schema: 1
name: node
description: Node.js runtime
homepage: https://nodejs.org
license: MIT
bins:
  - bin/node
  - bin/npm
  - bin/npx
versions:
  "22.2.0":
    platforms:
      linux-amd64:
        type: tar
        url: https://nodejs.org/dist/v22.2.0/node-v22.2.0-linux-x64.tar.xz
        checksum: sha256:5f4a1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab
      darwin-arm64:
        type: tar
        url: https://nodejs.org/dist/v22.2.0/node-v22.2.0-darwin-arm64.tar.gz
        checksum: sha256:9a2c1234567890abcdef1234567890abcdef1234567890abcdef1234567890cd
`

	m, err := LoadFromBytes([]byte(yamlData))
	if err != nil {
		t.Fatalf("LoadFromBytes() failed: %v", err)
	}
	
	if m.Schema != 1 {
		t.Errorf("Schema = %d, want 1", m.Schema)
	}
	if m.Name != "node" {
		t.Errorf("Name = %q, want %q", m.Name, "node")
	}
	if m.Description != "Node.js runtime" {
		t.Errorf("Description = %q, want %q", m.Description, "Node.js runtime")
	}
	if len(m.Bins) != 3 {
		t.Errorf("Bins length = %d, want 3", len(m.Bins))
	}
	if m.Bins[0] != "bin/node" {
		t.Errorf("Bins[0] = %q, want %q", m.Bins[0], "bin/node")
	}
	
	version, ok := m.Versions["22.2.0"]
	if !ok {
		t.Fatal("Version 22.2.0 not found")
	}
	
	linuxAsset, ok := version.Platforms["linux-amd64"]
	if !ok {
		t.Fatal("linux-amd64 platform not found")
	}
	if linuxAsset.Type != "tar" {
		t.Errorf("Asset type = %q, want %q", linuxAsset.Type, "tar")
	}
	if linuxAsset.URL != "https://nodejs.org/dist/v22.2.0/node-v22.2.0-linux-x64.tar.xz" {
		t.Errorf("Asset URL = %q, want expected URL", linuxAsset.URL)
	}
	if linuxAsset.Checksum != "sha256:5f4a1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab" {
		t.Errorf("Asset checksum = %q, want expected checksum", linuxAsset.Checksum)
	}
}

func TestManifestMinimal(t *testing.T) {
	yamlData := `
schema: 1
name: test
bins:
  - bin/test
versions:
  "1.0.0":
    platforms:
      linux-amd64:
        type: tar
        url: https://example.com/test.tar.gz
        checksum: sha256:abcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ef
`
	
	m, err := LoadFromBytes([]byte(yamlData))
	if err != nil {
		t.Fatalf("LoadFromBytes() failed: %v", err)
	}
	
	if m.Name != "test" {
		t.Errorf("Name = %q, want %q", m.Name, "test")
	}
	if m.Description != "" {
		t.Errorf("Description should be empty for minimal manifest, got %q", m.Description)
	}
}

func TestManifestMultipleVersions(t *testing.T) {
	yamlData := `
schema: 1
name: node
bins:
  - bin/node
versions:
  "22.2.0":
    platforms:
      linux-amd64:
        type: tar
        url: https://example.com/v22.2.0.tar.gz
        checksum: sha256:abcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ef
  "20.5.1":
    platforms:
      linux-amd64:
        type: tar
        url: https://example.com/v20.5.1.tar.gz
        checksum: sha256:efab1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab
`
	
	m, err := LoadFromBytes([]byte(yamlData))
	if err != nil {
		t.Fatalf("LoadFromBytes() failed: %v", err)
	}
	
	if len(m.Versions) != 2 {
		t.Errorf("Versions count = %d, want 2", len(m.Versions))
	}
	
	if _, ok := m.Versions["22.2.0"]; !ok {
		t.Error("Version 22.2.0 not found")
	}
	if _, ok := m.Versions["20.5.1"]; !ok {
		t.Error("Version 20.5.1 not found")
	}
}

