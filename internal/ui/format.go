package ui

import (
	"fmt"
	"time"
)

// Divide performs integer division safely, returning 0 if divisor is 0
func Divide(a, b int) int {
	if b == 0 {
		return 0
	}
	return a / b
}

// FormatDuration formats seconds into human-readable duration (e.g., "1h 30m")
func FormatDuration(seconds *int) string {
	if seconds == nil || *seconds <= 0 {
		return ""
	}

	total := *seconds
	hours := total / 3600
	minutes := (total % 3600) / 60
	secs := total % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %02dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %02ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

// FormatReadingTime formats seconds into reading time estimate (e.g., "15m")
func FormatReadingTime(seconds *int) string {
	if seconds == nil || *seconds <= 0 {
		return ""
	}

	minutes := *seconds / 60
	if *seconds%60 != 0 {
		minutes++
	}

	hours := minutes / 60
	remaining := minutes % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, remaining)
	}
	return fmt.Sprintf("%dm", minutes)
}

// FormatDate formats a time to "Jan 2, 2006" format
// Returns an empty string if the time is zero (0001-01-01)
func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("Jan 2, 2006")
}
