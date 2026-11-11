//go:build devsignin

package bootstrap_test

// The tests in this file wire the hosted edition, which needs sign-in on, so
// they carry the devsignin tag, as hosted_edition_test.go explains. They use
// the sample Item API, so `task init --remove-sample` deletes the file (see
// scripts/initproject/manifest.go).

import (
	"context"
	"errors"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

func TestNewEmbedded_WithSignInOnListsNoItemsWithoutAWorkspaceInTheContext(t *testing.T) {
	a, err := bootstrap.NewEmbedded(config.Config{Storage: config.StorageMemory, SignIn: config.SignInDev})
	if err != nil {
		t.Fatalf("NewEmbedded: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })

	_, err = a.Items.ListItems(context.Background())

	if !errors.Is(err, domain.ErrNoWorkspace) {
		t.Errorf("ListItems error = %v, want ErrNoWorkspace", err)
	}
}
