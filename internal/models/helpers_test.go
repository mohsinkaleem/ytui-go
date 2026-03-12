package models

import "testing"

func TestFormatResumeCount(t *testing.T) {
	tests := []struct {
		count int
		want  string
	}{
		{0, "0 unfinished downloads"},
		{1, "1 unfinished download"},
		{5, "5 unfinished downloads"},
	}
	for _, tt := range tests {
		got := FormatResumeCount(tt.count)
		if got != tt.want {
			t.Errorf("FormatResumeCount(%d) = %q, want %q", tt.count, got, tt.want)
		}
	}
}

func TestFormatViewCount(t *testing.T) {
	tests := []struct {
		count int64
		want  string
	}{
		{0, "0"},
		{999, "999"},
		{1_000, "1.0K"},
		{1_500, "1.5K"},
		{1_000_000, "1.0M"},
		{2_500_000, "2.5M"},
		{1_000_000_000, "1.0B"},
		{3_700_000_000, "3.7B"},
	}
	for _, tt := range tests {
		got := formatViewCount(tt.count)
		if got != tt.want {
			t.Errorf("formatViewCount(%d) = %q, want %q", tt.count, got, tt.want)
		}
	}
}
