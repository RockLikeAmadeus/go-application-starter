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

// SessionNotEndedError is what SignOut returns when the server did not confirm
// ending the session but the token was forgotten all the same. Cause is the
// server's error.
type SessionNotEndedError struct{ Cause error }

func (e *SessionNotEndedError) Error() string { return e.Cause.Error() }

func (e *SessionNotEndedError) Unwrap() error { return e.Cause }

// Where a token can be kept, as TokenKept and a TokenStore's Save name it.
const (
	// KeptInCredentialStore is the operating system's credential store.
	KeptInCredentialStore = "credential store"
	// KeptInFile is a file in the settings folder.
	KeptInFile = "file"
)

// Result is what a finished sign-in leaves.
type Result struct {
	// Name is the signed-in account's name.
	Name string
	// TokenKept says where the token was kept: KeptInCredentialStore or
	// KeptInFile.
	TokenKept string
}

// BrowserSignIn signs in through the browser, handing the address to open, and
// returns the session token the server gave this client.
type BrowserSignIn interface {
	BrowserSignIn(ctx context.Context, open func(address string) error) (token string, err error)
}

// TokenStore keeps the session token for the server.
type TokenStore interface {
	// Save keeps the token and says where it was kept.
	Save(token string) (kept string, err error)
	// Forget drops the kept token.
	Forget() error
}

// SessionEnder ends the server's session for the token the client holds.
type SessionEnder interface {
	EndSession(ctx context.Context) error
}

// Client signs in to, and out of, one server.
type Client struct {
	session SessionReader
	ender   SessionEnder
	browser BrowserSignIn
	store   TokenStore
}

// New returns a Client that asks session about the server, signs in through
// browser, and keeps the token in store. The session reader must send the
// token the store keeps, so it names the account once the token is saved.
func New(session SessionReader, ender SessionEnder, browser BrowserSignIn, store TokenStore) *Client {
	return &Client{session: session, ender: ender, browser: browser, store: store}
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
	token, err := c.browser.BrowserSignIn(ctx, open)
	if err != nil {
		return Result{}, err
	}
	kept, err := c.store.Save(token)
	if err != nil {
		return Result{}, err
	}
	signedIn, err := c.session.CurrentSession(ctx)
	if err != nil {
		return Result{}, err
	}
	return Result{Name: signedIn.Name, TokenKept: kept}, nil
}

// CurrentSession reads the server's current session: whether it signs people
// in, and who is signed in.
func (c *Client) CurrentSession(ctx context.Context) (Session, error) {
	return c.session.CurrentSession(ctx)
}

// SignOut ends the server session and forgets the token. It forgets the token
// even when the server can't end the session. When only the server fails, the
// error is a *SessionNotEndedError; when forgetting fails, it holds that error
// too, and is not one.
func (c *Client) SignOut(ctx context.Context) error {
	endErr := c.ender.EndSession(ctx)
	forgetErr := c.store.Forget()
	if forgetErr == nil && endErr != nil {
		return &SessionNotEndedError{Cause: endErr}
	}
	return errors.Join(endErr, forgetErr)
}
