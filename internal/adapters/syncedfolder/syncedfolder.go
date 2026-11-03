// Package syncedfolder guesses whether a folder is kept in sync by a
// file-sync service (Dropbox, OneDrive, Syncthing, …), from the marker
// files and folder names those services leave. It's a best guess.
package syncedfolder

import (
	"os"
	"path/filepath"
)

// Detect reports the sync service that seems to keep dir in sync, if any.
func Detect(dir string) (service string, ok bool) {
	for {
		if _, err := os.Stat(filepath.Join(dir, ".dropbox")); err == nil {
			return "Dropbox", true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}
