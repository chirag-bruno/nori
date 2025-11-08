package manifest

// Manifest represents a package manifest
type Manifest struct {
	Schema      int               `yaml:"schema" json:"schema"`
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Homepage    string            `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	License     string            `yaml:"license,omitempty" json:"license,omitempty"`
	Bins        []string          `yaml:"bins" json:"bins"`
	Versions    map[string]Version `yaml:"versions" json:"versions"`
}

// Version represents a specific version of a package
type Version struct {
	Platforms map[string]Asset `yaml:"platforms" json:"platforms"`
}

// Asset represents a downloadable asset for a specific platform
type Asset struct {
	Type     string `yaml:"type" json:"type"`     // tar or zip
	URL      string `yaml:"url" json:"url"`       // HTTPS URL
	Checksum string `yaml:"checksum" json:"checksum"` // sha256:hex format
}

