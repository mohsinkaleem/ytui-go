# ytui

A beautiful terminal UI for [yt-dlp](https://github.com/yt-dlp/yt-dlp) — search, browse, download, and stream YouTube videos without leaving your terminal.

Built with the [Charm](https://charm.sh) stack (bubbletea + lipgloss + bubbles).

## Features

- **Search** — full-text YouTube search with sort options (relevance, upload date, view count, rating)
- **Format picker** — browse video/audio formats in an aligned table, pick a preset, or enter any yt-dlp format selector
- **Download queue** — concurrent downloads with progress bars, pause/resume/cancel/retry
- **Resumable** — unfinished downloads are saved on quit and can be continued with `/resume`
- **Playlist support** — browse and batch-download entire playlists
- **mpv integration** — stream videos directly in mpv
- **Slash commands** — `/download`, `/play`, `/playlist`, `/theme`, `/cookies`, and more
- **Theming** — Catppuccin Mocha & Latte, Nord, Gruvbox and Monochrome
- **Persistence** — history, download records, and settings stored via bbolt
- **Keyboard-driven** — full keyboard navigation with contextual status bar hints
- **Multi-select** — select multiple videos for batch downloads

## Prerequisites

- [yt-dlp](https://github.com/yt-dlp/yt-dlp#installation) in your `$PATH`
- [ffmpeg](https://ffmpeg.org/) (recommended, needed to merge separate video and audio streams)
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
ytui --theme latte           # use a theme for this session
ytui --download-dir ~/Videos # download to a folder for this session
ytui --version               # print version
```

Files are saved as `Title [video-id].ext`, so videos with the same title never overwrite each other.

### Search & Browse

Type a search query and press **Enter** to search YouTube, or paste a URL directly.

- **Video URL** → jumps straight to the format picker
- **Playlist URL** → opens the playlist browser
- **Search query** → shows search results

### Keyboard Shortcuts

#### Global

| Key | Action |
|---|---|
| `ctrl+c` | Quit (running downloads are saved for `/resume`) |
| `g` | Open the download queue (from lists, when downloads exist) |

#### Search Screen

| Key | Action |
|---|---|
| `Enter` | Search / open URL |
| `Tab` / `Shift+Tab` | Cycle sort option (completes a command while the command list is open) |
| `ctrl+s` | Toggle embed subtitles |
| `ctrl+t` | Toggle embed metadata |
| `ctrl+j` | Toggle embed chapters |
| `Up` / `Down` | Navigate history |
| `Esc` | Clear the input |
| `/` | Slash commands |

#### Video List

| Key | Action |
|---|---|
| `Enter` | Open format picker |
| `d` | Quick download (best format) — all selected videos, or the highlighted one |
| `p` | Play in mpv |
| `Space` | Toggle multi-select |
| `a` | Select / deselect all |
| `ctrl+y` | Copy URL to clipboard |
| `/` | Filter list |
| `b` | Back |
| `Esc` | Clear filter, or back |

#### Format List

| Key | Action |
|---|---|
| `Enter` | Download selected format (video-only formats get the best audio merged in) |
| `p` | Play selected format in mpv |
| `Tab` / `Shift+Tab` | Switch tab (Video / Audio / Custom) |
| `Up` / `Down` | Pick a preset (Custom tab) |
| `b` | Back |
| `Esc` | Home |

#### Download Screen

| Key | Action |
|---|---|
| `Up` / `Down` | Select a task in the queue |
| `p` | Pause / Resume |
| `c` | Cancel |
| `r` / `R` | Retry failed download / all failed downloads |
| `s` | Skip failed download |
| `b` | Back |
| `Esc` | Home |

#### Resume List

| Key | Action |
|---|---|
| `Enter` | Resume selected download |
| `d` | Resume all |
| `x` | Delete record |
| `b` / `Esc` | Back |

### Slash Commands

Type `/` in the search input to list all commands:

| Command | Description |
|---|---|
| `/download <url>` | Quick-download with best format |
| `/play <url>` | Stream in mpv |
| `/playlist <url>` | Open playlist browser |
| `/downloads` | Show the download queue |
| `/resume` | Show unfinished downloads |
| `/theme [name]` | Pick or switch theme |
| `/downloaddir [path]` | Show or set the download folder |
| `/cookies [browser\|path\|off]` | Use browser cookies (e.g. `chrome`, `firefox:Profile 1`) or a cookies.txt |
| `/clear` | Clear history |
| `/help` | List all commands |
| `/exit` | Exit app |

## Themes

ytui ships with five built-in themes:

- **mocha** — Catppuccin Mocha (default, dark)
- **latte** — Catppuccin Latte (light)
- **nord** — Nord (dark)
- **gruvbox** — Gruvbox (dark)
- **monochrome** — Minimal high contrast

Switch at runtime with `/theme` (saved) or at launch with `--theme <name>` (this session only).

## Project Structure

```
cmd/ytui/             CLI entry point (cobra)
internal/
  app/                Root model, Update, View, status bar key hints
  models/             Sub-models: search, video list, format list, download, resume, player
  types/              State constants, message types, shared data types
  styles/             Lipgloss styles, theme system, text truncation helpers
  slash/              Slash command registry with fuzzy matching
  ytdlp/              yt-dlp binary wrapper (search, metadata, formats, download, progress)
  download/           Download manager with worker pool, task lifecycle, state machine
  store/              bbolt persistence (history, downloads, settings)
  utils/              Cancellable metadata fetcher, mpv player, path helpers
```

## Architecture

ytui follows the [Elm architecture](https://guide.elm-lang.org/architecture/) via bubbletea:

- **Single root model** with flat composition — no nested `tea.Model` routing
- **String-typed state machine** — explicit state transitions in `update.go`
- **Typed messages as intent** — every action is a message, not a method call
- **Worker pool** for concurrent downloads; the UI polls task state, so downloads never block on the UI (e.g. while mpv has the terminal)
- **Context-aware status bar** — key hints adapt to the current screen and task state

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
| [cobra](https://github.com/spf13/cobra) | CLI flags and subcommands |
| [fuzzy](https://github.com/sahilm/fuzzy) | Slash command autocomplete |
| [clipboard](https://github.com/atotto/clipboard) | Clipboard support |
| [bbolt](https://go.etcd.io/bbolt) | Embedded K/V persistence |
| [xdg](https://github.com/adrg/xdg) | XDG-compliant paths |

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[MIT](LICENSE)
