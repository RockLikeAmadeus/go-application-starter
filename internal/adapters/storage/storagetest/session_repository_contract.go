package storagetest

import (
	"context"
	"errors"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// RunSessionRepositoryContract is the contract every ports.SessionRepository
// implementation must pass. newRepo returns a fresh, empty repository for
// each subtest.
func RunSessionRepositoryContract(t *testing.T, newRepo func(*testing.T) ports.SessionRepository) {
	ctx := context.Background()

	t.Run("FindByHash with no saved session returns ErrNotFound", func(t *testing.T) {
		repo := newRepo(t)
		if _, err := repo.FindByHash(ctx, domain.HashSessionToken("token-1")); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}
