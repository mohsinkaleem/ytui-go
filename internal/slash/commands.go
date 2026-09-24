package slash

import (
	"strings"

	"github.com/sahilm/fuzzy"
)

// Command represents a slash command
type Command struct {
	Name        string
	Description string
	Args        string
}

// Registry holds all registered slash commands
type Registry struct {
	commands []Command
}

// NewRegistry creates a command registry with built-in commands
func NewRegistry() *Registry {
	r := &Registry{}
	r.Register(Command{Name: "download", Description: "Quick-download with best format", Args: "<url>"})
	r.Register(Command{Name: "play", Description: "Stream in mpv", Args: "<url>"})
	r.Register(Command{Name: "playlist", Description: "Open playlist browser", Args: "<url>"})
	r.Register(Command{Name: "downloads", Description: "Show the download queue"})
	r.Register(Command{Name: "resume", Description: "Show & re-queue unfinished downloads"})
	r.Register(Command{Name: "theme", Description: "Switch color theme"})
	r.Register(Command{Name: "downloaddir", Description: "Set download folder", Args: "<path>"})
	r.Register(Command{Name: "cookies", Description: "Use browser cookies or a cookies.txt", Args: "<browser|path|off>"})
	r.Register(Command{Name: "clear", Description: "Clear history"})
	r.Register(Command{Name: "help", Description: "List all commands"})
	r.Register(Command{Name: "exit", Description: "Exit app"})
	return r
}

// Register adds a command to the registry
func (r *Registry) Register(cmd Command) {
	r.commands = append(r.commands, cmd)
}

// Match returns fuzzy-matched commands for the given input
func (r *Registry) Match(input string) []Command {
	input = strings.TrimPrefix(input, "/")
	if input == "" {
		return r.commands
	}

	matches := fuzzy.FindFrom(input, r)
	result := make([]Command, 0, len(matches))
	for _, m := range matches {
		result = append(result, r.commands[m.Index])
	}
	return result
}

// Get returns a command by exact name
func (r *Registry) Get(name string) (Command, bool) {
	name = strings.TrimPrefix(name, "/")
	for _, cmd := range r.commands {
		if cmd.Name == name {
			return cmd, true
		}
	}
	return Command{}, false
}

// All returns all registered commands
func (r *Registry) All() []Command {
	return r.commands
}

// ParseInput splits command input into name and args
func ParseInput(input string) (name, args string) {
	input = strings.TrimPrefix(input, "/")
	parts := strings.SplitN(input, " ", 2)
	name = parts[0]
	if len(parts) > 1 {
		args = strings.TrimSpace(parts[1])
	}
	return
}

// String implements fuzzy.Source
func (r *Registry) String(i int) string {
	return r.commands[i].Name
}

// Len implements fuzzy.Source
func (r *Registry) Len() int {
	return len(r.commands)
}
