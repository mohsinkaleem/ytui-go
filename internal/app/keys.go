package app

import (
	"github.com/charmbracelet/bubbles/key"
)

// Global key bindings
var (
	KeyQuit = key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "exit"),
	)
	KeyHelp = key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	)
	KeyBack = key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "back"),
	)
	KeyEnter = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	)
	KeySlash = key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "commands"),
	)
	KeyCopy = key.NewBinding(
		key.WithKeys("ctrl+y"),
		key.WithHelp("ctrl+y", "copy URL"),
	)
	KeySpace = key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "select"),
	)
	KeySelectAll = key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "select all"),
	)
	KeyDownload = key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "download"),
	)
	KeyPlay = key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "play"),
	)
	KeyPause = key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "pause"),
	)
	KeyResume = key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "resume"),
	)
	KeyCancel = key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "cancel"),
	)
	KeyTab = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
	)
	KeyShiftTab = key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev tab"),
	)
	KeyToggleSubs = key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "subtitles"),
	)
	KeyToggleMeta = key.NewBinding(
		key.WithKeys("ctrl+m"),
		key.WithHelp("ctrl+m", "metadata"),
	)
	KeyToggleChapters = key.NewBinding(
		key.WithKeys("ctrl+j"),
		key.WithHelp("ctrl+j", "chapters"),
	)
	KeyRetry = key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "retry"),
	)
	KeyRetryAll = key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "retry all"),
	)
	KeySkip = key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "skip"),
	)
	KeyUp = key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	)
	KeyDown = key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	)
	KeyEsc = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "home"),
	)
)

// GetStatusKeysText returns formatted key hints for the status bar
func GetStatusKeysText(state string) string {
	switch state {
	case "SearchInput":
		return FormatKeys(
			KeyEnter, "search",
			KeySlash, "commands",
			KeyTab, "sort",
		)
	case "VideoList":
		return FormatKeys(
			KeyEnter, "formats",
			KeyDownload, "download",
			KeyPlay, "play",
			KeySpace, "select",
			KeySelectAll, "select all",
			KeyEsc, "home",
		)
	case "FormatList":
		return FormatKeys(
			KeyEnter, "download",
			KeyPlay, "play",
			KeyTab, "next tab",
			KeyBack, "back",
			KeyEsc, "home",
		)
	case "Download":
		// Default – caller should prefer Model.currentKeysText() for live context
		return FormatKeys(
			KeyPause, "pause",
			KeyCancel, "cancel",
			KeyBack, "back",
			KeyEsc, "home",
		)
	case "Loading":
		return "esc/c: cancel"
	case "VideoPlaying":
		return "mpv running — close mpv to return"
	case "ResumeList":
		return FormatKeys(
			KeyEnter, "resume",
			KeyDownload, "resume all",
			KeyEsc, "home",
		)
	default:
		return ""
	}
}

// FormatKeys formats key bindings as "key: desc | key: desc"
func FormatKeys(bindings ...interface{}) string {
	result := ""
	for i := 0; i < len(bindings)-1; i += 2 {
		binding, ok := bindings[i].(key.Binding)
		if !ok {
			continue
		}
		desc, ok := bindings[i+1].(string)
		if !ok {
			continue
		}
		if result != "" {
			result += " | "
		}
		keys := binding.Keys()
		if len(keys) > 0 {
			result += keys[0] + ": " + desc
		}
	}
	return result
}
