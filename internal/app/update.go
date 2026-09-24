package app

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mohsinkaleem/ytui-go/internal/download"
	"github.com/mohsinkaleem/ytui-go/internal/models"
	"github.com/mohsinkaleem/ytui-go/internal/slash"
	"github.com/mohsinkaleem/ytui-go/internal/store"
	"github.com/mohsinkaleem/ytui-go/internal/styles"
	"github.com/mohsinkaleem/ytui-go/internal/types"
	"github.com/mohsinkaleem/ytui-go/internal/utils"
	"github.com/mohsinkaleem/ytui-go/internal/ytdlp"
)

// bestFormat is used for quick downloads.
const bestFormat = "bestvideo+bestaudio/best"

// browsers accepted by /cookies (yt-dlp --cookies-from-browser).
var browsers = []string{"brave", "chrome", "chromium", "edge", "firefox", "opera", "safari", "vivaldi", "whale"}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Width, m.Height = msg.Width, msg.Height
		w, h := m.contentSize()
		m.Search.SetSize(w, h)
		m.VideoList.SetSize(w, h)
		m.FormatList.SetSize(w, h)
		m.Download.SetSize(w, h)
		m.Player.SetSize(w, h)
		m.ResumeList.SetSize(w, h)
		return m, nil

	case tea.KeyMsg:
		next, cmd := m.handleKey(msg)
		tick := next.ensureTick()
		return next, tea.Batch(cmd, tick)

	case spinner.TickMsg:
		if m.State != types.StateLoading {
			return m, nil
		}
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd

	// Results are ignored once the user has cancelled the load.
	case types.SearchResultMsg:
		if m.State != types.StateLoading {
			return m, nil
		}
		m.State = m.PrevState
		switch {
		case msg.Err != nil:
			return m.withError(msg.Err.Error())
		case len(msg.Videos) == 0:
			return m.withError("No results for " + strconv.Quote(msg.Query))
		}
		m.VideoList.SetItems(msg.Videos, "", msg.Query)
		m.State = types.StateVideoList
		return m, nil

	case types.PlaylistResultMsg:
		if m.State != types.StateLoading {
			return m, nil
		}
		m.State = m.PrevState
		switch {
		case msg.Err != nil:
			return m.withError(msg.Err.Error())
		case len(msg.Videos) == 0:
			return m.withError("The playlist is empty")
		}
		m.VideoList.SetItems(msg.Videos, msg.Title, "")
		m.State = types.StateVideoList
		return m, nil

	case types.FormatResultMsg:
		if m.State != types.StateLoading {
			return m, nil
		}
		m.State = m.PrevState
		if msg.Err != nil {
			return m.withError(msg.Err.Error())
		}
		m.FormatBack = m.PrevState
		m.FormatList.SetFormats(msg.Formats, msg.Video)
		m.State = types.StateFormatList
		return m, nil

	case types.DownloadTickMsg:
		m.ticking = false
		if m.State == types.StateDownload {
			m.Download.SetTasks(m.DownloadMgr.GetTasks())
		}
		tick := m.ensureTick()
		return m, tick

	case types.MPVExitedMsg:
		m.State = m.PrevState
		if msg.Err != nil {
			return m.withError("mpv: " + msg.Err.Error())
		}
		return m, nil

	case types.ClearToastMsg:
		if msg.Seq == m.toastSeq {
			m.Toast = ""
		}
		return m, nil

	case list.FilterMatchesMsg: // results of the video list's async filtering
		var cmd tea.Cmd
		m.VideoList.List, cmd = m.VideoList.List.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "g":
		if m.canJumpToDownloads() {
			m.openDownloads("")
			return m, nil
		}
	}

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
	case types.StateResumeList:
		return m.updateResumeList(msg)
	}
	return m, nil
}

// --- State-specific update handlers ---

func (m Model) updateSearchInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	s := &m.Search
	key := msg.String()

	if s.ShowThemes {
		n := len(s.ThemeNames)
		switch key {
		case "up":
			s.ThemeSelected = (s.ThemeSelected + n - 1) % n
			return m, nil
		case "down":
			s.ThemeSelected = (s.ThemeSelected + 1) % n
			return m, nil
		case "enter":
			return m.runCommand("/theme " + s.ThemeNames[s.ThemeSelected])
		case "esc":
			s.CloseThemes()
			return m, nil
		}
		s.CloseThemes() // any other key dismisses the picker and is handled below
	}

	if s.ShowSlash && len(s.SlashMatches) > 0 {
		n := len(s.SlashMatches)
		switch key {
		case "up":
			s.SlashSelected = (s.SlashSelected + n - 1) % n
			return m, nil
		case "down":
			s.SlashSelected = (s.SlashSelected + 1) % n
			return m, nil
		case "enter", "tab":
			c := s.SlashMatches[s.SlashSelected]
			if key == "enter" && c.Args == "" {
				return m.runCommand("/" + c.Name)
			}
			s.SetInput("/" + c.Name + " ")
			m.refreshSlash()
			return m, nil
		}
	}

	switch key {
	case "enter":
		return m.submit()
	case "esc":
		s.ClearInput()
		s.ErrMsg = ""
		return m, nil
	case "tab":
		s.CycleSort(1)
		return m, nil
	case "shift+tab":
		s.CycleSort(-1)
		return m, nil
	case "ctrl+s":
		m.Settings.EmbedSubs = !m.Settings.EmbedSubs
		m.saveSettings()
		return m, nil
	case "ctrl+t": // not ctrl+m: terminals send that as enter
		m.Settings.EmbedMetadata = !m.Settings.EmbedMetadata
		m.saveSettings()
		return m, nil
	case "ctrl+j":
		m.Settings.EmbedChapters = !m.Settings.EmbedChapters
		m.saveSettings()
		return m, nil
	case "up":
		s.HistoryPrev()
		s.CloseSlash()
		return m, nil
	case "down":
		s.HistoryNext()
		s.CloseSlash()
		return m, nil
	}

	var cmd tea.Cmd
	s.Input, cmd = s.Input.Update(msg)
	s.ErrMsg = ""
	s.HistoryIndex = -1
	m.refreshSlash()
	return m, cmd
}

// refreshSlash updates the command dropdown, or the usage hint once arguments are being typed.
func (m *Model) refreshSlash() {
	s := &m.Search
	s.CloseSlash()
	val := s.Input.Value()
	if !strings.HasPrefix(val, "/") {
		return
	}
	name, _, hasArgs := strings.Cut(val[1:], " ")
	if !hasArgs {
		s.ShowSlash = true
		s.SlashMatches = m.SlashRegistry.Match(name)
		return
	}
	if c, ok := m.SlashRegistry.Get(name); ok && c.Args != "" {
		s.SlashHint = models.CommandUsage(c) + " — " + c.Description
	}
}

// submit runs the typed input: a slash command, a URL or a search.
func (m Model) submit() (Model, tea.Cmd) {
	input := strings.TrimSpace(m.Search.Input.Value())
	switch {
	case input == "":
		return m, nil
	case strings.HasPrefix(input, "/"):
		return m.runCommand(input)
	}

	m.remember(input)
	m.Search.CloseSlash()
	if ytdlp.IsURL(input) {
		if ytdlp.IsPlaylistURL(input) {
			return m.load("Loading playlist…", m.Fetcher.Playlist(input))
		}
		return m.load("Loading formats…", m.Fetcher.Formats(input))
	}
	return m.load("Searching for "+strconv.Quote(input)+"…", m.Fetcher.Search(input, m.Search.SortBy()))
}

// load shows the spinner while cmd fetches data; cancelling returns to the current screen.
func (m Model) load(label string, cmd tea.Cmd) (Model, tea.Cmd) {
	m.PrevState = m.State
	m.State = types.StateLoading
	m.LoadingMsg = label
	m.Search.ErrMsg = ""
	return m, tea.Batch(m.Spinner.Tick, cmd)
}

func (m Model) updateLoading(msg tea.KeyMsg) (Model, tea.Cmd) {
	if key := msg.String(); key == "esc" || key == "c" {
		m.Fetcher.Cancel()
		m.State = m.PrevState
	}
	return m, nil
}

func (m Model) updateVideoList(msg tea.KeyMsg) (Model, tea.Cmd) {
	vl := &m.VideoList
	var cmd tea.Cmd

	// While typing a filter every key belongs to the list
	if vl.IsFiltering() {
		vl.List, cmd = vl.List.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "enter":
		v, ok := vl.SelectedVideo()
		if !ok {
			return m, nil
		}
		return m.load("Loading formats for "+v.Title+"…", m.Fetcher.Formats(v.VideoURL()))

	case " ":
		cmd = vl.ToggleSelected()
		return m, cmd

	case "a":
		cmd = vl.SelectAll()
		return m, cmd

	case "d":
		videos := vl.GetSelectedVideos()
		if len(videos) == 0 {
			if v, ok := vl.SelectedVideo(); ok {
				videos = []types.VideoItem{v}
			}
		}
		if len(videos) == 0 {
			return m, nil
		}
		cmd = vl.ClearSelections()
		m.enqueue(videos, bestFormat)
		return m, cmd

	case "p":
		v, ok := vl.SelectedVideo()
		if !ok {
			return m, nil
		}
		return m.play(v, "")

	case "ctrl+y":
		v, ok := vl.SelectedVideo()
		if !ok {
			return m, nil
		}
		if err := clipboard.WriteAll(v.VideoURL()); err != nil {
			return m.withToast("Clipboard unavailable: "+err.Error(), true)
		}
		return m.withToast("Copied "+v.VideoURL(), false)

	case "esc":
		if vl.List.FilterState() == list.FilterApplied {
			vl.List.ResetFilter()
			return m, nil
		}
		m.State = types.StateSearchInput
		return m, nil

	case "b":
		m.State = types.StateSearchInput
		return m, nil
	}

	vl.List, cmd = vl.List.Update(msg)
	return m, cmd
}

func (m Model) updateFormatList(msg tea.KeyMsg) (Model, tea.Cmd) {
	fl := &m.FormatList
	key := msg.String()

	switch key {
	case "tab":
		fl.NextTab()
		return m, nil
	case "shift+tab":
		fl.PrevTab()
		return m, nil
	case "esc":
		m.State = types.StateSearchInput
		return m, nil
	case "enter":
		if id := fl.GetFormatID(); id != "" {
			m.enqueue([]types.VideoItem{fl.Video}, id)
		}
		return m, nil
	}

	var cmd tea.Cmd
	if fl.ActiveTab == types.FormatTabCustom {
		// Letters must reach the input ("bestaudio" starts with b), so only arrows move the preset.
		switch key {
		case "up":
			fl.MoveCombo(-1)
		case "down":
			fl.MoveCombo(1)
		default:
			fl.CustomInput, cmd = fl.CustomInput.Update(msg)
		}
		return m, cmd
	}

	switch key {
	case "p":
		return m.play(fl.Video, fl.GetFormatID())
	case "b":
		m.State = m.FormatBack
		return m, nil
	}
	fl.List, cmd = fl.List.Update(msg)
	return m, cmd
}

func (m Model) updateDownload(msg tea.KeyMsg) (Model, tea.Cmd) {
	task := m.Download.CurrentTask()

	switch msg.String() {
	case "up", "k":
		m.Download.MoveUp()

	case "down", "j":
		m.Download.MoveDown()

	case "p":
		if task != nil {
			switch task.GetState() {
			case download.StateDownloading:
				m.DownloadMgr.PauseTask(task.ID)
			case download.StatePaused:
				m.DownloadMgr.ResumeTask(task.ID)
			}
		}

	case "r":
		if task != nil && task.GetState() == download.StateFailed {
			m.DownloadMgr.ResumeTask(task.ID)
		}

	case "R":
		retried := 0
		for _, t := range m.Download.Tasks {
			if t.GetState() == download.StateFailed {
				m.DownloadMgr.ResumeTask(t.ID)
				retried++
			}
		}
		if retried > 0 {
			return m.withToast(fmt.Sprintf("Retrying %d failed downloads", retried), false)
		}

	case "c", "s": // cancel / skip
		if task != nil {
			m.DownloadMgr.CancelTask(task.ID)
		}

	case "esc":
		m.State = types.StateSearchInput

	case "b":
		m.State = m.PrevState
		if m.State == types.StateResumeList && len(m.ResumeList.Items) == 0 {
			m.State = types.StateSearchInput
		}
	}

	return m, nil
}

func (m Model) updateResumeList(msg tea.KeyMsg) (Model, tea.Cmd) {
	rl := &m.ResumeList

	switch msg.String() {
	case "up", "k":
		rl.MoveUp()

	case "down", "j":
		rl.MoveDown()

	case "enter":
		if rec := rl.SelectedItem(); rec != nil {
			task := m.resumeRecord(*rec)
			rl.RemoveSelected()
			m.openDownloads(task.ID)
		}

	case "d":
		if len(rl.Items) > 0 {
			first := m.resumeRecord(rl.Items[0])
			for _, rec := range rl.Items[1:] {
				m.resumeRecord(rec)
			}
			rl.SetItems(nil)
			m.openDownloads(first.ID)
		}

	case "x":
		if rec := rl.SelectedItem(); rec != nil {
			if m.Store != nil {
				m.Store.DeleteDownload(rec.ID)
			}
			rl.RemoveSelected()
			if len(rl.Items) == 0 {
				m.State = types.StateSearchInput
			}
		}

	case "esc", "b":
		m.State = types.StateSearchInput
	}

	return m, nil
}

// --- Helpers ---

// withToast shows a transient message in the status bar.
func (m Model) withToast(text string, isErr bool) (Model, tea.Cmd) {
	m.toastSeq++
	m.Toast, m.ToastErr = text, isErr
	return m, clearToastAfter(m.toastSeq)
}

// withError shows text inline on the search screen, or as an error toast elsewhere.
func (m Model) withError(text string) (Model, tea.Cmd) {
	if m.State == types.StateSearchInput {
		m.Search.ErrMsg = text
		return m, nil
	}
	return m.withToast(text, true)
}

func clearToastAfter(seq int) tea.Cmd {
	return tea.Tick(4*time.Second, func(time.Time) tea.Msg {
		return types.ClearToastMsg{Seq: seq}
	})
}

// ensureTick keeps the screen refreshing while downloads are queued or running.
func (m *Model) ensureTick() tea.Cmd {
	if m.ticking || !m.downloadsBusy() {
		return nil
	}
	m.ticking = true
	return tea.Tick(250*time.Millisecond, func(time.Time) tea.Msg {
		return types.DownloadTickMsg{}
	})
}

func (m *Model) downloadsBusy() bool {
	for _, t := range m.DownloadMgr.GetTasks() {
		if s := t.GetState(); s == download.StateQueued || s == download.StateDownloading {
			return true
		}
	}
	return false
}

// canJumpToDownloads reports whether "g" opens the downloads screen; it must
// still type into text fields.
func (m Model) canJumpToDownloads() bool {
	switch m.State {
	case types.StateVideoList:
		if m.VideoList.IsFiltering() {
			return false
		}
	case types.StateFormatList:
		if m.FormatList.ActiveTab == types.FormatTabCustom {
			return false
		}
	case types.StateResumeList:
	default:
		return false
	}
	return len(m.DownloadMgr.GetTasks()) > 0
}

// openDownloads shows the download screen, focusing the task with focusID if given.
func (m *Model) openDownloads(focusID string) {
	m.Download.SetTasks(m.DownloadMgr.GetTasks())
	m.Download.Focus(focusID)
	if m.State != types.StateDownload {
		m.PrevState = m.State
	}
	m.State = types.StateDownload
}

// enqueue queues videos for download and shows the download screen.
func (m *Model) enqueue(videos []types.VideoItem, formatID string) {
	tasks := m.DownloadMgr.EnqueueBatch(videos, formatID, m.downloadOpts())
	if len(tasks) > 0 {
		m.openDownloads(tasks[0].ID)
	}
}

// play streams v in mpv and returns to the current screen when it exits.
func (m Model) play(v types.VideoItem, formatID string) (Model, tea.Cmd) {
	if !m.PlayerManager.IsAvailable() {
		return m.withError("mpv not found — install it to stream videos (https://mpv.io)")
	}
	m.PrevState = m.State
	m.State = types.StateVideoPlaying
	m.Player.SetVideo(v, formatID)
	return m, m.PlayerManager.Play(v.VideoURL(), formatID)
}

// resumable returns stored unfinished downloads that aren't already in this session's queue.
func (m Model) resumable() []store.DownloadRecord {
	if m.Store == nil {
		return nil
	}
	records, _ := m.Store.GetIncomplete()
	return slices.DeleteFunc(records, func(r store.DownloadRecord) bool {
		return m.DownloadMgr.GetTask(r.ID) != nil
	})
}

// resumeRecord re-queues a stored download under its original ID and folder.
func (m *Model) resumeRecord(rec store.DownloadRecord) *download.Task {
	opts := m.downloadOpts()
	opts.EmbedSubs, opts.EmbedMetadata, opts.EmbedChapters = rec.EmbedSubs, rec.EmbedMeta, rec.EmbedChaps
	if rec.OutputPath != "" {
		opts.OutputDir = filepath.Dir(rec.OutputPath) // so yt-dlp finds the .part file
	}
	video := types.VideoItem{ID: rec.VideoID, Title: rec.Title, URL: rec.URL}
	task := download.NewTask(video, rec.FormatID, opts, m.YtdlpClient)
	task.ID = rec.ID
	task.CreatedAt = rec.CreatedAt
	m.DownloadMgr.Enqueue(task)
	return task
}

// remember adds input to the in-memory and persisted history.
func (m *Model) remember(input string) {
	m.Search.PushHistory(input)
	if m.Store == nil {
		return
	}
	kind := "search"
	switch {
	case strings.HasPrefix(input, "/"):
		kind = "command"
	case ytdlp.IsURL(input):
		kind = "url"
	}
	m.Store.SaveSearch(store.SearchEntry{URL: input, Title: input, Timestamp: time.Now(), Type: kind})
}

// runCommand executes a slash command such as "/theme latte".
func (m Model) runCommand(input string) (Model, tea.Cmd) {
	name, args := slash.ParseInput(input)
	s := &m.Search
	if !slices.Contains([]string{"clear", "exit", "quit", "help"}, name) {
		m.remember(strings.TrimSpace(input))
	}
	s.ErrMsg = ""
	s.CloseSlash()
	s.CloseThemes()

	switch name {
	case "exit", "quit":
		return m, tea.Quit

	case "help":
		s.SetInput("/")
		m.refreshSlash()
		return m, nil

	case "download":
		if args == "" {
			return m.withError("Usage: /download <url>")
		}
		s.ClearInput()
		m.enqueue([]types.VideoItem{{URL: args, Title: args}}, bestFormat)
		return m, nil

	case "play":
		if args == "" {
			return m.withError("Usage: /play <url>")
		}
		s.ClearInput()
		return m.play(types.VideoItem{URL: args, Title: args}, "")

	case "playlist":
		if args == "" {
			return m.withError("Usage: /playlist <url>")
		}
		s.ClearInput()
		return m.load("Loading playlist…", m.Fetcher.Playlist(args))

	case "downloads":
		s.ClearInput()
		if len(m.DownloadMgr.GetTasks()) == 0 {
			return m.withToast("No downloads yet", false)
		}
		m.openDownloads("")
		return m, nil

	case "resume":
		s.ClearInput()
		records := m.resumable()
		if len(records) == 0 {
			return m.withToast("No unfinished downloads", false)
		}
		m.ResumeList.SetItems(records)
		m.State = types.StateResumeList
		return m, nil

	case "theme":
		if args == "" {
			s.ClearInput()
			s.OpenThemes(styles.ThemeNames(), styles.CurrentTheme.Name)
			return m, nil
		}
		if !styles.LoadTheme(args) {
			return m.withError("Unknown theme " + strconv.Quote(args) + " (available: " + strings.Join(styles.ThemeNames(), ", ") + ")")
		}
		s.ClearInput()
		m.applyTheme()
		m.Settings.Theme = args
		m.saveSettings()
		return m.withToast("Theme: "+args, false)

	case "downloaddir":
		return m.setDownloadDir(args)

	case "cookies":
		return m.setCookies(args)

	case "clear":
		s.ClearInput()
		s.History = nil
		if m.Store != nil {
			m.Store.ClearHistory()
		}
		return m.withToast("History cleared", false)
	}

	return m.withError("Unknown command /" + name + " — type / to list commands")
}

func (m Model) setDownloadDir(arg string) (Model, tea.Cmd) {
	if arg == "" {
		m.Search.SetInput("/downloaddir ")
		m.refreshSlash()
		return m.withToast("Current folder: "+m.downloadDir(), false)
	}

	dir := utils.ExpandPath(arg)
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		msg := "Not a folder: " + dir
		if closest := utils.ClosestExistingDir(dir); err != nil && closest != "" {
			msg += " (closest existing: " + closest + ")"
		}
		return m.withError(msg)
	}
	if !strings.HasPrefix(arg, "~") {
		if abs, err := filepath.Abs(dir); err == nil {
			arg = abs // relative paths would depend on where ytui is started
		}
	}

	m.Search.ClearInput()
	m.Settings.DownloadDir = arg
	m.dirOverride = ""
	m.saveSettings()
	return m.withToast("Saving downloads to "+dir, false)
}

func (m Model) setCookies(arg string) (Model, tea.Cmd) {
	var browser, file, msg string
	switch path := utils.ExpandPath(arg); {
	case arg == "":
		current := cookiesLabel(m.Settings)
		if current == "" {
			current = "none"
		}
		m.Search.SetInput("/cookies ")
		m.refreshSlash()
		return m.withToast("Cookies: "+current+" — enter a browser, a cookies.txt path, or off", false)
	case arg == "off" || arg == "none" || arg == "clear":
		msg = "Cookies disabled"
	case isFile(path):
		file, msg = path, "Using cookies from "+path
	case isBrowser(arg):
		browser, msg = arg, "Using cookies from "+arg
	default:
		return m.withError("Unknown browser or file " + strconv.Quote(arg) + " (browsers: " + strings.Join(browsers, ", ") + ")")
	}

	m.YtdlpClient.SetCookiesFrom(browser)
	m.YtdlpClient.SetCookiesFile(file)
	m.Settings.CookiesFrom, m.Settings.CookiesFile = browser, file
	m.Search.ClearInput()
	m.saveSettings()
	return m.withToast(msg, false)
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// isBrowser accepts yt-dlp's BROWSER[+KEYRING][:PROFILE] syntax.
func isBrowser(arg string) bool {
	name, _, _ := strings.Cut(arg, ":")
	name, _, _ = strings.Cut(name, "+")
	return slices.Contains(browsers, strings.ToLower(name))
}
