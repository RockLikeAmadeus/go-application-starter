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

// A server that doesn't sign people in can't be signed in to: SignIn says so
// with a named error, which each client words for its own screen, and opens
// no browser.
func TestSignIn_RefusesAServerThatDoesNotSignPeopleIn(t *testing.T) {
	client := clientsignin.New(fakeSession{clientsignin.Session{IsSignInOn: false}})
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
