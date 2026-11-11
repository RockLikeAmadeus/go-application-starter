package sqlite_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlstore"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// openItemRepository opens a store on dir, closed when the test ends.
func openItemRepository(t *testing.T, dir string) *sqlstore.ItemRepository {
	t.Helper()
	return sqlstore.NewItemRepository(openStore(t, dir).DB())
}

func TestTwoStoresOnOneDataFolderShareItemsAndCanWriteAtOnce(t *testing.T) {
	// Two programs (say the desktop app and the CLI) open the same folder.
	dir := t.TempDir()
	repos := []*sqlstore.ItemRepository{openItemRepository(t, dir), openItemRepository(t, dir)}

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
