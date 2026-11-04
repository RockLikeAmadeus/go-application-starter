package syncedfolder_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/syncedfolder"
)

// folderInside makes, and returns, a data folder two levels below root.
func folderInside(t *testing.T, root string) string {
	t.Helper()
	dir := filepath.Join(root, "notes", "starter")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestDetect_AFolderInsideOneHoldingDotDropboxLooksSyncedByDropbox(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".dropbox"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	dir := folderInside(t, root)

	service, ok := syncedfolder.Detect(dir)
	if !ok || service != "Dropbox" {
		t.Errorf("Detect(%q) = (%q, %v), want (%q, true)", dir, service, ok, "Dropbox")
	}
}

func TestDetect_AFolderWithNoSignsAnywhereAboveItDoesNotLookSynced(t *testing.T) {
	dir := folderInside(t, t.TempDir())

	if service, ok := syncedfolder.Detect(dir); ok {
		t.Errorf("Detect(%q) = (%q, true), want (\"\", false)", dir, service)
	}
}

func TestDetect_AFolderInsideOneHoldingDotStfolderLooksSyncedBySyncthing(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".stfolder"), 0o750); err != nil {
		t.Fatal(err)
	}
	dir := folderInside(t, root)

	service, ok := syncedfolder.Detect(dir)
	if !ok || service != "Syncthing" {
		t.Errorf("Detect(%q) = (%q, %v), want (%q, true)", dir, service, ok, "Syncthing")
	}
}

func TestDetect_AFolderInsideOneHoldingASyncJournalLooksSyncedByNextcloud(t *testing.T) {
	for _, journal := range []string{".sync_7a3f9c2e81b4.db", "._sync_7a3f9c2e81b4.db"} {
		t.Run(journal, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, journal), nil, 0o600); err != nil {
				t.Fatal(err)
			}
			dir := folderInside(t, root)

			service, ok := syncedfolder.Detect(dir)
			if !ok || service != "Nextcloud" {
				t.Errorf("Detect(%q) = (%q, %v), want (%q, true)", dir, service, ok, "Nextcloud")
			}
		})
	}
}

func TestDetect_AFolderWhosePathHoldsGlobCharactersIsStillCheckedForMarkers(t *testing.T) {
	synced := filepath.Join(t.TempDir(), "notes [old]")
	dir := filepath.Join(synced, "starter")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(synced, ".dropbox"), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	service, ok := syncedfolder.Detect(dir)
	if !ok || service != "Dropbox" {
		t.Errorf("Detect(%q) = (%q, %v), want (%q, true)", dir, service, ok, "Dropbox")
	}
}

func TestDetect_AFolderInsideOneNamedForASyncServiceLooksSyncedByIt(t *testing.T) {
	for _, tc := range []struct{ folder, service string }{
		{"Dropbox", "Dropbox"},
		{"OneDrive", "OneDrive"},
		{"OneDrive - Contoso", "OneDrive"},
		{"Google Drive", "Google Drive"},
		{"My Drive", "Google Drive"},
		{"iCloud Drive", "iCloud Drive"},
		{"iCloudDrive", "iCloud Drive"},
		{"Mobile Documents", "iCloud Drive"},
	} {
		t.Run(tc.folder, func(t *testing.T) {
			dir := folderInside(t, filepath.Join(t.TempDir(), tc.folder))

			service, ok := syncedfolder.Detect(dir)
			if !ok || service != tc.service {
				t.Errorf("Detect(%q) = (%q, %v), want (%q, true)", dir, service, ok, tc.service)
			}
		})
	}
}
