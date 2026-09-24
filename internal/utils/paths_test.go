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

func TestClosestExistingDir(t *testing.T) {
	dir := t.TempDir()
	if got := ClosestExistingDir(filepath.Join(dir, "missing", "deeper")); got != dir {
		t.Errorf("ClosestExistingDir() = %q, want %q", got, dir)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		secs float64
		want string
	}{
		{0, "0:00"},
		{59, "0:59"},
		{754, "12:34"},
		{3723, "1:02:03"},
	}
	for _, tt := range tests {
		if got := formatDuration(tt.secs); got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.secs, got, tt.want)
		}
	}
}
