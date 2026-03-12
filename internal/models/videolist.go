package models

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mohsinkaleem/ytui-go/internal/styles"
	"github.com/mohsinkaleem/ytui-go/internal/types"
)

// VideoListModel represents the video list screen
type VideoListModel struct {
	List   list.Model
	Title  string
	Query  string
	Width  int
	Height int
}

// videoItemDelegate is a custom delegate for rendering video items
type videoItemDelegate struct{}

func (d videoItemDelegate) Height() int                             { return 2 }
func (d videoItemDelegate) Spacing() int                            { return 1 }
func (d videoItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d videoItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(types.VideoItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	// Title line
	title := item.Title
	maxLen := m.Width() - 6
	if maxLen < 10 {
		maxLen = 10
	}
	if len([]rune(title)) > maxLen {
		title = string([]rune(title)[:maxLen-3]) + "..."
	}

	// Detail line
	details := []string{}
	if item.Channel != "" {
		details = append(details, item.Channel)
	}
	if item.DurationString != "" {
		details = append(details, item.DurationString)
	}
	if item.ViewCount > 0 {
		details = append(details, formatViewCount(item.ViewCount)+" views")
	}
	detail := strings.Join(details, " • ")

	if isSelected {
		titleStr := styles.ListSelectedTitleStyle.Render(title)
		detailStr := styles.ListSelectedDescStyle.Render(detail)
		content := titleStr + "\n" + detailStr

		rendered := styles.ListSelectedItemStyle.Render(content)
		fmt.Fprint(w, rendered)
	} else {
		prefix := "   "
		if item.Selected {
			prefix = " " + styles.MultiSelectCheckStyle.Render("✓") + " "
		}

		titleStr := styles.ListTitleStyle.Render(title)
		detailStr := styles.ListDescStyle.Render(detail)

		fmt.Fprint(w, prefix+titleStr+"\n"+prefix+detailStr)
	}
}

// NewVideoListModel creates a new video list model
func NewVideoListModel() VideoListModel {
	delegate := videoItemDelegate{}
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)

	// Style the filter
	l.FilterInput.PromptStyle = styles.InputPromptStyle
	l.FilterInput.TextStyle = styles.InputStyle

	return VideoListModel{
		List: l,
	}
}

// SetItems sets the list items
func (m *VideoListModel) SetItems(items []list.Item, title, query string) {
	m.List.SetItems(items)
	m.Title = title
	m.Query = query
}

// SetSize updates dimensions
func (m *VideoListModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.List.SetSize(w, h-4) // reserve space for header
}

// SelectedVideo returns the currently highlighted video
func (m *VideoListModel) SelectedVideo() (types.VideoItem, bool) {
	item := m.List.SelectedItem()
	if item == nil {
		return types.VideoItem{}, false
	}
	v, ok := item.(types.VideoItem)
	return v, ok
}

// ToggleSelected toggles multi-select on the current item
func (m *VideoListModel) ToggleSelected() {
	idx := m.List.Index()
	items := m.List.Items()
	if idx < 0 || idx >= len(items) {
		return
	}
	if v, ok := items[idx].(types.VideoItem); ok {
		v.Selected = !v.Selected
		m.List.SetItem(idx, v)
	}
}

// SelectAll toggles all items
func (m *VideoListModel) SelectAll() {
	items := m.List.Items()
	// Check if all are already selected
	allSelected := true
	for _, item := range items {
		if v, ok := item.(types.VideoItem); ok && !v.Selected {
			allSelected = false
			break
		}
	}

	for i, item := range items {
		if v, ok := item.(types.VideoItem); ok {
			v.Selected = !allSelected
			m.List.SetItem(i, v)
		}
	}
}

// GetSelectedVideos returns all multi-selected videos
func (m *VideoListModel) GetSelectedVideos() []types.VideoItem {
	var selected []types.VideoItem
	for _, item := range m.List.Items() {
		if v, ok := item.(types.VideoItem); ok && v.Selected {
			selected = append(selected, v)
		}
	}
	return selected
}

// ClearSelections deselects all items in the list
func (m *VideoListModel) ClearSelections() {
	for i, item := range m.List.Items() {
		if v, ok := item.(types.VideoItem); ok && v.Selected {
			v.Selected = false
			m.List.SetItem(i, v)
		}
	}
}

// View renders the video list screen
func (m *VideoListModel) View() string {
	var b strings.Builder

	// Header
	headerText := "Search Results"
	if m.Query != "" {
		headerText = fmt.Sprintf("Search Results for: \"%s\"", m.Query)
	}
	if m.Title != "" {
		headerText = m.Title
	}
	header := styles.SectionHeaderStyle.Render(headerText)
	b.WriteString(header)
	b.WriteString("\n\n")

	// List
	b.WriteString(m.List.View())

	return b.String()
}

// IsFiltering returns true if the list is in filter mode
func (m *VideoListModel) IsFiltering() bool {
	return m.List.FilterState() == list.Filtering
}

func formatViewCount(count int64) string {
	switch {
	case count >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", float64(count)/1_000_000_000)
	case count >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(count)/1_000_000)
	case count >= 1_000:
		return fmt.Sprintf("%.1fK", float64(count)/1_000)
	default:
		return fmt.Sprintf("%d", count)
	}
}
