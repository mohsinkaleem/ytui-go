package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mohsinkaleem/ytui-go/internal/download"
	"github.com/mohsinkaleem/ytui-go/internal/slash"
	"github.com/mohsinkaleem/ytui-go/internal/store"
	"github.com/mohsinkaleem/ytui-go/internal/styles"
	"github.com/mohsinkaleem/ytui-go/internal/types"
	"github.com/mohsinkaleem/ytui-go/internal/utils"
	"github.com/mohsinkaleem/ytui-go/internal/ytdlp"
)

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// --- Window resize ---
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Search.SetSize(msg.Width-2, msg.Height-4)
		m.VideoList.SetSize(msg.Width-2, msg.Height-4)
		m.FormatList.SetSize(msg.Width-2, msg.Height-4)
		m.Download.SetSize(msg.Width-2, msg.Height-4)
		m.Player.SetSize(msg.Width-2, msg.Height-4)
		m.ResumeList.SetSize(msg.Width-2, msg.Height-4)
		return m, nil

	// --- Key handling ---
	case tea.KeyMsg:
		// Global keys
		switch msg.String() {
		case "ctrl+c":
			m.DownloadMgr.Shutdown()
			if m.Store != nil {
				m.Store.Close()
			}
			return m, tea.Quit
		case "g":
			// Jump to active downloads from any non-download/non-input screen
			if m.State != types.StateDownload && m.State != types.StateSearchInput && m.State != types.StateLoading {
				tasks := m.DownloadMgr.GetTasks()
				if len(tasks) > 0 {
					m.Download.SetTasks(tasks)
					m.PrevState = m.State
					m.State = types.StateDownload
					return m, nil
				}
			}
		}

		// State-specific key handling
		switch m.State {
		case types.StateSearchInput:
			return m.updateSearchInput(msg)
		case types.StateLoading:
			return m.updateLoading(msg)
		case types.StateVideoList:
			return m.updateVideoList(msg)
		case types.StateFormatList:
			return m.updateFormatList(msg)
		case types.StateDownload:
			return m.updateDownload(msg)
		case types.StateVideoPlaying:
			return m, nil
		case types.StateResumeList:
			return m.updateResumeList(msg)
		}

	// --- Spinner tick ---
	case spinner.TickMsg:
		if m.State == types.StateLoading {
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	// --- Search results ---
	case types.SearchResultMsg:
		if msg.Err != nil {
			m.ErrMsg = msg.Err.Error()
			m.State = types.StateSearchInput
			m.Search.ErrMsg = msg.Err.Error()
			return m, nil
		}
		m.PrevState = types.StateSearchInput
		m.State = types.StateVideoList
		m.VideoList.SetItems(msg.Videos, "", msg.Query)
		m.VideoList.ClearSelections()
		m.CurrentQuery = msg.Query
		// Save search history
		if m.Store != nil {
			m.Store.SaveSearch(store.SearchEntry{
				URL:       msg.Query,
				Title:     msg.Query,
				Timestamp: time.Now(),
				Type:      "search",
			})
		}
		// Update in-memory history (newest first, deduplicate)
		filtered := []string{msg.Query}
		for _, h := range m.Search.History {
			if h != msg.Query {
				filtered = append(filtered, h)
			}
		}
		m.Search.History = filtered
		m.Search.HistoryIndex = -1
		return m, nil

	// --- Format results ---
	case types.FormatResultMsg:
		if msg.Err != nil {
			m.ErrMsg = msg.Err.Error()
			m.State = types.StateSearchInput
			m.Search.ErrMsg = msg.Err.Error()
			return m, nil
		}
		m.PrevState = m.State
		if m.CurrentQuery != "" {
			m.PrevState = types.StateVideoList
		} else {
			m.PrevState = types.StateSearchInput
		}
		m.State = types.StateFormatList
		m.FormatList.SetFormats(msg.Formats, msg.Video)
		m.SelectedVideo = msg.Video
		return m, nil

	// --- Playlist results ---
	case types.PlaylistResultMsg:
		if msg.Err != nil {
			m.ErrMsg = msg.Err.Error()
			m.State = types.StateSearchInput
			m.Search.ErrMsg = msg.Err.Error()
			return m, nil
		}
		m.PrevState = types.StateSearchInput
		m.State = types.StateVideoList
		m.VideoList.SetItems(msg.Videos, msg.Title, "")
		m.VideoList.ClearSelections()
		// Save playlist to history
		if m.Store != nil && msg.Title != "" {
			m.Store.SaveSearch(store.SearchEntry{
				URL:       msg.Title,
				Title:     msg.Title,
				Timestamp: time.Now(),
				Type:      "playlist",
			})
			filtered := []string{msg.Title}
			for _, h := range m.Search.History {
				if h != msg.Title {
					filtered = append(filtered, h)
				}
			}
			m.Search.History = filtered
			m.Search.HistoryIndex = -1
		}
		return m, nil

	// --- Download progress ---
	case types.ProgressMsg:
		// Refresh task list on download screen so new tasks are visible
		if m.State == types.StateDownload {
			m.Download.SetTasks(m.DownloadMgr.GetTasks())
		}
		// Schedule a re-render tick to keep status bar indicator updated
		return m, m.scheduleDownloadTick()

	// --- Download complete ---
	case types.DownloadCompleteMsg:
		// Refresh task list
		if m.State == types.StateDownload {
			m.Download.SetTasks(m.DownloadMgr.GetTasks())
		}
		// Check if all downloads are done
		if m.Download.AllDone() {
			// Stay on download screen to show summary
		}
		return m, nil

	// --- Download tick (periodic refresh for indicators) ---
	case types.DownloadTickMsg:
		// Just re-render; schedule another tick if downloads active
		if m.DownloadMgr.ActiveCount() > 0 {
			return m, m.scheduleDownloadTick()
		}
		return m, nil

	// --- MPV exited ---
	case types.MPVExitedMsg:
		m.State = m.PrevState
		if m.State == "" {
			m.State = types.StateSearchInput
		}
		if msg.Err != nil {
			m.ErrMsg = "mpv: " + msg.Err.Error()
			m.Search.ErrMsg = m.ErrMsg
		}
		return m, nil

	// --- Toast ---
	case types.ShowToastMsg:
		m.ToastMsg = msg.Message
		return m, func() tea.Msg {
			time.Sleep(3 * time.Second)
			return types.ClearToastMsg{}
		}

	case types.ClearToastMsg:
		m.ToastMsg = ""
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

// --- State-specific update handlers ---

func (m Model) updateSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Theme picker takes priority
		if m.Search.ShowThemes && len(m.Search.ThemeNames) > 0 {
			selected := m.Search.ThemeNames[m.Search.ThemeSelected]
			return m.handleSlashCommand("/theme " + selected)
		}

		// If slash dropdown is visible and a command is selected, use it
		if m.Search.ShowSlash && len(m.Search.SlashMatches) > 0 {
			sel := m.Search.SlashMatches[m.Search.SlashSelected]
			if sel.Args == "" {
				// No args needed – execute immediately
				return m.handleSlashCommand("/" + sel.Name)
			}
			// Command needs args – fill input so user can type them
			m.Search.Input.SetValue("/" + sel.Name + " ")
			m.Search.Input.CursorEnd()
			m.Search.SlashMatches = m.SlashRegistry.Match(sel.Name)
			m.Search.SlashSelected = 0
			return m, nil
		}

		input := strings.TrimSpace(m.Search.Input.Value())
		if input == "" {
			return m, nil
		}

		// Check for slash commands
		if strings.HasPrefix(input, "/") {
			return m.handleSlashCommand(input)
		}

		m.Search.ErrMsg = ""
		m.ErrMsg = ""
		m.SelectedVideos = nil // clear any stale multi-selection

		// Detect URL vs search query
		if ytdlp.IsURL(input) {
			if ytdlp.IsPlaylistURL(input) {
				// Playlist
				m.State = types.StateLoading
				m.LoadingType = types.LoadingPlaylist
				return m, tea.Batch(m.Spinner.Tick, m.PlaylistMgr.FetchPlaylist(input))
			}
			// Single video URL → format list
			m.State = types.StateLoading
			m.LoadingType = types.LoadingFormats
			return m, tea.Batch(m.Spinner.Tick, m.FormatsManager.FetchFormats(input))
		}

		// Search query
		m.State = types.StateLoading
		m.LoadingType = types.LoadingSearch
		m.CurrentQuery = input
		return m, tea.Batch(m.Spinner.Tick, m.SearchManager.Search(input, m.Search.SortBy))

	case "tab":
		// Cycle sort option forward
		m.Search.SortIndex = (m.Search.SortIndex + 1) % len(types.SortOptions)
		m.Search.SortBy = types.SortOptions[m.Search.SortIndex]
		return m, nil

	case "shift+tab":
		// Cycle sort option backward
		m.Search.SortIndex--
		if m.Search.SortIndex < 0 {
			m.Search.SortIndex = len(types.SortOptions) - 1
		}
		m.Search.SortBy = types.SortOptions[m.Search.SortIndex]
		return m, nil

	case "ctrl+s":
		m.Search.EmbedSubs = !m.Search.EmbedSubs
		m.persistSettings()
		return m, nil

	case "ctrl+m":
		m.Search.EmbedMetadata = !m.Search.EmbedMetadata
		m.persistSettings()
		return m, nil

	case "ctrl+j":
		m.Search.EmbedChapters = !m.Search.EmbedChapters
		m.persistSettings()
		return m, nil

	case "up":
		// Theme picker navigation
		if m.Search.ShowThemes && len(m.Search.ThemeNames) > 0 {
			m.Search.ThemeSelected--
			if m.Search.ThemeSelected < 0 {
				m.Search.ThemeSelected = len(m.Search.ThemeNames) - 1
			}
			return m, nil
		}
		// Navigate slash dropdown when visible
		if m.Search.ShowSlash && len(m.Search.SlashMatches) > 0 {
			m.Search.SlashSelected--
			if m.Search.SlashSelected < 0 {
				m.Search.SlashSelected = len(m.Search.SlashMatches) - 1
			}
			return m, nil
		}
		// History navigation
		if len(m.Search.History) > 0 && m.Search.HistoryIndex < len(m.Search.History)-1 {
			m.Search.HistoryIndex++
			m.Search.Input.SetValue(m.Search.History[m.Search.HistoryIndex])
			m.Search.Input.CursorEnd()
		}
		return m, nil

	case "down":
		// Theme picker navigation
		if m.Search.ShowThemes && len(m.Search.ThemeNames) > 0 {
			m.Search.ThemeSelected = (m.Search.ThemeSelected + 1) % len(m.Search.ThemeNames)
			return m, nil
		}
		// Navigate slash dropdown when visible
		if m.Search.ShowSlash && len(m.Search.SlashMatches) > 0 {
			m.Search.SlashSelected = (m.Search.SlashSelected + 1) % len(m.Search.SlashMatches)
			return m, nil
		}
		if m.Search.HistoryIndex > 0 {
			m.Search.HistoryIndex--
			m.Search.Input.SetValue(m.Search.History[m.Search.HistoryIndex])
			m.Search.Input.CursorEnd()
		} else if m.Search.HistoryIndex == 0 {
			m.Search.HistoryIndex = -1
			m.Search.Input.SetValue("")
		}
		return m, nil

	case "esc":
		// Dismiss theme picker or slash dropdown
		if m.Search.ShowThemes {
			m.Search.ShowThemes = false
			m.Search.ThemeNames = nil
			return m, nil
		}
		if m.Search.ShowSlash {
			m.Search.ShowSlash = false
			m.Search.SlashMatches = nil
			m.Search.Input.SetValue("")
			return m, nil
		}
		return m, nil

	default:
		// Dismiss theme picker when typing
		if m.Search.ShowThemes {
			m.Search.ShowThemes = false
			m.Search.ThemeNames = nil
		}

		// Handle slash command autocomplete
		var cmd tea.Cmd
		m.Search.Input, cmd = m.Search.Input.Update(msg)

		// Update slash matches
		val := m.Search.Input.Value()
		if strings.HasPrefix(val, "/") && len(val) > 0 {
			m.Search.ShowSlash = true
			m.Search.SlashMatches = m.SlashRegistry.Match(val[1:])
			if m.Search.SlashSelected >= len(m.Search.SlashMatches) {
				m.Search.SlashSelected = 0
			}
		} else {
			m.Search.ShowSlash = false
			m.Search.SlashMatches = nil
		}

		return m, cmd
	}
}

func (m Model) updateLoading(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "c":
		// Cancel current operation
		switch m.LoadingType {
		case types.LoadingSearch:
			m.SearchManager.Cancel()
		case types.LoadingFormats:
			m.FormatsManager.Cancel()
		case types.LoadingPlaylist:
			m.PlaylistMgr.Cancel()
		}
		m.State = types.StateSearchInput
		return m, nil
	}
	return m, nil
}

func (m Model) updateVideoList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If filtering, let the list handle it
	if m.VideoList.IsFiltering() {
		var cmd tea.Cmd
		m.VideoList.List, cmd = m.VideoList.List.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "enter":
		video, ok := m.VideoList.SelectedVideo()
		if !ok {
			return m, nil
		}
		// Clear multi-selection: enter goes to format list for ONE video only.
		// Batch download is handled by the "d" shortcut in video list.
		m.SelectedVideos = nil
		m.SelectedVideo = video
		m.State = types.StateLoading
		m.LoadingType = types.LoadingFormats
		url := video.VideoURL()
		return m, tea.Batch(m.Spinner.Tick, m.FormatsManager.FetchFormats(url))

	case " ":
		m.VideoList.ToggleSelected()
		return m, nil

	case "a":
		m.VideoList.SelectAll()
		return m, nil

	case "d":
		// Quick download with best format
		selected := m.VideoList.GetSelectedVideos()
		if len(selected) == 0 {
			if v, ok := m.VideoList.SelectedVideo(); ok {
				selected = []types.VideoItem{v}
			}
		}
		if len(selected) > 0 {
			opts := m.getDownloadOpts()
			tasks := m.DownloadMgr.EnqueueBatch(selected, "bestvideo+bestaudio/best", ytdlp.DownloadOpts{
				EmbedSubs:     opts.EmbedSubs,
				EmbedMetadata: opts.EmbedMetadata,
				EmbedChapters: opts.EmbedChapters,
				OutputDir:     opts.OutputDir,
			})
			m.Download.SetTasks(tasks)
			m.PrevState = types.StateVideoList
			m.State = types.StateDownload
		}
		return m, nil

	case "p":
		video, ok := m.VideoList.SelectedVideo()
		if !ok {
			return m, nil
		}
		m.PrevState = types.StateVideoList
		m.State = types.StateVideoPlaying
		m.Player.SetVideo(video, "")
		url := video.VideoURL()
		return m, m.PlayerManager.Play(url, "")

	case "ctrl+y":
		video, ok := m.VideoList.SelectedVideo()
		if !ok {
			return m, nil
		}
		url := video.VideoURL()
		if err := clipboard.WriteAll(url); err != nil {
			return m, func() tea.Msg {
				return types.ShowToastMsg{Message: "Clipboard unavailable: " + err.Error()}
			}
		}
		return m, func() tea.Msg {
			return types.ShowToastMsg{Message: "Copied to clipboard"}
		}

	case "/":
		var cmd tea.Cmd
		m.VideoList.List, cmd = m.VideoList.List.Update(msg)
		return m, cmd

	case "esc":
		// Home: always go to search input
		m.State = types.StateSearchInput
		m.Search.Input.Focus()
		return m, nil

	case "b":
		// Back: go to previous state (for video list, that's search)
		m.State = types.StateSearchInput
		m.Search.Input.Focus()
		return m, nil

	default:
		var cmd tea.Cmd
		m.VideoList.List, cmd = m.VideoList.List.Update(msg)
		return m, cmd
	}
}

func (m Model) updateFormatList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If on custom tab, let input handle keys
	if m.FormatList.ActiveTab == types.FormatTabCustom {
		switch msg.String() {
		case "enter":
			formatID := m.FormatList.GetFormatID()
			return m.startDownload(formatID)
		case "tab":
			m.FormatList.NextTab()
			return m, nil
		case "shift+tab":
			m.FormatList.PrevTab()
			return m, nil
		case "up", "k":
			if m.FormatList.ComboIdx > 0 {
				m.FormatList.ComboIdx--
			}
			return m, nil
		case "down", "j":
			if m.FormatList.ComboIdx < len(m.FormatList.Combos)-1 {
				m.FormatList.ComboIdx++
			}
			return m, nil
		case "esc":
			// Home: always go to search input
			m.State = types.StateSearchInput
			m.Search.Input.Focus()
			return m, nil
		case "b":
			// Back: go to video list or search
			if m.CurrentQuery != "" {
				m.State = types.StateVideoList
			} else {
				m.State = types.StateSearchInput
				m.Search.Input.Focus()
			}
			return m, nil
		default:
			var cmd tea.Cmd
			m.FormatList.CustomInput, cmd = m.FormatList.CustomInput.Update(msg)
			return m, cmd
		}
	}

	switch msg.String() {
	case "enter":
		formatID := m.FormatList.GetFormatID()
		return m.startDownload(formatID)

	case "p":
		url := m.SelectedVideo.VideoURL()
		formatID := m.FormatList.GetFormatID()
		m.PrevState = types.StateFormatList
		m.State = types.StateVideoPlaying
		m.Player.SetVideo(m.SelectedVideo, formatID)
		return m, m.PlayerManager.Play(url, formatID)

	case "tab":
		m.FormatList.NextTab()
		return m, nil

	case "shift+tab":
		m.FormatList.PrevTab()
		return m, nil

	case "esc":
		// Home: always go to search input
		m.State = types.StateSearchInput
		m.Search.Input.Focus()
		return m, nil

	case "b":
		// Back: go to video list or search
		if m.CurrentQuery != "" {
			m.State = types.StateVideoList
		} else {
			m.State = types.StateSearchInput
			m.Search.Input.Focus()
		}
		return m, nil

	default:
		var cmd tea.Cmd
		m.FormatList.List, cmd = m.FormatList.List.Update(msg)
		return m, cmd
	}
}

func (m Model) updateDownload(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	task := m.Download.CurrentTask()

	switch msg.String() {
	case "up", "k":
		m.Download.MoveUp()
		return m, nil

	case "down", "j":
		m.Download.MoveDown()
		return m, nil

	case "p":
		if task != nil {
			state := task.GetState()
			if state == download.StateDownloading {
				m.DownloadMgr.PauseTask(task.ID)
			} else if state == download.StatePaused {
				m.DownloadMgr.ResumeTask(task.ID)
			}
		}
		return m, nil

	case "c":
		if task != nil {
			m.DownloadMgr.CancelTask(task.ID)
		}
		return m, nil

	case "r":
		// Retry failed task (current cursor)
		if task != nil && task.GetState() == download.StateFailed {
			m.DownloadMgr.ResumeTask(task.ID)
		}
		return m, nil

	case "R":
		// Retry ALL failed tasks in the queue
		retried := 0
		for _, t := range m.Download.Tasks {
			if t.GetState() == download.StateFailed {
				m.DownloadMgr.ResumeTask(t.ID)
				retried++
			}
		}
		if retried > 0 {
			return m, func() tea.Msg {
				return types.ShowToastMsg{Message: fmt.Sprintf("Retrying %d failed download(s)", retried)}
			}
		}
		return m, nil

	case "s":
		// Skip - cancel current and move on
		if task != nil {
			m.DownloadMgr.CancelTask(task.ID)
		}
		return m, nil

	case "esc":
		// Home: always go to search input
		m.State = types.StateSearchInput
		m.Search.Input.Focus()
		return m, nil

	case "b":
		// Back: go to previous state
		if m.PrevState != "" && m.PrevState != types.StateDownload {
			m.State = m.PrevState
			if m.State == types.StateSearchInput {
				m.Search.Input.Focus()
			}
		} else if m.CurrentQuery != "" {
			m.State = types.StateVideoList
		} else {
			m.State = types.StateSearchInput
			m.Search.Input.Focus()
		}
		return m, nil
	}

	return m, nil
}

func (m Model) updateResumeList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.ResumeList.MoveUp()
		return m, nil

	case "down", "j":
		m.ResumeList.MoveDown()
		return m, nil

	case "enter":
		rec := m.ResumeList.SelectedItem()
		if rec == nil {
			return m, nil
		}
		// Reconstruct a download task from the stored record
		video := types.VideoItem{
			ID:    rec.VideoID,
			Title: rec.Title,
			URL:   rec.URL,
		}
		opts := m.getDownloadOpts()
		task := download.NewTask(video, rec.FormatID, ytdlp.DownloadOpts{
			EmbedSubs:     rec.EmbedSubs,
			EmbedMetadata: rec.EmbedMeta,
			EmbedChapters: rec.EmbedChaps,
			OutputDir:     opts.OutputDir,
			ContinueDL:    true,
		}, m.YtdlpClient)
		m.PrevState = types.StateResumeList
		m.DownloadMgr.Enqueue(task)
		m.Download.SetTasks([]*download.Task{task})
		m.Download.Video = video
		m.State = types.StateDownload
		return m, nil

	case "d":
		// Resume all incomplete downloads at once
		if len(m.ResumeList.Items) > 0 {
			var tasks []*download.Task
			opts := m.getDownloadOpts()
			for _, rec := range m.ResumeList.Items {
				video := types.VideoItem{
					ID:    rec.VideoID,
					Title: rec.Title,
					URL:   rec.URL,
				}
				task := download.NewTask(video, rec.FormatID, ytdlp.DownloadOpts{
					EmbedSubs:     rec.EmbedSubs,
					EmbedMetadata: rec.EmbedMeta,
					EmbedChapters: rec.EmbedChaps,
					OutputDir:     opts.OutputDir,
					ContinueDL:    true,
				}, m.YtdlpClient)
				m.DownloadMgr.Enqueue(task)
				tasks = append(tasks, task)
			}
			m.PrevState = types.StateResumeList
			m.Download.SetTasks(tasks)
			m.State = types.StateDownload
		}
		return m, nil

	case "x":
		// Delete the selected incomplete download record
		rec := m.ResumeList.SelectedItem()
		if rec != nil && m.Store != nil {
			m.Store.DeleteDownload(rec.ID)
			// Refresh list
			incomplete, _ := m.Store.GetIncomplete()
			m.ResumeList.SetItems(incomplete)
			if len(incomplete) == 0 {
				m.State = types.StateSearchInput
				m.Search.Input.Focus()
			}
		}
		return m, nil

	case "esc":
		// Home: always go to search input
		m.State = types.StateSearchInput
		m.Search.Input.Focus()
		return m, nil

	case "b":
		// Back: for resume list, back is also search
		m.State = types.StateSearchInput
		m.Search.Input.Focus()
		return m, nil
	}

	return m, nil
}

// --- Helper methods ---

func (m Model) startDownload(formatID string) (tea.Model, tea.Cmd) {
	// Auto-merge: if the selected format is video-only, pair with best audio
	if m.FormatList.IsVideoOnlyFormat(formatID) {
		formatID = formatID + "+bestaudio"
	}

	opts := m.getDownloadOpts()
	dlOpts := ytdlp.DownloadOpts{
		EmbedSubs:     opts.EmbedSubs,
		EmbedMetadata: opts.EmbedMetadata,
		EmbedChapters: opts.EmbedChapters,
		OutputDir:     opts.OutputDir,
	}

	// Batch download if multiple videos were selected
	if len(m.SelectedVideos) > 1 {
		tasks := m.DownloadMgr.EnqueueBatch(m.SelectedVideos, formatID, dlOpts)
		m.Download.SetTasks(tasks)
		m.PrevState = types.StateFormatList
		m.State = types.StateDownload
		return m, nil
	}

	task := download.NewTask(m.SelectedVideo, formatID, dlOpts, m.YtdlpClient)
	m.DownloadMgr.Enqueue(task)
	m.Download.SetTasks([]*download.Task{task})
	m.Download.Video = m.SelectedVideo
	m.PrevState = types.StateFormatList
	m.State = types.StateDownload
	return m, nil
}

func (m Model) getDownloadOpts() types.DownloadOpts {
	outputDir := "~/Downloads"
	if m.Store != nil {
		if settings, err := m.Store.GetSettings(); err == nil && settings.DownloadDir != "" {
			outputDir = settings.DownloadDir
		}
	}
	return types.DownloadOpts{
		EmbedSubs:     m.Search.EmbedSubs,
		EmbedMetadata: m.Search.EmbedMetadata,
		EmbedChapters: m.Search.EmbedChapters,
		OutputDir:     outputDir,
	}
}

func (m Model) handleSlashCommand(input string) (tea.Model, tea.Cmd) {
	name, args := slash.ParseInput(input)

	switch name {
	case "exit", "quit":
		m.DownloadMgr.Shutdown()
		if m.Store != nil {
			m.Store.Close()
		}
		return m, tea.Quit

	case "help":
		return m, func() tea.Msg {
			return types.ShowToastMsg{Message: "/download <url> • /play <url> • /playlist <url> • /theme • /cookies <browser|file> • /downloaddir • /resume • /exit"}
		}

	case "cookies":
		validBrowsers := []string{"chrome", "firefox", "brave", "edge", "opera", "safari", "chromium", "vivaldi"}
		if args != "" {
			if args == "off" || args == "none" || args == "clear" {
				m.YtdlpClient.SetCookiesFrom("")
				m.YtdlpClient.SetCookiesFile("")
				if m.Store != nil {
					settings, _ := m.Store.GetSettings()
					settings.CookiesFrom = ""
					settings.CookiesFile = ""
					m.Store.SaveSettings(settings)
				}
				m.Search.Input.SetValue("")
				m.Search.ShowSlash = false
				return m, func() tea.Msg {
					return types.ShowToastMsg{Message: "Cookies disabled"}
				}
			}

			// Check if argument is a file path (cookies.txt)
			expanded := utils.ExpandPath(args)
			if info, err := os.Stat(expanded); err == nil && !info.IsDir() {
				// It's a file — use as cookies.txt
				m.YtdlpClient.SetCookiesFrom("")
				m.YtdlpClient.SetCookiesFile(expanded)
				if m.Store != nil {
					settings, _ := m.Store.GetSettings()
					settings.CookiesFrom = ""
					settings.CookiesFile = expanded
					m.Store.SaveSettings(settings)
				}
				m.Search.Input.SetValue("")
				m.Search.ShowSlash = false
				return m, func() tea.Msg {
					return types.ShowToastMsg{Message: "Using cookies from file: " + expanded}
				}
			}

			valid := false
			for _, b := range validBrowsers {
				if b == args {
					valid = true
					break
				}
			}
			if !valid {
				m.Search.ErrMsg = "Unknown browser: " + args + ". Valid: chrome, firefox, brave, edge, opera, safari, chromium, vivaldi, off, or path to cookies.txt"
				return m, nil
			}
			m.YtdlpClient.SetCookiesFrom(args)
			m.YtdlpClient.SetCookiesFile("")
			if m.Store != nil {
				settings, _ := m.Store.GetSettings()
				settings.CookiesFrom = args
				settings.CookiesFile = ""
				m.Store.SaveSettings(settings)
			}
			m.Search.Input.SetValue("")
			m.Search.ShowSlash = false
			return m, func() tea.Msg {
				return types.ShowToastMsg{Message: "Using cookies from " + args}
			}
		}
		current := m.YtdlpClient.CookiesFrom()
		if current == "" {
			if f := m.YtdlpClient.CookiesFile(); f != "" {
				current = "file: " + f
			} else {
				current = "none"
			}
		}
		m.Search.Input.SetValue("/cookies ")
		m.Search.Input.CursorEnd()
		m.Search.ShowSlash = false
		return m, func() tea.Msg {
			return types.ShowToastMsg{Message: "Current: " + current + " (browser name, path to cookies.txt, or off)"}
		}

	case "theme":
		if args != "" {
			if styles.LoadTheme(args) {
				m.Search.Input.SetValue("")
				m.Search.ShowSlash = false
				m.Search.ShowThemes = false
				// Persist theme selection
				if m.Store != nil {
					settings, _ := m.Store.GetSettings()
					settings.Theme = args
					m.Store.SaveSettings(settings)
				}
				return m, func() tea.Msg {
					return types.ShowToastMsg{Message: "Theme switched to " + args}
				}
			}
			avail := strings.Join(styles.ThemeNames(), ", ")
			m.Search.ErrMsg = "Unknown theme: " + args + ". Available: " + avail
			return m, nil
		}
		// Open theme picker
		m.Search.Input.SetValue("")
		m.Search.ShowSlash = false
		m.Search.ShowThemes = true
		m.Search.ThemeNames = styles.ThemeNames()
		m.Search.ThemeSelected = 0
		return m, nil

	case "clear":
		if m.Store != nil {
			m.Store.ClearHistory()
		}
		m.Search.Input.SetValue("")
		m.Search.ShowSlash = false
		return m, func() tea.Msg {
			return types.ShowToastMsg{Message: "Search history cleared"}
		}

	case "download":
		if args != "" {
			m.Search.Input.SetValue("")
			m.Search.ShowSlash = false
			// Direct download with best format — skip format selection
			video := types.VideoItem{URL: args, Title: args}
			opts := m.getDownloadOpts()
			dlOpts := ytdlp.DownloadOpts{
				EmbedSubs:     opts.EmbedSubs,
				EmbedMetadata: opts.EmbedMetadata,
				EmbedChapters: opts.EmbedChapters,
				OutputDir:     opts.OutputDir,
			}
			task := download.NewTask(video, "bestvideo+bestaudio/best", dlOpts, m.YtdlpClient)
			m.DownloadMgr.Enqueue(task)
			m.Download.SetTasks([]*download.Task{task})
			m.Download.Video = video
			m.PrevState = types.StateSearchInput
			m.State = types.StateDownload
			return m, nil
		}
		m.Search.ErrMsg = "Usage: /download <url>"
		return m, nil

	case "playlist":
		if args != "" {
			m.Search.Input.SetValue("")
			m.Search.ShowSlash = false
			m.State = types.StateLoading
			m.LoadingType = types.LoadingPlaylist
			return m, tea.Batch(m.Spinner.Tick, m.PlaylistMgr.FetchPlaylist(args))
		}
		m.Search.ErrMsg = "Usage: /playlist <url>"
		return m, nil

	case "play":
		if args != "" {
			m.Search.Input.SetValue("")
			m.Search.ShowSlash = false
			m.PrevState = types.StateSearchInput
			m.State = types.StateVideoPlaying
			v := types.VideoItem{URL: args, Title: args}
			m.Player.SetVideo(v, "")
			return m, m.PlayerManager.Play(args, "")
		}
		m.Search.ErrMsg = "Usage: /play <url>"
		return m, nil

	case "resume":
		// Navigate to resume list view
		m.Search.Input.SetValue("")
		m.Search.ShowSlash = false
		if m.Store != nil {
			incomplete, _ := m.Store.GetIncomplete()
			if len(incomplete) > 0 {
				m.ResumeList.SetItems(incomplete)
				m.State = types.StateResumeList
				return m, nil
			}
		}
		return m, func() tea.Msg {
			return types.ShowToastMsg{Message: "No unfinished downloads"}
		}

	case "downloaddir":
		if args != "" {
			expanded := utils.ExpandPath(args)
			info, err := os.Stat(expanded)
			if err != nil {
				// Path doesn't exist — suggest the closest existing parent
				suggestion := utils.ClosestExistingDir(expanded)
				m.Search.ErrMsg = "Directory not found: " + expanded
				if suggestion != "" && suggestion != expanded {
					m.Search.ErrMsg += " (did you mean: " + suggestion + "?)"
				}
				return m, nil
			}
			if !info.IsDir() {
				m.Search.ErrMsg = "Not a directory: " + expanded
				return m, nil
			}
			// Save to store
			if m.Store != nil {
				settings, _ := m.Store.GetSettings()
				settings.DownloadDir = args
				m.Store.SaveSettings(settings)
			}
			m.Search.Input.SetValue("")
			m.Search.ShowSlash = false
			return m, func() tea.Msg {
				return types.ShowToastMsg{Message: "Download folder set to " + expanded}
			}
		}
		// Show current download dir
		dir := "~/Downloads"
		if m.Store != nil {
			if settings, err := m.Store.GetSettings(); err == nil && settings.DownloadDir != "" {
				dir = settings.DownloadDir
			}
		}
		m.Search.Input.SetValue("/downloaddir ")
		m.Search.Input.CursorEnd()
		m.Search.ShowSlash = false
		return m, func() tea.Msg {
			return types.ShowToastMsg{Message: "Current: " + dir + " (enter an existing directory path)"}
		}

	}

	m.Search.Input.SetValue("")
	m.Search.ShowSlash = false
	return m, nil
}

func (m *Model) persistSettings() {
	if m.Store == nil {
		return
	}
	settings, _ := m.Store.GetSettings()
	settings.EmbedSubs = m.Search.EmbedSubs
	settings.EmbedMetadata = m.Search.EmbedMetadata
	settings.EmbedChapters = m.Search.EmbedChapters
	m.Store.SaveSettings(settings)
}

// scheduleDownloadTick returns a command that sends a DownloadTickMsg after a short delay.
func (m Model) scheduleDownloadTick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return types.DownloadTickMsg{}
	})
}
