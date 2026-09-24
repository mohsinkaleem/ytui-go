package download

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/mohsinkaleem/ytui-go/internal/types"
	"github.com/mohsinkaleem/ytui-go/internal/ytdlp"
)

// fakeYtdlp puts a yt-dlp stand-in on PATH that prints progress like the real
// one; FAKE_FAIL makes it fail, FAKE_DELAY slows it down.
func fakeYtdlp(t *testing.T, delay string, fail bool) *ytdlp.Client {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake yt-dlp is a shell script")
	}
	dir := t.TempDir()
	script := `#!/bin/sh
trap 'echo "ERROR: Interrupted by user"; exit 1' INT
echo "[youtube] abc: Downloading webpage"
echo "[download] Destination: /out/Fake [abc].f137.mp4"
i=0
while [ $i -le 10 ]; do
  echo "YTPROG|$((i*100))|1000|NA|2048|$((10-i))|downloading|Fake Title"
  i=$((i+1))
  sleep "$FAKE_DELAY"
done
if [ -n "$FAKE_FAIL" ]; then echo "ERROR: [youtube] abc: Video unavailable"; exit 1; fi
echo '[Merger] Merging formats into "/out/Fake [abc].mp4"'
`
	if err := os.WriteFile(filepath.Join(dir, "yt-dlp"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_DELAY", delay)
	if fail {
		t.Setenv("FAKE_FAIL", "1")
	}
	client, err := ytdlp.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func waitForState(t *testing.T, task *Task, want TaskState) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for task.GetState() != want {
		if time.Now().After(deadline) {
			t.Fatalf("state = %q, want %q", task.GetState(), want)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestDownloadCompletes(t *testing.T) {
	mgr := NewManager(1, fakeYtdlp(t, "0", false), nil)
	defer mgr.Shutdown()
	task := mgr.EnqueueBatch([]types.VideoItem{{URL: "https://youtu.be/abc", Title: "https://youtu.be/abc"}}, "best", ytdlp.DownloadOpts{})[0]

	waitForState(t, task, StateCompleted)
	if got := task.GetTitle(); got != "Fake Title" {
		t.Errorf("title = %q, want it updated from yt-dlp", got)
	}
	if got := task.GetOutputPath(); got != "/out/Fake [abc].mp4" {
		t.Errorf("output path = %q, want the merged file", got)
	}
	if p := task.GetProgress(); p.Percent != 1 {
		t.Errorf("percent = %v, want 1", p.Percent)
	}
}

func TestDownloadFailureKeepsYtdlpError(t *testing.T) {
	mgr := NewManager(1, fakeYtdlp(t, "0", true), nil)
	defer mgr.Shutdown()
	task := mgr.EnqueueBatch([]types.VideoItem{{URL: "u"}}, "best", ytdlp.DownloadOpts{})[0]

	waitForState(t, task, StateFailed)
	if err := task.GetError(); err == nil || err.Error() != "[youtube] abc: Video unavailable" {
		t.Errorf("error = %v, want the yt-dlp ERROR line", err)
	}
}

func TestDownloadPauseResumeAndShutdown(t *testing.T) {
	mgr := NewManager(1, fakeYtdlp(t, "0.05", false), nil)
	task := mgr.EnqueueBatch([]types.VideoItem{{URL: "u"}}, "best", ytdlp.DownloadOpts{})[0]

	waitForState(t, task, StateDownloading)
	mgr.PauseTask(task.ID)
	mgr.ResumeTask(task.ID) // immediately: must wait for the old run instead of racing it
	waitForState(t, task, StateDownloading)

	if n := mgr.Shutdown(); n != 1 {
		t.Errorf("Shutdown() = %d interrupted, want 1", n)
	}
	if got := task.GetState(); got != StatePaused {
		t.Errorf("state after shutdown = %q, want paused (resumable)", got)
	}
}

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
	if task.GetTitle() != "Test Video" {
		t.Errorf("Title = %q, want %q", task.GetTitle(), "Test Video")
	}
	if task.URL != "https://www.youtube.com/watch?v=abc123" {
		t.Errorf("URL = %q, want correct URL", task.URL)
	}
	if task.FormatID != "137" {
		t.Errorf("FormatID = %q, want %q", task.FormatID, "137")
	}
	if task.GetState() != StateQueued {
		t.Errorf("State = %q, want %q", task.GetState(), StateQueued)
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
	task.state = StateCompleted
	task.mu.Unlock()

	// Cancel should be a no-op
	task.Cancel()
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

func TestTaskFinish(t *testing.T) {
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	tests := []struct {
		name  string
		from  TaskState
		ctx   context.Context
		err   error
		want  TaskState
		hasEr bool
	}{
		{"success", StateDownloading, context.Background(), nil, StateCompleted, false},
		{"failure", StateDownloading, context.Background(), errors.New("boom"), StateFailed, true},
		{"shutdown leaves it resumable", StateDownloading, cancelled, errors.New("interrupted"), StatePaused, false},
		{"user pause is kept", StatePaused, cancelled, errors.New("interrupted"), StatePaused, false},
		{"user cancel is kept", StateCancelled, cancelled, errors.New("interrupted"), StateCancelled, false},
		{"quick resume is kept", StateQueued, cancelled, errors.New("interrupted"), StateQueued, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask(types.VideoItem{ID: "x"}, "best", ytdlp.DownloadOpts{}, nil)
			task.state = tt.from
			task.finish(tt.ctx, tt.err)
			if got := task.GetState(); got != tt.want {
				t.Errorf("state = %q, want %q", got, tt.want)
			}
			if (task.GetError() != nil) != tt.hasEr {
				t.Errorf("error = %v, want error: %v", task.GetError(), tt.hasEr)
			}
		})
	}
}

func TestTaskPrepareResume(t *testing.T) {
	for _, tt := range []struct {
		from TaskState
		want bool
	}{
		{StatePaused, true},
		{StateFailed, true},
		{StateQueued, false},
		{StateCompleted, false},
		{StateCancelled, false},
	} {
		task := NewTask(types.VideoItem{ID: "x"}, "best", ytdlp.DownloadOpts{}, nil)
		task.state = tt.from
		task.err = errors.New("old")
		if got := task.PrepareResume(); got != tt.want {
			t.Errorf("PrepareResume() from %q = %v, want %v", tt.from, got, tt.want)
		}
		if tt.want && (task.GetState() != StateQueued || task.GetError() != nil) {
			t.Errorf("from %q: state = %q, err = %v; want queued with no error", tt.from, task.GetState(), task.GetError())
		}
	}
}

func TestManagerShutdownIdle(t *testing.T) {
	mgr := NewManager(0, nil, nil)
	if n := mgr.Shutdown(); n != 0 {
		t.Errorf("Shutdown() = %d, want 0 interrupted", n)
	}
}

func TestManagerShutdownCountsQueued(t *testing.T) {
	mgr := NewManager(1, nil, nil)
	mgr.mu.Lock()
	mgr.closed = true // keep workers from starting the task
	mgr.mu.Unlock()

	task := NewTask(types.VideoItem{ID: "x"}, "best", ytdlp.DownloadOpts{}, nil)
	mgr.Enqueue(task)
	if n := mgr.Shutdown(); n != 1 {
		t.Errorf("Shutdown() = %d, want 1 interrupted", n)
	}
	if task.GetState() != StateQueued {
		t.Errorf("state = %q, want queued (resumable)", task.GetState())
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
