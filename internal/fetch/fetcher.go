package fetch

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	maxRetries = 3
	retryDelay = time.Second
	// No timeout - allow large binaries to download
	// Context cancellation still works for user-initiated cancellation
)

// Fetcher handles HTTP downloads with retries and checksum verification
type Fetcher struct {
	client *http.Client
}

// New creates a new fetcher
func New() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			// No timeout - allow large binaries to download
			// Context cancellation still works for user-initiated cancellation
		},
	}
}

// Fetch downloads data from a URL and verifies its checksum
func (f *Fetcher) Fetch(ctx context.Context, url, expectedChecksum string) ([]byte, error) {
	return f.FetchWithProgress(ctx, url, expectedChecksum, nil)
}

// FetchWithProgress downloads data from a URL with progress tracking
// progressWriter can be nil to disable progress tracking
func (f *Fetcher) FetchWithProgress(ctx context.Context, url, expectedChecksum string, progressWriter io.Writer) ([]byte, error) {
	var lastErr error
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay * time.Duration(attempt)):
			}
		}
		
		data, err := f.fetchOnce(ctx, url, progressWriter)
		if err != nil {
			lastErr = err
			// Retry on network errors or 5xx errors
			if isRetryableError(err) {
				continue
			}
			return nil, err
		}
		
		// Verify checksum
		if err := VerifyChecksum(data, expectedChecksum); err != nil {
			return nil, fmt.Errorf("checksum verification failed: %w", err)
		}
		
		return data, nil
	}
	
	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// fetchOnce performs a single HTTP GET request
func (f *Fetcher) fetchOnce(ctx context.Context, url string, progressWriter io.Writer) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	
	// Read with progress tracking if progressWriter is provided
	var reader io.Reader = resp.Body
	if progressWriter != nil {
		reader = io.TeeReader(resp.Body, progressWriter)
	}
	
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	
	return data, nil
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	// Retry on network errors or 5xx server errors
	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "HTTP 5") {
		return true
	}
	
	return false
}

// VerifyChecksum verifies that data matches the expected SHA256 checksum
func VerifyChecksum(data []byte, expected string) error {
	// Parse checksum format: sha256:hex
	if !strings.HasPrefix(expected, "sha256:") {
		return fmt.Errorf("invalid checksum format: must start with 'sha256:'")
	}
	
	expectedHex := strings.TrimPrefix(expected, "sha256:")
	if len(expectedHex) != 64 {
		return fmt.Errorf("invalid checksum length: expected 64 hex characters, got %d", len(expectedHex))
	}
	
	// Decode expected hex
	expectedBytes, err := hex.DecodeString(expectedHex)
	if err != nil {
		return fmt.Errorf("invalid checksum hex: %w", err)
	}
	
	// Compute actual checksum
	hash := sha256.Sum256(data)
	
	// Compare
	if !equalBytes(hash[:], expectedBytes) {
		return fmt.Errorf("checksum mismatch: expected %s, got sha256:%s",
			expected, hex.EncodeToString(hash[:]))
	}
	
	return nil
}

// equalBytes performs constant-time comparison of byte slices
func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	
	result := byte(0)
	for i := range a {
		result |= a[i] ^ b[i]
	}
	
	return result == 0
}

