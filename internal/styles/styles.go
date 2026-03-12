package styles

import "github.com/charmbracelet/lipgloss"

// Package-level style variables — rebuilt when theme changes
var (
	// Layout
	AppStyle     lipgloss.Style
	ContentStyle lipgloss.Style

	// Section headers
	SectionHeaderStyle lipgloss.Style
	SubtitleStyle      lipgloss.Style

	// Status bar
	StatusBarStyle  lipgloss.Style
	StatusKeyStyle  lipgloss.Style
	StatusDescStyle lipgloss.Style
	StatusSepStyle  lipgloss.Style

	// Input
	InputStyle       lipgloss.Style
	InputPromptStyle lipgloss.Style
	InputPlaceholder lipgloss.Style

	// Text variants
	MutedStyle   lipgloss.Style
	BoldStyle    lipgloss.Style
	ErrorStyle   lipgloss.Style
	SuccessStyle lipgloss.Style
	WarningStyle lipgloss.Style
	InfoStyle    lipgloss.Style
	AccentStyle  lipgloss.Style
	PinkStyle    lipgloss.Style

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

	// Tab bar
	TabActiveStyle   lipgloss.Style
	TabInactiveStyle lipgloss.Style
	TabGapStyle      lipgloss.Style

	// Progress
	ProgressStyle     lipgloss.Style
	ProgressFillColor lipgloss.Color

	// Toast
	ToastStyle lipgloss.Style

	// Download queue
	QueueItemPendingStyle  lipgloss.Style
	QueueItemActiveStyle   lipgloss.Style
	QueueItemCompleteStyle lipgloss.Style
	QueueItemErrorStyle    lipgloss.Style
	QueueItemPausedStyle   lipgloss.Style

	// Logo
	LogoStyle    lipgloss.Style
	LogoSubStyle lipgloss.Style

	// Borders
	BorderStyle lipgloss.Style

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
	ContentStyle = lipgloss.NewStyle()

	// Section headers
	SectionHeaderStyle = lipgloss.NewStyle().
		Foreground(t.Text).
		Bold(true).
		MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
		Foreground(t.Secondary)

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
		Foreground(t.Secondary).
		Background(t.Surface)

	StatusKeyStyle = lipgloss.NewStyle().
		Foreground(t.Pink).
		Bold(true)

	StatusDescStyle = lipgloss.NewStyle().
		Foreground(t.Secondary)

	StatusSepStyle = lipgloss.NewStyle().
		Foreground(t.Muted)

	// Input
	InputStyle = lipgloss.NewStyle().
		Foreground(t.Text)

	InputPromptStyle = lipgloss.NewStyle().
		Foreground(t.Pink).
		Bold(true)

	InputPlaceholder = lipgloss.NewStyle().
		Foreground(t.Muted)

	// Text variants
	MutedStyle = lipgloss.NewStyle().Foreground(t.Muted)
	BoldStyle = lipgloss.NewStyle().Bold(true).Foreground(t.Text)
	ErrorStyle = lipgloss.NewStyle().Foreground(t.Error)
	SuccessStyle = lipgloss.NewStyle().Foreground(t.Success)
	WarningStyle = lipgloss.NewStyle().Foreground(t.Warning)
	InfoStyle = lipgloss.NewStyle().Foreground(t.Info)
	AccentStyle = lipgloss.NewStyle().Foreground(t.Accent)
	PinkStyle = lipgloss.NewStyle().Foreground(t.Pink)

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
		Foreground(t.Success)

	// Tab bar
	TabActiveStyle = lipgloss.NewStyle().
		Foreground(t.Base).
		Background(t.Accent).
		Padding(0, 2).
		Bold(true)

	TabInactiveStyle = lipgloss.NewStyle().
		Foreground(t.Secondary).
		Padding(0, 2)

	TabGapStyle = lipgloss.NewStyle().
		Foreground(t.Muted)

	// Progress
	ProgressStyle = lipgloss.NewStyle()
	ProgressFillColor = t.Info

	// Toast
	ToastStyle = lipgloss.NewStyle().
		Foreground(t.Info)

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

	// Borders
	BorderStyle = lipgloss.NewStyle().
		Foreground(t.Muted)

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
	SlashSelectedStyle = lipgloss.NewStyle().Foreground(t.Accent)
	SlashNormalStyle = lipgloss.NewStyle().Foreground(t.Subtext)
	SlashDescStyle = lipgloss.NewStyle().Foreground(t.Muted)

	// Speed / path
	SpeedStyle = lipgloss.NewStyle().Foreground(t.Success).Italic(true)
	PathStyle = lipgloss.NewStyle().Foreground(t.Muted)
}
