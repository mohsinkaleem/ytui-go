package download

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mohsinkaleem/ytui-go/internal/types"
	"github.com/mohsinkaleem/ytui-go/internal/ytdlp"
)

// Task represents a single download task. Exported fields are immutable once
// the task is enqueued; everything else is guarded by mu.
type Task struct {
	ID        string
	VideoID   string
	URL       string
	FormatID  string
	Opts      ytdlp.DownloadOpts
	CreatedAt time.Time

	mu          sync.Mutex
	title       string
	outputPath  string
	state       TaskState
	progress    ytdlp.Progress
	err         error
	completedAt time.Time
	cancel      context.CancelFunc

	runMu  sync.Mutex // serializes runs so a quick pause/resume can't overlap
	client *ytdlp.Client
}

// NewTask creates a new download task
func NewTask(video types.VideoItem, formatID string, opts ytdlp.DownloadOpts, client *ytdlp.Client) *Task {
	return &Task{
		ID:        uuid.New().String(),
		VideoID:   video.ID,
		URL:       video.VideoURL(),
		FormatID:  formatID,
		Opts:      opts,
		CreatedAt: time.Now(),
		title:     video.Title,
		state:     StateQueued,
		client:    client,
	}
}

// Start runs yt-dlp and blocks until it exits. If ctx is cancelled (manager
// shutdown) the task is left paused so it can be resumed later.
func (t *Task) Start(ctx context.Context) {
	t.runMu.Lock()
	defer t.runMu.Unlock()

	t.mu.Lock()
	if !ValidTransition(t.state, StateDownloading) {
		t.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	t.state = StateDownloading
	t.err = nil
	t.cancel = cancel
	t.mu.Unlock()

	t.finish(ctx, t.run(ctx))
}

// run executes yt-dlp, recording progress, title and output path as reported.
func (t *Task) run(ctx context.Context) error {
	cmd, output, err := t.client.Download(ctx, t.URL, t.FormatID, t.Opts)
	if err != nil {
		return err
	}
	defer output.Close()

	var errLine, lastLine string
	scanner := bufio.NewScanner(output)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if p, err := ytdlp.ParseProgressLine(line); err == nil {
			t.mu.Lock()
			t.progress = *p
			if p.Title != "" {
				t.title = p.Title
			}
			t.mu.Unlock()
			continue
		}
		if path := ytdlp.ParseOutputPath(line); path != "" {
			t.mu.Lock()
			t.outputPath = path
			t.mu.Unlock()
			continue
		}
		switch {
		case strings.HasPrefix(line, "ERROR:"):
			if errLine == "" {
				errLine = strings.TrimSpace(strings.TrimPrefix(line, "ERROR:"))
			}
		case line != "" && !strings.HasPrefix(line, "["):
			lastLine = line // e.g. the last line of a Python traceback
		}
	}
	if scanner.Err() != nil {
		_, _ = io.Copy(io.Discard, output) // keep draining so yt-dlp never blocks on a full pipe
	}

	if err := cmd.Wait(); err != nil {
		switch {
		case errLine != "":
			return errors.New(errLine)
		case lastLine != "":
			return errors.New(lastLine)
		}
		return err
	}
	return nil
}

// finish records the outcome of a run unless the user paused or cancelled it meanwhile.
func (t *Task) finish(ctx context.Context, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cancel = nil
	if t.state != StateDownloading {
		return
	}
	switch {
	case err == nil:
		t.state = StateCompleted
		t.completedAt = time.Now()
	case ctx.Err() != nil:
		t.state = StatePaused
	default:
		t.state = StateFailed
		t.err = err
	}
}

// Pause stops a running download; yt-dlp keeps the partial file.
func (t *Task) Pause() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !ValidTransition(t.state, StatePaused) {
		return
	}
	t.state = StatePaused
	if t.cancel != nil {
		t.cancel()
	}
}

// PrepareResume moves a paused or failed task back to queued so the manager
// can re-schedule it (yt-dlp continues partial files by default).
// Returns true when the state was reset.
func (t *Task) PrepareResume() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.state != StatePaused && t.state != StateFailed {
		return false
	}
	t.state = StateQueued
	t.err = nil
	return true
}

// Cancel stops the task for good.
func (t *Task) Cancel() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state == StateCompleted || t.state == StateCancelled {
		return
	}
	t.state = StateCancelled
	if t.cancel != nil {
		t.cancel()
	}
}

// GetState returns the current task state (thread-safe)
func (t *Task) GetState() TaskState {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.state
}

// GetTitle returns the video title, updated from yt-dlp once known (thread-safe)
func (t *Task) GetTitle() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.title
}

// GetProgress returns the current progress (thread-safe)
func (t *Task) GetProgress() ytdlp.Progress {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.progress
}

// GetError returns the task error (thread-safe)
func (t *Task) GetError() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.err
}

// GetOutputPath returns the output path (thread-safe)
func (t *Task) GetOutputPath() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.outputPath
}

// GetCompletedAt returns the completion time (thread-safe)
func (t *Task) GetCompletedAt() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.completedAt
}

// String returns a human-readable description
func (t *Task) String() string {
	state := t.GetState()
	return fmt.Sprintf("[%s] %s (%s)", state.Icon(), t.GetTitle(), state)
}
