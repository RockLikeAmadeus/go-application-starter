package cli_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/cli"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/clientsignin"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// Signing in stores a token only: it never picks the server, so with none set
// there is nothing to sign in to, however the command line is otherwise set up.
func TestLoginIsRefusedWithoutAServer(t *testing.T) {
	want := "starter login needs a server to sign in to, and none is set. " +
		"Give one with --server <url>, with the STARTER_SERVER_URL environment variable, or with server_url in the settings file. " +
		"Signing in stores a session token only; it never changes which server or data the command line uses, " +
		"so the same server must be set whenever you run starter afterwards."

	out, err := runCLI(t, "--storage", "memory", "login")

	if err == nil || err.Error() != want {
		t.Errorf("login error = %v, want %q", err, want)
	}
	if out != "" {
		t.Errorf("login output = %q, want nothing", out)
	}
}

// signInOffServer plays a server with sign-in off: its session says so, and
// nothing else may be asked of it.
type signInOffServer struct{}

func (signInOffServer) BrowserSignIn(context.Context, func(string) error) (string, error) {
	return "", errors.New("no browser sign-in is expected")
}

func (signInOffServer) CurrentSession(context.Context) (clientsignin.Session, error) {
	return clientsignin.Session{}, nil
}

// A server that doesn't sign people in has nothing to sign in to, and the
// refusal names the server, so a mistyped address is easy to spot.
func TestLoginIsRefusedByAServerThatDoesNotSignPeopleIn(t *testing.T) {
	server := signInOffServer{}
	newSignIn := func(config.Config) (*clientsignin.Client, error) {
		return clientsignin.New(server, nil, server, nil), nil
	}
	var out bytes.Buffer
	cmd := cli.NewRootCmd(bootstrap.New, bootstrap.NewBilling, newSignIn, computer(t.TempDir(), nil), &out)
	cmd.SetArgs([]string{"--storage", "memory", "--server", "https://starter.example.com", "login"})

	err := cmd.ExecuteContext(context.Background())

	want := "The server at https://starter.example.com does not support sign-in, so there is nothing to sign in to. " +
		"No browser was opened, and no session token was stored."
	if err == nil || err.Error() != want {
		t.Errorf("login error = %v, want %q", err, want)
	}
	if out.Len() != 0 {
		t.Errorf("login output = %q, want nothing", out.String())
	}
}
