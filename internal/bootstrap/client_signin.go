package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/browsersignin"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/clientsignin"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/tokenstore"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// NewClientSignIn builds the client that signs the command line in to, and
// out of, the server cfg names, keeping the token in the credential store
// cfg picks. It also returns the opener that shows the browser sign-in
// address, so the command line imports no browser adapter.
func NewClientSignIn(cfg config.Config) (*clientsignin.Client, func(address string) error, error) {
	folder, err := config.AppFolder(config.Env{UserConfigDir: os.UserConfigDir})
	if err != nil {
		return nil, nil, fmt.Errorf("find the settings folder: %w", err)
	}
	return NewClientSignInFor(cfg, folder, CommandLine)
}

// Program is the program a session token belongs to. Each keeps its own
// token for a server, so signing in or out of one leaves the others alone.
type Program int

const (
	CommandLine Program = iota
	DesktopOrMobileApp
)

// NewClientSignInFor is NewClientSignIn for a program, keeping its token in
// the settings folder given.
func NewClientSignInFor(cfg config.Config, folder string, program Program) (*clientsignin.Client, func(address string) error, error) {
	session, err := httpclient.NewSessionService(cfg.ServerURL, httpclient.WithTokenSource(tokenSource(cfg, folder, program)))
	if err != nil {
		return nil, nil, fmt.Errorf("create API client: %w", err)
	}
	client := clientsignin.New(
		sessionReader{session},
		session,
		browserSignIn{cfg.ServerURL},
		newServerTokenStore(cfg, folder, program),
	)
	return client, browsersignin.Open, nil
}

// sessionReader adapts httpclient's session to clientsignin's.
type sessionReader struct{ *httpclient.SessionService }

func (r sessionReader) CurrentSession(ctx context.Context) (clientsignin.Session, error) {
	session, err := r.SessionService.CurrentSession(ctx)
	return clientsignin.Session{IsSignInOn: session.IsSignInOn, Name: session.Name}, err
}

// browserSignIn signs in to one server through browsersignin.
type browserSignIn struct{ serverURL string }

func (b browserSignIn) BrowserSignIn(ctx context.Context, isSignUp bool, open func(address string) error) (string, error) {
	attempt, err := browsersignin.Start(ctx, b.serverURL, isSignUp, open)
	if err != nil {
		return "", err
	}
	defer attempt.Close()
	return attempt.Wait(ctx)
}

// serverTokenStore keeps the session token of one server in the credential
// store cfg picks.
type serverTokenStore struct {
	serverURL string
	save      func(serverURL, token string) (string, error)
	forget    func(serverURL string) error
}

func newServerTokenStore(cfg config.Config, folder string, program Program) serverTokenStore {
	if cfg.CredentialStore == config.CredentialStoreFile {
		file := fileStore(folder, program)
		return serverTokenStore{cfg.ServerURL, func(serverURL, token string) (string, error) {
			return clientsignin.KeptInFile, file.Save(serverURL, token)
		}, file.Delete}
	}
	system := systemStore(folder, program)
	return serverTokenStore{cfg.ServerURL, func(serverURL, token string) (string, error) {
		kept, err := system.Save(serverURL, token)
		if kept == tokenstore.InKeyring {
			return clientsignin.KeptInCredentialStore, err
		}
		return clientsignin.KeptInFile, err
	}, system.Delete}
}

func (s serverTokenStore) Save(token string) (string, error) { return s.save(s.serverURL, token) }
func (s serverTokenStore) Forget() error                     { return s.forget(s.serverURL) }

func fileStore(folder string, program Program) *tokenstore.File {
	if program == DesktopOrMobileApp {
		return tokenstore.NewAppFile(folder)
	}
	return tokenstore.NewFile(folder)
}

func systemStore(folder string, program Program) *tokenstore.System {
	if program == DesktopOrMobileApp {
		return tokenstore.NewAppSystem(folder)
	}
	return tokenstore.NewSystem(folder)
}

// NewAppClientSignIn builds the sign-in the desktop and mobile apps' in-process
// handler uses in remote mode, keeping the app's own token in the settings
// folder given.
func NewAppClientSignIn(cfg config.Config, folder string) (httpapi.ClientSignIn, error) {
	client, _, err := NewClientSignInFor(cfg, folder, DesktopOrMobileApp)
	if err != nil {
		return nil, err
	}
	return appSignIn{client}, nil
}

// appSignIn adapts clientsignin's client to the interface httpapi declares.
type appSignIn struct{ client *clientsignin.Client }

func (a appSignIn) CurrentSession(ctx context.Context) (httpapi.ClientSession, error) {
	session, err := a.client.CurrentSession(ctx)
	return httpapi.ClientSession{IsSignInOn: session.IsSignInOn, Name: session.Name}, err
}

func (a appSignIn) SignInThroughBrowser(ctx context.Context, isSignUp bool, open func(address string) error) error {
	signIn := a.client.SignIn
	if isSignUp {
		signIn = a.client.SignUp
	}
	_, err := signIn(ctx, open)
	if errors.Is(err, clientsignin.ErrCancelled) {
		return httpapi.ErrSignInCancelled
	}
	return err
}

func (a appSignIn) CancelBrowserSignIn() { a.client.Cancel() }

func (a appSignIn) SignOut(ctx context.Context) error {
	err := a.client.SignOut(ctx)
	var notEnded *clientsignin.SessionNotEndedError
	if errors.As(err, &notEnded) {
		return fmt.Errorf("%w: %w", httpapi.ErrSessionNotEnded, err)
	}
	return err
}
