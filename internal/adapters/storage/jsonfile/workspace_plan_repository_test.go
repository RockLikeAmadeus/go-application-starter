package jsonfile_test

import (
	"context"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/jsonfile"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/storagetest"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

func TestWorkspacePlanRepositoryContract(t *testing.T) {
	storagetest.RunWorkspacePlanRepositoryContract(t, func(t *testing.T) ports.WorkspacePlanRepository {
		return jsonfile.NewWorkspacePlanRepository(openStore(t, t.TempDir()))
	})
}

func TestWorkspacePlansSurviveReopen(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	store, err := jsonfile.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := jsonfile.NewWorkspacePlanRepository(store).SavePlan(ctx, "ws-a", "pro"); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	got, err := jsonfile.NewWorkspacePlanRepository(openStore(t, dir)).PlanOf(ctx, "ws-a")
	if err != nil || got != "pro" {
		t.Fatalf("after reopen PlanOf = %q, %v, want \"pro\"", got, err)
	}
}
