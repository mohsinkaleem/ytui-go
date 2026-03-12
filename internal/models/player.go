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

	b.WriteString("\n\n")
	b.WriteString(styles.AccentStyle.Render("♫ Now Playing"))
	b.WriteString("\n\n")
	b.WriteString(styles.VideoTitleStyle.Render(m.Video.Title))
	b.WriteString("\n")
	if m.Video.Channel != "" {
		b.WriteString(styles.VideoDetailStyle.Render("📺 " + m.Video.Channel))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(styles.MutedStyle.Render("mpv is running... press q in mpv to return"))

	return b.String()
}
