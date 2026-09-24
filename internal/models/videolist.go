package models

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

	titleStyle, descStyle := styles.ListTitleStyle, styles.ListDescStyle
	if index == m.Index() {
		titleStyle, descStyle = styles.ListSelectedTitleStyle, styles.ListSelectedDescStyle
	}
	width := m.Width() - 3 // left border/indent

	title := titleStyle.Render(styles.Truncate(item.Title, width))
	if item.Selected {
		title = styles.MultiSelectCheckStyle.Render("✓ ") + titleStyle.Render(styles.Truncate(item.Title, width-2))
	}

	detail := descStyle.Render(styles.Truncate(videoDetails(item), width))
	if item.IsLive {
		detail = styles.LiveBadgeStyle.Render("● LIVE ") + descStyle.Render(styles.Truncate(videoDetails(item), width-7))
	}

	if index == m.Index() {
		fmt.Fprint(w, styles.ListSelectedItemStyle.Render(title+"\n"+detail))
	} else {
		fmt.Fprint(w, styles.ListItemStyle.Render(title+"\n"+detail))
	}
}

// NewVideoListModel creates a new video list model
func NewVideoListModel() VideoListModel {
	l := newList(videoItemDelegate{})
	l.SetFilteringEnabled(true)
	l.SetStatusBarItemName("video", "videos")
	l.FilterInput.Prompt = "Filter: "
	l.Styles.TitleBar = l.Styles.TitleBar.PaddingLeft(3) // align the filter with the items

	m := VideoListModel{List: l}
	m.ApplyTheme()
	return m
}

// newList creates a bare list whose keys don't clash with the app's own.
func newList(delegate list.ItemDelegate) list.Model {
	l := list.New(nil, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()         // the app owns quitting (ctrl+c); "q" must not exit
	l.KeyMap.GoToStart.SetKeys("home") // "g" opens the downloads screen
	return l
}

// ApplyTheme re-applies the current theme to the filter input.
func (m *VideoListModel) ApplyTheme() {
	styleInput(&m.List.FilterInput)
}

// SetItems replaces the list contents and resets filter and cursor
func (m *VideoListModel) SetItems(items []list.Item, title, query string) {
	m.List.ResetFilter()
	m.List.SetItems(items)
	m.List.ResetSelected()
	m.Title = title
	m.Query = query
}

// SetSize updates dimensions
func (m *VideoListModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.List.SetSize(w, h-1) // header; the list's own first row is the blank/filter line
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
func (m *VideoListModel) ToggleSelected() tea.Cmd {
	idx := m.List.GlobalIndex() // Index() is relative to the filtered view
	items := m.List.Items()
	if idx < 0 || idx >= len(items) {
		return nil
	}
	v, ok := items[idx].(types.VideoItem)
	if !ok {
		return nil
	}
	v.Selected = !v.Selected
	return m.List.SetItem(idx, v)
}

// SelectAll selects every item, or clears the selection if all are selected
func (m *VideoListModel) SelectAll() tea.Cmd {
	return m.setAllSelected(m.SelectedCount() < len(m.List.Items()))
}

// ClearSelections deselects all items in the list
func (m *VideoListModel) ClearSelections() tea.Cmd {
	return m.setAllSelected(false)
}

func (m *VideoListModel) setAllSelected(selected bool) tea.Cmd {
	var cmd tea.Cmd
	for i, item := range m.List.Items() {
		if v, ok := item.(types.VideoItem); ok && v.Selected != selected {
			v.Selected = selected
			cmd = m.List.SetItem(i, v) // each returns the same refilter command
		}
	}
	return cmd
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

// SelectedCount returns the number of multi-selected videos
func (m *VideoListModel) SelectedCount() int {
	return len(m.GetSelectedVideos())
}

// View renders the video list screen
func (m *VideoListModel) View() string {
	header := m.Title
	if header == "" {
		header = fmt.Sprintf("Results for %q", m.Query)
	}

	n := len(m.List.Items())
	meta := fmt.Sprintf("  %d videos", n)
	if n == 1 {
		meta = "  1 video"
	}
	selected := ""
	if count := m.SelectedCount(); count > 0 {
		selected = fmt.Sprintf(" · %d selected", count)
	}

	header = styles.Truncate(header, m.Width-lipgloss.Width(meta+selected))
	return styles.SectionHeaderStyle.Render(header) +
		styles.MutedStyle.Render(meta) +
		styles.MultiSelectCheckStyle.Render(selected) +
		"\n" + m.List.View()
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
