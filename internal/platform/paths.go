package platform

import (
	"os"
	"path/filepath"
)

// NoriRoot returns the root directory for nori (~/.nori)
func NoriRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home is unavailable
		return ".nori"
	}
	return filepath.Join(home, ".nori")
}

// InstallsDir returns the directory where packages are installed
func InstallsDir() string {
	return filepath.Join(NoriRoot(), "installs")
}

// ShimsDir returns the directory where shims are created
func ShimsDir() string {
	return filepath.Join(NoriRoot(), "shims")
}

// RegistryDir returns the directory where registry data is cached
func RegistryDir() string {
	return filepath.Join(NoriRoot(), "registry")
}

// ConfigDir returns the directory where configuration files are stored
func ConfigDir() string {
	return filepath.Join(NoriRoot(), "config")
}

// InstallPath returns the full path for a package installation
func InstallPath(pkg, version, platform string) string {
	return filepath.Join(InstallsDir(), pkg, version, platform)
}

// PackageManifestPath returns the path to a cached package manifest
func PackageManifestPath(pkg string) string {
	return filepath.Join(RegistryDir(), "packages", pkg+".yaml")
}

// IndexPath returns the path to the cached registry index
func IndexPath() string {
	return filepath.Join(RegistryDir(), "index.yaml")
}

// ActiveConfigPath returns the path to the active versions configuration
func ActiveConfigPath() string {
	return filepath.Join(ConfigDir(), "active.yaml")
}

