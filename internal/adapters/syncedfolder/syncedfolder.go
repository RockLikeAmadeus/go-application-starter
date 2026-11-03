// Package syncedfolder guesses whether a folder is kept in sync by a
// file-sync service (Dropbox, OneDrive, Syncthing, …), from the marker
// files and folder names those services leave. It's a best guess.
package syncedfolder

import (
	"os"
	"path/filepath"
)

// markers are the files or folders a sync service leaves at the top of the
// folder it keeps in sync, each with the service's name.
var markers = []struct{ name, service string }{
	{".dropbox", "Dropbox"},
	{".stfolder", "Syncthing"},
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

// markedBy reports the service whose marker dir holds, if any.
func markedBy(dir string) (service string, ok bool) {
	for _, m := range markers {
		if _, err := os.Stat(filepath.Join(dir, m.name)); err == nil {
			return m.service, true
		}
	}
	return "", false
}
