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
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/oidc/oidctest"
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
	// identity is the identity that signing in as name gives the account.
	identity(name string) (domain.Identity, error)
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

func (d *directSignIn) identity(name string) (domain.Identity, error) { return devIdentity(name) }

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

// browsersOverHTTP keeps each account's session cookie in its own cookie
// jar, as separate browsers do, and points the harness's driving ports at
// the server through the current one. A way of signing in over HTTP embeds
// it and adds signIn and identity.
type browsersOverHTTP struct {
	h       *harness
	baseURL string
	// browsers holds a client, with its own cookie jar, for each name
	// signed in as.
	browsers map[string]*http.Client
	current  *http.Client
}

func newBrowsersOverHTTP(h *harness, baseURL string) browsersOverHTTP {
	return browsersOverHTTP{h: h, baseURL: baseURL, browsers: map[string]*http.Client{}, current: newBrowser(nil)}
}

// newBrowser returns a client keeping cookies in jar, or none when jar is
// nil, that doesn't follow redirects, so a step sees where it was sent.
func newBrowser(jar http.CookieJar) *http.Client {
	return &http.Client{
		Jar:           jar,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

// browserFor returns the browser that signs in as name, with a new cookie
// jar the first time.
func (b *browsersOverHTTP) browserFor(name string) (*http.Client, error) {
	browser, ok := b.browsers[name]
	if !ok {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, err
		}
		browser = newBrowser(jar)
		b.browsers[name] = browser
	}
	return browser, nil
}

func (b *browsersOverHTTP) signOut(ctx context.Context) (context.Context, error) {
	resp, err := b.current.Post(b.baseURL+"/auth/sign-out", "", nil)
	if err != nil {
		return ctx, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		return ctx, fmt.Errorf("POST /auth/sign-out: status %d, want %d", resp.StatusCode, http.StatusSeeOther)
	}
	return ctx, nil
}

func (b *browsersOverHTTP) signedInName(ctx context.Context) (string, error) {
	sessions, err := httpclient.NewSessionService(b.baseURL, apiclient.WithHTTPClient(b.current))
	if err != nil {
		return "", err
	}
	session, err := sessions.CurrentSession(ctx)
	if err != nil {
		return "", err
	}
	return session.Name, nil
}

func (b *browsersOverHTTP) signedOut(ctx context.Context) (context.Context, error) {
	return ctx, b.use(newBrowser(nil))
}

// use points the harness's driving ports at the server through browser.
func (b *browsersOverHTTP) use(browser *http.Client) error {
	b.current = browser
	return b.h.driveOverHTTP(b.baseURL, apiclient.WithHTTPClient(browser))
}

// devSignInOverHTTP signs in through the server's dev sign-in form, as a
// browser does.
type devSignInOverHTTP struct {
	browsersOverHTTP
}

func newDevSignInOverHTTP(h *harness, baseURL string) *devSignInOverHTTP {
	return &devSignInOverHTTP{newBrowsersOverHTTP(h, baseURL)}
}

func (d *devSignInOverHTTP) identity(name string) (domain.Identity, error) { return devIdentity(name) }

func (d *devSignInOverHTTP) signIn(ctx context.Context, name string) (context.Context, error) {
	browser, err := d.browserFor(name)
	if err != nil {
		return ctx, err
	}
	resp, err := browser.PostForm(d.baseURL+"/auth/dev", url.Values{"name": {name}})
	if err != nil {
		return ctx, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		return ctx, fmt.Errorf("POST /auth/dev: status %d, want %d", resp.StatusCode, http.StatusSeeOther)
	}
	return ctx, d.use(browser)
}

// oidcSignInOverHTTP signs in through the server's OIDC sign-in, as a
// browser does, at a fake sign-in provider that knows a person for each name.
type oidcSignInOverHTTP struct {
	browsersOverHTTP
	provider *oidctest.Provider
}

func newOIDCSignInOverHTTP(h *harness, baseURL string, provider *oidctest.Provider) *oidcSignInOverHTTP {
	return &oidcSignInOverHTTP{browsersOverHTTP: newBrowsersOverHTTP(h, baseURL), provider: provider}
}

// personNamed is who the fake provider signs in for name.
func personNamed(name string) oidctest.Person {
	return oidctest.Person{Subject: "subject-" + name, Name: name}
}

func (o *oidcSignInOverHTTP) identity(name string) (domain.Identity, error) {
	person := personNamed(name)
	return domain.NewIdentity(o.provider.Issuer(), person.Subject, person.Name)
}

func (o *oidcSignInOverHTTP) signIn(ctx context.Context, name string) (context.Context, error) {
	return o.signInAs(ctx, name, personNamed(name))
}

// signInAs signs in person in the browser kept for browserName.
func (o *oidcSignInOverHTTP) signInAs(ctx context.Context, browserName string, person oidctest.Person) (context.Context, error) {
	browser, err := o.browserFor(browserName)
	if err != nil {
		return ctx, err
	}
	o.provider.SetPerson(person)
	if err := followRedirects(browser, o.baseURL+"/auth/sign-in"); err != nil {
		return ctx, err
	}
	return ctx, o.use(browser)
}

// followRedirects gets address with browser and goes wherever the server and
// the provider send it, until an answer that isn't a redirect.
func followRedirects(browser *http.Client, address string) error {
	const maxHops = 10
	for range maxHops {
		resp, err := browser.Get(address)
		if err != nil {
			return err
		}
		resp.Body.Close()
		if resp.StatusCode < 300 || resp.StatusCode >= 400 {
			return nil
		}
		next, err := resp.Location()
		if err != nil {
			return err
		}
		address = next.String()
	}
	return fmt.Errorf("GET %s: more than %d redirects", address, maxHops)
}
