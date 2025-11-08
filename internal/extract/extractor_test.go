package extract

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func createTestTar(t *testing.T) []byte {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	
	// Add a file
	hdr := &tar.Header{
		Name: "test.txt",
		Size: 11,
		Mode: 0644,
	}
	tw.WriteHeader(hdr)
	tw.Write([]byte("hello world"))
	tw.Close()
	
	return buf.Bytes()
}

func createTestTarGz(t *testing.T) []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	
	hdr := &tar.Header{
		Name: "test.txt",
		Size: 11,
		Mode: 0644,
	}
	tw.WriteHeader(hdr)
	tw.Write([]byte("hello world"))
	tw.Close()
	gw.Close()
	
	return buf.Bytes()
}

func createTestTarWithDir(t *testing.T) []byte {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	
	// Add a directory
	hdr := &tar.Header{
		Name:     "mypackage/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	}
	tw.WriteHeader(hdr)
	
	// Add a file in the directory
	hdr = &tar.Header{
		Name: "mypackage/test.txt",
		Size: 11,
		Mode: 0644,
	}
	tw.WriteHeader(hdr)
	tw.Write([]byte("hello world"))
	tw.Close()
	
	return buf.Bytes()
}

func createTestZip(t *testing.T) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	
	w, _ := zw.Create("test.txt")
	w.Write([]byte("hello world"))
	zw.Close()
	
	return buf.Bytes()
}

func createTestZipWithDir(t *testing.T) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	
	w, _ := zw.Create("mypackage/")
	w.Write([]byte{})
	
	w, _ = zw.Create("mypackage/test.txt")
	w.Write([]byte("hello world"))
	zw.Close()
	
	return buf.Bytes()
}

func TestExtractTar(t *testing.T) {
	data := createTestTar(t)
	hash := sha256.Sum256(data)
	checksum := "sha256:" + hex.EncodeToString(hash[:])
	
	extractor := New()
	extractDir, err := extractor.Extract(data, "tar", checksum)
	if err != nil {
		t.Fatalf("Extract() failed: %v", err)
	}
	defer os.RemoveAll(extractDir)
	
	// Check that file exists
	testFile := filepath.Join(extractDir, "test.txt")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("test.txt not found in extracted directory")
	}
	
	content, _ := os.ReadFile(testFile)
	if string(content) != "hello world" {
		t.Errorf("File content = %q, want %q", string(content), "hello world")
	}
}

func TestExtractTarGz(t *testing.T) {
	data := createTestTarGz(t)
	hash := sha256.Sum256(data)
	checksum := "sha256:" + hex.EncodeToString(hash[:])
	
	extractor := New()
	extractDir, err := extractor.Extract(data, "tar", checksum)
	if err != nil {
		t.Fatalf("Extract() failed: %v", err)
	}
	defer os.RemoveAll(extractDir)
	
	testFile := filepath.Join(extractDir, "test.txt")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("test.txt not found in extracted directory")
	}
}

func TestExtractZip(t *testing.T) {
	data := createTestZip(t)
	hash := sha256.Sum256(data)
	checksum := "sha256:" + hex.EncodeToString(hash[:])
	
	extractor := New()
	extractDir, err := extractor.Extract(data, "zip", checksum)
	if err != nil {
		t.Fatalf("Extract() failed: %v", err)
	}
	defer os.RemoveAll(extractDir)
	
	testFile := filepath.Join(extractDir, "test.txt")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("test.txt not found in extracted directory")
	}
}

func TestDetectRoot(t *testing.T) {
	// Create a temp directory with a single top-level directory
	tmpDir := t.TempDir()
	
	// Case 1: Single top-level directory
	pkgDir := filepath.Join(tmpDir, "mypackage-v1.0.0")
	os.MkdirAll(pkgDir, 0755)
	os.WriteFile(filepath.Join(pkgDir, "file.txt"), []byte("test"), 0644)
	
	root, err := DetectRoot(tmpDir)
	if err != nil {
		t.Fatalf("DetectRoot() failed: %v", err)
	}
	
	if root != pkgDir {
		t.Errorf("DetectRoot() = %q, want %q", root, pkgDir)
	}
	
	// Case 2: Multiple top-level items - should return the extract dir
	tmpDir2 := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir2, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir2, "file2.txt"), []byte("test"), 0644)
	
	root2, err := DetectRoot(tmpDir2)
	if err != nil {
		t.Fatalf("DetectRoot() failed: %v", err)
	}
	
	if root2 != tmpDir2 {
		t.Errorf("DetectRoot() = %q, want %q", root2, tmpDir2)
	}
}

func TestExtractPathTraversal(t *testing.T) {
	// Create a tar with path traversal attempt
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	
	// Try to extract to parent directory
	hdr := &tar.Header{
		Name: "../evil.txt",
		Size: 5,
		Mode: 0644,
	}
	tw.WriteHeader(hdr)
	tw.Write([]byte("evil"))
	tw.Close()
	
	data := buf.Bytes()
	hash := sha256.Sum256(data)
	checksum := "sha256:" + hex.EncodeToString(hash[:])
	
	extractor := New()
	_, err := extractor.Extract(data, "tar", checksum)
	if err == nil {
		t.Error("Extract() should reject path traversal attempts")
	}
}

func TestExtractAbsolutePath(t *testing.T) {
	// Create a tar with absolute path
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	
	hdr := &tar.Header{
		Name: "/etc/passwd",
		Size: 5,
		Mode: 0644,
	}
	tw.WriteHeader(hdr)
	tw.Write([]byte("evil"))
	tw.Close()
	
	data := buf.Bytes()
	hash := sha256.Sum256(data)
	checksum := "sha256:" + hex.EncodeToString(hash[:])
	
	extractor := New()
	_, err := extractor.Extract(data, "tar", checksum)
	if err == nil {
		t.Error("Extract() should reject absolute paths")
	}
}

