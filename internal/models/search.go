package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/mohsinkaleem/ytui-go/internal/slash"
	"github.com/mohsinkaleem/ytui-go/internal/styles"
	"github.com/mohsinkaleem/ytui-go/internal/types"
)

// SearchModel represents the search input screen
type SearchModel struct {
	Input         textinput.Model
	SortBy        types.SortBy
	SortIndex     int
	EmbedSubs     bool
	EmbedMetadata bool
	EmbedChapters bool
	ErrMsg        string
	Width         int
	Height        int

	// Slash command state
	ShowSlash     bool
	SlashMatches  []slash.Command
	SlashSelected int
	SlashRegistry *slash.Registry

	// Theme picker state
	ShowThemes    bool
	ThemeNames    []string
	ThemeSelected int

	// History
	History      []string
	HistoryIndex int
}

// NewSearchModel creates a new search model
func NewSearchModel(registry *slash.Registry) SearchModel {
	ti := textinput.New()
	ti.Placeholder = "paste url or search..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 40
	ti.Prompt = "$ "
	ti.PromptStyle = styles.InputPromptStyle
	ti.TextStyle = styles.InputStyle
	ti.PlaceholderStyle = styles.InputPlaceholder

	return SearchModel{
		Input:         ti,
		SortBy:        types.SortRelevance,
		SortIndex:     0,
		SlashRegistry: registry,
		HistoryIndex:  -1,
	}
}

// SetSize updates dimensions
func (m *SearchModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.Input.Width = w - 6
}

// View renders the search screen
func (m *SearchModel) View() string {
	var b strings.Builder

	// Logo
	logo := styles.LogoStyle.Render("ytui-go")
	subtitle := styles.LogoSubStyle.Render("YouTube from your terminal")
	b.WriteString("\n")
	b.WriteString(logo)
	b.WriteString("\n")
	b.WriteString(subtitle)
	b.WriteString("\n\n")

	// Input
	b.WriteString(m.Input.View())
	b.WriteString("\n")
	maxW := m.Width - 4
	if maxW > 40 {
		maxW = 40
	}
	if maxW < 1 {
		maxW = 1
	}
	b.WriteString(styles.MutedStyle.Render(strings.Repeat("-", maxW)))
	b.WriteString("\n")

	// Theme picker dropdown
	if m.ShowThemes && len(m.ThemeNames) > 0 {
		b.WriteString("\n")
		b.WriteString(styles.BoldStyle.Render("  Select theme:"))
		b.WriteString("\n")
		for i, name := range m.ThemeNames {
			if i == m.ThemeSelected {
				b.WriteString(styles.SlashSelectedStyle.Render("  > " + name))
			} else {
				b.WriteString(styles.SlashNormalStyle.Render("    " + name))
			}
			b.WriteString("\n")
		}
		b.WriteString(styles.MutedStyle.Render("  ↑↓ navigate  enter select  esc cancel"))
		b.WriteString("\n")
	} else if m.ShowSlash && len(m.SlashMatches) > 0 {
		// Slash command dropdown
		b.WriteString("\n")
		maxVisible := 5
		total := len(m.SlashMatches)
		if total < maxVisible {
			maxVisible = total
		}
		// Scroll so the selected item is always visible
		start := m.SlashSelected - maxVisible + 1
		if start < 0 {
			start = 0
		}
		if m.SlashSelected < start {
			start = m.SlashSelected
		}
		end := start + maxVisible
		if end > total {
			end = total
		}
		for i := start; i < end; i++ {
			cmd := m.SlashMatches[i]
			if i == m.SlashSelected {
				b.WriteString(styles.SlashSelectedStyle.Render("  > /" + cmd.Name))
			} else {
				b.WriteString(styles.SlashNormalStyle.Render("    /" + cmd.Name))
			}
			if cmd.Description != "" {
				b.WriteString("  ")
				b.WriteString(styles.SlashDescStyle.Render(cmd.Description))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Sort By
	b.WriteString("\n")
	b.WriteString(styles.SortLabelStyle.Render("Sort By: "))
	b.WriteString(styles.SortValueStyle.Render("> " + m.SortBy.String()))
	b.WriteString("\n")

	// Download Options
	b.WriteString("\n")
	b.WriteString(styles.BoldStyle.Render("Download Options"))
	b.WriteString("\n")

	type opt struct {
		name    string
		enabled bool
		key     string
	}
	options := []opt{
		{"Embed subtitles", m.EmbedSubs, "ctrl+s"},
		{"Embed metadata", m.EmbedMetadata, "ctrl+m"},
		{"Embed chapters", m.EmbedChapters, "ctrl+j"},
	}

	for _, o := range options {
		icon := "o"
		style := styles.OptionOffStyle
		if o.enabled {
			icon = "*"
			style = styles.OptionOnStyle
		}
		b.WriteString(fmt.Sprintf("  %s  ", style.Render(icon)))
		b.WriteString(style.Render(o.name))
		b.WriteString("  ")
		b.WriteString(styles.OptionKeyStyle.Render(o.key))
		b.WriteString("\n")
	}

	// Error
	if m.ErrMsg != "" {
		b.WriteString("\n")
		b.WriteString(styles.ErrorStyle.Render("! " + m.ErrMsg))
	}

	// Search history hint
	if !m.ShowSlash && !m.ShowThemes && len(m.History) > 0 {
		b.WriteString("\n")
		b.WriteString(styles.MutedStyle.Render("  ↑↓ browse search history"))
		maxShow := 5
		if len(m.History) < maxShow {
			maxShow = len(m.History)
		}
		b.WriteString("\n")
		for i := 0; i < maxShow; i++ {
			prefix := "    "
			style := styles.MutedStyle
			if i == m.HistoryIndex {
				prefix = "  > "
				style = styles.BoldStyle
			}
			b.WriteString(style.Render(prefix + m.History[i]))
			b.WriteString("\n")
		}
	}

	return b.String()
}
