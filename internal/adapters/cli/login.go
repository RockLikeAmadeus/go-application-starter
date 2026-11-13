package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/clientsignin"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// errLoginNeedsAServer is what `starter login` says when no server is set:
// signing in stores a token only, so the server must come from the settings.
var errLoginNeedsAServer = errors.New("starter login needs a server to sign in to, and none is set. " +
	"Give one with --server <url>, with the STARTER_SERVER_URL environment variable, or with server_url in the settings file. " +
	"Signing in stores a session token only; it never changes which server or data the command line uses, " +
	"so the same server must be set whenever you run starter afterwards.")

// newLoginCmd signs in to the server the settings name.
func newLoginCmd(loadConfig func() (config.Config, error), newSignIn SignInFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Sign in to the server in your browser",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			if cfg.ServerURL == "" {
				return errLoginNeedsAServer
			}
			client, err := newSignIn(cfg)
			if err != nil {
				return err
			}
			_, err = client.SignIn(cmd.Context(), func(string) error { return nil })
			if errors.Is(err, clientsignin.ErrServerDoesNotSignPeopleIn) {
				return fmt.Errorf("The server at %s does not support sign-in, so there is nothing to sign in to. "+
					"No browser was opened, and no session token was stored.", cfg.ServerURL)
			}
			return err
		},
	}
}
