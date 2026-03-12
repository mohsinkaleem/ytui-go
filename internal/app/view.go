package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mohsinkaleem/ytui-go/internal/download"
	"github.com/mohsinkaleem/ytui-go/internal/styles"
	"github.com/mohsinkaleem/ytui-go/internal/types"
)

// View implements tea.Model
func (m Model) View() string {
	if m.Width == 0 {
		return "Initializing..."
	}

	// Content area
	var content string
	switch m.State {
	case types.StateSearchInput:
		content = m.Search.View()
	case types.StateLoading:
		content = m.viewLoading()
	case types.StateVideoList:
		content = m.VideoList.View()
	case types.StateFormatList:
		content = m.FormatList.View()
	case types.StateDownload:
		content = m.Download.View()
	case types.StateVideoPlaying:
		content = m.Player.View()
	case types.StateResumeList:
		content = m.ResumeList.View()
	default:
		content = "Unknown state"
	}

	contentHeight := m.Height - 3 // reserve for status bar + padding
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Ensure content fills available space
	contentStyle := lipgloss.NewStyle().
		Height(contentHeight).
		Width(m.Width - 2)
	renderedContent := contentStyle.Render(content)

	// Status bar
	statusBar := m.viewStatusBar()

	// Compose
	return styles.AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, renderedContent, statusBar),
	)
}

func (m Model) viewLoading() string {
	var msg string
	switch m.LoadingType {
	case types.LoadingSearch:
		msg = fmt.Sprintf("Searching for \"%s\"…", m.CurrentQuery)
	case types.LoadingFormats:
		msg = "Loading formats…"
	case types.LoadingPlaylist:
		msg = "Loading playlist…"
	case types.LoadingPlay:
		msg = "Starting playback…"
	default:
		msg = "Loading…"
	}

	return fmt.Sprintf("\n\n  %s %s\n\n  %s",
		m.Spinner.View(),
		styles.AccentStyle.Render(msg),
		styles.MutedStyle.Render("Press esc to cancel"),
	)
}

func (m Model) viewStatusBar() string {
	width := m.Width - 2
	if width < 10 {
		width = 10
	}

	// Left: key hints
	left := m.currentKeysText()

	// Right: toast, error, or download indicator
	right := ""
	if m.ToastMsg != "" {
		right = styles.ToastStyle.Render("🛈 " + m.ToastMsg)
	} else if m.ErrMsg != "" {
		right = styles.ErrorStyle.Render("⚠ " + m.ErrMsg)
	}

	// Show active download indicator when not on download screen
	if m.State != types.StateDownload {
		if indicator := m.downloadIndicator(); indicator != "" {
			if right != "" {
				right = indicator + "  " + right
			} else {
				right = indicator
			}
		}
	}

	// Build status bar
	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right)
	gap := width - leftLen - rightLen
	if gap < 0 {
		gap = 0
	}

	bar := styles.StatusBarStyle.
		Width(width).
		Render(left + strings.Repeat(" ", gap) + right)

	return bar
}

// downloadIndicator returns a short status string when downloads are running.
func (m Model) downloadIndicator() string {
	if m.DownloadMgr == nil {
		return ""
	}
	tasks := m.DownloadMgr.GetTasks()
	if len(tasks) == 0 {
		return ""
	}

	active := 0
	queued := 0
	completed := 0
	for _, t := range tasks {
		switch t.GetState() {
		case download.StateDownloading:
			active++
		case download.StateQueued:
			queued++
		case download.StateCompleted:
			completed++
		}
	}

	if active == 0 && queued == 0 {
		return ""
	}

	parts := []string{}
	if active > 0 {
		// Show progress of first active task
		for _, t := range tasks {
			if t.GetState() == download.StateDownloading {
				p := t.GetProgress()
				title := t.Title
				if len([]rune(title)) > 25 {
					title = string([]rune(title)[:22]) + "..."
				}
				pct := fmt.Sprintf("%.0f%%", p.Percent*100)
				parts = append(parts, fmt.Sprintf("⇣ %s %s", title, pct))
				break
			}
		}
		if active > 1 {
			parts = append(parts, fmt.Sprintf("+%d more", active-1))
		}
	}
	if queued > 0 {
		parts = append(parts, fmt.Sprintf("%d queued", queued))
	}

	indicator := strings.Join(parts, " • ")
	return styles.InfoStyle.Render(indicator + "  (g: downloads)")
}

// currentKeysText returns context-aware key hints for the current state.
// For the Download state it adapts based on the current task's live state.
func (m Model) currentKeysText() string {
	if m.State == types.StateDownload {
		task := m.Download.CurrentTask()
		hasMultiple := len(m.Download.Tasks) > 1
		if task != nil {
			switch task.GetState() {
			case download.StateDownloading:
				if hasMultiple {
					return FormatKeys(
						KeyUp, "up", KeyDown, "down",
						KeyPause, "pause",
						KeyCancel, "cancel",
						KeyBack, "back",
						KeyEsc, "home",
					)
				}
				return FormatKeys(
					KeyPause, "pause",
					KeyCancel, "cancel",
					KeyBack, "back",
					KeyEsc, "home",
				)
			case download.StatePaused:
				if hasMultiple {
					return FormatKeys(
						KeyUp, "up", KeyDown, "down",
						KeyResume, "resume",
						KeyCancel, "cancel",
						KeyBack, "back",
						KeyEsc, "home",
					)
				}
				return FormatKeys(
					KeyResume, "resume",
					KeyCancel, "cancel",
					KeyBack, "back",
					KeyEsc, "home",
				)
			case download.StateFailed:
				if hasMultiple {
					if m.Download.HasFailedTasks() {
						return FormatKeys(
							KeyUp, "up", KeyDown, "down",
							KeyRetry, "retry",
							KeyRetryAll, "retry all",
							KeySkip, "skip",
							KeyBack, "back",
							KeyEsc, "home",
						)
					}
					return FormatKeys(
						KeyUp, "up", KeyDown, "down",
						KeyRetry, "retry",
						KeySkip, "skip",
						KeyBack, "back",
						KeyEsc, "home",
					)
				}
				return FormatKeys(
					KeyRetry, "retry",
					KeySkip, "skip",
					KeyBack, "back",
					KeyEsc, "home",
				)
			case download.StateQueued:
				if hasMultiple {
					return FormatKeys(
						KeyUp, "up", KeyDown, "down",
						KeyCancel, "cancel",
						KeyBack, "back",
						KeyEsc, "home",
					)
				}
				return FormatKeys(
					KeyCancel, "cancel",
					KeyBack, "back",
					KeyEsc, "home",
				)
			case download.StateCompleted, download.StateCancelled:
				if m.Download.AllDone() {
					return FormatKeys(KeyBack, "back", KeyEsc, "home")
				}
				if hasMultiple {
					return FormatKeys(
						KeyUp, "up", KeyDown, "down",
						KeyBack, "back",
						KeyEsc, "home",
					)
				}
				return FormatKeys(KeyBack, "back", KeyEsc, "home")
			}
		}
	}
	return GetStatusKeysText(string(m.State))
}
