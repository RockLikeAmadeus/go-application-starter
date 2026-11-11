package sqlstore

import "time"

// timeFormat keeps full precision and a fixed width, so stored times sort in
// time order as text.
const timeFormat = "2006-01-02T15:04:05.000000000Z"

// formatTime is the time as it's stored: UTC text in timeFormat.
func formatTime(t time.Time) string {
	return t.UTC().Format(timeFormat)
}

// parseTime reads a time stored by formatTime.
func parseTime(stored string) (time.Time, error) {
	return time.Parse(timeFormat, stored)
}
