package utils

import (
	"context"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mohsinkaleem/ytui-go/internal/types"
	"github.com/mohsinkaleem/ytui-go/internal/ytdlp"
)

// PlaylistManager wraps yt-dlp playlist fetch
type PlaylistManager struct {
	client *ytdlp.Client
	cancel context.CancelFunc
}

// NewPlaylistManager creates a new playlist manager
func NewPlaylistManager(client *ytdlp.Client) *PlaylistManager {
	return &PlaylistManager{client: client}
}

// FetchPlaylist starts an async playlist fetch
func (pm *PlaylistManager) FetchPlaylist(url string) tea.Cmd {
	pm.Cancel()

	ctx, cancel := context.WithCancel(context.Background())
	pm.cancel = cancel

	return func() tea.Msg {
		info, err := pm.client.GetPlaylistMetadata(ctx, url)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return types.PlaylistResultMsg{Err: err}
		}

		items := make([]list.Item, len(info.Entries))
		for i, v := range info.Entries {
			durStr := v.DurationString
			if durStr == "" && v.Duration > 0 {
				durStr = formatDuration(v.Duration)
			}
			items[i] = types.VideoItem{
				ID:             v.ID,
				Title:          v.Title,
				Channel:        v.Channel,
				Duration:       durStr,
				DurationString: durStr,
				ViewCount:      v.ViewCount,
				URL:            v.WebpageURL,
				IsLive:         v.IsLive,
			}
		}

		return types.PlaylistResultMsg{Videos: items, Title: info.Title}
	}
}

// Cancel cancels any in-progress playlist fetch
func (pm *PlaylistManager) Cancel() {
	if pm.cancel != nil {
		pm.cancel()
		pm.cancel = nil
	}
}
