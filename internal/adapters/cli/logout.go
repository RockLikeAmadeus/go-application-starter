package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// errLogoutNeedsAServer is what `starter logout` says when no server is set:
// the stored session token belongs to one server, which the settings name.
var errLogoutNeedsAServer = errors.New("starter logout needs a server to sign out of, and none is set. " +
	"Give one with --server <url>, with the STARTER_SERVER_URL environment variable, or with server_url in the settings file.")

// newLogoutCmd signs out of the server the settings name.
func newLogoutCmd(loadConfig func() (config.Config, error), newSignIn SignInFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Sign out of the server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			if cfg.ServerURL == "" {
				return errLogoutNeedsAServer
			}
			client, _, err := newSignIn(cfg)
			if err != nil {
				return err
			}
			if err := client.SignOut(cmd.Context()); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Signed out of %s. The server ended the session, "+
				"and the session token was removed from this computer. Run starter login to sign in again.\n", cfg.ServerURL)
			return err
		},
	}
}
