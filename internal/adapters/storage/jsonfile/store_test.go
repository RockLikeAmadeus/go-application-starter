package jsonfile_test

import (
	"errors"
	"os"
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

func TestStoreIsReadyWhileItsDataDirectoryIsThere(t *testing.T) {
	store := openStore(t, t.TempDir())

	if err := store.Ready(t.Context()); err != nil {
		t.Errorf("Ready error = %v, want nil", err)
	}
}

func TestStoreIsNotReadyOnceItsDataDirectoryIsRemoved(t *testing.T) {
	dir := t.TempDir()
	store := openStore(t, dir)
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}

	if err := store.Ready(t.Context()); err == nil {
		t.Error("Ready error = nil, want an error once the data directory is gone")
	}
}
