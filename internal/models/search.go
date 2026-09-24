package models

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/mohsinkaleem/ytui-go/internal/slash"
	"github.com/mohsinkaleem/ytui-go/internal/styles"
	"github.com/mohsinkaleem/ytui-go/internal/types"
)

// maxBoxWidth caps the search box width on wide terminals.
const maxBoxWidth = 72

// SearchModel represents the search input screen
type SearchModel struct {
	Input     textinput.Model
	SortIndex int
	ErrMsg    string
	Width     int
	Height    int

	// Settings shown on screen
	EmbedSubs     bool
	EmbedMetadata bool
	EmbedChapters bool
	DownloadDir   string
	Cookies       string

	// Slash command state
	ShowSlash     bool
	SlashMatches  []slash.Command
	SlashSelected int
	SlashHint     string // usage line shown while typing a command's arguments

	// Theme picker state
	ShowThemes    bool
	ThemeNames    []string
	ThemeSelected int

	// History
	History      []string
	HistoryIndex int
}

// NewSearchModel creates a new search model
func NewSearchModel() SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Search YouTube or paste a URL…  (/ for commands)"
	ti.Focus()
	ti.CharLimit = 500
	ti.Prompt = "❯ "

	m := SearchModel{Input: ti, HistoryIndex: -1}
	m.ApplyTheme()
	m.SetSize(maxBoxWidth, 24)
	return m
}

// ApplyTheme re-applies the current theme to the input.
func (m *SearchModel) ApplyTheme() {
	styleInput(&m.Input)
}

// SetSize updates dimensions
func (m *SearchModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	// box border (2) + padding (2) + prompt + cursor cell
	m.Input.Width = max(m.boxWidth()-4-lipgloss.Width(m.Input.Prompt)-1, 5)
}

func (m *SearchModel) boxWidth() int {
	return max(min(m.Width, maxBoxWidth), 12)
}

// SortBy returns the selected sort order.
func (m *SearchModel) SortBy() types.SortBy {
	return types.SortOptions[m.SortIndex]
}

// CycleSort moves the sort selection by delta, wrapping around.
func (m *SearchModel) CycleSort(delta int) {
	n := len(types.SortOptions)
	m.SortIndex = ((m.SortIndex+delta)%n + n) % n
}

// SetInput replaces the input text and moves the cursor to the end.
func (m *SearchModel) SetInput(s string) {
	m.Input.SetValue(s)
	m.Input.CursorEnd()
}

// ClearInput empties the input and closes the command dropdown.
func (m *SearchModel) ClearInput() {
	m.SetInput("")
	m.CloseSlash()
	m.HistoryIndex = -1
}

// CloseSlash hides the command dropdown and usage hint.
func (m *SearchModel) CloseSlash() {
	m.ShowSlash, m.SlashMatches, m.SlashSelected, m.SlashHint = false, nil, 0, ""
}

// OpenThemes shows the theme picker with the current theme preselected.
func (m *SearchModel) OpenThemes(names []string, current string) {
	m.ShowThemes = true
	m.ThemeNames = names
	m.ThemeSelected = max(slices.Index(names, current), 0)
}

// CloseThemes hides the theme picker.
func (m *SearchModel) CloseThemes() {
	m.ShowThemes, m.ThemeNames = false, nil
}

// PushHistory records entry as the most recent history item.
func (m *SearchModel) PushHistory(entry string) {
	rest := slices.DeleteFunc(m.History, func(h string) bool { return h == entry })
	m.History = append([]string{entry}, rest...)
	m.HistoryIndex = -1
}

// HistoryPrev recalls the next older history entry.
func (m *SearchModel) HistoryPrev() {
	if m.HistoryIndex < len(m.History)-1 {
		m.HistoryIndex++
		m.SetInput(m.History[m.HistoryIndex])
	}
}

// HistoryNext recalls the next newer entry, clearing the input past the newest.
func (m *SearchModel) HistoryNext() {
	switch {
	case m.HistoryIndex > 0:
		m.HistoryIndex--
		m.SetInput(m.History[m.HistoryIndex])
	case m.HistoryIndex == 0:
		m.HistoryIndex = -1
		m.SetInput("")
	}
}

// View renders the search screen
func (m *SearchModel) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styles.LogoStyle.Render("ytui-go"))
	b.WriteString("  ")
	b.WriteString(styles.LogoSubStyle.Render("YouTube from your terminal"))
	b.WriteString("\n\n")

	b.WriteString(styles.InputBoxStyle.Width(m.boxWidth() - 2).Render(m.Input.View()))
	b.WriteString("\n")

	// The dropdowns replace the rest of the screen so they never push it off-screen.
	switch {
	case m.ShowThemes:
		b.WriteString(m.themesView())
		return b.String()
	case m.ShowSlash:
		b.WriteString(m.slashView())
		return b.String()
	case m.SlashHint != "":
		b.WriteString(styles.MutedStyle.Render(styles.Truncate("  "+m.SlashHint, m.Width)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.sortView())
	b.WriteString("\n\n")
	b.WriteString(m.optionsView())

	if m.ErrMsg != "" {
		b.WriteString("\n")
		b.WriteString(styles.ErrorStyle.Width(m.Width).Render("✗ " + m.ErrMsg))
		b.WriteString("\n")
	}

	return b.String()
}

func (m *SearchModel) themesView() string {
	var b strings.Builder
	b.WriteString(styles.BoldStyle.Render("  Theme"))
	b.WriteString("\n")
	for i, name := range m.ThemeNames {
		if i == m.ThemeSelected {
			b.WriteString(styles.SlashSelectedStyle.Render("  ❯ " + name))
		} else {
			b.WriteString(styles.SlashNormalStyle.Render("    " + name))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m *SearchModel) slashView() string {
	if len(m.SlashMatches) == 0 {
		return styles.MutedStyle.Render("  No matching command") + "\n"
	}

	const maxVisible = 8
	total := len(m.SlashMatches)
	start := max(0, min(m.SlashSelected-maxVisible+1, total-maxVisible))
	end := min(start+maxVisible, total)

	nameW := 0
	for _, c := range m.SlashMatches[start:end] {
		nameW = max(nameW, lipgloss.Width(CommandUsage(c)))
	}

	var b strings.Builder
	for i := start; i < end; i++ {
		c := m.SlashMatches[i]
		name := padRight(CommandUsage(c), nameW)
		if i == m.SlashSelected {
			b.WriteString(styles.SlashSelectedStyle.Render("  ❯ " + name))
		} else {
			b.WriteString(styles.SlashNormalStyle.Render("    " + name))
		}
		b.WriteString(styles.SlashDescStyle.Render(styles.Truncate("  "+c.Description, m.Width-nameW-4)))
		b.WriteString("\n")
	}
	if total > maxVisible {
		b.WriteString(styles.MutedStyle.Render(fmt.Sprintf("    %d/%d", m.SlashSelected+1, total)))
		b.WriteString("\n")
	}
	return b.String()
}

func (m *SearchModel) sortView() string {
	parts := make([]string, len(types.SortOptions))
	for i, s := range types.SortOptions {
		if i == m.SortIndex {
			parts[i] = styles.SortValueStyle.Render(s.String())
		} else {
			parts[i] = styles.MutedStyle.Render(s.String())
		}
	}
	return styles.SortLabelStyle.Render("Sort by  ") +
		strings.Join(parts, styles.MutedStyle.Render(" · ")) +
		"  " + styles.OptionKeyStyle.Render("tab")
}

func (m *SearchModel) optionsView() string {
	var b strings.Builder
	b.WriteString(styles.BoldStyle.Render("Download options"))
	b.WriteString("\n")

	options := []struct {
		name    string
		enabled bool
		key     string
	}{
		{"Embed subtitles", m.EmbedSubs, "ctrl+s"},
		{"Embed metadata", m.EmbedMetadata, "ctrl+t"},
		{"Embed chapters", m.EmbedChapters, "ctrl+j"},
	}
	for _, o := range options {
		icon, style := "○", styles.OptionOffStyle
		if o.enabled {
			icon, style = "●", styles.OptionOnStyle
		}
		fmt.Fprintf(&b, "  %s %s  %s\n", style.Render(icon), style.Render(padRight(o.name, 16)), styles.OptionKeyStyle.Render(o.key))
	}

	settings := []struct{ label, value string }{
		{"Save to", m.DownloadDir},
		{"Cookies", m.Cookies},
	}
	for _, s := range settings {
		if s.value == "" {
			continue
		}
		fmt.Fprintf(&b, "  %s %s\n", styles.SortLabelStyle.Render(padRight(s.label, 8)), styles.TextStyle.Render(styles.TruncateLeft(s.value, m.Width-11)))
	}
	return b.String()
}

// CommandUsage renders a command with its argument placeholder, e.g. "/play <url>".
func CommandUsage(c slash.Command) string {
	if c.Args == "" {
		return "/" + c.Name
	}
	return "/" + c.Name + " " + c.Args
}
