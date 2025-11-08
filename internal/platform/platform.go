package platform

import "runtime"

// Platform represents an OS-architecture combination
type Platform struct {
	OS   string
	Arch string
}

// Detect returns the current platform
func Detect() Platform {
	return Platform{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// Normalize converts OS and architecture to the normalized format "os-arch"
func Normalize(os, arch string) string {
	return os + "-" + arch
}

// String returns the normalized platform string
func (p Platform) String() string {
	return Normalize(p.OS, p.Arch)
}
