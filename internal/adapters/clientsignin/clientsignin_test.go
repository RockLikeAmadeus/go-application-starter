package clientsignin_test

import (
	"context"
	"errors"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/clientsignin"
)

type fakeSession struct{ session clientsignin.Session }

func (f fakeSession) CurrentSession(context.Context) (clientsignin.Session, error) {
	return f.session, nil
}

// fakeServer plays the server's session, which names the account once a
// token is saved, the browser sign-in, and the token store.
type fakeServer struct {
	name       string
	token      string
	kept       string
	savedToken string

	endError     error
	wasEnded     bool
	wasForgotten bool
	wasSignUp    bool
	forgetError  error
}

func (f *fakeServer) EndSession(context.Context) error {
	f.wasEnded = true
	return f.endError
}

func (f *fakeServer) Forget() error {
	f.wasForgotten = true
	return f.forgetError
}

func (f *fakeServer) CurrentSession(context.Context) (clientsignin.Session, error) {
	if f.savedToken == "" {
		return clientsignin.Session{IsSignInOn: true}, nil
	}
	return clientsignin.Session{IsSignInOn: true, Name: f.name}, nil
}

func (f *fakeServer) BrowserSignIn(_ context.Context, isSignUp bool, _ func(string) error) (string, error) {
	f.wasSignUp = isSignUp
	return f.token, nil
}

func (f *fakeServer) Save(token string) (string, error) {
	f.savedToken = token
	return f.kept, nil
}

// A server that doesn't sign people in can't be signed in to: SignIn says so
// with a named error, which each client words for its own screen, and opens
// no browser.
func TestSignIn_RefusesAServerThatDoesNotSignPeopleIn(t *testing.T) {
	client := clientsignin.New(fakeSession{clientsignin.Session{IsSignInOn: false}}, &fakeServer{}, &fakeServer{}, &fakeServer{})
	wasOpened := false
	open := func(string) error {
		wasOpened = true
		return nil
	}

	_, err := client.SignIn(context.Background(), open)

	if !errors.Is(err, clientsignin.ErrServerDoesNotSignPeopleIn) {
		t.Errorf("SignIn error = %v, want ErrServerDoesNotSignPeopleIn", err)
	}
	if wasOpened {
		t.Error("SignIn opened the browser for a server that doesn't sign people in")
	}
}

// Signing in runs the browser sign-in, saves the token it returns, and reports
// the account's name (read with the saved token) and where the token was kept.
func TestSignIn_SavesTheBrowsersTokenAndReportsTheAccountAndWhereItWasKept(t *testing.T) {
	server := &fakeServer{name: "Ada", token: "tok-1", kept: "file"}
	client := clientsignin.New(server, server, server, server)

	result, err := client.SignIn(context.Background(), func(string) error { return nil })

	if err != nil {
		t.Fatalf("SignIn error = %v", err)
	}
	if server.savedToken != "tok-1" {
		t.Errorf("saved token = %q, want tok-1", server.savedToken)
	}
	if result.Name != "Ada" || result.TokenKept != "file" {
		t.Errorf("result = %+v, want name Ada and token kept in file", result)
	}
}

// Creating an account goes through the browser's sign-up, then keeps the token
// and reports the account as SignIn does.
func TestSignUp_SignsUpThroughTheBrowserAndKeepsTheToken(t *testing.T) {
	server := &fakeServer{name: "Ada", token: "tok-2", kept: "file"}
	client := clientsignin.New(server, server, server, server)

	result, err := client.SignUp(context.Background(), func(string) error { return nil })

	if err != nil {
		t.Fatalf("SignUp error = %v", err)
	}
	if !server.wasSignUp {
		t.Error("SignUp did not ask the browser for sign-up")
	}
	if server.savedToken != "tok-2" {
		t.Errorf("saved token = %q, want tok-2", server.savedToken)
	}
	if result.Name != "Ada" || result.TokenKept != "file" {
		t.Errorf("result = %+v, want name Ada and token kept in file", result)
	}
}

// Signing out ends the server's session and forgets the token, and forgets
// the token even when the server refuses or can't be reached, so the client
// never keeps a token it was told to drop. The server's error is still
// returned, for the caller to word.
func TestSignOut_EndsTheServerSessionAndForgetsTheTokenEvenWhenTheServerFails(t *testing.T) {
	serverDown := errors.New("server unreachable")
	for name, endError := range map[string]error{"server answers": nil, "server fails": serverDown} {
		t.Run(name, func(t *testing.T) {
			server := &fakeServer{endError: endError}
			client := clientsignin.New(server, server, server, server)

			err := client.SignOut(context.Background())

			if !server.wasEnded {
				t.Error("SignOut did not end the server session")
			}
			if !server.wasForgotten {
				t.Error("SignOut did not forget the token")
			}
			if !errors.Is(err, endError) {
				t.Errorf("SignOut error = %v, want %v", err, endError)
			}
		})
	}
}

// When the server fails but the token is forgotten, SignOut says so with a
// named error, so a client can tell it from a token that couldn't be
// forgotten, which is not that error.
func TestSignOut_TellsAServerFailureWithTheTokenForgottenFromAForgetFailure(t *testing.T) {
	serverDown := errors.New("server unreachable")
	locked := errors.New("credential store locked")
	for name, tc := range map[string]struct {
		server  *fakeServer
		wantNot bool
	}{
		"token forgotten":       {&fakeServer{endError: serverDown}, false},
		"token not forgotten":   {&fakeServer{endError: serverDown, forgetError: locked}, true},
		"only forgetting fails": {&fakeServer{forgetError: locked}, true},
	} {
		t.Run(name, func(t *testing.T) {
			client := clientsignin.New(tc.server, tc.server, tc.server, tc.server)

			err := client.SignOut(context.Background())

			var notEnded *clientsignin.SessionNotEndedError
			if got := errors.As(err, &notEnded); got == tc.wantNot {
				t.Errorf("SignOut error = %v, is SessionNotEndedError = %v, want %v", err, got, !tc.wantNot)
			}
			if !tc.wantNot && !errors.Is(err, serverDown) {
				t.Errorf("SignOut error = %v, want it to wrap the server's error", err)
			}
		})
	}
}
