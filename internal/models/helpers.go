package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/mohsinkaleem/ytui-go/internal/styles"
	"github.com/mohsinkaleem/ytui-go/internal/types"
)

// paddedIcon returns a state icon padded to exactly 2 display columns.
func paddedIcon(icon string) string {
	return padRight(icon, 2)
}

// padRight pads s with spaces to width display columns.
func padRight(s string, width int) string {
	if w := lipgloss.Width(s); w < width {
		return s + strings.Repeat(" ", width-w)
	}
	return s
}

// videoDetails joins a video's channel, duration and view count.
func videoDetails(v types.VideoItem) string {
	var parts []string
	if v.Channel != "" {
		parts = append(parts, v.Channel)
	}
	if v.DurationString != "" {
		parts = append(parts, v.DurationString)
	}
	if v.ViewCount > 0 {
		parts = append(parts, formatViewCount(v.ViewCount)+" views")
	}
	return strings.Join(parts, " • ")
}

// styleInput applies the current theme to a text input and uses a steady cursor.
func styleInput(ti *textinput.Model) {
	ti.PromptStyle = styles.InputPromptStyle
	ti.TextStyle = styles.InputStyle
	ti.PlaceholderStyle = styles.InputPlaceholder
	ti.Cursor.Style = styles.AccentStyle
	ti.Cursor.SetMode(cursor.CursorStatic)
}

// formatBytes converts a byte count to a human-readable string (e.g. "1.5 GB").
func formatBytes(n int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case n >= GB:
		return fmt.Sprintf("%.1f GB", float64(n)/float64(GB))
	case n >= MB:
		return fmt.Sprintf("%.1f MB", float64(n)/float64(MB))
	case n >= KB:
		return fmt.Sprintf("%.1f KB", float64(n)/float64(KB))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
