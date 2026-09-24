package models

import (
	"strings"

	"github.com/mohsinkaleem/ytui-go/internal/styles"
	"github.com/mohsinkaleem/ytui-go/internal/types"
)

// PlayerModel represents the "now playing" state
type PlayerModel struct {
	Video    types.VideoItem
	FormatID string
	Width    int
	Height   int
}

// NewPlayerModel creates a new player model
func NewPlayerModel() PlayerModel {
	return PlayerModel{}
}

// SetVideo sets the currently playing video
func (m *PlayerModel) SetVideo(video types.VideoItem, formatID string) {
	m.Video = video
	m.FormatID = formatID
}

// SetSize updates dimensions
func (m *PlayerModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

// View renders the now playing screen
func (m *PlayerModel) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styles.AccentStyle.Render("▶ Playing in mpv"))
	b.WriteString("\n\n")
	b.WriteString(styles.VideoTitleStyle.Render(styles.Truncate(m.Video.Title, m.Width)))
	b.WriteString("\n")
	if details := videoDetails(m.Video); details != "" {
		b.WriteString(styles.VideoDetailStyle.Render(styles.Truncate(details, m.Width)))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(styles.MutedStyle.Render("Quit mpv (q) to return."))

	return b.String()
}
