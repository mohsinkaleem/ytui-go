package download

// TaskState represents the state of a download task
type TaskState string

const (
	StateQueued      TaskState = "queued"
	StateDownloading TaskState = "downloading"
	StateCompleted   TaskState = "completed"
	StateFailed      TaskState = "failed"
	StateCancelled   TaskState = "cancelled"
	StatePaused      TaskState = "paused"
)

// ValidTransition checks if a state transition is allowed
func ValidTransition(from, to TaskState) bool {
	switch from {
	case StateQueued:
		return to == StateDownloading || to == StateCancelled
	case StateDownloading:
		return to == StateCompleted || to == StateFailed || to == StateCancelled || to == StatePaused
	case StatePaused:
		return to == StateDownloading || to == StateCancelled || to == StateQueued
	case StateFailed:
		return to == StateQueued // retry
	case StateCompleted, StateCancelled:
		return false
	}
	return false
}

// IsTerminal returns true if the state is a final state
func (s TaskState) IsTerminal() bool {
	return s == StateCompleted || s == StateCancelled || s == StateFailed
}

// IsActive returns true if the task is currently doing work
func (s TaskState) IsActive() bool {
	return s == StateDownloading
}

// Icon returns a unicode icon for the state
func (s TaskState) Icon() string {
	switch s {
	case StateQueued:
		return "○"
	case StateDownloading:
		return "↓"
	case StateCompleted:
		return "✓"
	case StateFailed:
		return "✗"
	case StateCancelled:
		return "→"
	case StatePaused:
		return "⏸"
	default:
		return "?"
	}
}

// Label returns a human-readable label for the state
func (s TaskState) Label() string {
	switch s {
	case StateQueued:
		return "Queued"
	case StateDownloading:
		return "⇣ Downloading"
	case StateCompleted:
		return "✓ Complete"
	case StateFailed:
		return "✕ Failed"
	case StateCancelled:
		return "✕ Cancelled"
	case StatePaused:
		return "⏸ Paused"
	default:
		return string(s)
	}
}
