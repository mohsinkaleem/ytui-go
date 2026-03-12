package utils

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mohsinkaleem/ytui-go/internal/types"
	"github.com/mohsinkaleem/ytui-go/internal/ytdlp"
)

// SearchManager wraps yt-dlp search with cancellation
type SearchManager struct {
	client *ytdlp.Client
	cancel context.CancelFunc
}

// NewSearchManager creates a new search manager
func NewSearchManager(client *ytdlp.Client) *SearchManager {
	return &SearchManager{client: client}
}

// Search starts an async search, returning a tea.Cmd
func (sm *SearchManager) Search(query string, sortBy types.SortBy) tea.Cmd {
	// Cancel any existing search
	sm.Cancel()

	ctx, cancel := context.WithCancel(context.Background())
	sm.cancel = cancel

	return func() tea.Msg {
		videos, err := sm.client.Search(ctx, query, string(sortBy))
		if err != nil {
			if ctx.Err() != nil {
				return nil // cancelled
			}
			return types.SearchResultMsg{Err: err, Query: query}
		}

		items := make([]list.Item, len(videos))
		for i, v := range videos {
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
				UploadDate:     v.UploadDate,
				URL:            v.WebpageURL,
				IsLive:         v.IsLive,
			}
		}

		return types.SearchResultMsg{Videos: items, Query: query}
	}
}

// Cancel cancels any in-progress search
func (sm *SearchManager) Cancel() {
	if sm.cancel != nil {
		sm.cancel()
		sm.cancel = nil
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
