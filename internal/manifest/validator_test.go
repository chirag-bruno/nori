package manifest

import (
	"testing"
)

func TestValidateValidManifest(t *testing.T) {
	yamlData := `
schema: 1
name: node
description: Node.js runtime
homepage: https://nodejs.org
license: MIT
bins:
  - bin/node
  - bin/npm
versions:
  "22.2.0":
    platforms:
      linux-amd64:
        type: tar
        url: https://nodejs.org/dist/v22.2.0/node-v22.2.0-linux-x64.tar.xz
        checksum: sha256:5f4a1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab
`
	
	m, err := LoadFromBytes([]byte(yamlData))
	if err != nil {
		t.Fatalf("LoadFromBytes() failed: %v", err)
	}
	
	if err := Validate(m); err != nil {
		t.Errorf("Validate() failed for valid manifest: %v", err)
	}
}

func TestValidateInvalidSchema(t *testing.T) {
	yamlData := `
schema: 2
name: node
bins:
  - bin/node
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
	
	if err := Validate(m); err == nil {
		t.Error("Validate() should fail for invalid schema")
	}
}

func TestValidateMissingName(t *testing.T) {
	yamlData := `
schema: 1
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
	
	if err := Validate(m); err == nil {
		t.Error("Validate() should fail for missing name")
	}
}

func TestValidateEmptyBins(t *testing.T) {
	yamlData := `
schema: 1
name: test
bins: []
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
	
	if err := Validate(m); err == nil {
		t.Error("Validate() should fail for empty bins")
	}
}

func TestValidateInvalidPlatform(t *testing.T) {
	yamlData := `
schema: 1
name: test
bins:
  - bin/test
versions:
  "1.0.0":
    platforms:
      invalid-platform:
        type: tar
        url: https://example.com/test.tar.gz
        checksum: sha256:abcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ef
`
	
	m, err := LoadFromBytes([]byte(yamlData))
	if err != nil {
		t.Fatalf("LoadFromBytes() failed: %v", err)
	}
	
	if err := Validate(m); err == nil {
		t.Error("Validate() should fail for invalid platform")
	}
}

func TestValidateInvalidAssetType(t *testing.T) {
	yamlData := `
schema: 1
name: test
bins:
  - bin/test
versions:
  "1.0.0":
    platforms:
      linux-amd64:
        type: rar
        url: https://example.com/test.rar
        checksum: sha256:abcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ef
`
	
	m, err := LoadFromBytes([]byte(yamlData))
	if err != nil {
		t.Fatalf("LoadFromBytes() failed: %v", err)
	}
	
	if err := Validate(m); err == nil {
		t.Error("Validate() should fail for invalid asset type")
	}
}

func TestValidateNonHTTPSURL(t *testing.T) {
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
        url: http://example.com/test.tar.gz
        checksum: sha256:abcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ef
`
	
	m, err := LoadFromBytes([]byte(yamlData))
	if err != nil {
		t.Fatalf("LoadFromBytes() failed: %v", err)
	}
	
	if err := Validate(m); err == nil {
		t.Error("Validate() should fail for non-HTTPS URL")
	}
}

func TestValidateInvalidChecksumFormat(t *testing.T) {
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
        checksum: md5:abcd1234
`
	
	m, err := LoadFromBytes([]byte(yamlData))
	if err != nil {
		t.Fatalf("LoadFromBytes() failed: %v", err)
	}
	
	if err := Validate(m); err == nil {
		t.Error("Validate() should fail for invalid checksum format")
	}
}

func TestValidateInvalidChecksumLength(t *testing.T) {
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
        checksum: sha256:abcd1234
`
	
	m, err := LoadFromBytes([]byte(yamlData))
	if err != nil {
		t.Fatalf("LoadFromBytes() failed: %v", err)
	}
	
	if err := Validate(m); err == nil {
		t.Error("Validate() should fail for invalid checksum length")
	}
}

func TestValidateInvalidVersionFormat(t *testing.T) {
	yamlData := `
schema: 1
name: test
bins:
  - bin/test
versions:
  "v1.0.0":
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
	
	if err := Validate(m); err == nil {
		t.Error("Validate() should fail for invalid version format")
	}
}

