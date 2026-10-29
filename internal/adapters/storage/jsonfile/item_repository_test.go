package jsonfile_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/jsonfile"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/storagetest"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
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

func TestItemRepositoryContract(t *testing.T) {
	storagetest.RunItemRepositoryContract(t, func(t *testing.T) ports.ItemRepository {
		return jsonfile.NewItemRepository(openStore(t, t.TempDir()))
	})
}

func TestItemsSurviveReopen(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	item := domain.Item{ID: "a", Name: "Alpha", CreatedAt: time.Unix(0, 0).UTC(), UpdatedAt: time.Unix(0, 0).UTC()}

	store, _ := jsonfile.Open(dir)
	if err := jsonfile.NewItemRepository(store).Save(ctx, item); err != nil {
		t.Fatalf("Save: %v", err)
	}
	store.Close()

	got, err := jsonfile.NewItemRepository(openStore(t, dir)).FindByID(ctx, "a")
	if err != nil || got.Name != "Alpha" {
		t.Fatalf("after reopen got %+v, %v", got, err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "items.json"))
	if !strings.Contains(string(data), `"version": 1`) {
		t.Errorf("items.json missing version header:\n%s", data)
	}
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
