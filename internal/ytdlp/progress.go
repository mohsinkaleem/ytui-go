package ytdlp

import (
	"fmt"
	"strconv"
	"strings"
)

// progressMarker is a literal sentinel embedded in every progress line we emit.
// By placing it after the section specifier "download:" (which yt-dlp strips),
// it becomes the first token in the rendered output and lets us identify our
// lines versus yt-dlp's own [info]/[download] status messages.
const progressMarker = "YTPROG"

// progressTemplate is passed verbatim to yt-dlp --progress-template.
// "download:" is the yt-dlp section specifier and is NOT included in the output.
// Everything after it is the rendered template that yt-dlp writes to stderr.
const progressTemplate = `download:` + progressMarker + `|%(progress.downloaded_bytes)s|%(progress.total_bytes)s|%(progress.total_bytes_estimate)s|%(progress.speed)s|%(progress.eta)s|%(progress.status)s`

// ParseProgressLine parses a yt-dlp progress template line.
// Expected format: YTPROG|downloaded|total|total_estimate|speed|eta|status
func ParseProgressLine(line string) (*Progress, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, progressMarker+"|") {
		return nil, fmt.Errorf("not a progress line")
	}
	line = strings.TrimPrefix(line, progressMarker+"|")

	parts := strings.Split(line, "|")
	if len(parts) < 6 {
		return nil, fmt.Errorf("unexpected format: %d parts", len(parts))
	}

	p := &Progress{
		Status: strings.TrimSpace(parts[5]),
	}

	p.DownloadedBytes = parseNA(parts[0])
	totalBytes := parseNA(parts[1])
	totalEstimate := parseNA(parts[2])

	if totalBytes > 0 {
		p.TotalBytes = totalBytes
	} else if totalEstimate > 0 {
		p.TotalBytes = totalEstimate
	}

	p.Speed = formatSpeed(parts[3])
	p.ETA = formatETA(parts[4])

	if p.TotalBytes > 0 && p.DownloadedBytes > 0 {
		p.Percent = float64(p.DownloadedBytes) / float64(p.TotalBytes)
		if p.Percent > 1.0 {
			p.Percent = 1.0
		}
	}

	return p, nil
}

func parseNA(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "NA" || s == "None" || s == "none" {
		return 0
	}
	// Try float first (yt-dlp can output float)
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(f)
}

func formatSpeed(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || s == "NA" || s == "None" || s == "none" {
		return ""
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return s
	}
	const (
		KB = 1024.0
		MB = 1024.0 * KB
		GB = 1024.0 * MB
	)
	switch {
	case f >= GB:
		return fmt.Sprintf("%.1f GB/s", f/GB)
	case f >= MB:
		return fmt.Sprintf("%.1f MB/s", f/MB)
	case f >= KB:
		return fmt.Sprintf("%.1f KB/s", f/KB)
	default:
		return fmt.Sprintf("%.0f B/s", f)
	}
}

func formatETA(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || s == "NA" || s == "None" || s == "none" {
		return ""
	}
	secs, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return s
	}
	total := int(secs)
	if total < 60 {
		return fmt.Sprintf("%ds", total)
	}
	if total < 3600 {
		return fmt.Sprintf("%d:%02d", total/60, total%60)
	}
	return fmt.Sprintf("%d:%02d:%02d", total/3600, (total%3600)/60, total%60)
}

// HumanBytes formats bytes to human readable string
func HumanBytes(b int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case b >= int64(GB):
		return fmt.Sprintf("%.1f GB", float64(b)/float64(GB))
	case b >= int64(MB):
		return fmt.Sprintf("%.1f MB", float64(b)/float64(MB))
	case b >= int64(KB):
		return fmt.Sprintf("%.1f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
