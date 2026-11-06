package storagetest

import (
	"context"
	"errors"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// RunWorkspacePlanRepositoryContract is the contract every
// ports.WorkspacePlanRepository implementation must pass. newRepo returns a
// fresh, empty repository for each subtest.
func RunWorkspacePlanRepositoryContract(t *testing.T, newRepo func(*testing.T) ports.WorkspacePlanRepository) {
	ctx := context.Background()
	const mine domain.WorkspaceID = "ws-a"

	t.Run("PlanOf with no stored plan returns ErrNotFound", func(t *testing.T) {
		repo := newRepo(t)
		if _, err := repo.PlanOf(ctx, mine); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("PlanOf returns the saved plan", func(t *testing.T) {
		repo := newRepo(t)
		mustSavePlan(t, repo, mine, "pro")
		assertPlan(t, repo, mine, "pro")
	})

	t.Run("SavePlan replaces the saved plan", func(t *testing.T) {
		repo := newRepo(t)
		mustSavePlan(t, repo, mine, "pro")
		mustSavePlan(t, repo, mine, "free")
		assertPlan(t, repo, mine, "free")
	})
}

func mustSavePlan(t *testing.T, repo ports.WorkspacePlanRepository, workspace domain.WorkspaceID, plan string) {
	t.Helper()
	if err := repo.SavePlan(context.Background(), workspace, plan); err != nil {
		t.Fatalf("SavePlan(%q, %q): %v", workspace, plan, err)
	}
}

func assertPlan(t *testing.T, repo ports.WorkspacePlanRepository, workspace domain.WorkspaceID, want string) {
	t.Helper()
	got, err := repo.PlanOf(context.Background(), workspace)
	if err != nil {
		t.Fatalf("PlanOf(%q): %v", workspace, err)
	}
	if got != want {
		t.Errorf("PlanOf(%q) = %q, want %q", workspace, got, want)
	}
}
