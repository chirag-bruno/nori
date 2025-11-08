package registry

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chirag-bruno/nori/internal/manifest"
	"github.com/chirag-bruno/nori/internal/platform"
	"gopkg.in/yaml.v3"
)

const (
	defaultRegistryURL = "https://raw.githubusercontent.com/chirag-bruno/nori-registry/main"
)

// PackageMeta represents package metadata from the index
type PackageMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// Index represents the registry index
type Index struct {
	Packages []PackageMeta `yaml:"packages"`
}

// Registry represents a registry client
type Registry struct {
	BaseURL string
	client  *http.Client
}

// New creates a new registry client with the given base URL
func New(baseURL string) *Registry {
	return &Registry{
		BaseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewFromEnv creates a new registry client using NORI_REGISTRY_URL env var or default
func NewFromEnv() *Registry {
	baseURL := os.Getenv("NORI_REGISTRY_URL")
	if baseURL == "" {
		baseURL = defaultRegistryURL
	}
	return New(baseURL)
}

// Update fetches the registry index and caches package manifests
func (r *Registry) Update(ctx context.Context) error {
	// Fetch index.yaml
	indexURL := strings.TrimSuffix(r.BaseURL, "/") + "/index.yaml"
	indexData, err := r.fetch(ctx, indexURL)
	if err != nil {
		return fmt.Errorf("failed to fetch index: %w", err)
	}
	
	// Parse index
	var index Index
	if err := yaml.Unmarshal(indexData, &index); err != nil {
		return fmt.Errorf("failed to parse index: %w", err)
	}
	
	// Ensure registry directory exists
	registryDir := platform.RegistryDir()
	if err := os.MkdirAll(registryDir, 0755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}
	
	// Save index.yaml
	indexPath := platform.IndexPath()
	if err := os.WriteFile(indexPath, indexData, 0644); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}
	
	// Fetch and cache each package manifest
	packagesDir := filepath.Join(registryDir, "packages")
	if err := os.MkdirAll(packagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create packages directory: %w", err)
	}
	
	for _, pkg := range index.Packages {
		manifestURL := strings.TrimSuffix(r.BaseURL, "/") + "/packages/" + pkg.Name + ".yaml"
		manifestData, err := r.fetch(ctx, manifestURL)
		if err != nil {
			// Log error but continue with other packages
			fmt.Printf("Warning: failed to fetch manifest for %s: %v\n", pkg.Name, err)
			continue
		}
		
		// Validate manifest
		m, err := manifest.LoadFromBytes(manifestData)
		if err != nil {
			fmt.Printf("Warning: failed to parse manifest for %s: %v\n", pkg.Name, err)
			continue
		}
		
		if err := manifest.Validate(m); err != nil {
			fmt.Printf("Warning: invalid manifest for %s: %v\n", pkg.Name, err)
			continue
		}
		
		// Save manifest
		manifestPath := platform.PackageManifestPath(pkg.Name)
		if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
			fmt.Printf("Warning: failed to write manifest for %s: %v\n", pkg.Name, err)
			continue
		}
	}
	
	return nil
}

// LoadPackage loads a package manifest (from cache or remote)
func (r *Registry) LoadPackage(ctx context.Context, name string) (*manifest.Manifest, error) {
	// Try to load from cache first
	manifestPath := platform.PackageManifestPath(name)
	if data, err := os.ReadFile(manifestPath); err == nil {
		m, err := manifest.LoadFromBytes(data)
		if err == nil {
			// Validate cached manifest
			if err := manifest.Validate(m); err == nil {
				return m, nil
			}
		}
	}
	
	// If cache miss or invalid, fetch from remote
	manifestURL := strings.TrimSuffix(r.BaseURL, "/") + "/packages/" + name + ".yaml"
	manifestData, err := r.fetch(ctx, manifestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	
	m, err := manifest.LoadFromBytes(manifestData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}
	
	if err := manifest.Validate(m); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}
	
	// Cache the manifest
	manifestPath = platform.PackageManifestPath(name)
	registryDir := platform.RegistryDir()
	packagesDir := filepath.Join(registryDir, "packages")
	if err := os.MkdirAll(packagesDir, 0755); err == nil {
		_ = os.WriteFile(manifestPath, manifestData, 0644)
	}
	
	return m, nil
}

// Search searches the registry index for packages matching the query
func (r *Registry) Search(ctx context.Context, query string) ([]PackageMeta, error) {
	// Load index from cache or fetch
	indexPath := platform.IndexPath()
	var indexData []byte
	
	if data, err := os.ReadFile(indexPath); err == nil {
		indexData = data
	} else {
		// Fetch index
		indexURL := strings.TrimSuffix(r.BaseURL, "/") + "/index.yaml"
		var err error
		indexData, err = r.fetch(ctx, indexURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch index: %w", err)
		}
	}
	
	// Parse index
	var index Index
	if err := yaml.Unmarshal(indexData, &index); err != nil {
		return nil, fmt.Errorf("failed to parse index: %w", err)
	}
	
	// Search for matching packages
	query = strings.ToLower(query)
	var results []PackageMeta
	for _, pkg := range index.Packages {
		if strings.Contains(strings.ToLower(pkg.Name), query) ||
			strings.Contains(strings.ToLower(pkg.Description), query) {
			results = append(results, pkg)
		}
	}
	
	return results, nil
}

// fetch performs an HTTP GET request
func (r *Registry) fetch(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	return data, nil
}

