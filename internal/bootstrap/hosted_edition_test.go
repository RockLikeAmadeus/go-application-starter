//go:build devsignin

package bootstrap_test

// The tests in this file wire the hosted edition, which needs sign-in on. Only
// the dev sign-in exists so far, and only a devsignin build accepts it, so
// they carry that tag; `task test:go` runs them.

import (
	"context"
	"errors"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

func TestNewEmbedded_WithSignInOnHasNoWorkspaceWithoutOneInTheContext(t *testing.T) {
	a, err := bootstrap.NewEmbedded(config.Config{Storage: config.StorageMemory, SignIn: config.SignInDev})
	if err != nil {
		t.Fatalf("NewEmbedded: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })

	_, err = a.Entitlements.CurrentEntitlements(context.Background())

	if !errors.Is(err, domain.ErrNoWorkspace) {
		t.Errorf("CurrentEntitlements error = %v, want ErrNoWorkspace", err)
	}
}

func TestNewEmbedded_WithSignInOnSignsInAndFindsTheSessionsAccount(t *testing.T) {
	a, err := bootstrap.NewEmbedded(config.Config{Storage: config.StorageMemory, SignIn: config.SignInDev})
	if err != nil {
		t.Fatalf("NewEmbedded: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })
	if a.SignIn == nil {
		t.Fatal("SignIn = nil, want the sign-in service")
	}
	ctx := context.Background()
	identity := domain.Identity{Issuer: "dev", Subject: "Alpha", Name: "Alpha"}

	token, err := a.SignIn.SignIn(ctx, identity)
	if err != nil {
		t.Fatalf("SignIn: %v", err)
	}
	account, err := a.SignIn.CurrentAccount(ctx, token)

	if err != nil {
		t.Fatalf("CurrentAccount: %v", err)
	}
	if account.Identity != identity {
		t.Errorf("account's identity = %+v, want %+v", account.Identity, identity)
	}
}

// A session made on SQLite still signs its account in for the next program to
// open the same data folder.
func TestNewEmbedded_OnSQLiteASessionOutlivesTheAppThatMadeIt(t *testing.T) {
	assertSessionOutlivesTheAppThatMadeIt(t, config.Config{Storage: config.StorageSQLite, DataDir: t.TempDir(), SignIn: config.SignInDev})
}

func assertSessionOutlivesTheAppThatMadeIt(t *testing.T, cfg config.Config) {
	t.Helper()
	identity := domain.Identity{Issuer: "dev", Subject: "Alpha", Name: "Alpha"}

	first, err := bootstrap.NewEmbedded(cfg)
	if err != nil {
		t.Fatalf("first NewEmbedded: %v", err)
	}
	token, err := first.SignIn.SignIn(t.Context(), identity)
	if err != nil {
		t.Fatalf("SignIn: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	second, err := bootstrap.NewEmbedded(cfg)
	if err != nil {
		t.Fatalf("second NewEmbedded: %v", err)
	}
	t.Cleanup(func() { _ = second.Close() })
	account, err := second.SignIn.CurrentAccount(t.Context(), token)

	if err != nil {
		t.Fatalf("CurrentAccount after reopening: %v", err)
	}
	if account.Identity != identity {
		t.Errorf("account's identity after reopening = %+v, want %+v", account.Identity, identity)
	}
}
