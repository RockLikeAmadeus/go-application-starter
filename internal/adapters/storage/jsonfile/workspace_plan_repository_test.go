package jsonfile_test

import (
	"context"
	"os"
	"path/filepath"
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

func TestEachWorkspacesPlanIsKeptInItsOwnFileAndSavingOneLeavesTheOthersUntouched(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	repo := jsonfile.NewWorkspacePlanRepository(openStore(t, dir))
	if err := repo.SavePlan(ctx, "ws-a", "free"); err != nil {
		t.Fatalf("SavePlan ws-a: %v", err)
	}
	if err := repo.SavePlan(ctx, "ws-b", "pro"); err != nil {
		t.Fatalf("SavePlan ws-b: %v", err)
	}

	// A hand-edited layout that a rewrite of ws-b's file would lose.
	otherPlan := filepath.Join(dir, "workspace_plans", "ws-b.json")
	data, err := os.ReadFile(otherPlan)
	if err != nil {
		t.Fatalf("reading ws-b's file: %v", err)
	}
	edited := append([]byte("\n"), data...)
	if err := os.WriteFile(otherPlan, edited, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := repo.SavePlan(ctx, "ws-a", "pro"); err != nil {
		t.Fatalf("SavePlan ws-a again: %v", err)
	}

	got, err := os.ReadFile(otherPlan)
	if err != nil {
		t.Fatalf("reading ws-b's file after saving ws-a: %v", err)
	}
	if string(got) != string(edited) {
		t.Errorf("saving ws-a changed ws-b's file:\n%s\nwant:\n%s", got, edited)
	}
}
