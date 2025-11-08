package manifest

import (
	"fmt"
	"net/url"
	"regexp"
)

// Validate validates a manifest with basic YAML validation rules
func Validate(m *Manifest) error {
	// Validate required fields
	if m.Schema == 0 {
		return fmt.Errorf("missing required field: schema")
	}
	if m.Schema != 1 {
		return fmt.Errorf("unsupported schema version: %d (expected 1)", m.Schema)
	}

	if m.Name == "" {
		return fmt.Errorf("missing required field: name")
	}

	if len(m.Bins) == 0 {
		return fmt.Errorf("missing required field: bins (at least one binary required)")
	}

	if len(m.Versions) == 0 {
		return fmt.Errorf("missing required field: versions (at least one version required)")
	}

	// Validate name pattern
	namePattern := regexp.MustCompile(`^[a-z0-9][a-z0-9-_]{1,63}$`)
	if !namePattern.MatchString(m.Name) {
		return fmt.Errorf("invalid package name: must match pattern ^[a-z0-9][a-z0-9-_]{1,63}$")
	}

	// Validate bins
	for i, bin := range m.Bins {
		if bin == "" {
			return fmt.Errorf("empty binary path at index %d", i)
		}
	}

	// Validate version format and platform keys
	versionPattern := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
	platformPattern := regexp.MustCompile(`^(linux|darwin|windows)-(amd64|arm64)$`)

	for version, ver := range m.Versions {
		if !versionPattern.MatchString(version) {
			return fmt.Errorf("invalid version format %q: must be semver (e.g., 1.2.3)", version)
		}

		if len(ver.Platforms) == 0 {
			return fmt.Errorf("version %q has no platforms", version)
		}

		for platform, asset := range ver.Platforms {
			if !platformPattern.MatchString(platform) {
				return fmt.Errorf("invalid platform %q: must match pattern (linux|darwin|windows)-(amd64|arm64)", platform)
			}

			// Validate asset type
			if asset.Type != "tar" && asset.Type != "zip" {
				return fmt.Errorf("invalid asset type %q for %s/%s: must be 'tar' or 'zip'", asset.Type, version, platform)
			}

			// Validate URL is HTTPS
			if asset.URL == "" {
				return fmt.Errorf("missing URL for %s/%s", version, platform)
			}

			u, err := url.Parse(asset.URL)
			if err != nil {
				return fmt.Errorf("invalid URL %q for %s/%s: %w", asset.URL, version, platform, err)
			}
			if u.Scheme != "https" {
				return fmt.Errorf("URL must use HTTPS: %q for %s/%s", asset.URL, version, platform)
			}

			// Validate checksum format
			if asset.Checksum == "" {
				return fmt.Errorf("missing checksum for %s/%s", version, platform)
			}

			checksumPattern := regexp.MustCompile(`^sha256:[a-fA-F0-9]{64}$`)
			if !checksumPattern.MatchString(asset.Checksum) {
				return fmt.Errorf("invalid checksum format for %s/%s: must be sha256:hex (64 chars)", version, platform)
			}
		}
	}

	return nil
}

// ValidateVersion checks if a version exists for the given platform
func ValidateVersion(m *Manifest, version, platform string) error {
	ver, ok := m.Versions[version]
	if !ok {
		return fmt.Errorf("version %q not found for package %q", version, m.Name)
	}

	_, ok = ver.Platforms[platform]
	if !ok {
		return fmt.Errorf("platform %q not available for package %q version %q", platform, m.Name, version)
	}

	return nil
}

// GetAsset returns the asset for a specific version and platform
func (m *Manifest) GetAsset(version, platform string) (*Asset, error) {
	if err := ValidateVersion(m, version, platform); err != nil {
		return nil, err
	}

	asset := m.Versions[version].Platforms[platform]
	return &asset, nil
}
