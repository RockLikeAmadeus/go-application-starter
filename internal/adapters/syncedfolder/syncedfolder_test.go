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
