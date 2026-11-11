package acceptance

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient/apiclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/system"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/workspace"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/app"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
	"github.com/RockLikeAmadeus/go-application-starter/internal/testsupport"
)

// signInDriver is how a scenario signs in and out on a driver running the
// hosted edition. Each signs in with the dev sign-in's identity for a name.
type signInDriver interface {
	// signIn signs in as name, and returns ctx working as that account.
	signIn(ctx context.Context, name string) (context.Context, error)
	// signOut ends the current session, and returns ctx working as no one.
	signOut(ctx context.Context) (context.Context, error)
	// signedInName is the name of the account ctx works as.
	signedInName(ctx context.Context) (string, error)
	// signedOut returns ctx working as someone who hasn't signed in.
	signedOut(ctx context.Context) (context.Context, error)
}

// devIdentity is the identity the dev sign-in gives name.
func devIdentity(name string) (domain.Identity, error) {
	return domain.NewIdentity(httpapi.DevIssuer, name, name)
}

// newSignInService builds the sign-in use cases on accounts and sessions,
// driven by the harness's clock.
func (h *harness) newSignInService(accounts ports.AccountRepository, sessions ports.SessionRepository) *app.SignInService {
	return app.NewSignInService(accounts, sessions, system.Tokens{}, testsupport.NewSequentialIDs("account"), h.clock)
}

// useDirectSignIn signs scenarios in by calling the sign-in use cases on
// accounts and sessions directly, as the dev sign-in would.
func (h *harness) useDirectSignIn(accounts ports.AccountRepository, sessions ports.SessionRepository) {
	h.accounts = accounts
	h.signIn = &directSignIn{service: h.newSignInService(accounts, sessions)}
}

// directSignIn calls ports.SignInService and keeps the current session's
// token. Like the server's session middleware, it works in the workspace of
// the token's account, or in none when the token has no session.
type directSignIn struct {
	service ports.SignInService
	token   string
}

func (d *directSignIn) signIn(ctx context.Context, name string) (context.Context, error) {
	identity, err := devIdentity(name)
	if err != nil {
		return ctx, err
	}
	if d.token, err = d.service.SignIn(ctx, identity); err != nil {
		return ctx, err
	}
	return d.inSessionsWorkspace(ctx)
}

func (d *directSignIn) signOut(ctx context.Context) (context.Context, error) {
	if err := d.service.SignOut(ctx, d.token); err != nil {
		return ctx, err
	}
	return d.inSessionsWorkspace(ctx)
}

func (d *directSignIn) signedInName(ctx context.Context) (string, error) {
	account, err := d.service.CurrentAccount(ctx, d.token)
	if err != nil {
		return "", err
	}
	return account.Identity.Name, nil
}

func (d *directSignIn) signedOut(ctx context.Context) (context.Context, error) {
	d.token = ""
	return d.inSessionsWorkspace(ctx)
}

// inSessionsWorkspace returns ctx in the workspace of the token's account,
// or in none when the token has no session.
func (d *directSignIn) inSessionsWorkspace(ctx context.Context) (context.Context, error) {
	account, err := d.service.CurrentAccount(ctx, d.token)
	if errors.Is(err, domain.ErrNoWorkspace) {
		return workspace.WithID(ctx, ""), nil
	}
	if err != nil {
		return ctx, err
	}
	return workspace.WithID(ctx, account.WorkspaceID), nil
}

// devSignInOverHTTP signs in through the server's dev sign-in form, as a
// browser does, keeping each account's session cookie in its own cookie
// jar. The harness's driving ports send the current account's cookie.
type devSignInOverHTTP struct {
	h       *harness
	baseURL string
	// browsers holds a client, with its own cookie jar, for each name
	// signed in as.
	browsers map[string]*http.Client
	current  *http.Client
}

func newDevSignInOverHTTP(h *harness, baseURL string) *devSignInOverHTTP {
	return &devSignInOverHTTP{h: h, baseURL: baseURL, browsers: map[string]*http.Client{}, current: newBrowser(nil)}
}

// newBrowser returns a client keeping cookies in jar, or none when jar is
// nil, that doesn't follow redirects, so a step sees where it was sent.
func newBrowser(jar http.CookieJar) *http.Client {
	return &http.Client{
		Jar:           jar,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

func (d *devSignInOverHTTP) signIn(ctx context.Context, name string) (context.Context, error) {
	browser, ok := d.browsers[name]
	if !ok {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return ctx, err
		}
		browser = newBrowser(jar)
		d.browsers[name] = browser
	}
	if err := d.post(browser, "/auth/dev", url.Values{"name": {name}}); err != nil {
		return ctx, err
	}
	return ctx, d.use(browser)
}

func (d *devSignInOverHTTP) signOut(ctx context.Context) (context.Context, error) {
	return ctx, d.post(d.current, "/auth/sign-out", nil)
}

func (d *devSignInOverHTTP) signedInName(ctx context.Context) (string, error) {
	sessions, err := httpclient.NewSessionService(d.baseURL, apiclient.WithHTTPClient(d.current))
	if err != nil {
		return "", err
	}
	session, err := sessions.CurrentSession(ctx)
	if err != nil {
		return "", err
	}
	return session.Name, nil
}

func (d *devSignInOverHTTP) signedOut(ctx context.Context) (context.Context, error) {
	return ctx, d.use(newBrowser(nil))
}

// post posts form to path with browser, which the server answers by sending
// it on to the web app.
func (d *devSignInOverHTTP) post(browser *http.Client, path string, form url.Values) error {
	resp, err := browser.PostForm(d.baseURL+path, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		return fmt.Errorf("POST %s: status %d, want %d", path, resp.StatusCode, http.StatusSeeOther)
	}
	return nil
}

// use points the harness's driving ports at the server through browser.
func (d *devSignInOverHTTP) use(browser *http.Client) error {
	d.current = browser
	return d.h.driveOverHTTP(d.baseURL, apiclient.WithHTTPClient(browser))
}
