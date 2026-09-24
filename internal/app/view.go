package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mohsinkaleem/ytui-go/internal/download"
	"github.com/mohsinkaleem/ytui-go/internal/styles"
	"github.com/mohsinkaleem/ytui-go/internal/types"
)

// contentSize is the area available to screens: minus side padding and the status bar.
func (m Model) contentSize() (width, height int) {
	return max(m.Width-2, 1), max(m.Height-1, 1)
}

// View implements tea.Model
func (m Model) View() string {
	if m.Width == 0 {
		return ""
	}

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
	}

	// Clamp to the screen: taller output would scroll the top of the view away.
	_, h := m.contentSize()
	body := styles.AppStyle.Width(m.Width).Height(h).MaxHeight(h).Render(content)
	return body + "\n" + m.viewStatusBar()
}

func (m Model) viewLoading() string {
	w, _ := m.contentSize()
	return "\n" + m.Spinner.View() + styles.AccentStyle.Render(styles.Truncate(m.LoadingMsg, w-3))
}

func (m Model) viewStatusBar() string {
	width, _ := m.contentSize()

	// Right: toast, or the download indicator when not on the download screen
	var right string
	switch {
	case m.Toast != "" && m.ToastErr:
		right = styles.StatusErrorStyle.Render(styles.Truncate("✗ "+m.Toast, width))
	case m.Toast != "":
		right = styles.StatusInfoStyle.Render(styles.Truncate(m.Toast, width))
	case m.State != types.StateDownload:
		right = m.downloadIndicator()
	}

	leftWidth := width - lipgloss.Width(right)
	if right != "" {
		leftWidth -= 2
	}
	left := renderHints(leftWidth, m.keyHints()...)
	gap := max(width-lipgloss.Width(left)-lipgloss.Width(right), 0)

	return styles.StatusBarStyle.Width(m.Width).Padding(0, 1).
		Render(left + styles.StatusBarStyle.Render(strings.Repeat(" ", gap)) + right)
}

// downloadIndicator summarizes running downloads for the status bar.
func (m Model) downloadIndicator() string {
	var active, queued int
	var current *download.Task
	for _, t := range m.DownloadMgr.GetTasks() {
		switch t.GetState() {
		case download.StateDownloading:
			if current == nil {
				current = t
			}
			active++
		case download.StateQueued:
			queued++
		}
	}
	if active == 0 && queued == 0 {
		return ""
	}

	var parts []string
	if current != nil {
		pct := current.GetProgress().Percent * 100
		parts = append(parts, fmt.Sprintf("⇣ %s %.0f%%", styles.Truncate(current.GetTitle(), 24), pct))
		if active > 1 {
			parts = append(parts, fmt.Sprintf("+%d", active-1))
		}
	}
	if queued > 0 {
		parts = append(parts, fmt.Sprintf("%d queued", queued))
	}

	indicator := styles.StatusInfoStyle.Render(strings.Join(parts, " · "))
	switch {
	case m.canJumpToDownloads():
		indicator += styles.StatusKeyStyle.Render("  g") + styles.StatusDescStyle.Render(" downloads")
	case m.State == types.StateSearchInput:
		indicator += styles.StatusKeyStyle.Render("  /downloads")
	}
	return indicator
}
