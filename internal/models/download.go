package models

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"github.com/mohsinkaleem/ytui-go/internal/download"
	"github.com/mohsinkaleem/ytui-go/internal/styles"
	"github.com/mohsinkaleem/ytui-go/internal/types"
)

// paddedIcon returns a state icon padded to exactly 2 display columns.
func paddedIcon(icon string) string {
	w := lipgloss.Width(icon)
	if w < 2 {
		return icon + strings.Repeat(" ", 2-w)
	}
	return icon
}

// DownloadModel represents the download progress screen
type DownloadModel struct {
	Video      types.VideoItem
	Progress   progress.Model
	Tasks      []*download.Task
	CurrentIdx int
	CursorIdx  int // user-navigable cursor for queue selection
	Width      int
	Height     int
}

// NewDownloadModel creates a new download model
func NewDownloadModel() DownloadModel {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithoutPercentage(),
	)
	p.FullColor = string(styles.ProgressFillColor)

	return DownloadModel{
		Progress: p,
	}
}

// SetSize updates dimensions
func (m *DownloadModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.Progress.Width = w - 8
}

// SetTasks sets the download tasks, preserving cursor position when possible
func (m *DownloadModel) SetTasks(tasks []*download.Task) {
	prevCursor := m.CursorIdx
	m.Tasks = tasks
	m.CurrentIdx = 0
	if prevCursor >= 0 && prevCursor < len(tasks) {
		m.CursorIdx = prevCursor
	} else {
		m.CursorIdx = 0
	}
}

// CurrentTask returns the task at the user's cursor position, or first active.
func (m *DownloadModel) CurrentTask() *download.Task {
	if len(m.Tasks) == 0 {
		return nil
	}
	// If cursor points to a valid task, use it
	if m.CursorIdx >= 0 && m.CursorIdx < len(m.Tasks) {
		return m.Tasks[m.CursorIdx]
	}
	// Fallback: first active task
	for i, t := range m.Tasks {
		if t.GetState() == download.StateDownloading {
			m.CursorIdx = i
			return t
		}
	}
	return m.Tasks[0]
}

// MoveUp moves the queue cursor up.
func (m *DownloadModel) MoveUp() {
	if m.CursorIdx > 0 {
		m.CursorIdx--
	}
}

// MoveDown moves the queue cursor down.
func (m *DownloadModel) MoveDown() {
	if m.CursorIdx < len(m.Tasks)-1 {
		m.CursorIdx++
	}
}

// AllDone checks if all tasks are completed/cancelled/failed
func (m *DownloadModel) AllDone() bool {
	for _, t := range m.Tasks {
		state := t.GetState()
		if state != download.StateCompleted && state != download.StateCancelled && state != download.StateFailed {
			return false
		}
	}
	return true
}

// HasFailedTasks returns true if more than one task has failed state
func (m *DownloadModel) HasFailedTasks() bool {
	count := 0
	for _, t := range m.Tasks {
		if t.GetState() == download.StateFailed {
			count++
		}
	}
	return count > 1
}

// CompletionSummary returns counts of completed/failed/cancelled
func (m *DownloadModel) CompletionSummary() (completed, failed, cancelled int) {
	for _, t := range m.Tasks {
		switch t.GetState() {
		case download.StateCompleted:
			completed++
		case download.StateFailed:
			failed++
		case download.StateCancelled:
			cancelled++
		}
	}
	return
}

// View renders the download screen
func (m *DownloadModel) View() string {
	var b strings.Builder

	task := m.CurrentTask()
	if task == nil {
		return styles.MutedStyle.Render("No active downloads")
	}

	// ── Header: queue overview ──────────────────────────────────────────────
	if len(m.Tasks) > 1 {
		completed, failed, cancelled := m.CompletionSummary()
		done := completed + failed + cancelled
		active := 0
		queued := 0
		for _, t := range m.Tasks {
			switch t.GetState() {
			case download.StateDownloading:
				active++
			case download.StateQueued:
				queued++
			}
		}
		header := fmt.Sprintf("Queue  %d/%d done", done, len(m.Tasks))
		if active > 0 {
			header += fmt.Sprintf("  •  %d active", active)
		}
		if queued > 0 {
			header += fmt.Sprintf("  •  %d waiting", queued)
		}
		b.WriteString(styles.SectionHeaderStyle.Render(header))
		b.WriteString("\n\n")
	}

	// ── Current task ────────────────────────────────────────────────────────
	b.WriteString(styles.VideoTitleStyle.Render(task.Title))
	b.WriteString("\n")

	if task.FormatID != "" {
		b.WriteString(styles.MutedStyle.Render("format: " + task.FormatID))
		b.WriteString("\n")
	}

	progress := task.GetProgress()
	state := task.GetState()

	b.WriteString("\n")

	switch state {
	case download.StateQueued:
		// Show queue position
		pos := 1
		for i, t := range m.Tasks {
			if t.ID == task.ID {
				pos = i + 1
				break
			}
		}
		b.WriteString(styles.WarningStyle.Render(fmt.Sprintf("○ Waiting in queue  (position %d of %d)…", pos, len(m.Tasks))))
		b.WriteString("\n")

	case download.StateDownloading:
		b.WriteString(styles.AccentStyle.Render("⇣ Downloading"))
		b.WriteString("\n")

		pct := progress.Percent
		b.WriteString(m.Progress.ViewAs(pct))
		b.WriteString("\n")

		// Size + percentage line
		sizeLine := fmt.Sprintf("  %.0f%%", pct*100)
		if progress.DownloadedBytes > 0 {
			if progress.TotalBytes > 0 {
				sizeLine = fmt.Sprintf("  %s / %s  (%.0f%%)",
					formatBytes(progress.DownloadedBytes),
					formatBytes(progress.TotalBytes),
					pct*100,
				)
			} else {
				sizeLine = fmt.Sprintf("  %s downloaded  (%.0f%%)",
					formatBytes(progress.DownloadedBytes),
					pct*100,
				)
			}
		}
		b.WriteString(styles.SpeedStyle.Render(sizeLine))
		b.WriteString("\n")

		speedParts := []string{}
		if progress.Speed != "" {
			speedParts = append(speedParts, progress.Speed)
		}
		if progress.ETA != "" {
			speedParts = append(speedParts, "ETA "+progress.ETA)
		}
		if len(speedParts) > 0 {
			b.WriteString(styles.SpeedStyle.Render("  " + strings.Join(speedParts, "  •  ")))
			b.WriteString("\n")
		}

		if task.OutputPath != "" {
			b.WriteString(styles.PathStyle.Render("  → " + task.OutputPath))
			b.WriteString("\n")
		}

	case download.StatePaused:
		b.WriteString(styles.WarningStyle.Render("⏸ Paused"))
		b.WriteString("\n")

		b.WriteString(m.Progress.ViewAs(progress.Percent))
		b.WriteString("\n")

		// Size + percentage line
		sizeLine := fmt.Sprintf("  %.0f%% downloaded", progress.Percent*100)
		if progress.DownloadedBytes > 0 {
			if progress.TotalBytes > 0 {
				sizeLine = fmt.Sprintf("  %s / %s  (%.0f%%)",
					formatBytes(progress.DownloadedBytes),
					formatBytes(progress.TotalBytes),
					progress.Percent*100,
				)
			} else {
				sizeLine = fmt.Sprintf("  %s downloaded  (%.0f%%)",
					formatBytes(progress.DownloadedBytes),
					progress.Percent*100,
				)
			}
		}
		b.WriteString(styles.SpeedStyle.Render(sizeLine))
		b.WriteString("\n")

		if task.OutputPath != "" {
			b.WriteString(styles.PathStyle.Render("  → " + task.OutputPath))
			b.WriteString("\n")
		}

	case download.StateCompleted:
		b.WriteString(styles.SuccessStyle.Render("✓ Complete"))
		b.WriteString("\n")
		if task.OutputPath != "" {
			b.WriteString(styles.PathStyle.Render("→ " + task.OutputPath))
			b.WriteString("\n")
		}

	case download.StateFailed:
		b.WriteString(styles.ErrorStyle.Render("✕ Failed"))
		b.WriteString("\n")
		if task.GetError() != nil {
			b.WriteString(styles.ErrorStyle.Render("  " + task.GetError().Error()))
			b.WriteString("\n")
		}

	case download.StateCancelled:
		b.WriteString(styles.MutedStyle.Render("✕ Cancelled"))
		b.WriteString("\n")
	}

	// ── Completion summary (all done) ───────────────────────────────────────
	if m.AllDone() {
		completed, failed, cancelled := m.CompletionSummary()
		if len(m.Tasks) > 1 {
			b.WriteString("\n")
			summary := []string{}
			if completed > 0 {
				summary = append(summary, styles.SuccessStyle.Render(fmt.Sprintf("✓ %d completed", completed)))
			}
			if failed > 0 {
				summary = append(summary, styles.ErrorStyle.Render(fmt.Sprintf("✗ %d failed", failed)))
			}
			if cancelled > 0 {
				summary = append(summary, styles.WarningStyle.Render(fmt.Sprintf("→ %d skipped", cancelled)))
			}
			b.WriteString(strings.Join(summary, "  "))
			b.WriteString("\n")
		}
		// Show the output folder (from any completed task)
		for _, t := range m.Tasks {
			if t.GetState() == download.StateCompleted && t.OutputPath != "" {
				dir := filepath.Dir(t.OutputPath)
				if dir != "." && dir != "" {
					b.WriteString(styles.MutedStyle.Render("  Saved to: " + dir))
					b.WriteString("\n")
				}
				break
			}
		}

	}

	// ── Queue list ──────────────────────────────────────────────────────────
	if len(m.Tasks) > 1 {
		b.WriteString("\n")
		b.WriteString(styles.BoldStyle.Render("Queue:"))
		b.WriteString("  ")
		b.WriteString(styles.MutedStyle.Render("↑/↓ navigate  p pause/resume  c cancel  b back  esc home"))
		b.WriteString("\n")

		maxShow := 10
		if len(m.Tasks) < maxShow {
			maxShow = len(m.Tasks)
		}

		// Scrolling: keep cursor in view
		start := 0
		if m.CursorIdx >= maxShow {
			start = m.CursorIdx - maxShow + 1
		}
		end := start + maxShow
		if end > len(m.Tasks) {
			end = len(m.Tasks)
		}

		for i := start; i < end; i++ {
			t := m.Tasks[i]
			s := t.GetState()
			isCursor := i == m.CursorIdx
			var st lipgloss.Style

			switch s {
			case download.StateCompleted:
				st = styles.QueueItemCompleteStyle
			case download.StateDownloading:
				st = styles.QueueItemActiveStyle
			case download.StateFailed:
				st = styles.QueueItemErrorStyle
			case download.StatePaused:
				st = styles.QueueItemPausedStyle
			default:
				st = styles.QueueItemPendingStyle
			}

			title := t.Title
			maxLen := m.Width - 26
			if maxLen < 10 {
				maxLen = 10
			}
			if len([]rune(title)) > maxLen {
				title = string([]rune(title)[:maxLen-3]) + "..."
			}

			stateLabel := fmt.Sprintf("[%s]", string(s))
			iconCol := paddedIcon(s.Icon())

			// Build line: cursor(2) + icon(2) + space + state(padded 14) + title
			cursor := "  "
			if isCursor {
				cursor = styles.AccentStyle.Render("▸ ")
			}
			line := fmt.Sprintf("%s%s %-14s %s", cursor, iconCol, stateLabel, title)

			if isCursor {
				b.WriteString(styles.ListSelectedItemStyle.Render(st.Render(line)))
			} else {
				// 3 spaces matches ListSelectedItemStyle indent (border 1 + padding 2)
				b.WriteString("   " + st.Render(line))
			}

			// Show format info on a second line for the cursor item
			if isCursor && t.FormatID != "" {
				fmtLine := fmt.Sprintf("    fmt:%s", t.FormatID)
				if t.GetOutputPath() != "" {
					path := t.GetOutputPath()
					maxPath := m.Width - 12
					if maxPath > 60 {
						maxPath = 60
					}
					if len(path) > maxPath {
						path = "…" + path[len(path)-maxPath+1:]
					}
					fmtLine += "  •  → " + path
				}
				b.WriteString("\n" + styles.ListSelectedItemStyle.Render(styles.MutedStyle.Render(fmtLine)))
			}
			b.WriteString("\n")
		}

		if len(m.Tasks) > end {
			b.WriteString(styles.MutedStyle.Render(fmt.Sprintf("  … and %d more", len(m.Tasks)-end)))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// formatBytes converts a byte count to a human-readable string.
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
