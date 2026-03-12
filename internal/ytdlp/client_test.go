package ytdlp

import (
	"testing"
)

func TestIsURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"https://www.youtube.com/watch?v=abc", true},
		{"http://youtube.com/watch?v=abc", true},
		{"https://youtu.be/abc", true},
		{"www.youtube.com/watch?v=abc", true},
		{"youtube.com/watch?v=abc", true},
		{"some search query", false},
		{"", false},
		{"music video 2024", false},
		{"https://example.com/video", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsURL(tt.input); got != tt.want {
				t.Errorf("IsURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsPlaylistURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"https://www.youtube.com/playlist?list=PLabc", true},
		{"https://www.youtube.com/watch?v=abc&list=PLabc", true},
		{"https://www.youtube.com/watch?v=abc", false},
		{"some query", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsPlaylistURL(tt.input); got != tt.want {
				t.Errorf("IsPlaylistURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseError(t *testing.T) {
	tests := []struct {
		name   string
		stderr string
		want   string
	}{
		{
			"with ERROR prefix",
			"WARNING: some warning\nERROR: Video unavailable",
			"Video unavailable",
		},
		{
			"no ERROR prefix",
			"some error message\nanother line",
			"another line",
		},
		{
			"empty lines at end",
			"error happened\n\n\n",
			"error happened",
		},
		{
			"single line",
			"simple error",
			"simple error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseError(tt.stderr); got != tt.want {
				t.Errorf("parseError(%q) = %q, want %q", tt.stderr, got, tt.want)
			}
		})
	}
}

func TestCookiesFromAndFile(t *testing.T) {
	c := &Client{binaryPath: "/usr/bin/yt-dlp"}

	// Initially empty
	if got := c.CookiesFrom(); got != "" {
		t.Errorf("CookiesFrom() = %q, want empty", got)
	}
	if got := c.CookiesFile(); got != "" {
		t.Errorf("CookiesFile() = %q, want empty", got)
	}

	// Set browser
	c.SetCookiesFrom("chrome")
	if got := c.CookiesFrom(); got != "chrome" {
		t.Errorf("CookiesFrom() = %q, want %q", got, "chrome")
	}

	// Set file
	c.SetCookiesFile("/tmp/cookies.txt")
	if got := c.CookiesFile(); got != "/tmp/cookies.txt" {
		t.Errorf("CookiesFile() = %q, want %q", got, "/tmp/cookies.txt")
	}

	// Clear browser, file should still be set
	c.SetCookiesFrom("")
	if got := c.CookiesFrom(); got != "" {
		t.Errorf("CookiesFrom() = %q, want empty", got)
	}
	if got := c.CookiesFile(); got != "/tmp/cookies.txt" {
		t.Errorf("CookiesFile() = %q, want %q", got, "/tmp/cookies.txt")
	}

	// Clear file
	c.SetCookiesFile("")
	if got := c.CookiesFile(); got != "" {
		t.Errorf("CookiesFile() = %q, want empty", got)
	}
}
