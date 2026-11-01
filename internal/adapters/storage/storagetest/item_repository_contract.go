// Package storagetest is the contract test suite every ports.ItemRepository
// implementation must pass. Adding a new storage adapter (SQL, NoSQL, ...)
// should start by calling RunItemRepositoryContract from its own tests.
package storagetest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// Factory returns a fresh, empty repository for each subtest.
type Factory func(t *testing.T) ports.ItemRepository

var base = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// The workspace the contract's items belong to unless a case says otherwise,
// and another one to check that workspaces are kept apart.
const (
	workspace      domain.WorkspaceID = "ws-a"
	otherWorkspace domain.WorkspaceID = "ws-b"
)

func item(id, name string, offset time.Duration) domain.Item {
	return itemIn(workspace, id, name, offset)
}

func itemIn(ws domain.WorkspaceID, id, name string, offset time.Duration) domain.Item {
	return domain.Item{
		ID:          id,
		WorkspaceID: ws,
		Name:        name,
		CreatedAt:   base.Add(offset),
		UpdatedAt:   base.Add(offset),
	}
}

func RunItemRepositoryContract(t *testing.T, newRepo Factory) {
	ctx := context.Background()

	t.Run("FindByID returns saved item", func(t *testing.T) {
		repo := newRepo(t)
		want := item("a", "Alpha", 0)
		mustSave(t, repo, want)

		got, err := repo.FindByID(ctx, workspace, "a")
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		assertItemEqual(t, got, want)
	})

	t.Run("Save keeps the item's workspace", func(t *testing.T) {
		repo := newRepo(t)
		want := itemIn(otherWorkspace, "a", "Alpha", 0)
		mustSave(t, repo, want)

		got, err := repo.FindByID(ctx, otherWorkspace, "a")
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		assertItemEqual(t, got, want)
	})

	t.Run("FindByID missing returns ErrNotFound", func(t *testing.T) {
		repo := newRepo(t)
		if _, err := repo.FindByID(ctx, workspace, "missing"); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("FindByID of another workspace's item returns ErrNotFound", func(t *testing.T) {
		repo := newRepo(t)
		mustSave(t, repo, itemIn(otherWorkspace, "a", "Theirs", 0))

		if _, err := repo.FindByID(ctx, workspace, "a"); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("Save replaces existing item", func(t *testing.T) {
		repo := newRepo(t)
		mustSave(t, repo, item("a", "Alpha", 0))
		updated := item("a", "Alpha 2", 0)
		updated.UpdatedAt = base.Add(time.Hour)
		mustSave(t, repo, updated)

		got, _ := repo.FindByID(ctx, workspace, "a")
		assertItemEqual(t, got, updated)
		all, _ := repo.FindAll(ctx, workspace)
		if len(all) != 1 {
			t.Errorf("FindAll len = %d, want 1", len(all))
		}
	})

	t.Run("FindAll orders by CreatedAt then ID", func(t *testing.T) {
		repo := newRepo(t)
		mustSave(t, repo, item("c", "C", 2*time.Second))
		mustSave(t, repo, item("b", "B", time.Second))
		mustSave(t, repo, item("a", "A", time.Second))

		all, err := repo.FindAll(ctx, workspace)
		if err != nil {
			t.Fatalf("FindAll: %v", err)
		}
		var ids []string
		for _, it := range all {
			ids = append(ids, it.ID)
		}
		if len(ids) != 3 || ids[0] != "a" || ids[1] != "b" || ids[2] != "c" {
			t.Errorf("order = %v, want [a b c]", ids)
		}
	})

	t.Run("FindAll lists only the given workspace's items", func(t *testing.T) {
		repo := newRepo(t)
		mustSave(t, repo, item("a", "Mine", 0))
		mustSave(t, repo, itemIn(otherWorkspace, "b", "Theirs", time.Second))

		all, err := repo.FindAll(ctx, workspace)
		if err != nil {
			t.Fatalf("FindAll: %v", err)
		}
		if len(all) != 1 || all[0].ID != "a" {
			t.Errorf("FindAll(%s) = %+v, want only item a", workspace, all)
		}
	})

	t.Run("FindAll on empty repository returns empty slice", func(t *testing.T) {
		repo := newRepo(t)
		all, err := repo.FindAll(ctx, workspace)
		if err != nil || len(all) != 0 {
			t.Fatalf("FindAll = %v, %v; want empty", all, err)
		}
	})

	t.Run("Delete removes item", func(t *testing.T) {
		repo := newRepo(t)
		mustSave(t, repo, item("a", "A", 0))
		if err := repo.Delete(ctx, workspace, "a"); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if _, err := repo.FindByID(ctx, workspace, "a"); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("FindByID after delete err = %v, want ErrNotFound", err)
		}
	})

	t.Run("Delete missing returns ErrNotFound", func(t *testing.T) {
		repo := newRepo(t)
		if err := repo.Delete(ctx, workspace, "missing"); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("Delete of another workspace's item returns ErrNotFound and keeps it", func(t *testing.T) {
		repo := newRepo(t)
		theirs := itemIn(otherWorkspace, "a", "Theirs", 0)
		mustSave(t, repo, theirs)

		if err := repo.Delete(ctx, workspace, "a"); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("err = %v, want ErrNotFound", err)
		}
		got, err := repo.FindByID(ctx, otherWorkspace, "a")
		if err != nil {
			t.Fatalf("FindByID in its own workspace after the delete: %v", err)
		}
		assertItemEqual(t, got, theirs)
	})
}

func mustSave(t *testing.T, repo ports.ItemRepository, it domain.Item) {
	t.Helper()
	if err := repo.Save(context.Background(), it); err != nil {
		t.Fatalf("Save(%s): %v", it.ID, err)
	}
}

// assertItemEqual compares timestamps with Equal so adapters may round-trip
// through formats that drop the monotonic clock or change the location.
func assertItemEqual(t *testing.T, got, want domain.Item) {
	t.Helper()
	if got.ID != want.ID || got.WorkspaceID != want.WorkspaceID || got.Name != want.Name ||
		!got.CreatedAt.Equal(want.CreatedAt) || !got.UpdatedAt.Equal(want.UpdatedAt) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
