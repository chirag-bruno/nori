package registry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chirag-bruno/nori/internal/platform"
	"gopkg.in/yaml.v3"
)

func TestRegistryUpdate(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/index.yaml" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`packages:
  - name: node
    description: Node.js runtime
  - name: python
    description: Python
`))
			return
		}
		if r.URL.Path == "/packages/node.yaml" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`schema: 1
name: node
description: Node.js runtime
bins:
  - bin/node
versions:
  "22.2.0":
    platforms:
      linux-amd64:
        type: tar
        url: https://nodejs.org/dist/v22.2.0/node-v22.2.0-linux-x64.tar.xz
        checksum: sha256:5f4a1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab
`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create temp directory for registry cache
	tmpDir := t.TempDir()
	registryDir := filepath.Join(tmpDir, "registry")
	os.MkdirAll(registryDir, 0755)

	// Override the registry directory for testing
	originalRegistryDir := platform.RegistryDir
	defer func() {
		// Can't easily override, so we'll test with actual paths
		_ = originalRegistryDir
	}()

	reg := New(server.URL)

	ctx := context.Background()
	err := reg.Update(ctx)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}
}

func TestRegistryLoadPackage(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/packages/testnode.yaml" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`schema: 1
name: testnode
description: Node.js runtime
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
`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	reg := New(server.URL)

	ctx := context.Background()
	m, err := reg.LoadPackage(ctx, "testnode")
	if err != nil {
		t.Fatalf("LoadPackage() failed: %v", err)
	}

	if m.Name != "testnode" {
		t.Errorf("LoadPackage() name = %q, want %q", m.Name, "testnode")
	}
	if len(m.Bins) != 2 {
		t.Errorf("LoadPackage() bins count = %d, want 2. Bins: %v", len(m.Bins), m.Bins)
	}
}

func TestRegistrySearch(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/index.yaml" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`packages:
  - name: node
    description: Node.js runtime
  - name: python
    description: Python programming language
  - name: deno
    description: Deno runtime
`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	reg := New(server.URL)

	ctx := context.Background()

	// Test search for "node"
	results, err := reg.Search(ctx, "node")
	if err != nil {
		t.Fatalf("Search() failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Search('node') returned %d results, want 1", len(results))
	}
	if len(results) > 0 && results[0].Name != "node" {
		t.Errorf("Search('node') name = %q, want %q", results[0].Name, "node")
	}

	// Test search for "py" (should match python)
	results, err = reg.Search(ctx, "py")
	if err != nil {
		t.Fatalf("Search() failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Search('py') returned %d results, want 1", len(results))
	}
	if len(results) > 0 && results[0].Name != "python" {
		t.Errorf("Search('py') name = %q, want %q", results[0].Name, "python")
	}
}

func TestRegistryBaseURLFromEnv(t *testing.T) {
	// Test that registry URL can be loaded from environment
	originalURL := os.Getenv("NORI_REGISTRY_URL")
	defer func() {
		if originalURL != "" {
			os.Setenv("NORI_REGISTRY_URL", originalURL)
		} else {
			os.Unsetenv("NORI_REGISTRY_URL")
		}
	}()

	os.Setenv("NORI_REGISTRY_URL", "https://custom-registry.example.com")
	reg := NewFromEnv()

	if reg.BaseURL != "https://custom-registry.example.com" {
		t.Errorf("NewFromEnv() BaseURL = %q, want %q", reg.BaseURL, "https://custom-registry.example.com")
	}
}

func TestRegistryDefaultURL(t *testing.T) {
	// Test default URL when env var is not set
	originalURL := os.Getenv("NORI_REGISTRY_URL")
	defer func() {
		if originalURL != "" {
			os.Setenv("NORI_REGISTRY_URL", originalURL)
		} else {
			os.Unsetenv("NORI_REGISTRY_URL")
		}
	}()

	os.Unsetenv("NORI_REGISTRY_URL")
	reg := NewFromEnv()

	// Should have a default URL (not empty)
	if reg.BaseURL == "" {
		t.Error("NewFromEnv() BaseURL should not be empty when env var is not set")
	}
}

// TestGitHubURLConstruction verifies that URLs are constructed correctly for GitHub raw content
func TestGitHubURLConstruction(t *testing.T) {
	baseURL := "https://raw.githubusercontent.com/user/repo/main"
	reg := New(baseURL)

	// Test index URL construction
	expectedIndexURL := baseURL + "/index.yaml"
	actualIndexURL := strings.TrimSuffix(reg.BaseURL, "/") + "/index.yaml"
	if actualIndexURL != expectedIndexURL {
		t.Errorf("Index URL = %q, want %q", actualIndexURL, expectedIndexURL)
	}

	// Test manifest URL construction
	expectedManifestURL := baseURL + "/packages/node.yaml"
	actualManifestURL := strings.TrimSuffix(reg.BaseURL, "/") + "/packages/node.yaml"
	if actualManifestURL != expectedManifestURL {
		t.Errorf("Manifest URL = %q, want %q", actualManifestURL, expectedManifestURL)
	}

	// Test with trailing slash
	baseURLWithSlash := baseURL + "/"
	reg2 := New(baseURLWithSlash)
	actualIndexURL2 := strings.TrimSuffix(reg2.BaseURL, "/") + "/index.yaml"
	if actualIndexURL2 != expectedIndexURL {
		t.Errorf("Index URL with trailing slash = %q, want %q", actualIndexURL2, expectedIndexURL)
	}
}

// TestGitHubRawContentFormat verifies the expected GitHub repository structure
func TestGitHubRawContentFormat(t *testing.T) {
	// This test documents the expected GitHub repository structure
	// It doesn't make actual HTTP requests, but verifies URL format

	baseURL := "https://raw.githubusercontent.com/chirag-bruno/nori-registry/main"
	reg := New(baseURL)

	// Expected structure:
	// https://raw.githubusercontent.com/chirag-bruno/nori-registry/main/index.yaml
	// https://raw.githubusercontent.com/chirag-bruno/nori-registry/main/packages/neovim.yaml
	// https://raw.githubusercontent.com/chirag-bruno/nori-registry/main/packages/node.yaml

	indexURL := strings.TrimSuffix(reg.BaseURL, "/") + "/index.yaml"
	if !strings.HasPrefix(indexURL, "https://raw.githubusercontent.com/") {
		t.Errorf("Index URL should use GitHub raw content: %q", indexURL)
	}
	if !strings.HasSuffix(indexURL, "/index.yaml") {
		t.Errorf("Index URL should end with /index.yaml: %q", indexURL)
	}

	manifestURL := strings.TrimSuffix(reg.BaseURL, "/") + "/packages/node.yaml"
	if !strings.HasPrefix(manifestURL, "https://raw.githubusercontent.com/") {
		t.Errorf("Manifest URL should use GitHub raw content: %q", manifestURL)
	}
	if !strings.Contains(manifestURL, "/packages/") {
		t.Errorf("Manifest URL should contain /packages/ path: %q", manifestURL)
	}
	if !strings.HasSuffix(manifestURL, ".yaml") {
		t.Errorf("Manifest URL should end with .yaml: %q", manifestURL)
	}
}

// TestRegistryIntegrationWithGitHub is an optional integration test
// Set NORI_TEST_REGISTRY_URL to a real GitHub repository to run this test
// Example: NORI_TEST_REGISTRY_URL=https://raw.githubusercontent.com/user/repo/main
func TestRegistryIntegrationWithGitHub(t *testing.T) {
	testRegistryURL := os.Getenv("NORI_TEST_REGISTRY_URL")
	if testRegistryURL == "" {
		t.Skip("Skipping integration test. Set NORI_TEST_REGISTRY_URL to enable.")
	}

	// Verify it's a GitHub raw content URL
	if !strings.HasPrefix(testRegistryURL, "https://raw.githubusercontent.com/") {
		t.Fatalf("NORI_TEST_REGISTRY_URL must be a GitHub raw content URL, got: %q", testRegistryURL)
	}

	reg := New(testRegistryURL)
	ctx := context.Background()

	// Test fetching index via Search (which fetches index.yaml)
	indexURL := strings.TrimSuffix(reg.BaseURL, "/") + "/index.yaml"
	t.Logf("Testing index URL: %s", indexURL)

	// Search with empty query will fetch the index
	results, err := reg.Search(ctx, "")
	if err != nil {
		t.Fatalf("Failed to fetch index from GitHub (via Search): %v", err)
	}

	// Verify index was cached
	indexData, err := os.ReadFile(platform.IndexPath())
	if err != nil {
		t.Fatalf("Failed to read cached index: %v", err)
	}

	if len(indexData) == 0 {
		t.Error("Index data is empty")
	}

	// Test that index is valid YAML
	var index Index
	if err := yaml.Unmarshal(indexData, &index); err != nil {
		t.Fatalf("Index is not valid YAML: %v", err)
	}

	if len(index.Packages) == 0 {
		t.Log("Warning: Index contains no packages")
	}

	// Verify we got results from Search
	if len(results) != len(index.Packages) {
		t.Logf("Search returned %d results, index has %d packages", len(results), len(index.Packages))
	}

	// If there are packages, try loading the first one (tests manifest fetching)
	if len(index.Packages) > 0 {
		pkg := index.Packages[0]
		manifestURL := strings.TrimSuffix(reg.BaseURL, "/") + "/packages/" + pkg.Name + ".yaml"
		t.Logf("Testing manifest URL: %s", manifestURL)

		// Use LoadPackage which will fetch the manifest
		m, err := reg.LoadPackage(ctx, pkg.Name)
		if err != nil {
			t.Fatalf("Failed to fetch manifest for %s: %v", pkg.Name, err)
		}

		if m == nil {
			t.Errorf("Manifest for %s is nil", pkg.Name)
		} else if m.Name != pkg.Name {
			t.Errorf("Manifest name = %q, want %q", m.Name, pkg.Name)
		}
	}
}
