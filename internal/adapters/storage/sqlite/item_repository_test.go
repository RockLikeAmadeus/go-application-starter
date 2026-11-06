package sqlite_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlite"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/storagetest"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

func TestItemRepositoryContract(t *testing.T) {
	storagetest.RunItemRepositoryContract(t, func(t *testing.T) ports.ItemRepository {
		return newItemRepository(t)
	})
}

// newItemRepository opens a store in a fresh folder, closed when the test ends.
func newItemRepository(t *testing.T) *sqlite.ItemRepository {
	t.Helper()
	return openItemRepository(t, t.TempDir())
}

// openItemRepository opens a store on dir, closed when the test ends.
func openItemRepository(t *testing.T, dir string) *sqlite.ItemRepository {
	t.Helper()
	return sqlite.NewItemRepository(openStore(t, dir))
}

func TestItemRepositoryFindsASavedItemWithItsTimesInUTCAtFullPrecision(t *testing.T) {
	repo := newItemRepository(t)
	elsewhere := time.FixedZone("UTC+1", 60*60)
	want := domain.Item{
		ID:          "a",
		WorkspaceID: "ws-a",
		Name:        "Alpha",
		CreatedAt:   time.Date(2026, 3, 4, 5, 6, 7, 123456789, elsewhere),
		UpdatedAt:   time.Date(2026, 3, 4, 6, 0, 0, 1, elsewhere),
	}

	if err := repo.Save(t.Context(), want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := repo.FindByID(t.Context(), "ws-a", "a")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}

	if got.ID != want.ID || got.WorkspaceID != want.WorkspaceID || got.Name != want.Name {
		t.Errorf("got %+v, want %+v", got, want)
	}
	for _, tc := range []struct {
		field     string
		got, want time.Time
	}{
		{"CreatedAt", got.CreatedAt, want.CreatedAt},
		{"UpdatedAt", got.UpdatedAt, want.UpdatedAt},
	} {
		if !tc.got.Equal(tc.want) || tc.got.Location() != time.UTC {
			t.Errorf("%s = %v, want %v in UTC", tc.field, tc.got, tc.want.UTC())
		}
	}
}

func TestTwoStoresOnOneDataFolderShareItemsAndCanWriteAtOnce(t *testing.T) {
	// Two programs (say the desktop app and the CLI) open the same folder.
	dir := t.TempDir()
	repos := []*sqlite.ItemRepository{openItemRepository(t, dir), openItemRepository(t, dir)}

	const perStore = 50
	errs := make(chan error, len(repos)*perStore)
	var wg sync.WaitGroup
	for i, repo := range repos {
		for n := range perStore {
			wg.Go(func() {
				errs <- repo.Save(t.Context(), domain.Item{
					ID:          fmt.Sprintf("%d-%d", i, n),
					WorkspaceID: "ws",
					Name:        "Item",
					CreatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				})
			})
		}
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("Save while the other store writes: %v", err)
		}
	}

	for i, repo := range repos {
		items, err := repo.FindAll(t.Context(), "ws")
		if err != nil {
			t.Fatalf("FindAll on store %d: %v", i, err)
		}
		if len(items) != len(repos)*perStore {
			t.Errorf("store %d sees %d items, want %d (both stores' saves)", i, len(items), len(repos)*perStore)
		}
	}
}
