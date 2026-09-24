package types

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
)

// State constants — string-typed for readability
type State string

const (
	StateSearchInput  State = "SearchInput"
	StateLoading      State = "Loading"
	StateVideoList    State = "VideoList"
	StateFormatList   State = "FormatList"
	StateDownload     State = "Download"
	StateVideoPlaying State = "VideoPlaying"
	StateResumeList   State = "ResumeList"
)

// Sort options
type SortBy string

const (
	SortRelevance  SortBy = "relevance"
	SortUploadDate SortBy = "upload_date"
	SortViewCount  SortBy = "view_count"
	SortRating     SortBy = "rating"
)

var SortOptions = []SortBy{SortRelevance, SortUploadDate, SortViewCount, SortRating}

func (s SortBy) String() string {
	switch s {
	case SortRelevance:
		return "Relevance"
	case SortUploadDate:
		return "Upload date"
	case SortViewCount:
		return "View count"
	case SortRating:
		return "Rating"
	default:
		return string(s)
	}
}

// --- Video List Item (implements list.Item) ---

type VideoItem struct {
	ID             string
	Title          string
	Channel        string
	DurationString string
	ViewCount      int64
	URL            string
	IsLive         bool
	Selected       bool // for multi-select
}

func (v VideoItem) FilterValue() string { return v.Title }

// VideoURL returns the URL for the video, constructing from ID if needed
func (v VideoItem) VideoURL() string {
	if v.URL != "" {
		return v.URL
	}
	if v.ID != "" {
		return "https://www.youtube.com/watch?v=" + v.ID
	}
	return ""
}

// --- Format Tab ---

type FormatTab int

const (
	FormatTabVideo  FormatTab = 0
	FormatTabAudio  FormatTab = 1
	FormatTabCustom FormatTab = 2
)

// --- Result Messages ---

type SearchResultMsg struct {
	Videos []list.Item
	Query  string
	Err    error
}

type FormatResultMsg struct {
	Formats []FormatItem
	Video   VideoItem
	Err     error
}

type PlaylistResultMsg struct {
	Videos []list.Item
	Title  string
	Err    error
}

// DownloadTickMsg triggers a periodic refresh while downloads are running
type DownloadTickMsg struct{}

// ClearToastMsg hides the toast with the matching sequence number
type ClearToastMsg struct {
	Seq int
}

type MPVExitedMsg struct {
	Err error
}

// --- Format Item (implements list.Item) ---

type FormatItem struct {
	FormatID   string
	Ext        string
	Resolution string
	Width      int
	Height     int
	FPS        float64
	VCodec     string
	ACodec     string
	Filesize   int64
	TBR        float64
	VBR        float64
	ABR        float64
	FormatNote string
	Protocol   string
}

func (f FormatItem) FilterValue() string { return f.Resolution }

func (f FormatItem) IsVideoOnly() bool {
	return f.VCodec != "" && f.VCodec != "none" && (f.ACodec == "" || f.ACodec == "none")
}

func (f FormatItem) IsAudioOnly() bool {
	return f.ACodec != "" && f.ACodec != "none" && (f.VCodec == "" || f.VCodec == "none")
}

func (f FormatItem) HasVideoAndAudio() bool {
	return f.VCodec != "" && f.VCodec != "none" && f.ACodec != "" && f.ACodec != "none"
}

func (f FormatItem) HumanSize() string {
	if f.Filesize <= 0 {
		return "N/A"
	}
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case f.Filesize >= GB:
		return formatFloat(float64(f.Filesize)/float64(GB)) + "GB"
	case f.Filesize >= MB:
		return formatFloat(float64(f.Filesize)/float64(MB)) + "MB"
	case f.Filesize >= KB:
		return formatFloat(float64(f.Filesize)/float64(KB)) + "KB"
	default:
		return formatFloat(float64(f.Filesize)) + "B"
	}
}

func (f FormatItem) QualityLabel() string {
	if f.Height > 0 {
		return fmt.Sprintf("%dp", f.Height)
	}
	if f.ABR > 0 {
		return fmt.Sprintf("%.0fkbps", f.ABR)
	}
	return f.FormatNote
}

func formatFloat(f float64) string {
	s := fmt.Sprintf("%.1f", f)
	if s[len(s)-1] == '0' && s[len(s)-2] == '.' {
		return s[:len(s)-2]
	}
	return s
}

// FormatCombo represents a suggested format combination
type FormatCombo struct {
	Label    string
	FormatID string
	Size     int64 // estimated total size in bytes (0 = unknown)
}

// Ensure VideoItem and FormatItem satisfy list.Item at compile time
var (
	_ list.Item = VideoItem{}
	_ list.Item = FormatItem{}
)
