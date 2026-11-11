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
	ada := domain.Account{
		ID:          "account-1",
		Identity:    domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada"},
		WorkspaceID: "workspace-1",
	}

	t.Run("FindByIdentity with no saved account returns ErrNotFound", func(t *testing.T) {
		repo := newRepo(t)
		if _, err := repo.FindByIdentity(ctx, ada.Identity); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("FindByIdentity returns the account with the issuer and subject, whatever the name", func(t *testing.T) {
		repo := newRepo(t)
		if err := repo.Save(ctx, ada); err != nil {
			t.Fatalf("Save: %v", err)
		}
		got, err := repo.FindByIdentity(ctx, domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada Lovelace"})
		if err != nil {
			t.Fatalf("FindByIdentity: %v", err)
		}
		if got != ada {
			t.Fatalf("got %+v, want %+v", got, ada)
		}
	})

	t.Run("FindByIdentity doesn't match an account with the same name but another issuer or subject", func(t *testing.T) {
		repo := newRepo(t)
		if err := repo.Save(ctx, ada); err != nil {
			t.Fatalf("Save: %v", err)
		}
		for _, identity := range []domain.Identity{
			{Issuer: "other", Subject: "ada", Name: "Ada"},
			{Issuer: "dev", Subject: "someone-else", Name: "Ada"},
		} {
			if _, err := repo.FindByIdentity(ctx, identity); !errors.Is(err, domain.ErrNotFound) {
				t.Fatalf("FindByIdentity(%+v): err = %v, want ErrNotFound", identity, err)
			}
		}
	})

	t.Run("Find returns the saved account by its ID", func(t *testing.T) {
		repo := newRepo(t)
		grace := domain.Account{
			ID:          "account-2",
			Identity:    domain.Identity{Issuer: "dev", Subject: "grace", Name: "Grace"},
			WorkspaceID: "workspace-2",
		}
		for _, account := range []domain.Account{ada, grace} {
			if err := repo.Save(ctx, account); err != nil {
				t.Fatalf("Save(%s): %v", account.ID, err)
			}
		}
		got, err := repo.Find(ctx, grace.ID)
		if err != nil {
			t.Fatalf("Find: %v", err)
		}
		if got != grace {
			t.Fatalf("got %+v, want %+v", got, grace)
		}
	})

	t.Run("Find of an unknown ID returns ErrNotFound", func(t *testing.T) {
		repo := newRepo(t)
		if err := repo.Save(ctx, ada); err != nil {
			t.Fatalf("Save: %v", err)
		}
		if _, err := repo.Find(ctx, "account-unknown"); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("Save refuses another account with the same issuer and subject, and keeps the first", func(t *testing.T) {
		repo := newRepo(t)
		if err := repo.Save(ctx, ada); err != nil {
			t.Fatalf("Save: %v", err)
		}
		twin := domain.Account{
			ID:          "account-2",
			Identity:    domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada Lovelace"},
			WorkspaceID: "workspace-2",
		}
		if err := repo.Save(ctx, twin); err == nil {
			t.Fatal("Save of a second account with the same identity succeeded; want an error")
		}
		got, err := repo.FindByIdentity(ctx, ada.Identity)
		if err != nil {
			t.Fatalf("FindByIdentity: %v", err)
		}
		if got != ada {
			t.Fatalf("FindByIdentity: got %+v, want %+v", got, ada)
		}
	})

	t.Run("Save replaces the account with the same ID", func(t *testing.T) {
		repo := newRepo(t)
		renamed := ada
		renamed.Identity.Name = "Ada Lovelace"
		for _, account := range []domain.Account{ada, renamed} {
			if err := repo.Save(ctx, account); err != nil {
				t.Fatalf("Save(%+v): %v", account, err)
			}
		}
		got, err := repo.Find(ctx, ada.ID)
		if err != nil {
			t.Fatalf("Find: %v", err)
		}
		if got != renamed {
			t.Fatalf("Find: got %+v, want %+v", got, renamed)
		}
		got, err = repo.FindByIdentity(ctx, ada.Identity)
		if err != nil {
			t.Fatalf("FindByIdentity: %v", err)
		}
		if got != renamed {
			t.Fatalf("FindByIdentity: got %+v, want %+v", got, renamed)
		}
	})
}
