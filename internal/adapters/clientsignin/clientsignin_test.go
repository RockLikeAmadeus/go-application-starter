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
}

func (f *fakeServer) CurrentSession(context.Context) (clientsignin.Session, error) {
	if f.savedToken == "" {
		return clientsignin.Session{IsSignInOn: true}, nil
	}
	return clientsignin.Session{IsSignInOn: true, Name: f.name}, nil
}

func (f *fakeServer) BrowserSignIn(context.Context, func(string) error) (string, error) {
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
	client := clientsignin.New(fakeSession{clientsignin.Session{IsSignInOn: false}}, &fakeServer{}, &fakeServer{})
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
	client := clientsignin.New(server, server, server)

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
