package cli

import (
	"errors"
	"fmt"
	"io"

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
			client, open, err := newSignIn(cfg)
			if err != nil {
				return err
			}
			result, err := client.SignIn(cmd.Context(), openOrShowAddress(cmd.ErrOrStderr(), open))
			if errors.Is(err, clientsignin.ErrServerDoesNotSignPeopleIn) {
				return fmt.Errorf("The server at %s does not support sign-in, so there is nothing to sign in to. "+
					"No browser was opened, and no session token was stored.", cfg.ServerURL)
			}
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Signed in to %s as %s. %s\n", cfg.ServerURL, result.Name, keptSentence(cfg, result))
			return err
		},
	}
}

// keptSentence says where the session token is kept, and that it expires.
func keptSentence(cfg config.Config, result clientsignin.Result) string {
	const expiry = "and has a fixed lifetime for security purposes. Run starter logout to end the session before it expires."
	switch {
	case result.TokenKept == clientsignin.KeptInCredentialStore:
		return "The session token is kept in the operating system's credential store, " + expiry
	case cfg.CredentialStore == config.CredentialStoreFile:
		return "The session token is kept in a file in the settings folder, readable only by you, " + expiry
	default:
		return "The session token is kept in a file in the settings folder, readable only by you, " +
			"because the operating system's credential store is not available here, " + expiry
	}
}

// openOrShowAddress wraps open so that a browser that can't be opened doesn't
// end the sign-in: the address goes to stderr to be opened by hand, and the
// sign-in carries on waiting.
func openOrShowAddress(stderr io.Writer, open func(address string) error) func(address string) error {
	return func(address string) error {
		if open(address) != nil {
			fmt.Fprintf(stderr, "Could not open a browser. To sign in, open this address in a browser on this computer: %s. "+
				"Waiting for you to finish signing in there. Press Ctrl+C to cancel.\n", address)
		}
		return nil
	}
}
