// Package syncedfolder guesses whether a folder is kept in sync by a
// file-sync service (Dropbox, OneDrive, Syncthing, …), from the marker
// files and folder names those services leave. It's a best guess.
package syncedfolder

import (
	"os"
	"path/filepath"
)

// markers are the files or folders a sync service leaves at the top of the
// folder it keeps in sync, as filepath.Match patterns, each with the
// service's name.
var markers = []struct{ pattern, service string }{
	{".dropbox", "Dropbox"},
	{".stfolder", "Syncthing"},
	{".sync_*.db", "Nextcloud"},
	{"._sync_*.db", "Nextcloud"},
}

// Detect reports the sync service that seems to keep dir in sync, if any.
func Detect(dir string) (service string, ok bool) {
	for {
		if service, ok := markedBy(dir); ok {
			return service, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// markedBy reports the service whose marker dir holds, if any. It matches
// the names in dir rather than globbing under it, so `[`, `*` or `?` in dir's
// own path stay plain characters. A folder it can't read holds no markers.
func markedBy(dir string) (service string, ok bool) {
	entries, _ := os.ReadDir(dir)
	for _, m := range markers {
		for _, e := range entries {
			if matched, _ := filepath.Match(m.pattern, e.Name()); matched {
				return m.service, true
			}
		}
	}
	return "", false
}
