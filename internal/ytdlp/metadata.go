package ytdlp

// VideoInfo represents metadata for a single video
type VideoInfo struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Channel        string   `json:"channel"`
	Uploader       string   `json:"uploader"`
	Duration       float64  `json:"duration"`
	DurationString string   `json:"duration_string"`
	ViewCount      int64    `json:"view_count"`
	WebpageURL     string   `json:"webpage_url"`
	Formats        []Format `json:"formats"`
	IsLive         bool     `json:"is_live"`
	LiveStatus     string   `json:"live_status"`
	URL            string   `json:"url"`
}

// Format represents a single downloadable format
type Format struct {
	FormatID       string  `json:"format_id"`
	Ext            string  `json:"ext"`
	Resolution     string  `json:"resolution"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	FPS            float64 `json:"fps"`
	VCodec         string  `json:"vcodec"`
	ACodec         string  `json:"acodec"`
	Filesize       int64   `json:"filesize"`
	FilesizeApprox int64   `json:"filesize_approx"`
	TBR            float64 `json:"tbr"`
	VBR            float64 `json:"vbr"`
	ABR            float64 `json:"abr"`
	FormatNote     string  `json:"format_note"`
	Protocol       string  `json:"protocol"`
}

// PlaylistInfo represents metadata for a playlist
type PlaylistInfo struct {
	ID            string      `json:"id"`
	Title         string      `json:"title"`
	Channel       string      `json:"channel"`
	PlaylistCount int         `json:"playlist_count"`
	Entries       []VideoInfo `json:"entries"`
}

// DownloadOpts configures yt-dlp download behavior
type DownloadOpts struct {
	EmbedSubs     bool
	EmbedMetadata bool
	EmbedChapters bool
	OutputDir     string
}

// Progress represents parsed progress data from yt-dlp
type Progress struct {
	DownloadedBytes int64
	TotalBytes      int64
	Speed           string
	ETA             string
	Status          string
	Percent         float64
	Title           string
}

// IsVideoOnly checks if format has only video
func (f Format) IsVideoOnly() bool {
	return f.VCodec != "" && f.VCodec != "none" && (f.ACodec == "" || f.ACodec == "none")
}

// IsAudioOnly checks if format has only audio
func (f Format) IsAudioOnly() bool {
	return f.ACodec != "" && f.ACodec != "none" && (f.VCodec == "" || f.VCodec == "none")
}

// HasBoth checks if format has both video and audio
func (f Format) HasBoth() bool {
	return f.VCodec != "" && f.VCodec != "none" && f.ACodec != "" && f.ACodec != "none"
}

// EffectiveSize returns filesize or approx
func (f Format) EffectiveSize() int64 {
	if f.Filesize > 0 {
		return f.Filesize
	}
	return f.FilesizeApprox
}
