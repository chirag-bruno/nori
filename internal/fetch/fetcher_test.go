package fetch

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetch(t *testing.T) {
	// Create test data
	testData := []byte("hello, world")
	hash := sha256.Sum256(testData)
	expectedChecksum := "sha256:" + hex.EncodeToString(hash[:])
	
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()
	
	ctx := context.Background()
	fetcher := New()
	
	data, err := fetcher.Fetch(ctx, server.URL, expectedChecksum)
	if err != nil {
		t.Fatalf("Fetch() failed: %v", err)
	}
	
	if string(data) != string(testData) {
		t.Errorf("Fetch() data = %q, want %q", string(data), string(testData))
	}
}

func TestFetchChecksumMismatch(t *testing.T) {
	testData := []byte("hello, world")
	wrongChecksum := "sha256:" + hex.EncodeToString([]byte("wrong"))
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()
	
	ctx := context.Background()
	fetcher := New()
	
	_, err := fetcher.Fetch(ctx, server.URL, wrongChecksum)
	if err == nil {
		t.Error("Fetch() should fail on checksum mismatch")
	}
}

func TestFetchHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	
	ctx := context.Background()
	fetcher := New()
	
	_, err := fetcher.Fetch(ctx, server.URL, "sha256:abcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab")
	if err == nil {
		t.Error("Fetch() should fail on HTTP error")
	}
}

func TestFetchRetry(t *testing.T) {
	attempts := 0
	testData := []byte("hello, world")
	hash := sha256.Sum256(testData)
	expectedChecksum := "sha256:" + hex.EncodeToString(hash[:])
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()
	
	ctx := context.Background()
	fetcher := New()
	
	data, err := fetcher.Fetch(ctx, server.URL, expectedChecksum)
	if err != nil {
		t.Fatalf("Fetch() failed after retries: %v", err)
	}
	
	if string(data) != string(testData) {
		t.Errorf("Fetch() data = %q, want %q", string(data), string(testData))
	}
	
	if attempts != 3 {
		t.Errorf("Fetch() attempts = %d, want 3", attempts)
	}
}

func TestVerifyChecksum(t *testing.T) {
	testData := []byte("hello, world")
	hash := sha256.Sum256(testData)
	expectedChecksum := "sha256:" + hex.EncodeToString(hash[:])
	
	err := VerifyChecksum(testData, expectedChecksum)
	if err != nil {
		t.Errorf("VerifyChecksum() failed: %v", err)
	}
}

func TestVerifyChecksumInvalidFormat(t *testing.T) {
	testData := []byte("hello, world")
	
	tests := []string{
		"md5:abcd",
		"sha256:",
		"invalid",
		"sha256:abc", // too short
	}
	
	for _, checksum := range tests {
		t.Run(checksum, func(t *testing.T) {
			err := VerifyChecksum(testData, checksum)
			if err == nil {
				t.Errorf("VerifyChecksum() should fail for %q", checksum)
			}
		})
	}
}

func TestFetchTimeout(t *testing.T) {
	// Use a context with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	// Use an invalid URL that will timeout
	fetcher := New()
	
	// Use a URL that will hang (localhost with no server)
	_, err := fetcher.Fetch(ctx, "http://127.0.0.1:99999/invalid", "sha256:abcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab")
	if err == nil {
		t.Error("Fetch() should fail on timeout or connection error")
	}
	// Just verify we got an error - could be timeout or connection refused
}

