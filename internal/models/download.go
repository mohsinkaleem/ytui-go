package models

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"github.com/mohsinkaleem/ytui-go/internal/download"
	"github.com/mohsinkaleem/ytui-go/internal/styles"
	"github.com/mohsinkaleem/ytui-go/internal/ytdlp"
)

// DownloadModel represents the download progress screen
type DownloadModel struct {
	Progress  progress.Model
	Tasks     []*download.Task
	CursorIdx int // user-navigable cursor for queue selection
	Width     int
	Height    int
}

// NewDownloadModel creates a new download model
func NewDownloadModel() DownloadModel {
	m := DownloadModel{}
	m.ApplyTheme()
	return m
}

// ApplyTheme rebuilds the progress bar with the current theme's colors.
func (m *DownloadModel) ApplyTheme() {
	t := styles.CurrentTheme
	m.Progress = progress.New(
		progress.WithGradient(string(t.Accent), string(t.Pink)),
		progress.WithoutPercentage(),
	)
	m.Progress.EmptyColor = string(t.Overlay)
	m.SetSize(m.Width, m.Height)
}

// SetSize updates dimensions
func (m *DownloadModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.Progress.Width = max(min(w, 60), 10)
}

// SetTasks sets the download tasks, preserving cursor position when possible
func (m *DownloadModel) SetTasks(tasks []*download.Task) {
	m.Tasks = tasks
	m.CursorIdx = max(min(m.CursorIdx, len(tasks)-1), 0)
}

// Focus moves the cursor to the task with the given ID, if present.
func (m *DownloadModel) Focus(id string) {
	for i, t := range m.Tasks {
		if t.ID == id {
			m.CursorIdx = i
			return
		}
	}
}

// CurrentTask returns the task at the user's cursor position.
func (m *DownloadModel) CurrentTask() *download.Task {
	if m.CursorIdx < 0 || m.CursorIdx >= len(m.Tasks) {
		return nil
	}
	return m.Tasks[m.CursorIdx]
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
		if !t.GetState().IsTerminal() {
			return false
		}
	}
	return true
}

// FailedCount returns the number of failed tasks
func (m *DownloadModel) FailedCount() int {
	count := 0
	for _, t := range m.Tasks {
		if t.GetState() == download.StateFailed {
			count++
		}
	}
	return count
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
	task := m.CurrentTask()
	if task == nil {
		return styles.MutedStyle.Render("No downloads yet.")
	}

	var b strings.Builder
	b.WriteString(m.headerView())
	b.WriteString("\n\n")
	b.WriteString(m.taskView(task))
	if m.AllDone() {
		b.WriteString(m.summaryView())
	}
	if len(m.Tasks) > 1 {
		b.WriteString(m.queueView(m.Height - strings.Count(b.String(), "\n")))
	}
	return b.String()
}

func (m *DownloadModel) headerView() string {
	completed, failed, cancelled := m.CompletionSummary()
	active, queued := 0, 0
	for _, t := range m.Tasks {
		switch t.GetState() {
		case download.StateDownloading:
			active++
		case download.StateQueued:
			queued++
		}
	}
	meta := fmt.Sprintf("  %d/%d done", completed+failed+cancelled, len(m.Tasks))
	if active > 0 {
		meta += fmt.Sprintf(" · %d active", active)
	}
	if queued > 0 {
		meta += fmt.Sprintf(" · %d waiting", queued)
	}
	return styles.SectionHeaderStyle.Render("Downloads") + styles.MutedStyle.Render(meta)
}

// taskView renders the details of the task under the cursor.
func (m *DownloadModel) taskView(t *download.Task) string {
	var b strings.Builder
	state := t.GetState()

	b.WriteString(styles.VideoTitleStyle.Render(styles.Truncate(t.GetTitle(), m.Width)))
	b.WriteString("\n")
	b.WriteString(styles.MutedStyle.Render(styles.Truncate("format "+t.FormatID, m.Width)))
	b.WriteString("\n\n")

	switch state {
	case download.StateQueued:
		b.WriteString(styles.WarningStyle.Render("○ Waiting in queue…"))
		b.WriteString("\n")

	case download.StateDownloading, download.StatePaused:
		p := t.GetProgress()
		label := styles.AccentStyle.Render(state.Label())
		if state == download.StatePaused {
			label = styles.WarningStyle.Render(state.Label())
		}
		b.WriteString(label + "  " + styles.BoldStyle.Render(fmt.Sprintf("%.0f%%", p.Percent*100)))
		b.WriteString("\n")
		b.WriteString(m.Progress.ViewAs(p.Percent))
		b.WriteString("\n")
		if stats := progressStats(p, state == download.StateDownloading); stats != "" {
			b.WriteString(styles.SpeedStyle.Render(stats))
			b.WriteString("\n")
		}

	case download.StateCompleted:
		b.WriteString(styles.SuccessStyle.Render(state.Label()))
		b.WriteString("\n")

	case download.StateFailed:
		b.WriteString(styles.ErrorStyle.Render(state.Label()))
		b.WriteString("\n")
		if err := t.GetError(); err != nil {
			b.WriteString(styles.ErrorStyle.Width(m.Width).Render(err.Error()))
			b.WriteString("\n")
		}

	case download.StateCancelled:
		b.WriteString(styles.MutedStyle.Render(state.Label()))
		b.WriteString("\n")
	}

	if path := t.GetOutputPath(); path != "" {
		b.WriteString(styles.PathStyle.Render("→ " + styles.TruncateLeft(path, m.Width-2)))
		b.WriteString("\n")
	}
	return b.String()
}

// progressStats formats size, speed and ETA; speed/ETA only while live.
func progressStats(p ytdlp.Progress, live bool) string {
	var parts []string
	switch {
	case p.TotalBytes > 0:
		parts = append(parts, formatBytes(p.DownloadedBytes)+" / "+formatBytes(p.TotalBytes))
	case p.DownloadedBytes > 0:
		parts = append(parts, formatBytes(p.DownloadedBytes))
	}
	if live && p.Speed != "" {
		parts = append(parts, p.Speed)
	}
	if live && p.ETA != "" {
		parts = append(parts, "ETA "+p.ETA)
	}
	return strings.Join(parts, "  •  ")
}

func (m *DownloadModel) summaryView() string {
	var b strings.Builder
	if len(m.Tasks) > 1 {
		completed, failed, cancelled := m.CompletionSummary()
		var parts []string
		if completed > 0 {
			parts = append(parts, styles.SuccessStyle.Render(fmt.Sprintf("✓ %d completed", completed)))
		}
		if failed > 0 {
			parts = append(parts, styles.ErrorStyle.Render(fmt.Sprintf("✗ %d failed", failed)))
		}
		if cancelled > 0 {
			parts = append(parts, styles.WarningStyle.Render(fmt.Sprintf("→ %d skipped", cancelled)))
		}
		b.WriteString("\n")
		b.WriteString(strings.Join(parts, "  "))
		b.WriteString("\n")
	}
	for _, t := range m.Tasks {
		if path := t.GetOutputPath(); path != "" && t.GetState() == download.StateCompleted {
			b.WriteString(styles.MutedStyle.Render("Saved to " + styles.TruncateLeft(filepath.Dir(path), m.Width-9)))
			b.WriteString("\n")
			break
		}
	}
	return b.String()
}

// queueView renders as many queue rows as fit in avail lines, keeping the cursor visible.
func (m *DownloadModel) queueView(avail int) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(styles.BoldStyle.Render("Queue"))
	b.WriteString("\n")

	total := len(m.Tasks)
	rows := min(max(avail-3, 1), total) // blank + title + position line
	start := max(min(m.CursorIdx-rows/2, total-rows), 0)
	end := start + rows

	const labelW = 11
	titleW := m.Width - 3 - 3 - labelW - 1 // indent, icon + space, label, space
	for i := start; i < end; i++ {
		t := m.Tasks[i]
		s := t.GetState()
		label := string(s)
		switch s {
		case download.StateDownloading:
			label = fmt.Sprintf("%.0f%%", t.GetProgress().Percent*100)
		case download.StatePaused:
			label = fmt.Sprintf("paused %.0f%%", t.GetProgress().Percent*100)
		}
		line := paddedIcon(s.Icon()) + " " + padRight(label, labelW) + " " + styles.Truncate(t.GetTitle(), titleW)

		if i == m.CursorIdx {
			b.WriteString(styles.ListSelectedItemStyle.Render(queueStyle(s).Bold(true).Render(line)))
		} else {
			b.WriteString(styles.ListItemStyle.Render(queueStyle(s).Render(line)))
		}
		b.WriteString("\n")
	}

	if rows < total {
		b.WriteString(styles.MutedStyle.Render(fmt.Sprintf("   %d–%d of %d", start+1, end, total)))
		b.WriteString("\n")
	}
	return b.String()
}

func queueStyle(s download.TaskState) lipgloss.Style {
	switch s {
	case download.StateCompleted:
		return styles.QueueItemCompleteStyle
	case download.StateDownloading:
		return styles.QueueItemActiveStyle
	case download.StateFailed:
		return styles.QueueItemErrorStyle
	case download.StatePaused:
		return styles.QueueItemPausedStyle
	default:
		return styles.QueueItemPendingStyle
	}
}
