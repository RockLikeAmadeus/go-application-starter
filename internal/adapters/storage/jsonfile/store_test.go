package jsonfile_test

import (
	"errors"
	"runtime"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/jsonfile"
)

func openStore(t *testing.T, dir string) *jsonfile.Store {
	t.Helper()
	store, err := jsonfile.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestOpenFailsWhenDirectoryIsLocked(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file locking is not implemented on windows")
	}
	dir := t.TempDir()
	openStore(t, dir)

	if _, err := jsonfile.Open(dir); !errors.Is(err, jsonfile.ErrLocked) {
		t.Fatalf("second Open err = %v, want ErrLocked", err)
	}
}
