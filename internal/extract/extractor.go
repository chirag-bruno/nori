package extract

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chirag-bruno/nori/internal/fetch"
)

// Extractor handles safe extraction of archives
type Extractor struct {
	fetcher *fetch.Fetcher
}

// New creates a new extractor
func New() *Extractor {
	return &Extractor{
		fetcher: fetch.New(),
	}
}

// Extract extracts an archive to a temporary directory and returns the path
// assetType can be "tar" or "zip"
// For tar files, it auto-detects .tar, .tar.gz, .tgz, .tar.xz
func (e *Extractor) Extract(data []byte, assetType string, checksum string) (string, error) {
	// Verify checksum first
	if err := fetch.VerifyChecksum(data, checksum); err != nil {
		return "", fmt.Errorf("checksum verification failed: %w", err)
	}
	
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "nori-extract-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	
	// Extract based on type
	switch assetType {
	case "tar":
		if err := e.extractTar(data, tmpDir); err != nil {
			os.RemoveAll(tmpDir)
			return "", fmt.Errorf("failed to extract tar: %w", err)
		}
	case "zip":
		if err := e.extractZip(data, tmpDir); err != nil {
			os.RemoveAll(tmpDir)
			return "", fmt.Errorf("failed to extract zip: %w", err)
		}
	default:
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("unsupported asset type: %s", assetType)
	}
	
	return tmpDir, nil
}

// extractTar extracts a tar archive (handles .tar, .tar.gz, .tgz)
func (e *Extractor) extractTar(data []byte, destDir string) error {
	var reader io.Reader = bytes.NewReader(data)
	
	// Try to detect compression
	if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
		// Gzip compressed
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}
	// TODO: Add xz support if needed
	
	tr := tar.NewReader(reader)
	
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}
		
		// Validate and sanitize path
		path, err := sanitizePath(hdr.Name, destDir)
		if err != nil {
			return fmt.Errorf("invalid path %q: %w", hdr.Name, err)
		}
		
		// Create directory if needed
		if hdr.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(path, os.FileMode(hdr.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}
		
		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}
		
		// Extract file
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		
		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return fmt.Errorf("failed to write file: %w", err)
		}
		f.Close()
	}
	
	return nil
}

// extractZip extracts a zip archive
func (e *Extractor) extractZip(data []byte, destDir string) error {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}
	
	for _, file := range zipReader.File {
		// Validate and sanitize path
		path, err := sanitizePath(file.Name, destDir)
		if err != nil {
			return fmt.Errorf("invalid path %q: %w", file.Name, err)
		}
		
		// Create directory if needed
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, file.FileInfo().Mode()); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}
		
		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}
		
		// Extract file
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open zip file: %w", err)
		}
		
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.FileInfo().Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create file: %w", err)
		}
		
		if _, err := io.Copy(f, rc); err != nil {
			f.Close()
			rc.Close()
			return fmt.Errorf("failed to write file: %w", err)
		}
		
		f.Close()
		rc.Close()
	}
	
	return nil
}

// sanitizePath validates and sanitizes a path to prevent path traversal attacks
func sanitizePath(name, destDir string) (string, error) {
	// Clean the path
	clean := filepath.Clean(name)
	
	// Reject absolute paths
	if filepath.IsAbs(clean) {
		return "", fmt.Errorf("absolute paths are not allowed")
	}
	
	// Reject paths with ".."
	if strings.Contains(clean, "..") {
		return "", fmt.Errorf("path traversal (..) is not allowed")
	}
	
	// Join with destination directory
	fullPath := filepath.Join(destDir, clean)
	
	// Ensure the resolved path is still within destDir
	rel, err := filepath.Rel(destDir, fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve relative path: %w", err)
	}
	
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path escapes destination directory")
	}
	
	return fullPath, nil
}

// DetectRoot detects the archive root directory
// If there's a single top-level directory, returns its path
// Otherwise returns the extract directory itself
func DetectRoot(extractDir string) (string, error) {
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		return "", fmt.Errorf("failed to read extract directory: %w", err)
	}
	
	// Filter out hidden files and count directories
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			dirs = append(dirs, entry.Name())
		}
	}
	
	// If exactly one top-level directory, use it as root
	if len(dirs) == 1 {
		return filepath.Join(extractDir, dirs[0]), nil
	}
	
	// Otherwise, the extract directory is the root
	return extractDir, nil
}

