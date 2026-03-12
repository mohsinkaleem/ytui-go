package download

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mohsinkaleem/ytui-go/internal/store"
	"github.com/mohsinkaleem/ytui-go/internal/types"
	"github.com/mohsinkaleem/ytui-go/internal/ytdlp"
)

// Manager manages a pool of download workers
type Manager struct {
	mu            sync.Mutex
	shutdownOnce  sync.Once
	tasks         []*Task
	taskMap       map[string]*Task
	maxConcurrent int
	activeCount   int
	program       *tea.Program
	client        *ytdlp.Client
	store         *store.Store
	queue         chan *Task
	done          chan struct{}
}

// NewManager creates a download manager
func NewManager(maxConcurrent int, client *ytdlp.Client, st *store.Store) *Manager {
	if maxConcurrent <= 0 {
		maxConcurrent = 3
	}
	return &Manager{
		taskMap:       make(map[string]*Task),
		maxConcurrent: maxConcurrent,
		client:        client,
		store:         st,
		queue:         make(chan *Task, 100),
		done:          make(chan struct{}),
	}
}

// SetProgram sets the tea.Program reference for sending messages
func (m *Manager) SetProgram(p *tea.Program) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.program = p
}

// Start launches worker goroutines
func (m *Manager) Start() {
	for i := 0; i < m.maxConcurrent; i++ {
		go m.worker()
	}
}

func (m *Manager) worker() {
	for {
		select {
		case task := <-m.queue:
			if task == nil {
				return
			}
			m.mu.Lock()
			m.activeCount++
			p := m.program
			m.mu.Unlock()

			if p != nil {
				task.Start(p)
			}

			m.mu.Lock()
			m.activeCount--
			m.mu.Unlock()

			// Persist final state
			if m.store != nil {
				m.persistTask(task)
			}

		case <-m.done:
			return
		}
	}
}

// Enqueue adds a task to the download queue
func (m *Manager) Enqueue(task *Task) {
	m.mu.Lock()
	m.tasks = append(m.tasks, task)
	m.taskMap[task.ID] = task
	m.mu.Unlock()

	// Persist to store
	if m.store != nil {
		record := store.DownloadRecord{
			ID:         task.ID,
			VideoID:    task.VideoID,
			Title:      task.Title,
			URL:        task.URL,
			FormatID:   task.FormatID,
			OutputPath: task.OutputPath,
			State:      string(task.State),
			EmbedSubs:  task.Opts.EmbedSubs,
			EmbedMeta:  task.Opts.EmbedMetadata,
			EmbedChaps: task.Opts.EmbedChapters,
			CreatedAt:  task.CreatedAt,
		}
		m.store.SaveDownload(record)
	}

	m.queue <- task
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

// ActiveCount returns the number of actively downloading tasks
func (m *Manager) ActiveCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.activeCount
}

// PauseTask pauses a task
func (m *Manager) PauseTask(id string) {
	m.mu.Lock()
	task := m.taskMap[id]
	m.mu.Unlock()
	if task != nil {
		task.Pause()
		m.persistTask(task)
	}
}

// ResumeTask re-queues a paused or failed task without blocking the caller.
func (m *Manager) ResumeTask(id string) {
	m.mu.Lock()
	task := m.taskMap[id]
	m.mu.Unlock()
	if task != nil && task.PrepareResume() {
		m.persistTask(task)
		// Send back to the worker pool queue (buffered, non-blocking for typical sizes)
		m.queue <- task
	}
}

// CancelTask cancels a task
func (m *Manager) CancelTask(id string) {
	m.mu.Lock()
	task := m.taskMap[id]
	m.mu.Unlock()
	if task != nil {
		task.Cancel(false)
		m.persistTask(task)
	}
}

// Shutdown cancels all active downloads and persists state
func (m *Manager) Shutdown() {
	m.shutdownOnce.Do(func() { close(m.done) })
	m.mu.Lock()
	tasks := make([]*Task, len(m.tasks))
	copy(tasks, m.tasks)
	m.mu.Unlock()

	for _, t := range tasks {
		state := t.GetState()
		if state == StateDownloading || state == StateQueued {
			t.Cancel(false)
		}
		m.persistTask(t)
	}
}

func (m *Manager) persistTask(task *Task) {
	if m.store == nil {
		return
	}
	record := store.DownloadRecord{
		ID:          task.ID,
		VideoID:     task.VideoID,
		Title:       task.Title,
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
	m.store.SaveDownload(record)
}

// CurrentTask returns the first actively downloading task, or nil
func (m *Manager) CurrentTask() *Task {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.tasks {
		if t.GetState() == StateDownloading {
			return t
		}
	}
	return nil
}

// TotalSpeed returns a formatted total download speed across all active tasks
func (m *Manager) TotalSpeed() string {
	// For now, return the current task's speed
	if t := m.CurrentTask(); t != nil {
		return t.GetProgress().Speed
	}
	return ""
}
