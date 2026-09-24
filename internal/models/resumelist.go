package models

import (
	"fmt"
	"slices"
	"strings"

	"github.com/mohsinkaleem/ytui-go/internal/download"
	"github.com/mohsinkaleem/ytui-go/internal/store"
	"github.com/mohsinkaleem/ytui-go/internal/styles"
)

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

// RemoveSelected drops the highlighted record from the list.
func (m *ResumeListModel) RemoveSelected() {
	if m.SelectedItem() == nil {
		return
	}
	m.Items = slices.Delete(m.Items, m.Selected, m.Selected+1)
	m.Selected = max(min(m.Selected, len(m.Items)-1), 0)
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
	b.WriteString(styles.SectionHeaderStyle.Render("Unfinished downloads"))
	b.WriteString(styles.MutedStyle.Render(fmt.Sprintf("  %d", count)))
	b.WriteString("\n\n")

	if count == 0 {
		b.WriteString(styles.MutedStyle.Render("Nothing to resume."))
		b.WriteString("\n")
		return b.String()
	}

	// Two lines per item; header, blank and position line take three.
	rows := min(max((m.Height-3)/2, 1), count)
	start := max(min(m.Selected-rows+1, count-rows), 0)
	end := start + rows
	width := m.Width - 3 // left border/indent

	for i := start; i < end; i++ {
		rec := m.Items[i]
		title := paddedIcon(download.TaskState(rec.State).Icon()) + " " + padRight(rec.State, 11) + " " + rec.Title

		detail := "format " + rec.FormatID
		switch {
		case rec.State == string(download.StateFailed) && rec.Error != "":
			detail += " • " + rec.Error
		case rec.OutputPath != "":
			detail += " • → " + styles.TruncateLeft(rec.OutputPath, width-len(detail)-8)
		}

		titleLine := styles.Truncate(title, width)
		detailLine := styles.Truncate("   "+detail, width)
		if i == m.Selected {
			b.WriteString(styles.ListSelectedItemStyle.Render(
				styles.ListSelectedTitleStyle.Render(titleLine) + "\n" + styles.ListSelectedDescStyle.Render(detailLine)))
		} else {
			b.WriteString(styles.ListItemStyle.Render(
				styles.TextStyle.Render(titleLine) + "\n" + styles.ListDescStyle.Render(detailLine)))
		}
		b.WriteString("\n")
	}

	if rows < count {
		b.WriteString(styles.MutedStyle.Render(fmt.Sprintf("   %d–%d of %d", start+1, end, count)))
		b.WriteString("\n")
	}

	return b.String()
}
