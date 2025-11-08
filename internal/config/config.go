package config

import (
	"fmt"
	"os"

	"github.com/chirag-bruno/nori/internal/platform"
	"gopkg.in/yaml.v3"
)

// ActiveConfig represents the active versions configuration
type ActiveConfig map[string]string

// GetActive returns the active version for a package
func GetActive(pkg string) (string, error) {
	active, err := loadActive()
	if err != nil {
		return "", err
	}
	
	return active[pkg], nil
}

// SetActive sets the active version for a package
func SetActive(pkg, version string) error {
	active, err := loadActive()
	if err != nil {
		active = make(ActiveConfig)
	}
	
	active[pkg] = version
	
	return saveActive(active)
}

// ListActive returns all active versions
func ListActive() (ActiveConfig, error) {
	return loadActive()
}

// loadActive loads the active.yaml file
func loadActive() (ActiveConfig, error) {
	activePath := platform.ActiveConfigPath()
	
	data, err := os.ReadFile(activePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(ActiveConfig), nil
		}
		return nil, fmt.Errorf("failed to read active config: %w", err)
	}
	
	var active ActiveConfig
	if err := yaml.Unmarshal(data, &active); err != nil {
		return nil, fmt.Errorf("failed to parse active config: %w", err)
	}
	
	if active == nil {
		active = make(ActiveConfig)
	}
	
	return active, nil
}

// saveActive saves the active.yaml file
func saveActive(active ActiveConfig) error {
	activePath := platform.ActiveConfigPath()
	
	// Ensure config directory exists
	configDir := platform.ConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	data, err := yaml.Marshal(active)
	if err != nil {
		return fmt.Errorf("failed to marshal active config: %w", err)
	}
	
	if err := os.WriteFile(activePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write active config: %w", err)
	}
	
	return nil
}

