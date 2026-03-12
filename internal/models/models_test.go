package models

import (
	"strings"
	"testing"

	"github.com/mohsinkaleem/ytui-go/internal/slash"
	"github.com/mohsinkaleem/ytui-go/internal/types"
)

func TestNewSearchModelDefaults(t *testing.T) {
	r := slash.NewRegistry()
	m := NewSearchModel(r)

	if m.SortBy != types.SortRelevance {
		t.Errorf("default SortBy = %q, want %q", m.SortBy, types.SortRelevance)
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

func TestSearchModelSetSize(t *testing.T) {
	r := slash.NewRegistry()
	m := NewSearchModel(r)
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
	r := slash.NewRegistry()
	m := NewSearchModel(r)
	m.SetSize(80, 24)
	view := m.View()

	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestSearchModelViewShowsError(t *testing.T) {
	r := slash.NewRegistry()
	m := NewSearchModel(r)
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

	items := []types.VideoItem{
		{ID: "1", Title: "Video A"},
		{ID: "2", Title: "Video B"},
	}
	listItems := make([]interface{ FilterValue() string }, len(items))
	for i, v := range items {
		listItems[i] = v
	}
	// Use the actual list.Item interface
	m.SetItems(nil, "Test Playlist", "query")
	if m.Title != "Test Playlist" {
		t.Errorf("Title = %q, want %q", m.Title, "Test Playlist")
	}
	if m.Query != "query" {
		t.Errorf("Query = %q, want %q", m.Query, "query")
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

	// No items selected, should return "best"
	id := m.GetFormatID()
	if id != "best" {
		t.Errorf("empty format list GetFormatID() = %q, want %q", id, "best")
	}

	// Custom tab with input
	m.ActiveTab = types.FormatTabCustom
	m.CustomInput.SetValue("140+137")
	id = m.GetFormatID()
	if id != "140+137" {
		t.Errorf("custom tab GetFormatID() = %q, want %q", id, "140+137")
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
