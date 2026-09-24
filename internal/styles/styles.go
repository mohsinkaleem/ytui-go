package styles

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Package-level style variables — rebuilt when theme changes
var (
	// Layout
	AppStyle lipgloss.Style

	// Section headers
	SectionHeaderStyle lipgloss.Style

	// Status bar; every segment carries the bar background so inner resets don't punch holes in it
	StatusBarStyle   lipgloss.Style
	StatusKeyStyle   lipgloss.Style
	StatusDescStyle  lipgloss.Style
	StatusSepStyle   lipgloss.Style
	StatusInfoStyle  lipgloss.Style
	StatusErrorStyle lipgloss.Style

	// Input
	InputStyle       lipgloss.Style
	InputPromptStyle lipgloss.Style
	InputPlaceholder lipgloss.Style
	InputBoxStyle    lipgloss.Style

	// Text variants
	TextStyle    lipgloss.Style
	MutedStyle   lipgloss.Style
	BoldStyle    lipgloss.Style
	ErrorStyle   lipgloss.Style
	SuccessStyle lipgloss.Style
	WarningStyle lipgloss.Style
	InfoStyle    lipgloss.Style
	AccentStyle  lipgloss.Style

	// Spinner
	SpinnerStyle lipgloss.Style

	// List items
	ListTitleStyle         lipgloss.Style
	ListDescStyle          lipgloss.Style
	ListSelectedTitleStyle lipgloss.Style
	ListSelectedDescStyle  lipgloss.Style
	ListItemStyle          lipgloss.Style
	ListSelectedItemStyle  lipgloss.Style
	MultiSelectCheckStyle  lipgloss.Style
	LiveBadgeStyle         lipgloss.Style
	TableHeaderStyle       lipgloss.Style

	// Tab bar
	TabActiveStyle   lipgloss.Style
	TabInactiveStyle lipgloss.Style

	// Download queue
	QueueItemPendingStyle  lipgloss.Style
	QueueItemActiveStyle   lipgloss.Style
	QueueItemCompleteStyle lipgloss.Style
	QueueItemErrorStyle    lipgloss.Style
	QueueItemPausedStyle   lipgloss.Style

	// Logo
	LogoStyle    lipgloss.Style
	LogoSubStyle lipgloss.Style

	// Option toggles
	OptionOnStyle  lipgloss.Style
	OptionOffStyle lipgloss.Style
	OptionKeyStyle lipgloss.Style

	// Sort
	SortLabelStyle lipgloss.Style
	SortValueStyle lipgloss.Style

	// Video info header
	VideoTitleStyle  lipgloss.Style
	VideoDetailStyle lipgloss.Style

	// Slash command dropdown
	SlashSelectedStyle lipgloss.Style
	SlashNormalStyle   lipgloss.Style
	SlashDescStyle     lipgloss.Style

	// Speed / ETA
	SpeedStyle lipgloss.Style
	PathStyle  lipgloss.Style
)

func init() {
	rebuildStyles()
}

func rebuildStyles() {
	t := CurrentTheme

	// Layout
	AppStyle = lipgloss.NewStyle().Padding(0, 1)

	// Section headers
	SectionHeaderStyle = lipgloss.NewStyle().
		Foreground(t.Text).
		Bold(true)

	// Status bar
	bar := lipgloss.NewStyle().Background(t.Surface)
	StatusBarStyle = bar.Foreground(t.Secondary)
	StatusKeyStyle = bar.Foreground(t.Pink).Bold(true)
	StatusDescStyle = bar.Foreground(t.Secondary)
	StatusSepStyle = bar.Foreground(t.Muted)
	StatusInfoStyle = bar.Foreground(t.Info)
	StatusErrorStyle = bar.Foreground(t.Error).Bold(true)

	// Input
	InputStyle = lipgloss.NewStyle().
		Foreground(t.Text)

	InputPromptStyle = lipgloss.NewStyle().
		Foreground(t.Pink).
		Bold(true)

	InputPlaceholder = lipgloss.NewStyle().
		Foreground(t.Muted)

	InputBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Accent).
		Padding(0, 1)

	// Text variants
	TextStyle = lipgloss.NewStyle().Foreground(t.Text)
	MutedStyle = lipgloss.NewStyle().Foreground(t.Muted)
	BoldStyle = lipgloss.NewStyle().Bold(true).Foreground(t.Text)
	ErrorStyle = lipgloss.NewStyle().Foreground(t.Error)
	SuccessStyle = lipgloss.NewStyle().Foreground(t.Success)
	WarningStyle = lipgloss.NewStyle().Foreground(t.Warning)
	InfoStyle = lipgloss.NewStyle().Foreground(t.Info)
	AccentStyle = lipgloss.NewStyle().Foreground(t.Accent)

	// Spinner
	SpinnerStyle = lipgloss.NewStyle().Foreground(t.Pink)

	// List items — selected uses left border indicator
	ListItemStyle = lipgloss.NewStyle().PaddingLeft(3)
	ListSelectedItemStyle = lipgloss.NewStyle().
		PaddingLeft(2).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(t.Accent)

	ListTitleStyle = lipgloss.NewStyle().
		Foreground(t.Text).
		Bold(true)

	ListDescStyle = lipgloss.NewStyle().
		Foreground(t.Secondary)

	ListSelectedTitleStyle = lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	ListSelectedDescStyle = lipgloss.NewStyle().
		Foreground(t.Pink)

	MultiSelectCheckStyle = lipgloss.NewStyle().
		Foreground(t.Success).
		Bold(true)

	LiveBadgeStyle = lipgloss.NewStyle().
		Foreground(t.Error).
		Bold(true)

	TableHeaderStyle = lipgloss.NewStyle().
		Foreground(t.Muted).
		Bold(true)

	// Tab bar
	TabActiveStyle = lipgloss.NewStyle().
		Foreground(t.Base).
		Background(t.Accent).
		Padding(0, 2).
		Bold(true)

	TabInactiveStyle = lipgloss.NewStyle().
		Foreground(t.Secondary).
		Padding(0, 2)

	// Download queue
	QueueItemPendingStyle = lipgloss.NewStyle().Foreground(t.Muted)
	QueueItemActiveStyle = lipgloss.NewStyle().Foreground(t.Accent)
	QueueItemCompleteStyle = lipgloss.NewStyle().Foreground(t.Success)
	QueueItemErrorStyle = lipgloss.NewStyle().Foreground(t.Error)
	QueueItemPausedStyle = lipgloss.NewStyle().Foreground(t.Warning)

	// Logo
	LogoStyle = lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	LogoSubStyle = lipgloss.NewStyle().
		Foreground(t.Secondary)

	// Options
	OptionOnStyle = lipgloss.NewStyle().Foreground(t.Success)
	OptionOffStyle = lipgloss.NewStyle().Foreground(t.Muted)
	OptionKeyStyle = lipgloss.NewStyle().Foreground(t.Secondary)

	// Sort
	SortLabelStyle = lipgloss.NewStyle().Foreground(t.Secondary)
	SortValueStyle = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)

	// Video info header
	VideoTitleStyle = lipgloss.NewStyle().Foreground(t.Text).Bold(true)
	VideoDetailStyle = lipgloss.NewStyle().Foreground(t.Secondary)

	// Slash commands
	SlashSelectedStyle = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	SlashNormalStyle = lipgloss.NewStyle().Foreground(t.Subtext)
	SlashDescStyle = lipgloss.NewStyle().Foreground(t.Muted)

	// Speed / path
	SpeedStyle = lipgloss.NewStyle().Foreground(t.Success).Italic(true)
	PathStyle = lipgloss.NewStyle().Foreground(t.Muted)
}

// Truncate shortens s to at most width terminal cells, ending with "…".
func Truncate(s string, width int) string {
	return ansi.Truncate(s, max(width, 1), "…")
}

// TruncateLeft keeps the last width cells of s, starting with "…".
func TruncateLeft(s string, width int) string {
	width = max(width, 1)
	if over := ansi.StringWidth(s) - width; over > 0 {
		return ansi.TruncateLeft(s, over+1, "…")
	}
	return s
}
