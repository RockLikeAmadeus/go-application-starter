package acceptance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/desktop"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/mobile"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/tokenstore"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// remoteApp is the desktop or mobile app opened on the driver's server, with
// its own settings folder, as features/remote-mode.feature uses it.
type remoteApp struct {
	// handler answers the app's requests: the desktop app's window handler,
	// or the mobile app's API handler, which its requests reach over the
	// bridge.
	handler  http.Handler
	isMobile bool
	closeApp func() error
	dir      string
	// tokenFolder is the folder that holds the app's session token.
	tokenFolder string
	// status is the answer to the latest request the app made.
	status int
	// signingInAs is who the browser signs in as for the sign-in in progress.
	signingInAs string
	// browserErrors gets the result of each browser's part of a sign-in.
	browserErrors chan error
}

func (a *remoteApp) close() {
	_ = a.closeApp()
	_ = os.RemoveAll(a.dir)
}

func (w *world) registerRemoteModeSteps(sc *godog.ScenarioContext) {
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		for _, a := range w.remoteApps {
			a.close()
		}
		return ctx, err
	})
	const app = `(desktop app|mobile app)`
	sc.Step(`^the `+app+` tries to see my workspace's plan(?: without signing in)?$`, w.theAppTriesToSeeThePlan)
	sc.Step(`^I (?:have )?signed in from the `+app+` as "([^"]*)"$`, w.iSignInFromTheApp)
	sc.Step(`^I sign in from the `+app+` as "([^"]*)"$`, w.iSignInFromTheApp)
	sc.Step(`^the `+app+` is signed in as "([^"]*)"$`, w.theAppIsSignedInAs)
	sc.Step(`^the `+app+` says my workspace's plan is "([^"]*)"$`, w.theAppSaysThePlanIs)
	sc.Step(`^I open the `+app+` on the server$`, w.iOpenTheApp)
	sc.Step(`^the `+app+` offers to sign in through the browser$`, w.theAppOffersToSignInThroughTheBrowser)
	sc.Step(`^I sign out from the `+app+`$`, w.iSignOutFromTheApp)
	sc.Step(`^the `+app+` keeps no session for the server$`, w.theAppKeepsNoSessionForTheServer)
	// sample:start
	sc.Step(`^I create an item named "([^"]*)" from the `+app+`$`, w.iCreateAnItemFromTheApp)
	// sample:end
	sc.Step(`^the `+app+` tells me to sign in$`, w.theAppTellsMeToSignIn)
}

// theApp is the scenario's desktop or mobile app, named as the feature names
// it, opened on the server when first needed.
func (w *world) theApp(name string) (*remoteApp, error) {
	if a, ok := w.remoteApps[name]; ok {
		return a, nil
	}
	dir, err := os.MkdirTemp("", "acceptance-remote-app-*")
	if err != nil {
		return nil, err
	}
	local := &localComputer{dir: dir, env: map[string]string{
		"STARTER_SERVER_URL":       w.apiURL,
		"STARTER_CREDENTIAL_STORE": "file",
	}}
	// The opener is the system browser, which the scenario's sign-in drives.
	opened := &remoteApp{dir: filepath.Clean(dir), browserErrors: make(chan error, 8), isMobile: name == "mobile app"}
	openInBrowser := func(address string) error { return w.openInTheBrowser(opened, address) }
	if opened.isMobile {
		// The phone's private storage is its only folder.
		app, err := mobile.Open(local.getenv, dir, nil, openInBrowser)
		if err != nil {
			_ = os.RemoveAll(dir)
			return nil, fmt.Errorf("open the %s on the server: %w", name, err)
		}
		opened.handler, opened.closeApp = app.API, app.Close
		opened.tokenFolder = filepath.Join(dir, "starter")
	} else {
		app, err := desktop.Open(local.getenv, local.userConfigDir, nil, openInBrowser)
		if err != nil {
			_ = os.RemoveAll(dir)
			return nil, fmt.Errorf("open the %s on the server: %w", name, err)
		}
		opened.handler, opened.closeApp = app.Handler, app.Close
		opened.tokenFolder, err = config.AppFolder(local.configEnv())
		if err != nil {
			return nil, err
		}
	}
	if w.remoteApps == nil {
		w.remoteApps = map[string]*remoteApp{}
	}
	w.remoteApps[name] = opened
	return opened, nil
}

// openInTheBrowser is the app's opener: the person named by the step signs
// in at address a moment later, and the browser returns to the app's
// loopback port.
func (w *world) openInTheBrowser(a *remoteApp, address string) error {
	browser, ok := w.signIn.(nativeSignInBrowser)
	if !ok {
		return errors.New("this driver's sign-in can't finish an app's: tag the scenario @remote-mode")
	}
	name := a.signingInAs
	go func() { a.browserErrors <- finishInTheBrowser(browser, name, address) }()
	return nil
}

// request sends the app a request and returns the answer's status and body.
// The mobile app's requests go over its bridge, as the phone's do.
func (a *remoteApp) request(ctx context.Context, method, path, body string) (int, string) {
	if a.isMobile {
		reply := mobile.Serve(a.handler, method, path, body)
		a.status = reply.Status
		return reply.Status, reply.Body
	}
	req := httptest.NewRequestWithContext(ctx, method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	a.handler.ServeHTTP(rec, req)
	a.status = rec.Code
	return rec.Code, rec.Body.String()
}

// sessionName is the name the app is signed in as, empty when no one is.
func (a *remoteApp) sessionName(ctx context.Context) (string, error) {
	var session struct {
		Name string `json:"name"`
	}
	status, body := a.request(ctx, http.MethodGet, "/api/v1/session", "")
	if err := json.Unmarshal([]byte(body), &session); err != nil {
		return "", fmt.Errorf("read the app's session (status %d): %w", status, err)
	}
	return session.Name, nil
}

func (w *world) iSignInFromTheApp(ctx context.Context, name, person string) error {
	a, err := w.theApp(name)
	if err != nil {
		return err
	}
	a.signingInAs = person
	ctx, cancel := context.WithTimeout(ctx, signInTimeout)
	defer cancel()
	status, body := a.request(ctx, http.MethodPost, "/api/v1/browser-sign-in", `{"isSignUp":false}`)
	if status != http.StatusOK {
		return fmt.Errorf("signing in from the %s answered status %d: %.120s", name, status, body)
	}
	return nil
}

func (w *world) theAppIsSignedInAs(ctx context.Context, name, person string) error {
	a, err := w.theApp(name)
	if err != nil {
		return err
	}
	got, err := a.sessionName(ctx)
	if err != nil {
		return err
	}
	if got != person {
		return fmt.Errorf("the %s is signed in as %q, want %q", name, got, person)
	}
	return nil
}

func (w *world) theAppSaysThePlanIs(ctx context.Context, name, plan string) error {
	a, err := w.theApp(name)
	if err != nil {
		return err
	}
	status, body := a.request(ctx, http.MethodGet, "/api/v1/entitlements", "")
	var got struct {
		Plan string `json:"plan"`
	}
	if err := json.Unmarshal([]byte(body), &got); err != nil || got.Plan != plan {
		return fmt.Errorf("the %s says the plan is %q (status %d, error %w), want %q", name, got.Plan, status, err, plan)
	}
	return nil
}

func (w *world) theAppOffersToSignInThroughTheBrowser(ctx context.Context, name string) error {
	a, err := w.theApp(name)
	if err != nil {
		return err
	}
	var session struct {
		IsSignInOn   bool   `json:"isSignInOn"`
		SignInMethod string `json:"signInMethod"`
	}
	_, body := a.request(ctx, http.MethodGet, "/api/v1/session", "")
	if err := json.Unmarshal([]byte(body), &session); err != nil {
		return err
	}
	if !session.IsSignInOn || session.SignInMethod != "browser" {
		return fmt.Errorf("the %s's session is %+v, want sign-in on through the browser", name, session)
	}
	return nil
}

func (w *world) iSignOutFromTheApp(ctx context.Context, name string) error {
	a, err := w.theApp(name)
	if err != nil {
		return err
	}
	status, _ := a.request(ctx, http.MethodPost, "/auth/sign-out", "")
	if status != http.StatusSeeOther && status != http.StatusFound {
		return fmt.Errorf("signing out of the %s answered status %d", name, status)
	}
	return nil
}

func (w *world) theAppKeepsNoSessionForTheServer(name string) error {
	a, err := w.theApp(name)
	if err != nil {
		return err
	}
	if _, err := tokenstore.NewAppFile(a.tokenFolder).Read(w.apiURL); !errors.Is(err, tokenstore.ErrNoToken) {
		return fmt.Errorf("the %s keeps a token for the server (read error: %w)", name, err)
	}
	return nil
}

// sample:start

func (w *world) iCreateAnItemFromTheApp(ctx context.Context, item, name string) error {
	a, err := w.theApp(name)
	if err != nil {
		return err
	}
	body, err := json.Marshal(map[string]string{"name": item})
	if err != nil {
		return err
	}
	status, reply := a.request(ctx, http.MethodPost, "/api/v1/items", string(body))
	if status != http.StatusCreated {
		return fmt.Errorf("adding an item from the %s answered status %d: %.120s", name, status, reply)
	}
	return nil
}

// sample:end

func (w *world) iOpenTheApp(name string) error {
	_, err := w.theApp(name)
	return err
}

func (w *world) theAppTriesToSeeThePlan(ctx context.Context, name string) error {
	a, err := w.theApp(name)
	if err != nil {
		return err
	}
	a.request(ctx, http.MethodGet, "/api/v1/entitlements", "")
	return nil
}

func (w *world) theAppTellsMeToSignIn(name string) error {
	a, ok := w.remoteApps[name]
	if !ok {
		return fmt.Errorf("the %s was not opened", name)
	}
	if a.status != http.StatusUnauthorized {
		return fmt.Errorf("the %s answered status %d, want %d", name, a.status, http.StatusUnauthorized)
	}
	return nil
}
