package app

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mohsinkaleem/ytui-go/internal/download"
	"github.com/mohsinkaleem/ytui-go/internal/store"
	"github.com/mohsinkaleem/ytui-go/internal/types"
	"github.com/mohsinkaleem/ytui-go/internal/ytdlp"
)

func newTestModel(t *testing.T, width, height int) Model {
	t.Helper()
	mgr := download.NewManager(1, nil, nil)
	t.Cleanup(func() { mgr.Shutdown() })
	m := New(Config{Downloads: mgr, Settings: store.DefaultSettings()})
	return update(m, tea.WindowSizeMsg{Width: width, Height: height})
}

func update(m Model, msg tea.Msg) Model {
	next, _ := m.Update(msg)
	return next.(Model)
}

// press sends keys; single runes are typed, anything else is a named key.
func press(m Model, keys ...string) Model {
	named := map[string]tea.KeyType{
		"enter": tea.KeyEnter, "esc": tea.KeyEsc, "tab": tea.KeyTab, "up": tea.KeyUp,
		"down": tea.KeyDown, "space": tea.KeySpace, "ctrl+t": tea.KeyCtrlT,
	}
	for _, k := range keys {
		if kt, ok := named[k]; ok {
			m = update(m, tea.KeyMsg{Type: kt})
			continue
		}
		for _, r := range k {
			m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		}
	}
	return m
}

func sampleVideos(n int) []list.Item {
	items := make([]list.Item, n)
	for i := range items {
		items[i] = types.VideoItem{
			ID:             fmt.Sprintf("id%d", i),
			Title:          fmt.Sprintf("A fairly long video title number %d that keeps going and going", i),
			Channel:        "Some Channel",
			DurationString: "12:34",
			ViewCount:      1_234_567,
			IsLive:         i == 1,
		}
	}
	return items
}

func sampleFormats() []types.FormatItem {
	return []types.FormatItem{
		{FormatID: "137", Ext: "mp4", Width: 1920, Height: 1080, FPS: 30, VCodec: "avc1.640028", ACodec: "none", Filesize: 150 << 20},
		{FormatID: "22", Ext: "mp4", Width: 1280, Height: 720, FPS: 30, VCodec: "avc1.64001F", ACodec: "mp4a.40.2", Filesize: 80 << 20},
		{FormatID: "251", Ext: "webm", VCodec: "none", ACodec: "opus", ABR: 160, Filesize: 5 << 20},
	}
}

// screens drives the model into every screen with sample data.
func screens(t *testing.T, width, height int) map[string]Model {
	base := newTestModel(t, width, height)
	out := map[string]Model{"search": base}

	out["commands"] = press(base, "/")
	out["search error"] = update(func() Model { m, _ := base.load("x", nil); return m }(), types.SearchResultMsg{
		Err: fmt.Errorf("yt-dlp: [youtube] abc: Sign in to confirm your age. This video may be inappropriate for some users."),
	})

	loading, _ := base.load("Searching for \"a query that is rather long to fit on a narrow screen\"…", nil)
	out["loading"] = loading

	list := update(loading, types.SearchResultMsg{Videos: sampleVideos(30), Query: "lofi"})
	out["videos"] = press(list, "space")

	formatsLoading, _ := list.load("Loading formats…", nil)
	formats := update(formatsLoading, types.FormatResultMsg{Formats: sampleFormats(), Video: types.VideoItem{Title: "Title", Channel: "Chan"}})
	out["formats"] = formats
	out["formats custom"] = press(formats, "tab", "tab")

	dl := base
	var tasks []*download.Task
	for i := range 20 {
		tasks = append(tasks, download.NewTask(types.VideoItem{ID: fmt.Sprint(i), Title: fmt.Sprintf("Queued video %d", i)}, "bestvideo+bestaudio/best", ytdlp.DownloadOpts{}, nil))
	}
	tasks[3].Cancel()
	dl.Download.SetTasks(tasks)
	dl.State = types.StateDownload
	out["downloads"] = dl

	resume := base
	var records []store.DownloadRecord
	for i := range 15 {
		records = append(records, store.DownloadRecord{ID: fmt.Sprint(i), Title: "Unfinished video", State: "paused", FormatID: "22", OutputPath: "/Users/someone/Downloads/Unfinished video [abc].mp4"})
	}
	resume.ResumeList.SetItems(records)
	resume.State = types.StateResumeList
	out["resume"] = resume

	playing := base
	playing.Player.SetVideo(types.VideoItem{Title: "Now playing", Channel: "Chan"}, "")
	playing.State = types.StateVideoPlaying
	out["playing"] = playing
	return out
}

func TestViewFitsTerminal(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {50, 14}, {140, 45}} {
		w, h := size[0], size[1]
		for name, m := range screens(t, w, h) {
			view := m.View()
			if got := lipgloss.Height(view); got != h {
				t.Errorf("%dx%d %s: height %d, want %d", w, h, name, got, h)
			}
			for i, line := range strings.Split(view, "\n") {
				if lw := lipgloss.Width(line); lw > w {
					t.Errorf("%dx%d %s: line %d is %d wide: %q", w, h, name, i, lw, line)
				}
			}
			if testing.Verbose() && w == 80 {
				t.Logf("── %s ──\n%s", name, view)
			}
		}
	}
}

func TestStaleResultIgnoredAfterCancel(t *testing.T) {
	m := newTestModel(t, 80, 24)
	m, _ = m.load("Searching…", nil)
	m = press(m, "esc")
	m = update(m, types.SearchResultMsg{Videos: sampleVideos(3), Query: "late"})
	if m.State != types.StateSearchInput {
		t.Errorf("state = %s, want search input (late results must be ignored)", m.State)
	}
}

func TestFormatErrorReturnsToVideoList(t *testing.T) {
	m := newTestModel(t, 80, 24)
	m, _ = m.load("x", nil)
	m = update(m, types.SearchResultMsg{Videos: sampleVideos(3), Query: "q"})
	m, _ = m.load("Loading formats…", nil)
	m = update(m, types.FormatResultMsg{Err: fmt.Errorf("boom")})
	if m.State != types.StateVideoList || !m.ToastErr || m.Toast != "boom" {
		t.Errorf("state = %s toast = %q (err %v), want video list with error toast", m.State, m.Toast, m.ToastErr)
	}
}

func TestFormatBackReturnsToOrigin(t *testing.T) {
	m := newTestModel(t, 80, 24)
	m, _ = m.load("x", nil)
	m = update(m, types.FormatResultMsg{Formats: sampleFormats()})
	m = press(m, "b")
	if m.State != types.StateSearchInput {
		t.Errorf("back from formats opened from search = %s", m.State)
	}
}

func TestCustomFormatAcceptsLetters(t *testing.T) {
	m := newTestModel(t, 80, 24)
	m, _ = m.load("x", nil)
	m = update(m, types.FormatResultMsg{Formats: sampleFormats()})
	m = press(m, "tab", "tab", "bestvideo+bestaudio/b")
	if m.State != types.StateFormatList {
		t.Fatalf("typing left the format screen: %s", m.State)
	}
	if got := m.FormatList.GetFormatID(); got != "bestvideo+bestaudio/b" {
		t.Errorf("custom format = %q", got)
	}
}

func TestQuitKeyDoesNotLeaveVideoList(t *testing.T) {
	m := newTestModel(t, 80, 24)
	m, _ = m.load("x", nil)
	m = update(m, types.SearchResultMsg{Videos: sampleVideos(3), Query: "q"})
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		if _, quit := cmd().(tea.QuitMsg); quit {
			t.Fatal("q in the video list quit the app")
		}
	}
	if next.(Model).State != types.StateVideoList {
		t.Errorf("state = %s", next.(Model).State)
	}
}

func TestSlashCommands(t *testing.T) {
	m := newTestModel(t, 80, 24)

	m = press(m, "/")
	if !m.Search.ShowSlash || len(m.Search.SlashMatches) == 0 {
		t.Fatal("typing / should open the command list")
	}
	m = press(m, "esc")
	if m.Search.ShowSlash || m.Search.Input.Value() != "" {
		t.Error("esc should close the command list and clear the input")
	}
	if m = press(m, "lofi"); m.Search.Input.Value() != "lofi" {
		t.Errorf("input after esc = %q, want lofi", m.Search.Input.Value())
	}

	m = press(m, "esc", "/bogus", "enter")
	if !strings.Contains(m.Search.ErrMsg, "Unknown command") {
		t.Errorf("ErrMsg = %q, want unknown command error", m.Search.ErrMsg)
	}
	m = press(m, "x")
	if m.Search.ErrMsg != "" {
		t.Error("typing should clear the error")
	}

	m = press(m, "esc", "/theme", "enter")
	if !m.Search.ShowThemes {
		t.Error("/theme should open the theme picker")
	}
	picked := m.Search.ThemeNames[(m.Search.ThemeSelected+1)%len(m.Search.ThemeNames)]
	m = press(m, "down", "enter")
	if m.Search.ShowThemes || m.Settings.Theme != picked {
		t.Errorf("picker open = %v, theme = %q; want closed with %q", m.Search.ShowThemes, m.Settings.Theme, picked)
	}
	m = press(m, "/download ")
	if m.Search.SlashHint == "" {
		t.Error("typing arguments should show the usage hint")
	}
}

func TestToggleMetadataWithCtrlT(t *testing.T) {
	m := newTestModel(t, 80, 24)
	m = press(m, "ctrl+t")
	if !m.Settings.EmbedMetadata || !m.Search.EmbedMetadata {
		t.Error("ctrl+t should toggle embed metadata")
	}
}

func TestSearchRecordsHistory(t *testing.T) {
	m := newTestModel(t, 80, 24)
	m = press(m, "lofi", "enter")
	if m.State != types.StateLoading || len(m.Search.History) != 1 || m.Search.History[0] != "lofi" {
		t.Errorf("state = %s history = %v", m.State, m.Search.History)
	}
}
