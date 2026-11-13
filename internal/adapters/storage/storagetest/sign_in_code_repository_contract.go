package storagetest

import (
	"context"
	"errors"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// RunSignInCodeRepositoryContract is the contract every
// ports.SignInCodeRepository implementation must pass. newRepo returns a
// fresh, empty repository for each subtest.
func RunSignInCodeRepositoryContract(t *testing.T, newRepo func(*testing.T) ports.SignInCodeRepository) {
	ctx := context.Background()
	newCode := func(code, accountID string) domain.SignInCode {
		issued := domain.NewSignInCode(accountID, "challenge-"+code, baseTime)
		issued.CodeHash = domain.HashSignInCode(code)
		return issued
	}

	t.Run("Take with no saved code returns ErrNotFound", func(t *testing.T) {
		repo := newRepo(t)
		if _, err := repo.Take(ctx, domain.HashSignInCode("code-1")); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("Take returns the saved code with the hash", func(t *testing.T) {
		repo := newRepo(t)
		ada, grace := newCode("code-1", "account-1"), newCode("code-2", "account-2")
		for _, code := range []domain.SignInCode{ada, grace} {
			if err := repo.Save(ctx, code); err != nil {
				t.Fatalf("Save(%s): %v", code.AccountID, err)
			}
		}
		got, err := repo.Take(ctx, grace.CodeHash)
		if err != nil {
			t.Fatalf("Take: %v", err)
		}
		assertSignInCodeEqual(t, got, grace)
	})

	t.Run("Take deletes the code, so a second Take returns ErrNotFound", func(t *testing.T) {
		repo := newRepo(t)
		code := newCode("code-1", "account-1")
		if err := repo.Save(ctx, code); err != nil {
			t.Fatalf("Save: %v", err)
		}
		if _, err := repo.Take(ctx, code.CodeHash); err != nil {
			t.Fatalf("first Take: %v", err)
		}
		if _, err := repo.Take(ctx, code.CodeHash); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("second Take: err = %v, want ErrNotFound", err)
		}
	})
}

// assertSignInCodeEqual compares the expiry with Equal so adapters may
// round-trip through formats that drop the monotonic clock or change the
// location.
func assertSignInCodeEqual(t *testing.T, got, want domain.SignInCode) {
	t.Helper()
	same := got.CodeHash == want.CodeHash &&
		got.AccountID == want.AccountID &&
		got.Challenge == want.Challenge &&
		got.ExpiresAt.Equal(want.ExpiresAt)
	if !same {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
