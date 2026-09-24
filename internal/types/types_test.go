package types

import (
	"testing"
)

func TestStateConstants(t *testing.T) {
	states := []State{
		StateSearchInput, StateLoading, StateVideoList,
		StateFormatList, StateDownload, StateVideoPlaying, StateResumeList,
	}
	seen := make(map[State]bool)
	for _, s := range states {
		if s == "" {
			t.Errorf("state constant should not be empty")
		}
		if seen[s] {
			t.Errorf("duplicate state constant: %s", s)
		}
		seen[s] = true
	}
}

func TestSortByString(t *testing.T) {
	tests := []struct {
		sort SortBy
		want string
	}{
		{SortRelevance, "Relevance"},
		{SortUploadDate, "Upload date"},
		{SortViewCount, "View count"},
		{SortRating, "Rating"},
		{SortBy("unknown"), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.sort.String(); got != tt.want {
			t.Errorf("SortBy(%q).String() = %q, want %q", tt.sort, got, tt.want)
		}
	}
}

func TestVideoItemFilterValue(t *testing.T) {
	v := VideoItem{Title: "Test Video", Channel: "Test Channel"}
	if got := v.FilterValue(); got != "Test Video" {
		t.Errorf("VideoItem.FilterValue() = %q, want %q", got, "Test Video")
	}
}

func TestFormatItemFilterValue(t *testing.T) {
	f := FormatItem{Resolution: "1920x1080", FormatID: "137"}
	if got := f.FilterValue(); got != "1920x1080" {
		t.Errorf("FormatItem.FilterValue() = %q, want %q", got, "1920x1080")
	}
}

func TestFormatItemIsVideoOnly(t *testing.T) {
	tests := []struct {
		name   string
		item   FormatItem
		expect bool
	}{
		{"video only", FormatItem{VCodec: "h264", ACodec: "none"}, true},
		{"both", FormatItem{VCodec: "h264", ACodec: "aac"}, false},
		{"audio only", FormatItem{VCodec: "none", ACodec: "aac"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.IsVideoOnly(); got != tt.expect {
				t.Errorf("IsVideoOnly() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestFormatItemIsAudioOnly(t *testing.T) {
	tests := []struct {
		name   string
		item   FormatItem
		expect bool
	}{
		{"audio only", FormatItem{VCodec: "none", ACodec: "aac"}, true},
		{"both", FormatItem{VCodec: "h264", ACodec: "aac"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.IsAudioOnly(); got != tt.expect {
				t.Errorf("IsAudioOnly() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestFormatItemHumanSize(t *testing.T) {
	tests := []struct {
		name     string
		filesize int64
		want     string
	}{
		{"zero", 0, "N/A"},
		{"bytes", 500, "500B"},
		{"kilobytes", 1536, "1.5KB"},
		{"megabytes", 10 * 1024 * 1024, "10MB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := FormatItem{Filesize: tt.filesize}
			if got := f.HumanSize(); got != tt.want {
				t.Errorf("HumanSize() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatItemQualityLabel(t *testing.T) {
	tests := []struct {
		name string
		item FormatItem
		want string
	}{
		{"height", FormatItem{Height: 1080}, "1080p"},
		{"abr only", FormatItem{ABR: 128}, "128kbps"},
		{"empty", FormatItem{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.QualityLabel(); got != tt.want {
				t.Errorf("QualityLabel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatTabConstants(t *testing.T) {
	if FormatTabVideo != 0 {
		t.Error("FormatTabVideo should be 0")
	}
	if FormatTabAudio != 1 {
		t.Error("FormatTabAudio should be 1")
	}
	if FormatTabCustom != 2 {
		t.Error("FormatTabCustom should be 2")
	}
}

func TestVideoItemVideoURL(t *testing.T) {
	tests := []struct {
		name string
		item VideoItem
		want string
	}{
		{"has URL", VideoItem{URL: "https://example.com/video", ID: "abc"}, "https://example.com/video"},
		{"no URL, has ID", VideoItem{ID: "abc123"}, "https://www.youtube.com/watch?v=abc123"},
		{"empty", VideoItem{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.VideoURL(); got != tt.want {
				t.Errorf("VideoURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatItemHasVideoAndAudio(t *testing.T) {
	tests := []struct {
		name string
		item FormatItem
		want bool
	}{
		{"both", FormatItem{VCodec: "h264", ACodec: "aac"}, true},
		{"video only", FormatItem{VCodec: "h264", ACodec: "none"}, false},
		{"audio only", FormatItem{VCodec: "none", ACodec: "aac"}, false},
		{"empty", FormatItem{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.HasVideoAndAudio(); got != tt.want {
				t.Errorf("HasVideoAndAudio() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatItemHumanSizeGigabytes(t *testing.T) {
	f := FormatItem{Filesize: 2 * 1024 * 1024 * 1024}
	got := f.HumanSize()
	if got != "2GB" {
		t.Errorf("HumanSize() = %q, want %q", got, "2GB")
	}
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{1.0, "1"},
		{1.5, "1.5"},
		{0.0, "0"},
		{100.0, "100"},
		{99.9, "99.9"},
	}
	for _, tt := range tests {
		got := formatFloat(tt.input)
		if got != tt.want {
			t.Errorf("formatFloat(%f) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSortOptionsLength(t *testing.T) {
	if len(SortOptions) != 4 {
		t.Errorf("SortOptions length = %d, want 4", len(SortOptions))
	}
}
