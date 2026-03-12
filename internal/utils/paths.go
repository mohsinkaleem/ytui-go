package utils

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

// ConfigDir returns the XDG config directory for ytui-go
func ConfigDir() string {
	return filepath.Join(xdg.ConfigHome, "ytui-go")
}

// DataDir returns the XDG data directory for ytui-go
func DataDir() string {
	return filepath.Join(xdg.DataHome, "ytui-go")
}

// DefaultDownloadDir returns the default download directory
func DefaultDownloadDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, "Downloads")
}

// ExpandPath expands ~ to home directory
func ExpandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	return os.MkdirAll(ExpandPath(path), 0o755)
}

// ClosestExistingDir walks up the path to find the closest existing directory.
// Returns "" if no parent exists (shouldn't happen on a real filesystem).
func ClosestExistingDir(path string) string {
	path = ExpandPath(path)
	for {
		parent := filepath.Dir(path)
		if parent == path {
			// Reached root
			break
		}
		info, err := os.Stat(parent)
		if err == nil && info.IsDir() {
			return parent
		}
		path = parent
	}
	return ""
}
