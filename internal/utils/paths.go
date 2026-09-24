package utils

import (
	"os"
	"path/filepath"
)

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
