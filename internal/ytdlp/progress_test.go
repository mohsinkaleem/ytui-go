package ytdlp

import (
	"testing"
)

func TestParseProgressLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantErr bool
		check   func(*testing.T, *Progress)
	}{
		{
			name: "valid basic line",
			line: "YTPROG|5242880|10485760|10485760|1048576|5|downloading",
			check: func(t *testing.T, p *Progress) {
				if p.DownloadedBytes != 5242880 {
					t.Errorf("DownloadedBytes = %d, want 5242880", p.DownloadedBytes)
				}
				if p.TotalBytes != 10485760 {
					t.Errorf("TotalBytes = %d, want 10485760", p.TotalBytes)
				}
				if p.Status != "downloading" {
					t.Errorf("Status = %q, want %q", p.Status, "downloading")
				}
				if p.Percent < 0.49 || p.Percent > 0.51 {
					t.Errorf("Percent = %f, want ~0.5", p.Percent)
				}
			},
		},
		{
			name: "uses estimate when total is NA",
			line: "YTPROG|1024|NA|2048|512|10|downloading",
			check: func(t *testing.T, p *Progress) {
				if p.TotalBytes != 2048 {
					t.Errorf("TotalBytes = %d, want 2048 (estimate)", p.TotalBytes)
				}
				if p.Percent < 0.49 || p.Percent > 0.51 {
					t.Errorf("Percent = %f, want ~0.5", p.Percent)
				}
			},
		},
		{
			name: "NA values handled",
			line: "YTPROG|NA|NA|NA|NA|NA|downloading",
			check: func(t *testing.T, p *Progress) {
				if p.DownloadedBytes != 0 {
					t.Errorf("DownloadedBytes = %d, want 0", p.DownloadedBytes)
				}
				if p.TotalBytes != 0 {
					t.Errorf("TotalBytes = %d, want 0", p.TotalBytes)
				}
				if p.Percent != 0 {
					t.Errorf("Percent = %f, want 0", p.Percent)
				}
			},
		},
		{
			name: "status finished",
			line: "YTPROG|10485760|10485760|10485760|0|0|finished",
			check: func(t *testing.T, p *Progress) {
				if p.Percent != 1.0 {
					t.Errorf("Percent = %f, want 1.0", p.Percent)
				}
				if p.Status != "finished" {
					t.Errorf("Status = %q, want %q", p.Status, "finished")
				}
			},
		},
		{
			name:    "not a progress line",
			line:    "some random output",
			wantErr: true,
		},
		{
			name:    "too few parts",
			line:    "YTPROG|100|200",
			wantErr: true,
		},
		{
			name: "float bytes",
			line: "YTPROG|5242880.5|10485761.0|10485761.0|1048576.3|5.2|downloading",
			check: func(t *testing.T, p *Progress) {
				if p.DownloadedBytes != 5242880 {
					t.Errorf("DownloadedBytes = %d, want 5242880", p.DownloadedBytes)
				}
			},
		},
		{
			name: "percent capped at 1.0",
			line: "YTPROG|20000|10000|10000|1000|0|downloading",
			check: func(t *testing.T, p *Progress) {
				if p.Percent != 1.0 {
					t.Errorf("Percent = %f, want 1.0 (capped)", p.Percent)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := ParseProgressLine(tt.line)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.check(t, p)
		})
	}
}

func TestFormatSpeed(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"NA", ""},
		{"None", ""},
		{"", ""},
		{"500", "500 B/s"},
		{"1024", "1.0 KB/s"},
		{"1048576", "1.0 MB/s"},
		{"1073741824", "1.0 GB/s"},
		{"invalid", "invalid"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := formatSpeed(tt.input); got != tt.want {
				t.Errorf("formatSpeed(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatETA(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"NA", ""},
		{"None", ""},
		{"", ""},
		{"30", "30s"},
		{"90", "1:30"},
		{"3661", "1:01:01"},
		{"invalid", "invalid"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := formatETA(tt.input); got != tt.want {
				t.Errorf("formatETA(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHumanBytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := HumanBytes(tt.bytes); got != tt.want {
				t.Errorf("HumanBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}
