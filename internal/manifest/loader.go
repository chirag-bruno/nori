package manifest

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadFromBytes loads a manifest from YAML bytes
func LoadFromBytes(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return &m, nil
}

// LoadFromFile loads a manifest from a file path
func LoadFromFile(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", path, err)
	}
	return LoadFromBytes(data)
}

