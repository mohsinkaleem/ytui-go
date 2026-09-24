package download

import (
	"context"
	"sync"
	"time"

	"github.com/mohsinkaleem/ytui-go/internal/store"
	"github.com/mohsinkaleem/ytui-go/internal/types"
	"github.com/mohsinkaleem/ytui-go/internal/ytdlp"
)

// Manager runs queued tasks on a fixed pool of workers.
type Manager struct {
	mu      sync.Mutex
	cond    *sync.Cond
	tasks   []*Task
	taskMap map[string]*Task
	pending []*Task // FIFO of tasks waiting for a worker
	closed  bool

	ctx     context.Context
	stop    context.CancelFunc
	workers sync.WaitGroup

	client *ytdlp.Client
	store  *store.Store
}

// NewManager creates a download manager and starts its workers.
func NewManager(maxConcurrent int, client *ytdlp.Client, st *store.Store) *Manager {
	if maxConcurrent <= 0 {
		maxConcurrent = 3
	}
	ctx, stop := context.WithCancel(context.Background())
	m := &Manager{
		taskMap: make(map[string]*Task),
		ctx:     ctx,
		stop:    stop,
		client:  client,
		store:   st,
	}
	m.cond = sync.NewCond(&m.mu)
	for i := 0; i < maxConcurrent; i++ {
		m.workers.Add(1)
		go m.worker()
	}
	return m
}

func (m *Manager) worker() {
	defer m.workers.Done()
	for {
		m.mu.Lock()
		for len(m.pending) == 0 && !m.closed {
			m.cond.Wait()
		}
		if m.closed {
			m.mu.Unlock()
			return
		}
		task := m.pending[0]
		m.pending = m.pending[1:]
		m.mu.Unlock()

		task.Start(m.ctx)
		m.persistTask(task)
	}
}

// schedule appends a task to the pending queue without blocking.
func (m *Manager) schedule(task *Task) {
	m.mu.Lock()
	m.pending = append(m.pending, task)
	m.mu.Unlock()
	m.cond.Signal()
}

// Enqueue adds a task to the download queue
func (m *Manager) Enqueue(task *Task) {
	m.mu.Lock()
	m.tasks = append(m.tasks, task)
	m.taskMap[task.ID] = task
	m.mu.Unlock()

	m.persistTask(task)
	m.schedule(task)
}

// EnqueueBatch enqueues multiple tasks
func (m *Manager) EnqueueBatch(videos []types.VideoItem, formatID string, opts ytdlp.DownloadOpts) []*Task {
	var tasks []*Task
	for _, v := range videos {
		task := NewTask(v, formatID, opts, m.client)
		m.Enqueue(task)
		tasks = append(tasks, task)
	}
	return tasks
}

// GetTask returns a task by ID
func (m *Manager) GetTask(id string) *Task {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.taskMap[id]
}

// GetTasks returns all tasks
func (m *Manager) GetTasks() []*Task {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]*Task, len(m.tasks))
	copy(result, m.tasks)
	return result
}

// PauseTask pauses a task
func (m *Manager) PauseTask(id string) {
	if task := m.GetTask(id); task != nil {
		task.Pause()
		m.persistTask(task)
	}
}

// ResumeTask re-queues a paused or failed task.
func (m *Manager) ResumeTask(id string) {
	if task := m.GetTask(id); task != nil && task.PrepareResume() {
		m.persistTask(task)
		m.schedule(task)
	}
}

// CancelTask cancels a task
func (m *Manager) CancelTask(id string) {
	if task := m.GetTask(id); task != nil {
		task.Cancel()
		m.persistTask(task)
	}
}

// Shutdown stops all downloads, leaving unfinished ones resumable, and
// returns how many were interrupted.
func (m *Manager) Shutdown() int {
	m.mu.Lock()
	m.closed = true
	tasks := append([]*Task(nil), m.tasks...)
	m.mu.Unlock()
	m.cond.Broadcast()
	m.stop()

	// Give yt-dlp a moment to exit cleanly so no process outlives the app.
	done := make(chan struct{})
	go func() {
		m.workers.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(6 * time.Second):
	}

	interrupted := 0
	for _, t := range tasks {
		if s := t.GetState(); s == StateQueued || s == StatePaused || s == StateDownloading {
			interrupted++
		}
		m.persistTask(t)
	}
	return interrupted
}

func (m *Manager) persistTask(task *Task) {
	if m.store == nil {
		return
	}
	record := store.DownloadRecord{
		ID:          task.ID,
		VideoID:     task.VideoID,
		Title:       task.GetTitle(),
		URL:         task.URL,
		FormatID:    task.FormatID,
		OutputPath:  task.GetOutputPath(),
		State:       string(task.GetState()),
		EmbedSubs:   task.Opts.EmbedSubs,
		EmbedMeta:   task.Opts.EmbedMetadata,
		EmbedChaps:  task.Opts.EmbedChapters,
		CreatedAt:   task.CreatedAt,
		CompletedAt: task.GetCompletedAt(),
	}
	if err := task.GetError(); err != nil {
		record.Error = err.Error()
	}
	_ = m.store.SaveDownload(record)
}
