package ytdlp

import (
	"testing"

	"github.com/mohsinkaleem/ytui-go/internal/types"
)

func TestFormatIsVideoOnly(t *testing.T) {
	tests := []struct {
		name   string
		format Format
		want   bool
	}{
		{"video only", Format{VCodec: "h264", ACodec: "none"}, true},
		{"video with empty audio", Format{VCodec: "vp9", ACodec: ""}, true},
		{"both codecs", Format{VCodec: "h264", ACodec: "aac"}, false},
		{"none both", Format{VCodec: "none", ACodec: "none"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.format.IsVideoOnly(); got != tt.want {
				t.Errorf("IsVideoOnly() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatIsAudioOnly(t *testing.T) {
	tests := []struct {
		name   string
		format Format
		want   bool
	}{
		{"audio only", Format{VCodec: "none", ACodec: "opus"}, true},
		{"audio empty video", Format{VCodec: "", ACodec: "aac"}, true},
		{"both", Format{VCodec: "h264", ACodec: "aac"}, false},
		{"video only", Format{VCodec: "h264", ACodec: "none"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.format.IsAudioOnly(); got != tt.want {
				t.Errorf("IsAudioOnly() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatHasBoth(t *testing.T) {
	tests := []struct {
		name   string
		format Format
		want   bool
	}{
		{"both", Format{VCodec: "h264", ACodec: "aac"}, true},
		{"video only", Format{VCodec: "h264", ACodec: "none"}, false},
		{"audio only", Format{VCodec: "none", ACodec: "aac"}, false},
		{"empty", Format{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.format.HasBoth(); got != tt.want {
				t.Errorf("HasBoth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatEffectiveSize(t *testing.T) {
	tests := []struct {
		name   string
		format Format
		want   int64
	}{
		{"prefers filesize", Format{Filesize: 1000, FilesizeApprox: 2000}, 1000},
		{"fallback to approx", Format{FilesizeApprox: 2000}, 2000},
		{"zero", Format{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.format.EffectiveSize(); got != tt.want {
				t.Errorf("EffectiveSize() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSortFormats(t *testing.T) {
	formats := []Format{
		{FormatID: "1", Height: 720, ABR: 0, Filesize: 100},
		{FormatID: "2", Height: 1080, ABR: 0, Filesize: 200},
		{FormatID: "3", Height: 1080, ABR: 0, Filesize: 300},
		{FormatID: "4", Height: 480, ABR: 0, Filesize: 50},
	}

	sorted := SortFormats(formats)

	if sorted[0].FormatID != "3" {
		t.Errorf("first should be format 3 (1080p, biggest), got %s", sorted[0].FormatID)
	}
	if sorted[1].FormatID != "2" {
		t.Errorf("second should be format 2 (1080p, smaller), got %s", sorted[1].FormatID)
	}
	if sorted[2].FormatID != "1" {
		t.Errorf("third should be format 1 (720p), got %s", sorted[2].FormatID)
	}
	if sorted[3].FormatID != "4" {
		t.Errorf("fourth should be format 4 (480p), got %s", sorted[3].FormatID)
	}

	// Verify original is not modified
	if formats[0].FormatID != "1" {
		t.Error("original slice should not be modified")
	}
}

func TestSortFormatsByABR(t *testing.T) {
	formats := []Format{
		{FormatID: "a1", Height: 0, ABR: 128},
		{FormatID: "a2", Height: 0, ABR: 320},
		{FormatID: "a3", Height: 0, ABR: 64},
	}

	sorted := SortFormats(formats)

	if sorted[0].FormatID != "a2" {
		t.Errorf("first should be a2 (320kbps), got %s", sorted[0].FormatID)
	}
	if sorted[1].FormatID != "a1" {
		t.Errorf("second should be a1 (128kbps), got %s", sorted[1].FormatID)
	}
	if sorted[2].FormatID != "a3" {
		t.Errorf("third should be a3 (64kbps), got %s", sorted[2].FormatID)
	}
}

func TestFilterByType(t *testing.T) {
	formats := []Format{
		{FormatID: "1", VCodec: "h264", ACodec: "aac"},  // both
		{FormatID: "2", VCodec: "h264", ACodec: "none"}, // video only
		{FormatID: "3", VCodec: "none", ACodec: "opus"}, // audio only
		{FormatID: "4", VCodec: "vp9", ACodec: ""},      // video only (empty audio)
	}

	t.Run("video tab includes both and video-only", func(t *testing.T) {
		filtered := FilterByType(formats, types.FormatTabVideo)
		if len(filtered) != 3 {
			t.Errorf("expected 3 video formats, got %d", len(filtered))
		}
	})

	t.Run("audio tab includes audio-only", func(t *testing.T) {
		filtered := FilterByType(formats, types.FormatTabAudio)
		if len(filtered) != 1 {
			t.Errorf("expected 1 audio format, got %d", len(filtered))
		}
		if filtered[0].FormatID != "3" {
			t.Errorf("expected format 3, got %s", filtered[0].FormatID)
		}
	})

	t.Run("custom tab includes all", func(t *testing.T) {
		filtered := FilterByType(formats, types.FormatTabCustom)
		if len(filtered) != 4 {
			t.Errorf("expected 4 formats, got %d", len(filtered))
		}
	})
}

func TestSuggestBestFormats(t *testing.T) {
	formats := []Format{
		{FormatID: "137", Height: 1080, Resolution: "1920x1080", VCodec: "h264", ACodec: "none", Filesize: 500},
		{FormatID: "136", Height: 720, Resolution: "1280x720", VCodec: "h264", ACodec: "none", Filesize: 300},
		{FormatID: "251", Height: 0, VCodec: "none", ACodec: "opus", ABR: 160, Filesize: 100},
	}

	combos := SuggestBestFormats(formats)

	if len(combos) == 0 {
		t.Fatal("expected at least one combo")
	}

	// First should always be "Best"
	if combos[0].Label != "Best" {
		t.Errorf("first combo should be 'Best', got %q", combos[0].Label)
	}
	if combos[0].FormatID != "bestvideo+bestaudio/best" {
		t.Errorf("first combo FormatID should be best combo string")
	}

	// Should include 1080p and 720p
	found1080, found720, foundAudio := false, false, false
	for _, c := range combos {
		if c.Label == "1920x1080" {
			found1080 = true
		}
		if c.Label == "1280x720" {
			found720 = true
		}
		if c.Label == "Audio only (best)" {
			foundAudio = true
		}
	}
	if !found1080 {
		t.Error("should include 1080p combo")
	}
	if !found720 {
		t.Error("should include 720p combo")
	}
	if !foundAudio {
		t.Error("should include audio-only combo")
	}
}

func TestConvertFormatsToItems(t *testing.T) {
	formats := []Format{
		{
			FormatID:   "137",
			Ext:        "mp4",
			Resolution: "1920x1080",
			Width:      1920,
			Height:     1080,
			FPS:        30,
			VCodec:     "h264",
			ACodec:     "none",
			Filesize:   1000,
			TBR:        4000,
			VBR:        3900,
			ABR:        0,
			FormatNote: "1080p",
			Protocol:   "https",
		},
	}

	items := ConvertFormatsToItems(formats)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	item := items[0]
	if item.FormatID != "137" {
		t.Errorf("FormatID = %q, want %q", item.FormatID, "137")
	}
	if item.Height != 1080 {
		t.Errorf("Height = %d, want 1080", item.Height)
	}
	if item.Ext != "mp4" {
		t.Errorf("Ext = %q, want %q", item.Ext, "mp4")
	}
	if item.VCodec != "h264" {
		t.Errorf("VCodec = %q, want %q", item.VCodec, "h264")
	}
}
