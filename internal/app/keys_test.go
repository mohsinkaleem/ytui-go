package app

import (
	"testing"

	"github.com/mohsinkaleem/ytui-go/internal/types"
)

func TestRenderHints(t *testing.T) {
	if got := renderHints(80, "enter", "search", "/", "commands"); got != "enter search • / commands" {
		t.Errorf("renderHints = %q", got)
	}
	if got := renderHints(80, "", "mpv is running"); got != "mpv is running" {
		t.Errorf("renderHints without key = %q", got)
	}
}

func TestRenderHintsDropsWhatDoesNotFit(t *testing.T) {
	if got := renderHints(14, "enter", "search", "/", "commands"); got != "enter search" {
		t.Errorf("renderHints(14) = %q, want only the first hint", got)
	}
	if got := renderHints(80); got != "" {
		t.Errorf("renderHints() with no pairs = %q, want empty", got)
	}
}

func TestKeyHintsForEveryState(t *testing.T) {
	m := newTestModel(t, 80, 24)
	states := []types.State{
		types.StateSearchInput, types.StateLoading, types.StateVideoList, types.StateFormatList,
		types.StateDownload, types.StateVideoPlaying, types.StateResumeList,
	}
	for _, s := range states {
		m.State = s
		hints := m.keyHints()
		if len(hints) == 0 || len(hints)%2 != 0 {
			t.Errorf("keyHints(%s) = %v, want non-empty key/description pairs", s, hints)
		}
	}
}
