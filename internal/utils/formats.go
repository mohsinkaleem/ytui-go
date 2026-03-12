package utils

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mohsinkaleem/ytui-go/internal/types"
	"github.com/mohsinkaleem/ytui-go/internal/ytdlp"
)

// FormatsManager wraps yt-dlp format fetch with cancellation
type FormatsManager struct {
	client *ytdlp.Client
	cancel context.CancelFunc
}

// NewFormatsManager creates a new formats manager
func NewFormatsManager(client *ytdlp.Client) *FormatsManager {
	return &FormatsManager{client: client}
}

// FetchFormats starts an async format fetch, returning a tea.Cmd
func (fm *FormatsManager) FetchFormats(url string) tea.Cmd {
	fm.Cancel()

	ctx, cancel := context.WithCancel(context.Background())
	fm.cancel = cancel

	return func() tea.Msg {
		info, err := fm.client.GetVideoMetadata(ctx, url)
		if err != nil {
			if ctx.Err() != nil {
				return nil // cancelled
			}
			return types.FormatResultMsg{Err: err}
		}

		formats := ytdlp.SortFormats(info.Formats)
		items := ytdlp.ConvertFormatsToItems(formats)

		durStr := info.DurationString
		if durStr == "" && info.Duration > 0 {
			durStr = formatDuration(info.Duration)
		}

		video := types.VideoItem{
			ID:             info.ID,
			Title:          info.Title,
			Channel:        info.Channel,
			Duration:       durStr,
			DurationString: durStr,
			ViewCount:      info.ViewCount,
			URL:            url,
		}

		return types.FormatResultMsg{Formats: items, Video: video}
	}
}

// Cancel cancels any in-progress format fetch
func (fm *FormatsManager) Cancel() {
	if fm.cancel != nil {
		fm.cancel()
		fm.cancel = nil
	}
}
