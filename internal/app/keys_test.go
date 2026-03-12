package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
)

func TestKeyBindingsNotEmpty(t *testing.T) {
	bindings := []struct {
		name    string
		binding key.Binding
	}{
		{"KeyQuit", KeyQuit},
		{"KeyHelp", KeyHelp},
		{"KeyBack", KeyBack},
		{"KeyEnter", KeyEnter},
		{"KeySlash", KeySlash},
		{"KeyCopy", KeyCopy},
		{"KeySpace", KeySpace},
		{"KeySelectAll", KeySelectAll},
		{"KeyDownload", KeyDownload},
		{"KeyPlay", KeyPlay},
		{"KeyPause", KeyPause},
		{"KeyCancel", KeyCancel},
		{"KeyTab", KeyTab},
		{"KeyShiftTab", KeyShiftTab},
		{"KeyToggleSubs", KeyToggleSubs},
		{"KeyToggleMeta", KeyToggleMeta},
		{"KeyToggleChapters", KeyToggleChapters},
		{"KeyRetry", KeyRetry},
		{"KeySkip", KeySkip},
	}
	for _, tt := range bindings {
		t.Run(tt.name, func(t *testing.T) {
			keys := tt.binding.Keys()
			if len(keys) == 0 {
				t.Errorf("%s has no keys", tt.name)
			}
		})
	}
}

func TestGetStatusKeysText(t *testing.T) {
	states := []string{"SearchInput", "VideoList", "FormatList", "Download", "Loading", "VideoPlaying", "ResumeList"}
	for _, state := range states {
		t.Run(state, func(t *testing.T) {
			got := GetStatusKeysText(state)
			if state == "Loading" || state == "VideoPlaying" || state == "ResumeList" {
				if got == "" {
					t.Errorf("GetStatusKeysText(%q) should return non-empty", state)
				}
				return
			}
			// All interactive states should mention at least one key hint
			if got == "" {
				t.Errorf("GetStatusKeysText(%q) should return non-empty", state)
			}
		})
	}
	// Unknown state should return empty
	if got := GetStatusKeysText("nonexistent"); got != "" {
		t.Errorf("GetStatusKeysText(nonexistent) should be empty, got %q", got)
	}
}

func TestFormatKeysOutput(t *testing.T) {
	result := FormatKeys(KeyQuit, "exit", KeyHelp, "help")
	if result == "" {
		t.Error("FormatKeys should produce non-empty output")
	}
	if !strings.Contains(result, "exit") {
		t.Errorf("result should contain exit, got %q", result)
	}
	if !strings.Contains(result, "|") {
		t.Errorf("result should contain separator, got %q", result)
	}
}

func TestFormatKeysEmpty(t *testing.T) {
	result := FormatKeys()
	if result != "" {
		t.Errorf("FormatKeys() with no args should be empty, got %q", result)
	}
}
