package sqlstore_test

import (
	"testing"
	"time"

	"github.com/pressly/goose/v3"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlstore"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/storagetest"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

func TestItemRepositoryContractOnSQLite(t *testing.T) {
	storagetest.RunItemRepositoryContract(t, func(t *testing.T) ports.ItemRepository {
		return newItemRepositoryOnSQLite(t)
	})
}

// newItemRepositoryOnSQLite returns a repository on a fresh, migrated SQLite
// database, closed when the test ends.
func newItemRepositoryOnSQLite(t *testing.T) *sqlstore.ItemRepository {
	t.Helper()
	db := openSQLite(t)
	if err := sqlstore.Migrate(t.Context(), db, goose.DialectSQLite3); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return sqlstore.NewItemRepository(db)
}

func TestItemRepositoryFindsASavedItemWithItsTimesInUTCAtFullPrecision(t *testing.T) {
	repo := newItemRepositoryOnSQLite(t)
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
