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
}
