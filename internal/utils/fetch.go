package utils

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mohsinkaleem/ytui-go/internal/types"
	"github.com/mohsinkaleem/ytui-go/internal/ytdlp"
)

// Fetcher runs one cancellable yt-dlp metadata request at a time.
// Its methods must be called from the UI goroutine.
type Fetcher struct {
	client *ytdlp.Client
	cancel context.CancelFunc
}

// NewFetcher creates a fetcher backed by client.
func NewFetcher(client *ytdlp.Client) *Fetcher {
	return &Fetcher{client: client}
}

// Cancel aborts the in-flight request, if any.
func (f *Fetcher) Cancel() {
	if f.cancel != nil {
		f.cancel()
		f.cancel = nil
	}
}

func (f *Fetcher) begin() context.Context {
	f.Cancel()
	ctx, cancel := context.WithCancel(context.Background())
	f.cancel = cancel
	return ctx
}

// Search returns a command that searches YouTube.
func (f *Fetcher) Search(query string, sortBy types.SortBy) tea.Cmd {
	ctx := f.begin()
	return func() tea.Msg {
		videos, err := f.client.Search(ctx, query, string(sortBy))
		if ctx.Err() != nil {
			return nil
		}
		return types.SearchResultMsg{Videos: toListItems(videos), Query: query, Err: err}
	}
}

// Playlist returns a command that fetches a playlist's entries.
func (f *Fetcher) Playlist(url string) tea.Cmd {
	ctx := f.begin()
	return func() tea.Msg {
		info, err := f.client.GetPlaylistMetadata(ctx, url)
		if ctx.Err() != nil {
			return nil
		}
		if err != nil {
			return types.PlaylistResultMsg{Err: err}
		}
		return types.PlaylistResultMsg{Videos: toListItems(info.Entries), Title: info.Title}
	}
}

// Formats returns a command that fetches a video's metadata and formats.
func (f *Fetcher) Formats(url string) tea.Cmd {
	ctx := f.begin()
	return func() tea.Msg {
		info, err := f.client.GetVideoMetadata(ctx, url)
		if ctx.Err() != nil {
			return nil
		}
		if err != nil {
			return types.FormatResultMsg{Err: err}
		}
		video := toVideoItem(*info)
		video.URL = url
		formats := ytdlp.ConvertFormatsToItems(ytdlp.SortFormats(info.Formats))
		return types.FormatResultMsg{Formats: formats, Video: video}
	}
}

func toListItems(videos []ytdlp.VideoInfo) []list.Item {
	items := make([]list.Item, len(videos))
	for i, v := range videos {
		items[i] = toVideoItem(v)
	}
	return items
}

func toVideoItem(v ytdlp.VideoInfo) types.VideoItem {
	duration := v.DurationString
	if duration == "" && v.Duration > 0 {
		duration = formatDuration(v.Duration)
	}
	channel := v.Channel
	if channel == "" {
		channel = v.Uploader
	}
	return types.VideoItem{
		ID:             v.ID,
		Title:          v.Title,
		Channel:        channel,
		DurationString: duration,
		ViewCount:      v.ViewCount,
		URL:            v.WebpageURL,
		IsLive:         v.IsLive || v.LiveStatus == "is_live",
	}
}

// formatDuration converts total seconds (float64) into a human-readable "H:MM:SS" or "M:SS" string.
func formatDuration(secs float64) string {
	total := int(secs)
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}
