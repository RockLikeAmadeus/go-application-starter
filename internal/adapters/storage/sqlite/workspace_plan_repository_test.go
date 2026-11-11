package sqlite_test

import (
	"context"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlite"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlstore"
)

func TestWorkspacePlansSurviveReopen(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	store, err := sqlite.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := sqlstore.NewWorkspacePlanRepository(store.DB()).SavePlan(ctx, "ws-a", "pro"); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	got, err := sqlstore.NewWorkspacePlanRepository(openStore(t, dir).DB()).PlanOf(ctx, "ws-a")
	if err != nil || got != "pro" {
		t.Fatalf("after reopen PlanOf = %q, %v, want \"pro\"", got, err)
	}
}
