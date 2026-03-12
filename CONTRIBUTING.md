# Contributing to ytui

Thanks for your interest in contributing! Here's how to get started.

## Development Setup

1. **Prerequisites**: Go 1.25+, [yt-dlp](https://github.com/yt-dlp/yt-dlp), and optionally [mpv](https://mpv.io/).

2. **Clone and build**:
   ```sh
   git clone https://github.com/mohsinkaleem/ytui-go.git
   cd ytui-go
   make build
   ```

3. **Run locally**:
   ```sh
   make run
   ```

## Making Changes

1. Fork the repository and create a feature branch from `main`.
2. Make your changes with clear, focused commits.
3. Run the checks before submitting:
   ```sh
   make fmt       # format code
   make vet       # static analysis
   make test      # run tests
   ```
4. Open a pull request against `main`.

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`).
- Keep functions focused and concise.
- Add tests for new functionality.
- Use meaningful commit messages.

## Project Layout

- `cmd/ytui/` — CLI entry point
- `internal/app/` — Root bubbletea model, Update, View, key bindings
- `internal/models/` — Sub-models (search, video list, format list, download, player)
- `internal/types/` — Shared types and message definitions
- `internal/styles/` — Lipgloss styles and theme system
- `internal/ytdlp/` — yt-dlp binary wrapper
- `internal/download/` — Download manager and task lifecycle
- `internal/store/` — bbolt persistence layer

## Reporting Issues

Open a [GitHub issue](https://github.com/mohsinkaleem/ytui-go/issues) with:
- A clear description of the problem or feature request
- Steps to reproduce (for bugs)
- Your OS and terminal emulator

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
