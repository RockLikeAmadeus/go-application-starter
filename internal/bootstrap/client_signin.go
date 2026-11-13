package bootstrap

import (
	"errors"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/clientsignin"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// NewClientSignIn builds the client that signs the command line in to the
// server cfg names. It isn't wired yet: task 26c of docs/wip/cli-login.md
// passes in the real pieces.
func NewClientSignIn(config.Config) (*clientsignin.Client, error) {
	return nil, errors.New("signing in is not available yet")
}
