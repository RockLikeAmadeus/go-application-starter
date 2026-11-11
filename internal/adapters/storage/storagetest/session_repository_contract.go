package storagetest

import (
	"context"
	"errors"
	"testing"
	"time"

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

	t.Run("FindByHash returns the saved session with the hash", func(t *testing.T) {
		repo := newRepo(t)
		adaSession := domain.NewSession("token-1", "account-1", baseTime)
		graceSession := domain.NewSession("token-2", "account-2", baseTime.Add(time.Hour))
		for _, session := range []domain.Session{adaSession, graceSession} {
			if err := repo.Save(ctx, session); err != nil {
				t.Fatalf("Save(%s): %v", session.AccountID, err)
			}
		}
		got, err := repo.FindByHash(ctx, graceSession.TokenHash)
		if err != nil {
			t.Fatalf("FindByHash: %v", err)
		}
		assertSessionEqual(t, got, graceSession)
	})

	t.Run("Save replaces the session with the same token hash", func(t *testing.T) {
		repo := newRepo(t)
		first := domain.NewSession("token-1", "account-1", baseTime)
		replacement := domain.NewSession("token-1", "account-1", baseTime.Add(time.Hour))
		for _, session := range []domain.Session{first, replacement} {
			if err := repo.Save(ctx, session); err != nil {
				t.Fatalf("Save(expires %s): %v", session.ExpiresAt, err)
			}
		}
		got, err := repo.FindByHash(ctx, first.TokenHash)
		if err != nil {
			t.Fatalf("FindByHash: %v", err)
		}
		assertSessionEqual(t, got, replacement)
	})

	t.Run("Delete removes the session, so FindByHash returns ErrNotFound", func(t *testing.T) {
		repo := newRepo(t)
		session := domain.NewSession("token-1", "account-1", baseTime)
		if err := repo.Save(ctx, session); err != nil {
			t.Fatalf("Save: %v", err)
		}
		if err := repo.Delete(ctx, session.TokenHash); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if _, err := repo.FindByHash(ctx, session.TokenHash); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("FindByHash after Delete: err = %v, want ErrNotFound", err)
		}
	})

	t.Run("Delete of a missing session returns ErrNotFound", func(t *testing.T) {
		repo := newRepo(t)
		if err := repo.Delete(ctx, domain.HashSessionToken("token-1")); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}

// assertSessionEqual compares the expiry with Equal so adapters may
// round-trip through formats that drop the monotonic clock or change the
// location.
func assertSessionEqual(t *testing.T, got, want domain.Session) {
	t.Helper()
	if got.TokenHash != want.TokenHash || got.AccountID != want.AccountID || !got.ExpiresAt.Equal(want.ExpiresAt) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
