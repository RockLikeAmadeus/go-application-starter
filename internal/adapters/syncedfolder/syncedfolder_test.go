package syncedfolder_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/syncedfolder"
)

func TestDetect_AFolderInsideOneHoldingDotDropboxLooksSyncedByDropbox(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".dropbox"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "notes", "starter")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}

	service, ok := syncedfolder.Detect(dir)
	if !ok || service != "Dropbox" {
		t.Errorf("Detect(%q) = (%q, %v), want (%q, true)", dir, service, ok, "Dropbox")
	}
}
