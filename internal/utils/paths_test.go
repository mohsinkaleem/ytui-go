package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home dir")
	}
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"tilde expand", "~/Downloads", filepath.Join(home, "Downloads")},
		{"tilde only", "~", home},
		{"absolute", "/usr/local/bin", "/usr/local/bin"},
		{"relative", "relative/path", "relative/path"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExpandPath(tt.input); got != tt.want {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEnsureDir(t *testing.T) {
	dir := t.TempDir()
	newDir := filepath.Join(dir, "a", "b", "c")
	err := EnsureDir(newDir)
	if err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}
	info, err := os.Stat(newDir)
	if err != nil {
		t.Fatalf("directory should exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("should be a directory")
	}
}

func TestDefaultDownloadDir(t *testing.T) {
	dir := DefaultDownloadDir()
	if dir == "" || dir == "." {
		t.Skip("could not determine home directory")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("DefaultDownloadDir() = %q, should be absolute", dir)
	}
}

func TestConfigDir(t *testing.T) {
	dir := ConfigDir()
	if dir == "" {
		t.Error("ConfigDir should not be empty")
	}
}

func TestDataDir(t *testing.T) {
	dir := DataDir()
	if dir == "" {
		t.Error("DataDir should not be empty")
	}
}
