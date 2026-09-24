package models

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/mohsinkaleem/ytui-go/internal/types"
)

func TestNewSearchModelDefaults(t *testing.T) {
	m := NewSearchModel()

	if m.SortBy() != types.SortRelevance {
		t.Errorf("default SortBy = %q, want %q", m.SortBy(), types.SortRelevance)
	}
	if m.SortIndex != 0 {
		t.Errorf("default SortIndex = %d, want 0", m.SortIndex)
	}
	if m.HistoryIndex != -1 {
		t.Errorf("default HistoryIndex = %d, want -1", m.HistoryIndex)
	}
	if m.EmbedSubs || m.EmbedMetadata || m.EmbedChapters {
		t.Error("embed options should default to false")
	}
}

func TestSearchModelCycleSort(t *testing.T) {
	m := NewSearchModel()
	m.CycleSort(-1)
	if m.SortBy() != types.SortRating {
		t.Errorf("CycleSort(-1) = %q, want %q", m.SortBy(), types.SortRating)
	}
	m.CycleSort(1)
	if m.SortBy() != types.SortRelevance {
		t.Errorf("CycleSort(1) = %q, want %q", m.SortBy(), types.SortRelevance)
	}
}

func TestSearchModelHistory(t *testing.T) {
	m := NewSearchModel()
	m.PushHistory("a")
	m.PushHistory("b")
	m.PushHistory("a")
	if strings.Join(m.History, ",") != "a,b" {
		t.Fatalf("History = %v, want [a b]", m.History)
	}

	m.HistoryPrev()
	m.HistoryPrev()
	m.HistoryPrev() // stays on the oldest entry
	if m.Input.Value() != "b" {
		t.Errorf("after HistoryPrev input = %q, want b", m.Input.Value())
	}
	m.HistoryNext()
	m.HistoryNext()
	if m.Input.Value() != "" || m.HistoryIndex != -1 {
		t.Errorf("after HistoryNext past newest input = %q, index = %d", m.Input.Value(), m.HistoryIndex)
	}
}

func TestSearchModelSetSize(t *testing.T) {
	m := NewSearchModel()
	m.SetSize(80, 24)

	if m.Width != 80 {
		t.Errorf("Width = %d, want 80", m.Width)
	}
	if m.Height != 24 {
		t.Errorf("Height = %d, want 24", m.Height)
	}
	if m.Input.Width <= 0 {
		t.Error("Input.Width should be > 0 after SetSize")
	}
}

func TestSearchModelViewNonEmpty(t *testing.T) {
	m := NewSearchModel()
	m.SetSize(80, 24)
	view := m.View()

	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestSearchModelViewShowsError(t *testing.T) {
	m := NewSearchModel()
	m.SetSize(80, 24)
	m.ErrMsg = "something broke"
	view := m.View()

	if !strings.Contains(view, "something broke") {
		t.Error("View should show error message")
	}
}

func TestNewVideoListModel(t *testing.T) {
	m := NewVideoListModel()
	if m.List.FilteringEnabled() != true {
		t.Error("filtering should be enabled")
	}
}

func TestVideoListSetItems(t *testing.T) {
	m := NewVideoListModel()
	m.SetSize(80, 24)

	m.SetItems([]list.Item{types.VideoItem{ID: "1", Title: "Video A"}}, "Test Playlist", "query")
	if m.Title != "Test Playlist" {
		t.Errorf("Title = %q, want %q", m.Title, "Test Playlist")
	}
	if m.Query != "query" {
		t.Errorf("Query = %q, want %q", m.Query, "query")
	}
	if len(m.List.Items()) != 1 {
		t.Errorf("items = %d, want 1", len(m.List.Items()))
	}
}

func TestVideoListSelection(t *testing.T) {
	m := NewVideoListModel()
	m.SetSize(80, 24)
	m.SetItems([]list.Item{
		types.VideoItem{ID: "1", Title: "A"},
		types.VideoItem{ID: "2", Title: "B"},
	}, "", "q")

	m.ToggleSelected()
	if got := m.GetSelectedVideos(); len(got) != 1 || got[0].ID != "1" {
		t.Fatalf("after toggle selected = %v, want [1]", got)
	}
	m.SelectAll()
	if m.SelectedCount() != 2 {
		t.Errorf("after SelectAll count = %d, want 2", m.SelectedCount())
	}
	m.SelectAll() // all selected -> clears
	if m.SelectedCount() != 0 {
		t.Errorf("second SelectAll count = %d, want 0", m.SelectedCount())
	}
	if !strings.Contains(m.View(), "2 videos") {
		t.Error("header should show the video count")
	}
}

func TestVideoListQuitKeyDisabled(t *testing.T) {
	m := NewVideoListModel()
	m.SetItems([]list.Item{types.VideoItem{ID: "1", Title: "A"}}, "", "q") // SetItems re-evaluates key bindings
	if m.List.KeyMap.Quit.Enabled() {
		t.Error("the list's own quit binding (q) must be disabled")
	}
}

func TestVideoListViewHeader(t *testing.T) {
	m := NewVideoListModel()
	m.SetSize(80, 24)
	m.Query = "test query"
	view := m.View()

	if !strings.Contains(view, "test query") {
		t.Error("View should contain search query in header")
	}
}

func TestVideoListSelectedVideoEmpty(t *testing.T) {
	m := NewVideoListModel()
	m.SetSize(80, 24)

	_, ok := m.SelectedVideo()
	if ok {
		t.Error("should return false when no items")
	}
}

func TestNewFormatListModel(t *testing.T) {
	m := NewFormatListModel()
	if m.ActiveTab != types.FormatTabVideo {
		t.Errorf("default tab = %d, want %d", m.ActiveTab, types.FormatTabVideo)
	}
}

func TestFormatListTabCycling(t *testing.T) {
	m := NewFormatListModel()

	m.NextTab()
	if m.ActiveTab != types.FormatTabAudio {
		t.Errorf("after NextTab: tab = %d, want %d", m.ActiveTab, types.FormatTabAudio)
	}

	m.NextTab()
	if m.ActiveTab != types.FormatTabCustom {
		t.Errorf("after NextTab: tab = %d, want %d", m.ActiveTab, types.FormatTabCustom)
	}

	m.NextTab()
	if m.ActiveTab != types.FormatTabVideo {
		t.Errorf("after wraparound NextTab: tab = %d, want %d", m.ActiveTab, types.FormatTabVideo)
	}

	m.PrevTab()
	if m.ActiveTab != types.FormatTabCustom {
		t.Errorf("after PrevTab wraparound: tab = %d, want %d", m.ActiveTab, types.FormatTabCustom)
	}
}

func TestFormatListGetFormatID(t *testing.T) {
	m := NewFormatListModel()

	// Nothing selected
	if id := m.GetFormatID(); id != "" {
		t.Errorf("empty format list GetFormatID() = %q, want empty", id)
	}

	// Custom tab with input
	m.ActiveTab = types.FormatTabCustom
	m.CustomInput.SetValue("140+137")
	if id := m.GetFormatID(); id != "140+137" {
		t.Errorf("custom tab GetFormatID() = %q, want %q", id, "140+137")
	}
}

func TestFormatListTabsAndMerge(t *testing.T) {
	m := NewFormatListModel()
	m.SetSize(100, 30)
	m.SetFormats([]types.FormatItem{
		{FormatID: "137", Height: 1080, VCodec: "avc1.640028", ACodec: "none"},
		{FormatID: "18", Height: 360, VCodec: "avc1", ACodec: "mp4a.40.2"},
		{FormatID: "140", ACodec: "mp4a.40.2", VCodec: "none", ABR: 128},
	}, types.VideoItem{Title: "T"})

	if id := m.GetFormatID(); id != "137+bestaudio" {
		t.Errorf("video-only selection = %q, want 137+bestaudio", id)
	}
	m.List.Select(1)
	m.NextTab() // audio tab has one item; cursor must reset to it
	if id := m.GetFormatID(); id != "140" {
		t.Errorf("audio tab selection = %q, want 140", id)
	}
	if view := m.View(); !strings.Contains(view, "ID") || !strings.Contains(view, "mp4a") {
		t.Errorf("view should show table header and short codec names:\n%s", view)
	}
}

func TestFormatListAudioOnlySource(t *testing.T) {
	m := NewFormatListModel()
	m.SetFormats([]types.FormatItem{{FormatID: "mp3", ACodec: "mp3", VCodec: "none"}}, types.VideoItem{})
	if m.ActiveTab != types.FormatTabAudio {
		t.Errorf("ActiveTab = %d, want audio when there are no video formats", m.ActiveTab)
	}
}

func TestNewPlayerModel(t *testing.T) {
	m := NewPlayerModel()
	if m.Video.ID != "" {
		t.Error("new player should have empty video")
	}
}

func TestPlayerModelSetVideo(t *testing.T) {
	m := NewPlayerModel()
	v := types.VideoItem{ID: "abc", Title: "Test", Channel: "Chan"}
	m.SetVideo(v, "137")

	if m.Video.ID != "abc" {
		t.Errorf("Video.ID = %q, want %q", m.Video.ID, "abc")
	}
	if m.FormatID != "137" {
		t.Errorf("FormatID = %q, want %q", m.FormatID, "137")
	}
}

func TestPlayerModelView(t *testing.T) {
	m := NewPlayerModel()
	m.SetSize(80, 24)
	m.SetVideo(types.VideoItem{Title: "My Video", Channel: "My Channel"}, "")

	view := m.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestNewDownloadModel(t *testing.T) {
	m := NewDownloadModel()
	if m.Tasks != nil {
		t.Error("new download model should have nil tasks")
	}
}

func TestDownloadModelAllDoneEmpty(t *testing.T) {
	m := NewDownloadModel()
	// No tasks should trivially be "all done"
	if !m.AllDone() {
		t.Error("empty task list should count as all done")
	}
}

func TestDownloadModelCurrentTaskNil(t *testing.T) {
	m := NewDownloadModel()
	if m.CurrentTask() != nil {
		t.Error("should return nil with no tasks")
	}
}
