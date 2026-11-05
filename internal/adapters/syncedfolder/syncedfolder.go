// Package syncedfolder guesses whether a folder is kept in sync by a
// file-sync service (Dropbox, OneDrive, Syncthing, …), from the marker
// files and folder names those services leave. It's a best guess.
package syncedfolder

import (
	"os"
	"path/filepath"
	"strings"
)

// A syncSign is a filepath.Match pattern for a file or folder name that a sync
// service leaves, with the service's name.
type syncSign struct{ pattern, service string }

// markers are the files or folders a sync service leaves at the top of the
// folder it keeps in sync.
var markers = []syncSign{
	{".dropbox", "Dropbox"},
	{".stfolder", "Syncthing"},
	{".sync_*.db", "Nextcloud"},
	{"._sync_*.db", "Nextcloud"},
}

// folderNames are the names a sync service gives the folder it keeps in sync.
var folderNames = []syncSign{
	{"Dropbox", "Dropbox"},
	{"OneDrive", "OneDrive"},
	{"OneDrive - *", "OneDrive"}, // a work or school account: "OneDrive - <org>"
	{"Google Drive", "Google Drive"},
	{"My Drive", "Google Drive"},
	{"iCloud Drive", "iCloud Drive"},
	{"iCloudDrive", "iCloud Drive"},      // Windows
	{"Mobile Documents", "iCloud Drive"}, // macOS: ~/Library/Mobile Documents
}

// providerNames are the services' names for the providers in macOS's
// Library/CloudStorage whose folder prefix isn't already that name.
var providerNames = map[string]string{
	"GoogleDrive": "Google Drive",
}

// Detect reports the sync service that seems to keep dir in sync, if any.
// It checks dir and each folder above it. A relative dir is taken from the
// working directory, whose parents count too; if that can't be found, dir is
// checked as given.
func Detect(dir string) (service string, ok bool) {
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	for {
		if service, ok := markedBy(dir); ok {
			return service, true
		}
		if service, ok := namedFor(dir); ok {
			return service, true
		}
		if service, ok := cloudStorageProvider(dir); ok {
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
	for _, marker := range markers {
		for _, entry := range entries {
			if matches(marker, entry.Name()) {
				return marker.service, true
			}
		}
	}
	return "", false
}

// namedFor reports the service whose folder name dir has, if any.
func namedFor(dir string) (service string, ok bool) {
	for _, folderName := range folderNames {
		if matches(folderName, filepath.Base(dir)) {
			return folderName.service, true
		}
	}
	return "", false
}

// cloudStorageProvider reports the provider whose folder dir is in macOS's
// Library/CloudStorage, if it's one. Each provider names its folder
// "<Provider>-<account>", e.g. "Dropbox-Personal".
func cloudStorageProvider(dir string) (service string, ok bool) {
	parent := filepath.Dir(dir)
	if filepath.Base(parent) != "CloudStorage" || filepath.Base(filepath.Dir(parent)) != "Library" {
		return "", false
	}
	provider, _, _ := strings.Cut(filepath.Base(dir), "-")
	if name, ok := providerNames[provider]; ok {
		return name, true
	}
	return provider, true
}

func matches(sign syncSign, name string) bool {
	isMatch, _ := filepath.Match(sign.pattern, name)
	return isMatch
}
