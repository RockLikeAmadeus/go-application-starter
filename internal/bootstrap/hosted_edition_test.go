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
