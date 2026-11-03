package sqlite_test

import (
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlite"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

func TestItemRepositoryFindsASavedItemWithItsTimesInUTCAtFullPrecision(t *testing.T) {
	store, err := sqlite.Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer store.Close()
	repo := sqlite.NewItemRepository(store)
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
