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
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/tokenstore"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// remoteApp is the desktop app opened on the driver's server, with its own
// settings folder, as features/remote-mode.feature uses it.
type remoteApp struct {
	app *desktop.App
	dir string
	// status is the answer to the latest request the app made.
	status int
	// signingInAs is who the browser signs in as for the sign-in in progress.
	signingInAs string
	// browserErrors gets the result of each browser's part of a sign-in.
	browserErrors chan error
}

func (a *remoteApp) close() {
	_ = a.app.Close()
	_ = os.RemoveAll(a.dir)
}

func (w *world) registerRemoteModeSteps(sc *godog.ScenarioContext) {
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		if w.desktopApp != nil {
			w.desktopApp.close()
		}
		return ctx, err
	})
	sc.Step(`^the desktop app tries to see my workspace's plan(?: without signing in)?$`, w.theDesktopAppTriesToSeeThePlan)
	sc.Step(`^I (?:have )?signed in from the desktop app as "([^"]*)"$`, w.iSignInFromTheDesktopApp)
	sc.Step(`^I sign in from the desktop app as "([^"]*)"$`, w.iSignInFromTheDesktopApp)
	sc.Step(`^the desktop app is signed in as "([^"]*)"$`, w.theDesktopAppIsSignedInAs)
	sc.Step(`^the desktop app says my workspace's plan is "([^"]*)"$`, w.theDesktopAppSaysThePlanIs)
	sc.Step(`^I open the desktop app on the server$`, w.iOpenTheDesktopApp)
	sc.Step(`^the desktop app offers to sign in through the browser$`, w.theDesktopAppOffersToSignInThroughTheBrowser)
	sc.Step(`^I sign out from the desktop app$`, w.iSignOutFromTheDesktopApp)
	sc.Step(`^the desktop app keeps no session for the server$`, w.theDesktopAppKeepsNoSessionForTheServer)
	// sample:start
	sc.Step(`^I create an item named "([^"]*)" from the desktop app$`, w.iCreateAnItemFromTheDesktopApp)
	// sample:end
	sc.Step(`^the desktop app tells me to sign in$`, w.theDesktopAppTellsMeToSignIn)
}

// theDesktopApp is the scenario's desktop app, opened on the server when first
// needed.
func (w *world) theDesktopApp() (*remoteApp, error) {
	if w.desktopApp != nil {
		return w.desktopApp, nil
	}
	dir, err := os.MkdirTemp("", "acceptance-remote-desktop-*")
	if err != nil {
		return nil, err
	}
	local := &localComputer{dir: dir, env: map[string]string{
		"STARTER_SERVER_URL":       w.apiURL,
		"STARTER_CREDENTIAL_STORE": "file",
	}}
	// The opener is the system browser, which the scenario's sign-in drives.
	opened := &remoteApp{dir: filepath.Clean(dir), browserErrors: make(chan error, 8)}
	app, err := desktop.Open(local.getenv, local.userConfigDir, nil, func(address string) error {
		return w.openInTheBrowser(opened, address)
	})
	if err != nil {
		_ = os.RemoveAll(dir)
		return nil, fmt.Errorf("open the desktop app on the server: %w", err)
	}
	opened.app = app
	w.desktopApp = opened
	return w.desktopApp, nil
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

// request sends the app's handler a request and returns the answer.
func (a *remoteApp) request(ctx context.Context, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(ctx, method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	a.app.Handler.ServeHTTP(rec, req)
	a.status = rec.Code
	return rec
}

// sessionName is the name the app is signed in as, empty when no one is.
func (a *remoteApp) sessionName(ctx context.Context) (string, error) {
	var session struct {
		Name string `json:"name"`
	}
	rec := a.request(ctx, http.MethodGet, "/api/v1/session", "")
	if err := json.Unmarshal(rec.Body.Bytes(), &session); err != nil {
		return "", fmt.Errorf("read the app's session (status %d): %w", rec.Code, err)
	}
	return session.Name, nil
}

func (w *world) iSignInFromTheDesktopApp(ctx context.Context, name string) error {
	a, err := w.theDesktopApp()
	if err != nil {
		return err
	}
	a.signingInAs = name
	ctx, cancel := context.WithTimeout(ctx, signInTimeout)
	defer cancel()
	rec := a.request(ctx, http.MethodPost, "/api/v1/browser-sign-in", `{"isSignUp":false}`)
	if rec.Code != http.StatusOK {
		return fmt.Errorf("signing in from the desktop app answered status %d: %.120s", rec.Code, rec.Body.String())
	}
	return nil
}

func (w *world) theDesktopAppIsSignedInAs(ctx context.Context, name string) error {
	a, err := w.theDesktopApp()
	if err != nil {
		return err
	}
	got, err := a.sessionName(ctx)
	if err != nil {
		return err
	}
	if got != name {
		return fmt.Errorf("the desktop app is signed in as %q, want %q", got, name)
	}
	return nil
}

func (w *world) theDesktopAppSaysThePlanIs(ctx context.Context, plan string) error {
	a, err := w.theDesktopApp()
	if err != nil {
		return err
	}
	rec := a.request(ctx, http.MethodGet, "/api/v1/entitlements", "")
	var got struct {
		Plan string `json:"plan"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || got.Plan != plan {
		return fmt.Errorf("the desktop app says the plan is %q (status %d, error %w), want %q", got.Plan, rec.Code, err, plan)
	}
	return nil
}

func (w *world) theDesktopAppOffersToSignInThroughTheBrowser(ctx context.Context) error {
	a, err := w.theDesktopApp()
	if err != nil {
		return err
	}
	var session struct {
		IsSignInOn   bool   `json:"isSignInOn"`
		SignInMethod string `json:"signInMethod"`
	}
	rec := a.request(ctx, http.MethodGet, "/api/v1/session", "")
	if err := json.Unmarshal(rec.Body.Bytes(), &session); err != nil {
		return err
	}
	if !session.IsSignInOn || session.SignInMethod != "browser" {
		return fmt.Errorf("the desktop app's session is %+v, want sign-in on through the browser", session)
	}
	return nil
}

func (w *world) iSignOutFromTheDesktopApp(ctx context.Context) error {
	a, err := w.theDesktopApp()
	if err != nil {
		return err
	}
	rec := a.request(ctx, http.MethodPost, "/auth/sign-out", "")
	if rec.Code != http.StatusSeeOther && rec.Code != http.StatusFound {
		return fmt.Errorf("signing out of the desktop app answered status %d", rec.Code)
	}
	return nil
}

func (w *world) theDesktopAppKeepsNoSessionForTheServer() error {
	a, err := w.theDesktopApp()
	if err != nil {
		return err
	}
	local := &localComputer{dir: a.dir}
	folder, err := config.AppFolder(local.configEnv())
	if err != nil {
		return err
	}
	if _, err := tokenstore.NewAppFile(folder).Read(w.apiURL); !errors.Is(err, tokenstore.ErrNoToken) {
		return fmt.Errorf("the desktop app keeps a token for the server (read error: %w)", err)
	}
	return nil
}

// sample:start

func (w *world) iCreateAnItemFromTheDesktopApp(ctx context.Context, name string) error {
	a, err := w.theDesktopApp()
	if err != nil {
		return err
	}
	body, err := json.Marshal(map[string]string{"name": name})
	if err != nil {
		return err
	}
	rec := a.request(ctx, http.MethodPost, "/api/v1/items", string(body))
	if rec.Code != http.StatusCreated {
		return fmt.Errorf("adding an item from the desktop app answered status %d: %.120s", rec.Code, rec.Body.String())
	}
	return nil
}

// sample:end

func (w *world) iOpenTheDesktopApp() error {
	_, err := w.theDesktopApp()
	return err
}

func (w *world) theDesktopAppTriesToSeeThePlan() error {
	a, err := w.theDesktopApp()
	if err != nil {
		return err
	}
	rec := httptest.NewRecorder()
	a.app.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/entitlements", nil))
	a.status = rec.Code
	return nil
}

func (w *world) theDesktopAppTellsMeToSignIn() error {
	if w.desktopApp == nil {
		return errors.New("the desktop app was not opened")
	}
	if w.desktopApp.status != http.StatusUnauthorized {
		return fmt.Errorf("the desktop app answered status %d, want %d", w.desktopApp.status, http.StatusUnauthorized)
	}
	return nil
}
