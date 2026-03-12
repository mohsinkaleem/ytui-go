package store

import (
	"encoding/json"
	"time"
)

// SearchEntry represents a search history record
type SearchEntry struct {
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // "video", "playlist", "search"
}

// DownloadRecord represents a completed or in-progress download
type DownloadRecord struct {
	ID          string    `json:"id"`
	VideoID     string    `json:"video_id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	FormatID    string    `json:"format_id"`
	OutputPath  string    `json:"output_path"`
	State       string    `json:"state"` // "queued", "downloading", "completed", "failed", "cancelled", "paused"
	FileSize    int64     `json:"file_size"`
	EmbedSubs   bool      `json:"embed_subs"`
	EmbedMeta   bool      `json:"embed_meta"`
	EmbedChaps  bool      `json:"embed_chaps"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// Settings represents app-level persistent settings
type Settings struct {
	DownloadDir   string `json:"download_dir"`
	MaxConcurrent int    `json:"max_concurrent"`
	DefaultFormat string `json:"default_format"`
	Theme         string `json:"theme"`
	EmbedSubs     bool   `json:"embed_subs"`
	EmbedMetadata bool   `json:"embed_metadata"`
	EmbedChapters bool   `json:"embed_chapters"`
	MpvPath       string `json:"mpv_path"`
	CookiesFrom   string `json:"cookies_from"` // browser name for --cookies-from-browser (e.g. "chrome", "firefox", "brave")
	CookiesFile   string `json:"cookies_file"` // path to Netscape-format cookies.txt file for --cookies
}

// DefaultSettings returns sensible defaults
func DefaultSettings() Settings {
	return Settings{
		DownloadDir:   "~/Downloads",
		MaxConcurrent: 3,
		DefaultFormat: "best",
		Theme:         "mocha",
		EmbedSubs:     false,
		EmbedMetadata: false,
		EmbedChapters: false,
		MpvPath:       "mpv",
	}
}

// Marshal helpers
func (s *SearchEntry) Marshal() ([]byte, error)    { return json.Marshal(s) }
func (s *SearchEntry) Unmarshal(b []byte) error    { return json.Unmarshal(b, s) }
func (d *DownloadRecord) Marshal() ([]byte, error) { return json.Marshal(d) }
func (d *DownloadRecord) Unmarshal(b []byte) error { return json.Unmarshal(b, d) }
func (s *Settings) Marshal() ([]byte, error)       { return json.Marshal(s) }
func (s *Settings) Unmarshal(b []byte) error       { return json.Unmarshal(b, s) }
