package models

import "fmt"

// FormatResumeCount returns a formatted string for resume count
func FormatResumeCount(count int) string {
	if count == 1 {
		return "1 unfinished download"
	}
	return fmt.Sprintf("%d unfinished downloads", count)
}
