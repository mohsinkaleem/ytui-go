package slash

import (
	"testing"
)

func TestNewRegistryHasBuiltinCommands(t *testing.T) {
	r := NewRegistry()
	commands := r.All()
	if len(commands) == 0 {
		t.Fatal("registry should have builtin commands")
	}
	// Verify essential commands exist
	essential := []string{"download", "play", "exit", "help", "theme"}
	for _, name := range essential {
		_, ok := r.Get(name)
		if !ok {
			t.Errorf("missing essential command: %s", name)
		}
	}
}

func TestRegistryRegister(t *testing.T) {
	r := NewRegistry()
	initial := len(r.All())
	r.Register(Command{Name: "custom", Description: "Custom command"})
	if len(r.All()) != initial+1 {
		t.Error("Register should add a command")
	}
}

func TestRegistryGetExact(t *testing.T) {
	r := NewRegistry()
	cmd, ok := r.Get("download")
	if !ok {
		t.Fatal("expected download to be found")
	}
	if cmd.Name != "download" {
		t.Errorf("Name = %q, want download", cmd.Name)
	}
	cmd2, ok := r.Get("/theme")
	if !ok {
		t.Fatal("expected /theme to be found")
	}
	if cmd2.Name != "theme" {
		t.Errorf("Name = %q, want theme", cmd2.Name)
	}
	_, ok = r.Get("nonexistent")
	if ok {
		t.Error("expected nonexistent to not be found")
	}
}

func TestRegistryMatchEmpty(t *testing.T) {
	r := NewRegistry()
	matches := r.Match("")
	if len(matches) != len(r.All()) {
		t.Errorf("empty match should return all %d commands, got %d", len(r.All()), len(matches))
	}
}

func TestRegistryMatchFuzzy(t *testing.T) {
	r := NewRegistry()
	matches := r.Match("pl")
	names := make(map[string]bool)
	for _, m := range matches {
		names[m.Name] = true
	}
	if !names["play"] {
		t.Error("pl should match play")
	}
	if !names["playlist"] {
		t.Error("pl should match playlist")
	}
}

func TestRegistryMatchPrefersExactName(t *testing.T) {
	r := NewRegistry()
	for _, name := range []string{"play", "download"} {
		matches := r.Match(name)
		if len(matches) == 0 || matches[0].Name != name {
			t.Errorf("Match(%q) should rank the exact command first, got %v", name, matches)
		}
	}
}

func TestParseInput(t *testing.T) {
	tests := []struct {
		input    string
		wantName string
		wantArgs string
	}{
		{"/download https://example.com", "download", "https://example.com"},
		{"/theme mocha", "theme", "mocha"},
		{"/exit", "exit", ""},
		{"/play", "play", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name, args := ParseInput(tt.input)
			if name != tt.wantName {
				t.Errorf("name = %q, want %q", name, tt.wantName)
			}
			if args != tt.wantArgs {
				t.Errorf("args = %q, want %q", args, tt.wantArgs)
			}
		})
	}
}

func TestRegistryLen(t *testing.T) {
	r := NewRegistry()
	if r.Len() == 0 {
		t.Error("Len() should be > 0")
	}
}

func TestRegistryString(t *testing.T) {
	r := NewRegistry()
	if r.Len() > 0 && r.String(0) == "" {
		t.Error("String(0) should be non-empty")
	}
}
