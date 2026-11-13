package bootstrap

import (
	"context"
	"fmt"
	"os"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/browsersignin"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/clientsignin"
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
	session, err := httpclient.NewSessionService(cfg.ServerURL, httpclient.WithTokenSource(tokenSource(cfg, folder)))
	if err != nil {
		return nil, nil, fmt.Errorf("create API client: %w", err)
	}
	return clientsignin.New(sessionReader{session}, session, browserSignIn{cfg.ServerURL}, newServerTokenStore(cfg, folder)), browsersignin.Open, nil
}

// sessionReader adapts httpclient's session to clientsignin's.
type sessionReader struct{ *httpclient.SessionService }

func (r sessionReader) CurrentSession(ctx context.Context) (clientsignin.Session, error) {
	session, err := r.SessionService.CurrentSession(ctx)
	return clientsignin.Session{IsSignInOn: session.IsSignInOn, Name: session.Name}, err
}

// browserSignIn signs in to one server through browsersignin.
type browserSignIn struct{ serverURL string }

func (b browserSignIn) BrowserSignIn(ctx context.Context, open func(address string) error) (string, error) {
	attempt, err := browsersignin.Start(ctx, b.serverURL, open)
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

func newServerTokenStore(cfg config.Config, folder string) serverTokenStore {
	if cfg.CredentialStore == config.CredentialStoreFile {
		file := tokenstore.NewFile(folder)
		return serverTokenStore{cfg.ServerURL, func(serverURL, token string) (string, error) {
			return clientsignin.KeptInFile, file.Save(serverURL, token)
		}, file.Delete}
	}
	system := tokenstore.NewSystem(folder)
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
