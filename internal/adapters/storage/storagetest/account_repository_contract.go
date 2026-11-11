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

	t.Run("FindByIdentity returns the account with the issuer and subject, whatever the name", func(t *testing.T) {
		repo := newRepo(t)
		saved := domain.Account{
			ID:          "account-1",
			Identity:    domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada"},
			WorkspaceID: "workspace-1",
		}
		if err := repo.Save(ctx, saved); err != nil {
			t.Fatalf("Save: %v", err)
		}
		got, err := repo.FindByIdentity(ctx, domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada Lovelace"})
		if err != nil {
			t.Fatalf("FindByIdentity: %v", err)
		}
		if got != saved {
			t.Fatalf("got %+v, want %+v", got, saved)
		}
	})
}
