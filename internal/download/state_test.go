package download

import (
	"testing"
)

func TestValidTransition(t *testing.T) {
	tests := []struct {
		name string
		from TaskState
		to   TaskState
		want bool
	}{
		// From Queued
		{"queued -> downloading", StateQueued, StateDownloading, true},
		{"queued -> cancelled", StateQueued, StateCancelled, true},
		{"queued -> completed", StateQueued, StateCompleted, false},
		{"queued -> failed", StateQueued, StateFailed, false},
		{"queued -> paused", StateQueued, StatePaused, false},

		// From Downloading
		{"downloading -> completed", StateDownloading, StateCompleted, true},
		{"downloading -> failed", StateDownloading, StateFailed, true},
		{"downloading -> cancelled", StateDownloading, StateCancelled, true},
		{"downloading -> paused", StateDownloading, StatePaused, true},
		{"downloading -> queued", StateDownloading, StateQueued, false},

		// From Paused
		{"paused -> downloading", StatePaused, StateDownloading, true},
		{"paused -> cancelled", StatePaused, StateCancelled, true},
		{"paused -> completed", StatePaused, StateCompleted, false},

		// From Failed
		{"failed -> queued (retry)", StateFailed, StateQueued, true},
		{"failed -> downloading", StateFailed, StateDownloading, false},
		{"failed -> completed", StateFailed, StateCompleted, false},

		// From Completed (terminal)
		{"completed -> any", StateCompleted, StateDownloading, false},
		{"completed -> queued", StateCompleted, StateQueued, false},

		// From Cancelled (terminal)
		{"cancelled -> any", StateCancelled, StateDownloading, false},
		{"cancelled -> queued", StateCancelled, StateQueued, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidTransition(tt.from, tt.to); got != tt.want {
				t.Errorf("ValidTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestTaskStateIsTerminal(t *testing.T) {
	tests := []struct {
		state TaskState
		want  bool
	}{
		{StateQueued, false},
		{StateDownloading, false},
		{StatePaused, false},
		{StateFailed, true},
		{StateCompleted, true},
		{StateCancelled, true},
	}
	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if got := tt.state.IsTerminal(); got != tt.want {
				t.Errorf("IsTerminal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskStateIsActive(t *testing.T) {
	tests := []struct {
		state TaskState
		want  bool
	}{
		{StateQueued, false},
		{StateDownloading, true},
		{StatePaused, false},
		{StateFailed, false},
		{StateCompleted, false},
		{StateCancelled, false},
	}
	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if got := tt.state.IsActive(); got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskStateIcon(t *testing.T) {
	tests := []struct {
		state TaskState
		want  string
	}{
		{StateQueued, "○"},
		{StateDownloading, "↓"},
		{StateCompleted, "✓"},
		{StateFailed, "✗"},
		{StateCancelled, "→"},
		{StatePaused, "⏸"},
		{TaskState("unknown"), "?"},
	}
	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if got := tt.state.Icon(); got != tt.want {
				t.Errorf("Icon() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTaskStateLabel(t *testing.T) {
	tests := []struct {
		state TaskState
		want  string
	}{
		{StateQueued, "Queued"},
		{StateDownloading, "⇣ Downloading"},
		{StateCompleted, "✓ Complete"},
		{StateFailed, "✕ Failed"},
		{StateCancelled, "✕ Cancelled"},
		{StatePaused, "⏸ Paused"},
		{TaskState("custom"), "custom"},
	}
	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if got := tt.state.Label(); got != tt.want {
				t.Errorf("Label() = %q, want %q", got, tt.want)
			}
		})
	}
}
