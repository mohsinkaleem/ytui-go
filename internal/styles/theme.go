package styles

import (
	"sort"

	"github.com/charmbracelet/lipgloss"
)

// Theme defines named color slots for theming
type Theme struct {
	Name      string
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color
	Success   lipgloss.Color
	Error     lipgloss.Color
	Warning   lipgloss.Color
	Info      lipgloss.Color
	Muted     lipgloss.Color
	Base      lipgloss.Color
	Surface   lipgloss.Color
	Text      lipgloss.Color
	Subtext   lipgloss.Color
	Overlay   lipgloss.Color
	Pink      lipgloss.Color
}

// Built-in themes

var CatppuccinMocha = Theme{
	Name:      "mocha",
	Primary:   lipgloss.Color("#cba6f7"),
	Secondary: lipgloss.Color("#a6adc8"),
	Accent:    lipgloss.Color("#cba6f7"),
	Success:   lipgloss.Color("#a6e3a1"),
	Error:     lipgloss.Color("#f38ba8"),
	Warning:   lipgloss.Color("#f9e2af"),
	Info:      lipgloss.Color("#89dceb"),
	Muted:     lipgloss.Color("#6c7086"),
	Base:      lipgloss.Color("#1e1e2e"),
	Surface:   lipgloss.Color("#313244"),
	Text:      lipgloss.Color("#cdd6f4"),
	Subtext:   lipgloss.Color("#a6adc8"),
	Overlay:   lipgloss.Color("#45475a"),
	Pink:      lipgloss.Color("#f5c2e7"),
}

var CatppuccinLatte = Theme{
	Name:      "latte",
	Primary:   lipgloss.Color("#8839ef"),
	Secondary: lipgloss.Color("#6c6f85"),
	Accent:    lipgloss.Color("#8839ef"),
	Success:   lipgloss.Color("#40a02b"),
	Error:     lipgloss.Color("#d20f39"),
	Warning:   lipgloss.Color("#df8e1d"),
	Info:      lipgloss.Color("#04a5e5"),
	Muted:     lipgloss.Color("#9ca0b0"),
	Base:      lipgloss.Color("#eff1f5"),
	Surface:   lipgloss.Color("#ccd0da"),
	Text:      lipgloss.Color("#4c4f69"),
	Subtext:   lipgloss.Color("#6c6f85"),
	Overlay:   lipgloss.Color("#bcc0cc"),
	Pink:      lipgloss.Color("#ea76cb"),
}

var Monochrome = Theme{
	Name:      "monochrome",
	Primary:   lipgloss.Color("#ffffff"),
	Secondary: lipgloss.Color("#b0b0b0"),
	Accent:    lipgloss.Color("#ffffff"),
	Success:   lipgloss.Color("#00ff00"),
	Error:     lipgloss.Color("#ff0000"),
	Warning:   lipgloss.Color("#ffff00"),
	Info:      lipgloss.Color("#00ffff"),
	Muted:     lipgloss.Color("#808080"),
	Base:      lipgloss.Color("#000000"),
	Surface:   lipgloss.Color("#1a1a1a"),
	Text:      lipgloss.Color("#ffffff"),
	Subtext:   lipgloss.Color("#b0b0b0"),
	Overlay:   lipgloss.Color("#333333"),
	Pink:      lipgloss.Color("#ff69b4"),
}

var Nord = Theme{
	Name:      "nord",
	Primary:   lipgloss.Color("#88c0d0"),
	Secondary: lipgloss.Color("#81a1c1"),
	Accent:    lipgloss.Color("#88c0d0"),
	Success:   lipgloss.Color("#a3be8c"),
	Error:     lipgloss.Color("#bf616a"),
	Warning:   lipgloss.Color("#ebcb8b"),
	Info:      lipgloss.Color("#5e81ac"),
	Muted:     lipgloss.Color("#4c566a"),
	Base:      lipgloss.Color("#2e3440"),
	Surface:   lipgloss.Color("#3b4252"),
	Text:      lipgloss.Color("#eceff4"),
	Subtext:   lipgloss.Color("#d8dee9"),
	Overlay:   lipgloss.Color("#434c5e"),
	Pink:      lipgloss.Color("#b48ead"),
}

var Gruvbox = Theme{
	Name:      "gruvbox",
	Primary:   lipgloss.Color("#d79921"),
	Secondary: lipgloss.Color("#a89984"),
	Accent:    lipgloss.Color("#d79921"),
	Success:   lipgloss.Color("#b8bb26"),
	Error:     lipgloss.Color("#fb4934"),
	Warning:   lipgloss.Color("#fabd2f"),
	Info:      lipgloss.Color("#83a598"),
	Muted:     lipgloss.Color("#665c54"),
	Base:      lipgloss.Color("#282828"),
	Surface:   lipgloss.Color("#3c3836"),
	Text:      lipgloss.Color("#ebdbb2"),
	Subtext:   lipgloss.Color("#bdae93"),
	Overlay:   lipgloss.Color("#504945"),
	Pink:      lipgloss.Color("#d3869b"),
}

var themes = map[string]Theme{
	"mocha":      CatppuccinMocha,
	"latte":      CatppuccinLatte,
	"monochrome": Monochrome,
	"nord":       Nord,
	"gruvbox":    Gruvbox,
}

// CurrentTheme holds the active theme
var CurrentTheme = CatppuccinMocha

// LoadTheme switches to the named theme and rebuilds all style vars
func LoadTheme(name string) bool {
	t, ok := themes[name]
	if !ok {
		return false
	}
	CurrentTheme = t
	rebuildStyles()
	return true
}

// ThemeNames returns all available theme names in sorted order
func ThemeNames() []string {
	names := make([]string, 0, len(themes))
	for k := range themes {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
