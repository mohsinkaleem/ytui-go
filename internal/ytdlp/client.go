package ytdlp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// searchLimit is the number of results fetched per search.
const searchLimit = 25

// outputTemplate names downloaded files; the ID keeps same-titled videos apart.
const outputTemplate = "%(title)s [%(id)s].%(ext)s"

// sortFilters maps sort orders to YouTube search filters ("sp"), limited to videos.
var sortFilters = map[string]string{
	"upload_date": "CAISAhAB",
	"view_count":  "CAMSAhAB",
	"rating":      "CAESAhAB",
}

// Client wraps yt-dlp binary calls
type Client struct {
	binaryPath string

	mu          sync.RWMutex
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

// SetCookiesFrom configures the browser for --cookies-from-browser.
// Pass an empty string to disable.
func (c *Client) SetCookiesFrom(browser string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cookiesFrom = browser
}

// CookiesFrom returns the currently configured cookies-from-browser value.
func (c *Client) CookiesFrom() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cookiesFrom
}

// SetCookiesFile configures a Netscape-format cookies.txt file path for --cookies.
// It is only used when no browser is configured. Pass an empty string to disable.
func (c *Client) SetCookiesFile(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cookiesFile = path
}

// CookiesFile returns the currently configured cookies file path.
func (c *Client) CookiesFile() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cookiesFile
}

// cookieArgs returns the authentication flags for the configured cookies.
func (c *Client) cookieArgs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	switch {
	case c.cookiesFrom != "":
		return []string{"--cookies-from-browser", c.cookiesFrom}
	case c.cookiesFile != "":
		return []string{"--cookies", c.cookiesFile}
	}
	return nil
}

// Search runs a YouTube search and returns flat video entries.
func (c *Client) Search(ctx context.Context, query, sortBy string) ([]VideoInfo, error) {
	info, err := c.fetchPlaylist(ctx, searchTarget(query, sortBy), "--playlist-end", strconv.Itoa(searchLimit))
	if err != nil {
		return nil, err
	}
	return info.Entries, nil
}

// searchTarget returns the yt-dlp input for a search: ytsearch for relevance,
// otherwise YouTube's results page with a sort filter.
func searchTarget(query, sortBy string) string {
	if sp, ok := sortFilters[sortBy]; ok {
		return "https://www.youtube.com/results?search_query=" + url.QueryEscape(query) + "&sp=" + sp
	}
	return fmt.Sprintf("ytsearch%d:%s", searchLimit, query)
}

// GetVideoMetadata fetches full metadata (including formats) for a single video.
func (c *Client) GetVideoMetadata(ctx context.Context, videoURL string) (*VideoInfo, error) {
	out, err := c.run(ctx, []string{"-j", "--no-playlist", "--no-warnings"}, videoURL)
	if err != nil {
		return nil, err
	}

	var info VideoInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, fmt.Errorf("parse video metadata: %w", err)
	}
	return &info, nil
}

// GetPlaylistMetadata fetches playlist metadata with flat entries.
func (c *Client) GetPlaylistMetadata(ctx context.Context, playlistURL string) (*PlaylistInfo, error) {
	return c.fetchPlaylist(ctx, playlistURL)
}

func (c *Client) fetchPlaylist(ctx context.Context, target string, extra ...string) (*PlaylistInfo, error) {
	args := append([]string{"--flat-playlist", "-J", "--no-warnings"}, extra...)
	out, err := c.run(ctx, args, target)
	if err != nil {
		return nil, err
	}

	var info PlaylistInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, fmt.Errorf("parse playlist metadata: %w", err)
	}
	for i := range info.Entries {
		info.Entries[i].WebpageURL = entryURL(info.Entries[i])
	}
	return &info, nil
}

// entryURL returns the best URL for a flat playlist entry.
func entryURL(e VideoInfo) string {
	switch {
	case e.WebpageURL != "":
		return e.WebpageURL
	case strings.HasPrefix(e.URL, "http"):
		return e.URL
	case e.ID != "":
		return "https://www.youtube.com/watch?v=" + e.ID
	}
	return e.URL
}

// Download starts a download and returns the command plus a pipe carrying its
// combined stdout/stderr, which includes the --progress-template lines.
func (c *Client) Download(ctx context.Context, videoURL, formatID string, opts DownloadOpts) (*exec.Cmd, io.ReadCloser, error) {
	args := []string{
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

	if opts.OutputDir != "" {
		args = append(args, "-o", filepath.Join(expandHome(opts.OutputDir), outputTemplate))
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

	args = append(args, c.cookieArgs()...)
	args = append(args, "--", videoURL)

	cmd := exec.CommandContext(ctx, c.binaryPath, args...)
	// Interrupt instead of kill so yt-dlp stops ffmpeg and leaves resumable .part files.
	if runtime.GOOS != "windows" {
		cmd.Cancel = func() error { return cmd.Process.Signal(os.Interrupt) }
	}
	cmd.WaitDelay = 5 * time.Second

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

// run executes yt-dlp with args followed by target and returns stdout.
func (c *Client) run(ctx context.Context, args []string, target string) ([]byte, error) {
	args = append(append(c.cookieArgs(), args...), "--", target)
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

// IsPlaylistURL reports whether the URL carries a playlist ("list" query parameter).
func IsPlaylistURL(s string) bool {
	u, err := url.Parse(strings.TrimSpace(s))
	return err == nil && u.Query().Get("list") != ""
}
