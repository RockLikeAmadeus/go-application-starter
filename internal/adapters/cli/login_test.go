package cli_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/zalando/go-keyring"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/cli"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/clientsignin"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/tui"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
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

// Signing out forgets the token of the server the settings name, so with none
// set there is nothing to sign out of.
func TestLogoutIsRefusedWithoutAServer(t *testing.T) {
	want := "starter logout needs a server to sign out of, and none is set. " +
		"Give one with --server <url>, with the STARTER_SERVER_URL environment variable, or with server_url in the settings file."

	out, err := runCLI(t, "--storage", "memory", "logout")

	if err == nil || err.Error() != want {
		t.Errorf("logout error = %v, want %q", err, want)
	}
	if out != "" {
		t.Errorf("logout output = %q, want nothing", out)
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
	newSignIn := func(config.Config) (*clientsignin.Client, func(string) error, error) {
		return clientsignin.New(server, nil, server, nil), func(string) error { return nil }, nil
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

// signInServer plays a server that signs people in: its session names Ada
// once a request carries the token it hands out, its native sign-in goes
// straight back to the command line, and its token endpoint gives the token.
func signInServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/session", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Authorization") == "Bearer the-token" {
			_, _ = w.Write([]byte(`{"isSignInOn":true,"name":"Ada"}`))
			return
		}
		_, _ = w.Write([]byte(`{"isSignInOn":true}`))
	})
	mux.HandleFunc("GET /auth/native", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		http.Redirect(w, r, "http://127.0.0.1:"+query.Get("port")+"/callback?"+url.Values{
			"code": {"the-code"}, "state": {query.Get("state")},
		}.Encode(), http.StatusFound)
	})
	mux.HandleFunc("POST /auth/native/token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token":"the-token"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// A finished sign-in says who is signed in, and where the token is kept, in
// words that differ for the credential store, for a file used because that
// store is unavailable, and for a file chosen with credential_store.
func TestLoginSaysWhoIsSignedInAndWhereTheTokenIsKept(t *testing.T) {
	const lifetime = "and has a fixed lifetime for security purposes. Run starter logout to end the session before it expires."
	tests := []struct {
		name  string
		setup func()
		vars  map[string]string
		kept  string
	}{
		{"credential store", keyring.MockInit, nil,
			"The session token is kept in the operating system's credential store, " + lifetime},
		{"file fallback", func() { keyring.MockInitWithError(errors.New("no secret service")) }, nil,
			"The session token is kept in a file in the settings folder, readable only by you, " +
				"because the operating system's credential store is not available here, " + lifetime},
		{"file on purpose", keyring.MockInit, map[string]string{"STARTER_CREDENTIAL_STORE": "file"},
			"The session token is kept in a file in the settings folder, readable only by you, " + lifetime},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			settings := t.TempDir()
			t.Setenv("XDG_CONFIG_HOME", settings)
			browser := filepath.Join(t.TempDir(), "browser")
			if err := os.WriteFile(browser, []byte("#!/bin/sh\ncurl -s -L \"$1\" >/dev/null\n"), 0o755); err != nil {
				t.Fatal(err)
			}
			t.Setenv("BROWSER", browser)
			srv := signInServer(t)

			out, err := runCLIWith(t, computer(settings, tt.vars), "--storage", "memory", "--server", srv.URL, "login")

			if err != nil {
				t.Fatalf("login: %v", err)
			}
			want := "Signed in to " + srv.URL + " as Ada. " + tt.kept + "\n"
			if out != want {
				t.Errorf("login output = %q, want %q", out, want)
			}
		})
	}
}

// A browser that can't be opened doesn't end the sign-in: the address goes to
// stderr, so it can be opened by hand, and the command keeps waiting for the
// sign-in to finish there.
func TestLoginPrintsTheAddressWhenTheBrowserCannotOpenAndKeepsWaiting(t *testing.T) {
	keyring.MockInit()
	settings := t.TempDir()
	srv := signInServer(t)
	var address string
	newSignIn := func(cfg config.Config) (*clientsignin.Client, func(string) error, error) {
		client, _, err := bootstrap.NewClientSignIn(cfg)
		failing := func(a string) error {
			address = a
			go func() {
				// The user opening the address by hand.
				if response, err := http.Get(a); err == nil {
					response.Body.Close()
				}
			}()
			return errors.New("no browser here")
		}
		return client, failing, err
	}
	var out, errOut bytes.Buffer
	cmd := cli.NewRootCmd(bootstrap.New, bootstrap.NewBilling, newSignIn, computer(settings, nil), &out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"--storage", "memory", "--server", srv.URL, "login"})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := cmd.ExecuteContext(ctx)

	if err != nil {
		t.Fatalf("login: %v", err)
	}
	want := "Could not open a browser. To sign in, open this address in a browser on this computer: " + address + ". " +
		"Waiting for you to finish signing in there. Press Ctrl+C to cancel.\n"
	if errOut.String() != want {
		t.Errorf("login stderr = %q, want %q", errOut.String(), want)
	}
	if !strings.HasPrefix(out.String(), "Signed in to "+srv.URL+" as Ada.") {
		t.Errorf("login output = %q, want it to say Ada is signed in", out.String())
	}
}

// signOutPieces play a server that ends the session, and a store that forgets
// the token, recording that each was asked.
type signOutPieces struct {
	isSessionEnded bool
	isTokenGone    bool
	endError       error
	forgetError    error
}

func (p *signOutPieces) EndSession(context.Context) error { p.isSessionEnded = true; return p.endError }
func (p *signOutPieces) Forget() error                    { p.isTokenGone = true; return p.forgetError }
func (p *signOutPieces) Save(string) (string, error)      { return "", nil }

// Signing out ends the session on the server, forgets the token, and says
// both, naming the server.
func TestLogoutEndsTheSessionForgetsTheTokenAndSaysSo(t *testing.T) {
	pieces := &signOutPieces{}
	newSignIn := func(config.Config) (*clientsignin.Client, func(string) error, error) {
		return clientsignin.New(nil, pieces, nil, pieces), func(string) error { return nil }, nil
	}
	var out bytes.Buffer
	cmd := cli.NewRootCmd(bootstrap.New, bootstrap.NewBilling, newSignIn, computer(t.TempDir(), nil), &out)
	cmd.SetArgs([]string{"--storage", "memory", "--server", "https://starter.example.com", "logout"})

	err := cmd.ExecuteContext(context.Background())

	if err != nil {
		t.Fatalf("logout: %v", err)
	}
	want := "Signed out of https://starter.example.com. The server ended the session, " +
		"and the session token was removed from this computer. Run starter login to sign in again.\n"
	if out.String() != want {
		t.Errorf("logout output = %q, want %q", out.String(), want)
	}
	if !pieces.isSessionEnded || !pieces.isTokenGone {
		t.Errorf("session ended = %v, token gone = %v, want both", pieces.isSessionEnded, pieces.isTokenGone)
	}
}

// When the server fails or can't be reached, the token is still removed, and
// logout says that the session may remain valid on the server, ending with a
// failing status. A token that can't be removed is reported as it is.
func TestLogoutSaysWhenOnlyThisComputerWasSignedOutAndReportsATokenThatStaysAsItIs(t *testing.T) {
	for name, tc := range map[string]struct {
		pieces *signOutPieces
		want   string
	}{
		"server fails": {
			&signOutPieces{endError: errors.New("connection refused")},
			"Signed out of https://starter.example.com on this computer only: the session token was removed from this computer, " +
				"but the server did not confirm that it ended the session (connection refused). " +
				"The session may remain valid on the server until it expires, after a fixed lifetime for security purposes.",
		},
		"server refuses the session": {
			&signOutPieces{endError: domain.ErrNoWorkspace},
			"Signed out of https://starter.example.com on this computer only: the session token was removed from this computer, " +
				"but the server did not confirm that it ended the session (Please sign in to continue.). " +
				"The session may remain valid on the server until it expires, after a fixed lifetime for security purposes.",
		},
		"token stays": {
			&signOutPieces{forgetError: errors.New("keyring locked")},
			"keyring locked",
		},
	} {
		t.Run(name, func(t *testing.T) {
			newSignIn := func(config.Config) (*clientsignin.Client, func(string) error, error) {
				return clientsignin.New(nil, tc.pieces, nil, tc.pieces), func(string) error { return nil }, nil
			}
			var out bytes.Buffer
			cmd := cli.NewRootCmd(bootstrap.New, bootstrap.NewBilling, newSignIn, computer(t.TempDir(), nil), &out)
			cmd.SetArgs([]string{"--storage", "memory", "--server", "https://starter.example.com", "logout"})

			err := cmd.ExecuteContext(context.Background())

			if err == nil || err.Error() != tc.want {
				t.Errorf("logout error = %v, want %q", err, tc.want)
			}
			if out.Len() != 0 {
				t.Errorf("logout output = %q, want nothing", out.String())
			}
			if !tc.pieces.isTokenGone {
				t.Error("logout did not forget the token")
			}
		})
	}
}

// refusedAfterSavingSession reads a session that signs people in, then, once
// the token is saved, is refused for want of a session.
type refusedAfterSavingSession struct{ reads int }

func (s *refusedAfterSavingSession) CurrentSession(context.Context) (clientsignin.Session, error) {
	s.reads++
	if s.reads > 1 {
		return clientsignin.Session{}, domain.ErrNoWorkspace
	}
	return clientsignin.Session{IsSignInOn: true}, nil
}

// browserGivingToken signs in at once, with a token.
type browserGivingToken struct{}

func (browserGivingToken) BrowserSignIn(context.Context, func(string) error) (string, error) {
	return "the-token", nil
}

// A login whose own steps are refused for want of a session reports that
// error as it is, without the advice to run starter login, which the user is
// already doing.
func TestLoginReportsASessionRefusalAsItIsWithoutAdvisingLogin(t *testing.T) {
	session := &refusedAfterSavingSession{}
	pieces := &signOutPieces{}
	newSignIn := func(config.Config) (*clientsignin.Client, func(string) error, error) {
		return clientsignin.New(session, nil, browserGivingToken{}, pieces), func(string) error { return nil }, nil
	}
	var out bytes.Buffer
	cmd := cli.NewRootCmd(bootstrap.New, bootstrap.NewBilling, newSignIn, computer(t.TempDir(), nil), &out)
	cmd.SetArgs([]string{"--storage", "memory", "--server", "https://starter.example.com", "login"})

	err := cmd.ExecuteContext(context.Background())

	if !errors.Is(err, domain.ErrNoWorkspace) || err.Error() != domain.ErrNoWorkspace.Error() {
		t.Errorf("login error = %v, want %q as it is", err, domain.ErrNoWorkspace)
	}
}

// With a server, the terminal app is handed the sign-in the command line uses,
// and the server's address, so it can show the signed-out screen. On local
// data there is no sign-in to hand it.
func TestTheTerminalAppIsHandedTheSignInOnlyWithAServer(t *testing.T) {
	for _, tc := range []struct {
		name       string
		args       []string
		wantSignIn bool
		wantServer string
	}{
		{"a server", []string{"--server", "https://starter.example.com", "tui"}, true, "https://starter.example.com"},
		{"local data", []string{"--storage", "memory", "tui"}, false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got tui.Ports
			defer cli.RunTUIWith(func(p tui.Ports, _ []string, _ ...tea.ProgramOption) error {
				got = p
				return nil
			})()
			server := signInOffServer{}
			newSignIn := func(config.Config) (*clientsignin.Client, func(string) error, error) {
				return clientsignin.New(server, nil, server, nil), func(string) error { return nil }, nil
			}
			cmd := cli.NewRootCmd(bootstrap.New, bootstrap.NewBilling, newSignIn, computer(t.TempDir(), nil), &bytes.Buffer{})
			cmd.SetArgs(tc.args)

			if err := cmd.ExecuteContext(context.Background()); err != nil {
				t.Fatalf("tui: %v", err)
			}
			if (got.SignIn != nil) != tc.wantSignIn || (got.OpenBrowser != nil) != tc.wantSignIn {
				t.Errorf("SignIn set = %t, OpenBrowser set = %t, want both %t",
					got.SignIn != nil, got.OpenBrowser != nil, tc.wantSignIn)
			}
			if got.ServerURL != tc.wantServer {
				t.Errorf("ServerURL = %q, want %q", got.ServerURL, tc.wantServer)
			}
		})
	}
}
