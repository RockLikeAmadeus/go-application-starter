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

var baseTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// workspace is the workspace the contract's items belong to. Keeping
// workspaces apart is checked by RunWorkspaceIsolation.
const workspace domain.WorkspaceID = "ws-a"

func item(id, name string, offset time.Duration) domain.Item {
	return itemIn(workspace, id, name, offset)
}

func itemIn(itemWorkspace domain.WorkspaceID, id, name string, offset time.Duration) domain.Item {
	return domain.Item{
		ID:          id,
		WorkspaceID: itemWorkspace,
		Name:        name,
		CreatedAt:   baseTime.Add(offset),
		UpdatedAt:   baseTime.Add(offset),
	}
}

// RunItemRepositoryContract is the contract every ports.ItemRepository
// implementation must pass, workspace isolation included.
func RunItemRepositoryContract(t *testing.T, newRepo Factory) {
	ctx := context.Background()

	t.Run("workspace isolation", func(t *testing.T) {
		RunWorkspaceIsolation(t,
			func(t *testing.T) ports.WorkspaceRepository[domain.Item] { return newRepo(t) },
			RecordKind[domain.Item]{
				New: func(recordWorkspace domain.WorkspaceID, id string) domain.Item {
					return itemIn(recordWorkspace, id, "Item "+id, 0)
				},
				ID:        func(i domain.Item) string { return i.ID },
				Workspace: func(i domain.Item) domain.WorkspaceID { return i.WorkspaceID },
			})
	})

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

	t.Run("FindByID missing returns ErrNotFound", func(t *testing.T) {
		repo := newRepo(t)
		if _, err := repo.FindByID(ctx, workspace, "missing"); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("Save replaces existing item", func(t *testing.T) {
		repo := newRepo(t)
		mustSave(t, repo, item("a", "Alpha", 0))
		updated := item("a", "Alpha 2", 0)
		updated.UpdatedAt = baseTime.Add(time.Hour)
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
		for _, item := range all {
			ids = append(ids, item.ID)
		}
		if len(ids) != 3 || ids[0] != "a" || ids[1] != "b" || ids[2] != "c" {
			t.Errorf("order = %v, want [a b c]", ids)
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

}

func mustSave(t *testing.T, repo ports.ItemRepository, item domain.Item) {
	t.Helper()
	if err := repo.Save(context.Background(), item); err != nil {
		t.Fatalf("Save(%s): %v", item.ID, err)
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
