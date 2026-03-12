package download

import (
	"testing"

	"github.com/mohsinkaleem/ytui-go/internal/types"
	"github.com/mohsinkaleem/ytui-go/internal/ytdlp"
)

func TestNewTaskFields(t *testing.T) {
	video := types.VideoItem{
		ID:    "abc123",
		Title: "Test Video",
		URL:   "https://www.youtube.com/watch?v=abc123",
	}
	opts := ytdlp.DownloadOpts{
		EmbedSubs:     true,
		EmbedMetadata: false,
		OutputDir:     "/tmp",
	}

	task := NewTask(video, "137", opts, nil)

	if task.ID == "" {
		t.Error("task ID should not be empty")
	}
	if task.VideoID != "abc123" {
		t.Errorf("VideoID = %q, want %q", task.VideoID, "abc123")
	}
	if task.Title != "Test Video" {
		t.Errorf("Title = %q, want %q", task.Title, "Test Video")
	}
	if task.URL != "https://www.youtube.com/watch?v=abc123" {
		t.Errorf("URL = %q, want correct URL", task.URL)
	}
	if task.FormatID != "137" {
		t.Errorf("FormatID = %q, want %q", task.FormatID, "137")
	}
	if task.State != StateQueued {
		t.Errorf("State = %q, want %q", task.State, StateQueued)
	}
	if !task.Opts.EmbedSubs {
		t.Error("EmbedSubs should be true")
	}
	if task.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestNewTaskUsesVideoURL(t *testing.T) {
	// Video with no URL but has ID should use VideoURL()
	video := types.VideoItem{
		ID:    "xyz789",
		Title: "No URL Video",
	}

	task := NewTask(video, "best", ytdlp.DownloadOpts{}, nil)

	expected := "https://www.youtube.com/watch?v=xyz789"
	if task.URL != expected {
		t.Errorf("URL = %q, want %q (should use VideoURL())", task.URL, expected)
	}
}

func TestTaskGetStateConcurrent(t *testing.T) {
	video := types.VideoItem{ID: "test", Title: "Test"}
	task := NewTask(video, "best", ytdlp.DownloadOpts{}, nil)

	// Should be safe to call from multiple goroutines
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			_ = task.GetState()
			done <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestTaskGetProgressDefault(t *testing.T) {
	video := types.VideoItem{ID: "test", Title: "Test"}
	task := NewTask(video, "best", ytdlp.DownloadOpts{}, nil)

	progress := task.GetProgress()
	if progress.Percent != 0 {
		t.Errorf("default Percent = %f, want 0", progress.Percent)
	}
}

func TestTaskGetErrorDefault(t *testing.T) {
	video := types.VideoItem{ID: "test", Title: "Test"}
	task := NewTask(video, "best", ytdlp.DownloadOpts{}, nil)

	if task.GetError() != nil {
		t.Error("default error should be nil")
	}
}

func TestTaskString(t *testing.T) {
	video := types.VideoItem{ID: "test", Title: "My Video"}
	task := NewTask(video, "best", ytdlp.DownloadOpts{}, nil)

	s := task.String()
	if s == "" {
		t.Error("String() should not be empty")
	}
}

func TestTaskCancelFromTerminalState(t *testing.T) {
	video := types.VideoItem{ID: "test", Title: "Test"}
	task := NewTask(video, "best", ytdlp.DownloadOpts{}, nil)

	// Manually set to completed
	task.mu.Lock()
	task.State = StateCompleted
	task.mu.Unlock()

	// Cancel should be a no-op
	task.Cancel(false)
	if task.GetState() != StateCompleted {
		t.Errorf("State = %q, want %q (should not change from completed)", task.GetState(), StateCompleted)
	}
}

func TestTaskPauseInvalidTransition(t *testing.T) {
	video := types.VideoItem{ID: "test", Title: "Test"}
	task := NewTask(video, "best", ytdlp.DownloadOpts{}, nil)

	// Queued -> Paused is invalid
	task.Pause()
	if task.GetState() != StateQueued {
		t.Errorf("State = %q, want %q (pause from queued should be no-op)", task.GetState(), StateQueued)
	}
}

func TestManagerNewDefaults(t *testing.T) {
	mgr := NewManager(0, nil, nil)
	if mgr.maxConcurrent != 3 {
		t.Errorf("maxConcurrent = %d, want 3 (default)", mgr.maxConcurrent)
	}
}

func TestManagerGetTaskNotFound(t *testing.T) {
	mgr := NewManager(1, nil, nil)
	task := mgr.GetTask("nonexistent")
	if task != nil {
		t.Error("should return nil for nonexistent task")
	}
}

func TestManagerGetTasksEmpty(t *testing.T) {
	mgr := NewManager(1, nil, nil)
	tasks := mgr.GetTasks()
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestManagerActiveCountZero(t *testing.T) {
	mgr := NewManager(1, nil, nil)
	if mgr.ActiveCount() != 0 {
		t.Errorf("ActiveCount() = %d, want 0", mgr.ActiveCount())
	}
}
