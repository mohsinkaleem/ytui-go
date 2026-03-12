# ytui

A beautiful terminal UI for [yt-dlp](https://github.com/yt-dlp/yt-dlp) — search, browse, download, and stream YouTube videos without leaving your terminal.

Built with the [Charm](https://charm.sh) stack (bubbletea + lipgloss + bubbles).

## Features

- **Search** — full-text YouTube search with sort options (relevance, upload date, view count, rating)
- **Format picker** — browse video/audio formats with tabbed UI, or enter custom format strings
- **Download queue** — concurrent downloads with progress bars, pause/resume/cancel
- **Playlist support** — browse and batch-download entire playlists
- **mpv integration** — stream videos directly in mpv
- **Slash commands** — `/download`, `/play`, `/playlist`, `/theme`, and more
- **Theming** — ships with Catppuccin Mocha, Catppuccin Latte, and Monochrome themes
- **Persistence** — search history, download records, and settings stored via bbolt
- **Keyboard-driven** — full keyboard navigation with contextual status bar hints
- **Multi-select** — select multiple videos for batch operations

## Prerequisites

- [yt-dlp](https://github.com/yt-dlp/yt-dlp#installation) in your `$PATH`
- [mpv](https://mpv.io/) (optional, for streaming playback)

## Installation

### Pre-built binaries

Download the latest release for your platform from the [Releases](https://github.com/mohsinkaleem/ytui-go/releases) page.

### From source

Requires [Go](https://go.dev/dl/) 1.25+.

```sh
go install github.com/mohsinkaleem/ytui-go/cmd/ytui@latest
```

Or clone and build manually:

```sh
git clone https://github.com/mohsinkaleem/ytui-go.git
cd ytui-go
make install
```

### Build locally

```sh
make build
./bin/ytui
```

## Usage

```sh
ytui                         # launch the TUI
ytui --theme latte           # launch with a specific theme
ytui --download-dir ~/Videos # set download directory
ytui --version               # print version
```

### Search & Browse

Type a search query and press **Enter** to search YouTube, or paste a URL directly.

- **Video URL** → jumps straight to the format picker
- **Playlist URL** → opens the playlist browser
- **Search query** → shows search results

### Keyboard Shortcuts

#### Global

| Key | Action |
|---|---|
| `ctrl+c` | Quit |

#### Search Screen

| Key | Action |
|---|---|
| `Enter` | Search / open URL |
| `Tab` / `Shift+Tab` | Cycle sort option |
| `ctrl+s` | Toggle embed subtitles |
| `ctrl+m` | Toggle embed metadata |
| `ctrl+j` | Toggle embed chapters |
| `Up` / `Down` | Navigate search history |
| `/command` | Slash commands |

#### Video List

| Key | Action |
|---|---|
| `Enter` | Open format picker |
| `d` | Quick download (best format) |
| `p` | Play in mpv |
| `Space` | Toggle multi-select |
| `a` | Select / deselect all |
| `ctrl+y` | Copy URL to clipboard |
| `/` | Filter list |
| `b` / `Esc` | Back |

#### Format List

| Key | Action |
|---|---|
| `Enter` | Download selected format |
| `p` | Play selected format in mpv |
| `Tab` / `Shift+Tab` | Switch tab (Video / Audio / Custom) |
| `b` / `Esc` | Back |

#### Download Screen

| Key | Action |
|---|---|
| `p` | Pause / Resume |
| `c` | Cancel current download |
| `r` | Retry failed download |
| `s` | Skip current download |
| `b` / `Esc` | Back (when all done) |

### Slash Commands

Type `/` followed by a command name in the search input:

| Command | Description |
|---|---|
| `/download <url>` | Quick-download with best format |
| `/playlist <url>` | Open playlist browser |
| `/play <url>` | Stream in mpv |
| `/resume` | Show unfinished downloads |
| `/theme <name>` | Switch theme (`mocha`, `latte`, `monochrome`) |
| `/clear` | Clear search history |
| `/help` | Toggle help |
| `/exit` | Exit app |

## Themes

ytui ships with three built-in themes:

- **mocha** — Catppuccin Mocha (default, dark)
- **latte** — Catppuccin Latte (light)
- **monochrome** — Minimal black & white

Switch at runtime with `/theme <name>` or at launch with `--theme <name>`.

## Project Structure

```
cmd/ytui/             CLI entry point (cobra)
internal/
  app/                Root model, Update, View, key bindings
  models/             Sub-models: search, video list, format list, download, player
  types/              State constants, message types, shared data types
  styles/             Lipgloss styles and theme system
  slash/              Slash command registry with fuzzy matching
  ytdlp/              yt-dlp binary wrapper (search, metadata, formats, download, progress)
  download/           Download manager with worker pool, task lifecycle, state machine
  player/             mpv integration
  store/              bbolt persistence (search history, downloads, settings)
  utils/              Manager wrappers and XDG path helpers
```

## Architecture

ytui follows the [Elm architecture](https://guide.elm-lang.org/architecture/) via bubbletea:

- **Single root model** with flat composition — no nested `tea.Model` routing
- **String-typed state machine** — explicit state transitions in `update.go`
- **Typed messages as intent** — every action is a message, not a method call
- **Worker pool** for concurrent downloads with proper state machine transitions
- **Context-aware status bar** — key hints adapt to current state

## Development

```sh
make build       # build binary to bin/
make run         # build and run
make test        # run tests with race detector
make fmt         # format code
make vet         # static analysis
make lint        # lint (requires golangci-lint)
make clean       # remove build artifacts
```

## Tech Stack

| Dependency | Purpose |
|---|---|
| [bubbletea](https://github.com/charmbracelet/bubbletea) | Elm-architecture TUI framework |
| [lipgloss](https://github.com/charmbracelet/lipgloss) | Declarative terminal styling |
| [bubbles](https://github.com/charmbracelet/bubbles) | Stock components (list, textinput, progress, spinner) |
| [bubblezone](https://github.com/lrstanley/bubblezone) | Mouse click zones |
| [cobra](https://github.com/spf13/cobra) | CLI flags and subcommands |
| [fuzzy](https://github.com/sahilm/fuzzy) | Slash command autocomplete |
| [clipboard](https://github.com/atotto/clipboard) | Clipboard support |
| [bbolt](https://go.etcd.io/bbolt) | Embedded K/V persistence |
| [xdg](https://github.com/adrg/xdg) | XDG-compliant paths |

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[MIT](LICENSE)
