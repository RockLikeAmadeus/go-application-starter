// Package clientsignin signs a client application in to a server, and out
// again. It has no Cobra, Bubble Tea, or Wails import and imports no other
// adapter: it declares the small interfaces it needs, and bootstrap passes in
// the concrete adapters. The command line, the terminal app, and later the
// desktop app can all use it.
package clientsignin

import (
	"context"
	"errors"
)

// ErrServerDoesNotSignPeopleIn is what SignIn returns for a server that has
// sign-in off. Each client words it for its own screen.
var ErrServerDoesNotSignPeopleIn = errors.New("the server does not sign people in")

// Session is what the server says about the current session: whether it signs
// people in, and the signed-in account's name, empty when no one is.
type Session struct {
	IsSignInOn bool
	Name       string
}

// SessionReader reads the server's current session.
type SessionReader interface {
	CurrentSession(ctx context.Context) (Session, error)
}

// Result is what a finished sign-in leaves.
type Result struct{}

// Client signs in to, and out of, one server.
type Client struct {
	session SessionReader
}

// New returns a Client that asks session about the server.
func New(session SessionReader) *Client {
	return &Client{session: session}
}

// SignIn signs in to the server, handing the browser's address to open.
func (c *Client) SignIn(ctx context.Context, open func(address string) error) (Result, error) {
	session, err := c.session.CurrentSession(ctx)
	if err != nil {
		return Result{}, err
	}
	if !session.IsSignInOn {
		return Result{}, ErrServerDoesNotSignPeopleIn
	}
	return Result{}, nil
}
