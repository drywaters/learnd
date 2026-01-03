package enricher

import (
	"regexp"
	"strconv"
)

// ISO 8601 duration pattern (PT#H#M#S)
var durationPattern = regexp.MustCompile(`PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?`)

// parseDuration converts ISO 8601 duration to seconds
func parseDuration(duration string) int {
	matches := durationPattern.FindStringSubmatch(duration)
	if len(matches) == 0 {
		return 0
	}

	var hours, minutes, seconds int
	if matches[1] != "" {
		hours, _ = strconv.Atoi(matches[1])
	}
	if matches[2] != "" {
		minutes, _ = strconv.Atoi(matches[2])
	}
	if matches[3] != "" {
		seconds, _ = strconv.Atoi(matches[3])
	}

	return hours*3600 + minutes*60 + seconds
}
