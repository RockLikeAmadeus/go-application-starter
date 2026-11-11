package storagetest

import (
	"context"
	"errors"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// RunAccountRepositoryContract is the contract every ports.AccountRepository
// implementation must pass. newRepo returns a fresh, empty repository for
// each subtest.
func RunAccountRepositoryContract(t *testing.T, newRepo func(*testing.T) ports.AccountRepository) {
	ctx := context.Background()

	t.Run("FindByIdentity with no saved account returns ErrNotFound", func(t *testing.T) {
		repo := newRepo(t)
		identity := domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada"}
		if _, err := repo.FindByIdentity(ctx, identity); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}
