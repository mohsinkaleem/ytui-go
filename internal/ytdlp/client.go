package ytdlp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Client wraps yt-dlp binary calls
type Client struct {
	binaryPath  string
	cookiesFrom string // browser name for --cookies-from-browser
	cookiesFile string // path to Netscape-format cookies.txt for --cookies
}

// NewClient creates a new yt-dlp client, auto-discovering the binary
func NewClient() (*Client, error) {
	path, err := exec.LookPath("yt-dlp")
	if err != nil {
		return nil, fmt.Errorf("yt-dlp not found in PATH: %w", err)
	}
	return &Client{binaryPath: path}, nil
}

// NewClientWithPath creates a client with a specific binary path
func NewClientWithPath(path string) *Client {
	return &Client{binaryPath: path}
}

// SetCookiesFrom configures the browser for --cookies-from-browser.
// Valid values: "chrome", "firefox", "brave", "edge", "opera", "safari", "chromium", "vivaldi".
// Pass an empty string to disable.
func (c *Client) SetCookiesFrom(browser string) {
	c.cookiesFrom = browser
}

// CookiesFrom returns the currently configured cookies-from-browser value.
func (c *Client) CookiesFrom() string {
	return c.cookiesFrom
}

// SetCookiesFile configures a Netscape-format cookies.txt file path for --cookies.
// This is used as a fallback when --cookies-from-browser fails or is unavailable.
// Pass an empty string to disable.
func (c *Client) SetCookiesFile(path string) {
	c.cookiesFile = path
}

// CookiesFile returns the currently configured cookies file path.
func (c *Client) CookiesFile() string {
	return c.cookiesFile
}

// Search runs a yt-dlp search and returns video info
func (c *Client) Search(ctx context.Context, query string, sortBy string) ([]VideoInfo, error) {
	args := []string{
		fmt.Sprintf("ytsearch25:%s", query),
		"--flat-playlist",
		"-J",
		"--extractor-args", "youtube:player_skip=webpage",
		"--no-warnings",
	}

	if sortBy != "" && sortBy != "relevance" {
		args = append(args, "--extractor-args", fmt.Sprintf("youtube:sort=%s", sortBy))
	}

	out, err := c.run(ctx, args...)
	if err != nil {
		return nil, err
	}

	var result struct {
		Entries []VideoInfo `json:"entries"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parse search results: %w", err)
	}

	// Fill in URLs if missing
	for i := range result.Entries {
		if result.Entries[i].WebpageURL == "" && result.Entries[i].ID != "" {
			result.Entries[i].WebpageURL = "https://www.youtube.com/watch?v=" + result.Entries[i].ID
		}
		if result.Entries[i].URL == "" {
			result.Entries[i].URL = result.Entries[i].WebpageURL
		}
	}

	return result.Entries, nil
}

// GetVideoMetadata fetches full metadata for a single video
func (c *Client) GetVideoMetadata(ctx context.Context, url string) (*VideoInfo, error) {
	out, err := c.run(ctx, "-j", "--no-download", "--no-warnings", url)
	if err != nil {
		return nil, err
	}

	var info VideoInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, fmt.Errorf("parse video metadata: %w", err)
	}
	if info.URL == "" {
		info.URL = url
	}
	return &info, nil
}

// GetPlaylistMetadata fetches playlist metadata with flat entries
func (c *Client) GetPlaylistMetadata(ctx context.Context, url string) (*PlaylistInfo, error) {
	out, err := c.run(ctx, "--flat-playlist", "-J", "--no-warnings", url)
	if err != nil {
		return nil, err
	}

	var info PlaylistInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, fmt.Errorf("parse playlist metadata: %w", err)
	}

	// Fill in URLs
	for i := range info.Entries {
		if info.Entries[i].WebpageURL == "" && info.Entries[i].ID != "" {
			info.Entries[i].WebpageURL = "https://www.youtube.com/watch?v=" + info.Entries[i].ID
		}
		if info.Entries[i].URL == "" {
			info.Entries[i].URL = info.Entries[i].WebpageURL
		}
	}

	return &info, nil
}

// ListFormats fetches and returns all available formats for a URL
func (c *Client) ListFormats(ctx context.Context, url string) ([]Format, error) {
	info, err := c.GetVideoMetadata(ctx, url)
	if err != nil {
		return nil, err
	}
	return info.Formats, nil
}

// Download starts a download and returns the command + stderr pipe for progress.
// yt-dlp writes --progress-template output to stderr, so callers must read
// from the returned ReadCloser to receive progress lines.
func (c *Client) Download(ctx context.Context, url, formatID string, opts DownloadOpts) (*exec.Cmd, io.ReadCloser, error) {
	args := []string{
		url,
		"--newline",
		"--progress-template", progressTemplate,
		"--no-warnings",
	}

	if formatID != "" {
		args = append(args, "-f", formatID)
		// When merging video+audio streams, ensure mp4 output
		if strings.Contains(formatID, "+") {
			args = append(args, "--merge-output-format", "mp4")
		}
	}

	if opts.OutputTemplate != "" {
		args = append(args, "-o", expandHome(opts.OutputTemplate))
	} else if opts.OutputDir != "" {
		expanded := expandHome(opts.OutputDir)
		args = append(args, "-o", fmt.Sprintf("%s/%%(title)s.%%(ext)s", expanded))
	}

	if opts.EmbedSubs {
		args = append(args, "--embed-subs")
	}
	if opts.EmbedMetadata {
		args = append(args, "--embed-metadata")
	}
	if opts.EmbedChapters {
		args = append(args, "--embed-chapters")
	}
	if opts.ContinueDL {
		args = append(args, "--continue")
	}

	if c.cookiesFrom != "" {
		args = append(args, "--cookies-from-browser", c.cookiesFrom)
	} else if c.cookiesFile != "" {
		args = append(args, "--cookies", c.cookiesFile)
	}

	cmd := exec.CommandContext(ctx, c.binaryPath, args...)

	// yt-dlp may write --progress-template lines to stdout or stderr
	// depending on version and flags. Merge both into a single pipe
	// so the caller always receives progress lines.
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, fmt.Errorf("pipe: %w", err)
	}
	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Start(); err != nil {
		r.Close()
		w.Close()
		return nil, nil, fmt.Errorf("start download: %w", err)
	}

	// Close write end in parent; the child process holds its own copy.
	// When the child exits the reader will see EOF.
	w.Close()

	return cmd, r, nil
}

// BinaryPath returns the yt-dlp binary path
func (c *Client) BinaryPath() string {
	return c.binaryPath
}

// run executes yt-dlp with args and returns stdout
func (c *Client) run(ctx context.Context, args ...string) ([]byte, error) {
	if c.cookiesFrom != "" {
		args = append([]string{"--cookies-from-browser", c.cookiesFrom}, args...)
	} else if c.cookiesFile != "" {
		args = append([]string{"--cookies", c.cookiesFile}, args...)
	}
	cmd := exec.CommandContext(ctx, c.binaryPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return nil, fmt.Errorf("yt-dlp: %s", parseError(errMsg))
		}
		return nil, fmt.Errorf("yt-dlp: %w", err)
	}
	return out, nil
}

// parseError extracts user-friendly error from yt-dlp stderr
func parseError(stderr string) string {
	lines := strings.Split(stderr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "ERROR:") {
			return strings.TrimPrefix(line, "ERROR: ")
		}
	}
	// Return last non-empty line
	for i := len(lines) - 1; i >= 0; i-- {
		if s := strings.TrimSpace(lines[i]); s != "" {
			return s
		}
	}
	return stderr
}

// expandHome expands a leading ~ to the user's home directory.
// This is needed because os/exec does not invoke a shell so the shell's
// tilde expansion is unavailable.
func expandHome(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}

// IsURL checks if input looks like a URL
func IsURL(input string) bool {
	input = strings.TrimSpace(input)
	return strings.HasPrefix(input, "http://") ||
		strings.HasPrefix(input, "https://") ||
		strings.HasPrefix(input, "www.") ||
		strings.Contains(input, "youtube.com/") ||
		strings.Contains(input, "youtu.be/")
}

// IsPlaylistURL checks if URL is a playlist
func IsPlaylistURL(url string) bool {
	return strings.Contains(url, "playlist?list=") ||
		strings.Contains(url, "&list=")
}
