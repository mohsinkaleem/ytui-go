package download

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/mohsinkaleem/ytui-go/internal/types"
	dlpkg "github.com/mohsinkaleem/ytui-go/internal/ytdlp"
)

// Task represents a single download task
type Task struct {
	ID          string
	VideoID     string
	Title       string
	URL         string
	FormatID    string
	OutputPath  string
	State       TaskState
	Progress    types.ProgressMsg
	Error       error
	Opts        dlpkg.DownloadOpts
	CreatedAt   time.Time
	CompletedAt time.Time

	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
	client *dlpkg.Client
}

// NewTask creates a new download task
func NewTask(video types.VideoItem, formatID string, opts dlpkg.DownloadOpts, client *dlpkg.Client) *Task {
	return &Task{
		ID:        uuid.New().String(),
		VideoID:   video.ID,
		Title:     video.Title,
		URL:       video.VideoURL(),
		FormatID:  formatID,
		State:     StateQueued,
		Opts:      opts,
		CreatedAt: time.Now(),
		client:    client,
	}
}

// Start launches the download process and sends progress to the program.
// Blocks until the download completes, fails, or is cancelled.
func (t *Task) Start(program *tea.Program) {
	t.mu.Lock()
	if !ValidTransition(t.State, StateDownloading) {
		t.mu.Unlock()
		return
	}
	t.State = StateDownloading
	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.mu.Unlock()

	cmd, output, err := t.client.Download(t.ctx, t.URL, t.FormatID, t.Opts)
	if err != nil {
		t.mu.Lock()
		t.State = StateFailed
		t.Error = err
		t.mu.Unlock()
		program.Send(types.DownloadCompleteMsg{TaskID: t.ID, Err: err})
		return
	}
	defer output.Close()

	// Increase scanner buffer for very long yt-dlp output lines
	scanner := bufio.NewScanner(output)
	scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)

	// Collect non-progress lines so we can show the real error message
	var errLines []string
	for scanner.Scan() {
		line := scanner.Text()

		// Parse output path from yt-dlp destination lines
		if strings.HasPrefix(line, "[download] Destination:") {
			path := strings.TrimSpace(strings.TrimPrefix(line, "[download] Destination:"))
			if path != "" {
				t.mu.Lock()
				t.OutputPath = path
				t.mu.Unlock()
			}
			continue
		}

		progress, parseErr := dlpkg.ParseProgressLine(line)
		if parseErr != nil {
			// Collect lines that may contain error information
			trimmed := strings.TrimSpace(line)
			if strings.Contains(trimmed, "ERROR:") || strings.Contains(trimmed, "error:") {
				errLines = append(errLines, trimmed)
			} else if strings.HasPrefix(trimmed, "WARNING:") {
				// ignore warnings
			} else if trimmed != "" && !strings.HasPrefix(trimmed, "[") {
				// collect unrecognised non-status lines (might be Python tracebacks etc.)
				errLines = append(errLines, trimmed)
			}
			continue
		}

		msg := types.ProgressMsg{
			TaskID:          t.ID,
			Percent:         progress.Percent,
			DownloadedBytes: progress.DownloadedBytes,
			TotalBytes:      progress.TotalBytes,
			Speed:           progress.Speed,
			ETA:             progress.ETA,
			Status:          progress.Status,
		}

		t.mu.Lock()
		t.Progress = msg
		t.mu.Unlock()

		program.Send(msg)
	}

	cmdErr := cmd.Wait()
	t.mu.Lock()
	if cmdErr != nil {
		if t.ctx.Err() != nil {
			// Was cancelled/paused
			if t.State == StatePaused {
				t.mu.Unlock()
				return
			}
			t.State = StateCancelled
		} else {
			t.State = StateFailed
			// Prefer the captured yt-dlp error lines over the raw "exit status 1"
			if len(errLines) > 0 {
				t.Error = errors.New(strings.Join(errLines, "; "))
			} else {
				t.Error = fmt.Errorf("%w", cmdErr)
			}
		}
	} else {
		t.State = StateCompleted
		t.CompletedAt = time.Now()
	}
	t.mu.Unlock()

	program.Send(types.DownloadCompleteMsg{
		TaskID: t.ID,
		Err:    cmdErr,
	})
}

// Pause pauses the download by killing the process
func (t *Task) Pause() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !ValidTransition(t.State, StatePaused) {
		return
	}
	t.State = StatePaused
	if t.cancel != nil {
		t.cancel()
	}
}

// PrepareResume resets a paused or failed task back to StateQueued so the
// manager can re-enqueue it without blocking the UI goroutine.
// Returns true when the state was successfully reset.
func (t *Task) PrepareResume() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	switch t.State {
	case StatePaused:
		t.Opts.ContinueDL = true // yt-dlp --continue flag
		t.State = StateQueued
		return true
	case StateFailed:
		t.Error = nil
		t.State = StateQueued
		return true
	}
	return false
}

// Cancel cancels the download
func (t *Task) Cancel(deletePartial bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.State == StateCompleted || t.State == StateCancelled {
		return // already in terminal state
	}

	t.State = StateCancelled
	if t.cancel != nil {
		t.cancel()
	}

	if deletePartial && t.OutputPath != "" {
		os.Remove(t.OutputPath)
		os.Remove(t.OutputPath + ".part")
	}
}

// GetState returns the current task state (thread-safe)
func (t *Task) GetState() TaskState {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.State
}

// GetProgress returns the current progress (thread-safe)
func (t *Task) GetProgress() types.ProgressMsg {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Progress
}

// GetError returns the task error (thread-safe)
func (t *Task) GetError() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Error
}

// GetOutputPath returns the output path (thread-safe)
func (t *Task) GetOutputPath() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.OutputPath
}

// GetCompletedAt returns the completion time (thread-safe)
func (t *Task) GetCompletedAt() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.CompletedAt
}

// String returns a human-readable description
func (t *Task) String() string {
	return fmt.Sprintf("[%s] %s (%s)", t.State.Icon(), t.Title, t.State)
}
