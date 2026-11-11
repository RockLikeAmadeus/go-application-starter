//go:build !devsignin

package bootstrap_test

import (
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

func TestNewEmbedded_RefusesDevSignInWithoutTheDevSignInBuild(t *testing.T) {
	a, err := bootstrap.NewEmbedded(config.Config{Storage: config.StorageMemory, SignIn: config.SignInDev})
	if err == nil {
		_ = a.Close()
	}

	want := `The sign-in setting is "dev", but this program was built without the dev sign-in, which is only for development and tests. Please set the sign-in setting to "off", or build the program with the "devsignin" build tag.`
	if err == nil || err.Error() != want {
		t.Errorf("err = %v, want %q", err, want)
	}
}
