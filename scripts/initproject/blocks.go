package main

import (
	"fmt"
	"strings"
)

// RemoveBlocks deletes every block fenced by "<tag>:start" and "<tag>:end"
// marker lines, markers included. A marker is any line containing the text,
// so it works in whatever comment syntax the file uses (//, #, <!-- -->,
// {/* */}). Unmatched or nested markers are an error naming the line, so a
// mistyped marker never silently deletes the rest of a file.
func RemoveBlocks(text, tag string) (string, error) {
	start, end := tag+":start", tag+":end"
	var out strings.Builder
	inside, openedAt := false, 0
	for i, line := range strings.SplitAfter(text, "\n") {
		lineNo := i + 1
		switch {
		case strings.Contains(line, start):
			if inside {
				return "", fmt.Errorf("line %d: %s inside another %s block", lineNo, start, tag)
			}
			inside, openedAt = true, lineNo
		case strings.Contains(line, end):
			if !inside {
				return "", fmt.Errorf("line %d: %s without %s", lineNo, end, start)
			}
			inside = false
		case !inside:
			out.WriteString(line)
		}
	}
	if inside {
		return "", fmt.Errorf("line %d: %s without %s", openedAt, start, end)
	}
	return out.String(), nil
}
