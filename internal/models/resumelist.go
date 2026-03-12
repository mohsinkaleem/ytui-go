package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mohsinkaleem/ytui-go/internal/store"
	"github.com/mohsinkaleem/ytui-go/internal/styles"
)

// paddedResumeIcon returns a state icon padded to exactly 2 display columns.
func paddedResumeIcon(icon string) string {
	w := lipgloss.Width(icon)
	if w < 2 {
		return icon + strings.Repeat(" ", 2-w)
	}
	return icon
}

// ResumeListModel displays the list of incomplete (paused/failed/queued) downloads
// so the user can inspect and re-queue them.
type ResumeListModel struct {
	Items    []store.DownloadRecord
	Selected int
	Width    int
	Height   int
}

// NewResumeListModel creates an empty resume list model.
func NewResumeListModel() ResumeListModel {
	return ResumeListModel{}
}

// SetItems replaces the item list and resets the cursor.
func (m *ResumeListModel) SetItems(items []store.DownloadRecord) {
	m.Items = items
	m.Selected = 0
}

// SetSize updates the terminal dimensions.
func (m *ResumeListModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

// SelectedItem returns a pointer to the currently highlighted record, or nil.
func (m *ResumeListModel) SelectedItem() *store.DownloadRecord {
	if len(m.Items) == 0 || m.Selected < 0 || m.Selected >= len(m.Items) {
		return nil
	}
	return &m.Items[m.Selected]
}

// MoveUp moves the cursor up.
func (m *ResumeListModel) MoveUp() {
	if m.Selected > 0 {
		m.Selected--
	}
}

// MoveDown moves the cursor down.
func (m *ResumeListModel) MoveDown() {
	if m.Selected < len(m.Items)-1 {
		m.Selected++
	}
}

// View renders the incomplete-downloads list.
func (m *ResumeListModel) View() string {
	var b strings.Builder

	count := len(m.Items)
	headerText := fmt.Sprintf("Unfinished Downloads  (%d)", count)
	b.WriteString(styles.SectionHeaderStyle.Render(headerText))
	b.WriteString("\n\n")

	if count == 0 {
		b.WriteString(styles.MutedStyle.Render("  No unfinished downloads."))
		b.WriteString("\n")
		return b.String()
	}

	// Each item takes ~3 visual lines (title + detail + gap)
	linesPerItem := 3
	maxShow := (m.Height - 8) / linesPerItem
	if maxShow < 1 {
		maxShow = 1
	}
	if maxShow > count {
		maxShow = count
	}

	// Scrolling: keep selected item in view
	start := 0
	if m.Selected >= maxShow {
		start = m.Selected - maxShow + 1
	}
	end := start + maxShow
	if end > count {
		end = count
	}

	maxTitleLen := m.Width - 16
	if maxTitleLen < 20 {
		maxTitleLen = 20
	}

	for i := start; i < end; i++ {
		rec := m.Items[i]
		isSelected := i == m.Selected

		icon := paddedResumeIcon(resumeStateIcon(rec.State))
		stateTag := fmt.Sprintf("%-12s", "["+rec.State+"]")

		title := rec.Title
		if len([]rune(title)) > maxTitleLen {
			title = string([]rune(title)[:maxTitleLen-3]) + "..."
		}

		// Build detail parts on separate lines for clarity
		var detailParts []string
		if rec.FormatID != "" {
			detailParts = append(detailParts, "fmt:"+rec.FormatID)
		}
		if rec.OutputPath != "" {
			path := rec.OutputPath
			maxPath := m.Width - 10
			if maxPath > 70 {
				maxPath = 70
			}
			if len(path) > maxPath {
				path = "…" + path[len(path)-maxPath+1:]
			}
			detailParts = append(detailParts, "→ "+path)
		}
		detail := strings.Join(detailParts, "  •  ")

		if isSelected {
			titleLine := fmt.Sprintf("%s %s %s", icon, stateTag, title)
			rendered := styles.ListSelectedTitleStyle.Render(titleLine)
			if detail != "" {
				rendered += "\n" + styles.ListSelectedDescStyle.Render("     "+detail)
			}
			b.WriteString(styles.ListSelectedItemStyle.Render(rendered))
		} else {
			titleLine := fmt.Sprintf("%s %s ", icon, stateTag)
			// 3 spaces matches ListSelectedItemStyle indent (border 1 + padding 2)
			b.WriteString("   " + styles.ListTitleStyle.Render(titleLine) + styles.ListTitleStyle.Render(title))
			if detail != "" {
				b.WriteString("\n" + "   " + styles.ListDescStyle.Render("     "+detail))
			}
		}
		b.WriteString("\n")
	}

	if count > end {
		b.WriteString(styles.MutedStyle.Render(fmt.Sprintf("\n  … and %d more", count-end)))
		b.WriteString("\n")
	}

	// Footer hints
	b.WriteString("\n")
	b.WriteString(styles.MutedStyle.Render("  enter: resume selected  •  d: resume all  •  x: delete  •  esc: back"))
	b.WriteString("\n")

	return b.String()
}

// resumeStateIcon returns a unicode icon for a stored state string.
func resumeStateIcon(state string) string {
	switch state {
	case "queued":
		return "○"
	case "downloading":
		return "↓"
	case "completed":
		return "✓"
	case "failed":
		return "✗"
	case "cancelled":
		return "→"
	case "paused":
		return "⏸"
	default:
		return "?"
	}
}
