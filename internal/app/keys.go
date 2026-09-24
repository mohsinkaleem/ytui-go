package app

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/mohsinkaleem/ytui-go/internal/download"
	"github.com/mohsinkaleem/ytui-go/internal/styles"
	"github.com/mohsinkaleem/ytui-go/internal/types"
)

// keyHints returns alternating key/description pairs for the status bar,
// most important first. An empty key renders the description alone.
func (m Model) keyHints() []string {
	switch m.State {
	case types.StateSearchInput:
		switch {
		case m.Search.ShowThemes:
			return []string{"↑/↓", "select", "enter", "apply", "esc", "cancel"}
		case m.Search.ShowSlash:
			return []string{"↑/↓", "select", "tab", "complete", "enter", "run", "esc", "close"}
		}
		hints := []string{"enter", "search", "/", "commands", "tab", "sort"}
		if len(m.Search.History) > 0 {
			hints = append(hints, "↑", "history")
		}
		return append(hints, "ctrl+c", "quit")

	case types.StateLoading:
		return []string{"esc", "cancel"}

	case types.StateVideoList:
		if m.VideoList.IsFiltering() {
			return []string{"enter", "apply filter", "esc", "cancel"}
		}
		return []string{"enter", "formats", "d", "download", "p", "play", "space", "select", "a", "all", "/", "filter", "b", "back", "ctrl+y", "copy URL"}

	case types.StateFormatList:
		if m.FormatList.ActiveTab == types.FormatTabCustom {
			return []string{"↑/↓", "preset", "enter", "download", "tab", "switch tab", "esc", "home"}
		}
		return []string{"enter", "download", "p", "play", "tab", "switch tab", "b", "back", "esc", "home"}

	case types.StateDownload:
		return m.downloadHints()

	case types.StateVideoPlaying:
		return []string{"", "mpv is running — quit it to return"}

	case types.StateResumeList:
		return []string{"enter", "resume", "d", "resume all", "x", "delete", "b", "back"}
	}
	return nil
}

// downloadHints adapts to the state of the task under the cursor.
func (m Model) downloadHints() []string {
	var hints []string
	if len(m.Download.Tasks) > 1 {
		hints = append(hints, "↑/↓", "select")
	}
	if t := m.Download.CurrentTask(); t != nil {
		switch t.GetState() {
		case download.StateDownloading:
			hints = append(hints, "p", "pause", "c", "cancel")
		case download.StatePaused:
			hints = append(hints, "p", "resume", "c", "cancel")
		case download.StateQueued:
			hints = append(hints, "c", "cancel")
		case download.StateFailed:
			hints = append(hints, "r", "retry")
			if m.Download.FailedCount() > 1 {
				hints = append(hints, "R", "retry all")
			}
			hints = append(hints, "s", "skip")
		}
	}
	return append(hints, "b", "back", "esc", "home")
}

// renderHints renders key/description pairs, dropping trailing pairs that don't fit in width.
func renderHints(width int, pairs ...string) string {
	sep := styles.StatusSepStyle.Render(" • ")
	var out string
	for i := 0; i+1 < len(pairs); i += 2 {
		part := styles.StatusDescStyle.Render(pairs[i+1])
		if pairs[i] != "" {
			part = styles.StatusKeyStyle.Render(pairs[i]) + styles.StatusDescStyle.Render(" "+pairs[i+1])
		}
		if out != "" {
			part = sep + part
		}
		if lipgloss.Width(out)+lipgloss.Width(part) > width {
			break
		}
		out += part
	}
	return out
}
