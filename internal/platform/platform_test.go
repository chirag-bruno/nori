package platform

import (
	"runtime"
	"testing"
)

func TestDetect(t *testing.T) {
	p := Detect()

	if p.OS == "" {
		t.Error("Detect() should return a non-empty OS")
	}
	if p.Arch == "" {
		t.Error("Detect() should return a non-empty Arch")
	}

	// Verify it matches runtime
	if p.OS != runtime.GOOS {
		t.Errorf("Detect().OS = %q, want %q", p.OS, runtime.GOOS)
	}
	if p.Arch != runtime.GOARCH {
		t.Errorf("Detect().Arch = %q, want %q", p.Arch, runtime.GOARCH)
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		os   string
		arch string
		want string
	}{
		{"linux", "amd64", "linux-amd64"},
		{"linux", "arm64", "linux-arm64"},
		{"darwin", "amd64", "darwin-amd64"},
		{"darwin", "arm64", "darwin-arm64"},
		{"windows", "amd64", "windows-amd64"},
		{"windows", "arm64", "windows-arm64"},
		{"linux", "386", "linux-386"},
		{"darwin", "386", "darwin-386"},
	}

	for _, tt := range tests {
		t.Run(tt.os+"-"+tt.arch, func(t *testing.T) {
			got := Normalize(tt.os, tt.arch)
			if got != tt.want {
				t.Errorf("Normalize(%q, %q) = %q, want %q", tt.os, tt.arch, got, tt.want)
			}
		})
	}
}

func TestPlatformString(t *testing.T) {
	p := Platform{OS: "linux", Arch: "amd64"}
	want := "linux-amd64"
	got := p.String()
	if got != want {
		t.Errorf("Platform.String() = %q, want %q", got, want)
	}
}
