package jsonfile_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/jsonfile"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/storagetest"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

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

	got, err := jsonfile.NewItemRepository(openStore(t, dir)).FindByID(ctx, "", "a")
	if err != nil || got.Name != "Alpha" {
		t.Fatalf("after reopen got %+v, %v", got, err)
	}
}

func TestEachItemIsKeptInItsOwnFileAndSavingOneLeavesTheOthersUntouched(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	repo := jsonfile.NewItemRepository(openStore(t, dir))
	saveItem := func(id, name string) {
		t.Helper()
		if err := repo.Save(ctx, domain.Item{ID: id, Name: name}); err != nil {
			t.Fatalf("Save %s: %v", id, err)
		}
	}
	saveItem("a", "Alpha")
	saveItem("b", "Beta")

	// A hand-edited layout that a rewrite of b's file would lose.
	otherItem := filepath.Join(dir, "items", "b.json")
	data, err := os.ReadFile(otherItem)
	if err != nil {
		t.Fatalf("reading b's file: %v", err)
	}
	edited := append([]byte("\n"), data...)
	if err := os.WriteFile(otherItem, edited, 0o600); err != nil {
		t.Fatal(err)
	}

	saveItem("a", "Alpha renamed")

	got, err := os.ReadFile(otherItem)
	if err != nil {
		t.Fatalf("reading b's file after saving a: %v", err)
	}
	if string(got) != string(edited) {
		t.Errorf("saving a changed b's file:\n%s\nwant:\n%s", got, edited)
	}
}

func TestFindByIDReadsOnlyThatItemsFile(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	repo := jsonfile.NewItemRepository(openStore(t, dir))
	if err := repo.Save(ctx, domain.Item{ID: "a", Name: "Alpha"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	// Reading the whole table would stop at this file, which can't be decoded.
	if err := os.WriteFile(filepath.Join(dir, "items", "b.json"), []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := repo.FindByID(ctx, "", "a")
	if err != nil || got.Name != "Alpha" {
		t.Fatalf("FindByID = %+v, %v, want Alpha", got, err)
	}
}
