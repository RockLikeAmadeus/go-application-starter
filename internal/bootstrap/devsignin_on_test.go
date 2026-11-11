//go:build devsignin

package bootstrap_test

import (
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

func TestNewEmbedded_TurnsOnTheDevSignInInTheDevSignInBuild(t *testing.T) {
	a, err := bootstrap.NewEmbedded(config.Config{Storage: config.StorageMemory, SignIn: config.SignInDev})
	if err != nil {
		t.Fatalf("NewEmbedded: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })

	if !a.IsDevSignInOn {
		t.Error("IsDevSignInOn = false, want true")
	}
}
