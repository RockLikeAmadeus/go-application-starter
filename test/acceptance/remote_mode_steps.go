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
	"time"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/desktop"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/mobile"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/tokenstore"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
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
	// waits is whether the browser is left open and never signs in, so the
	// sign-in in progress waits until it is cancelled.
	waits bool
	// opened is told when the app has opened the sign-in address.
	opened chan struct{}
	// pending gets the answer to the sign-in the app is waiting on.
	pending chan int
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
		if w.isBuiltForServer {
			config.BuiltInServerURL = w.previousBuiltInServerURL
		}
		return ctx, err
	})
	const app = `(desktop app|mobile app)`
	sc.Step(`^the `+app+` tries to see my workspace's plan(?: without signing in)?$`, w.theAppTriesToSeeThePlan)
	sc.Step(`^the `+app+` was built for the server, with no server setting$`, w.theAppWasBuiltForTheServer)
	sc.Step(`^I (?:have )?signed in from the `+app+` as "([^"]*)"$`, w.iSignInFromTheApp)
	sc.Step(`^I sign in from the `+app+` as "([^"]*)"$`, w.iSignInFromTheApp)
	sc.Step(`^the `+app+` is signed in as "([^"]*)"$`, w.theAppIsSignedInAs)
	sc.Step(`^the `+app+` says my workspace's plan is "([^"]*)"$`, w.theAppSaysThePlanIs)
	sc.Step(`^I open the `+app+` on the server$`, w.iOpenTheApp)
	sc.Step(`^the `+app+` offers to sign in through the browser$`, w.theAppOffersToSignInThroughTheBrowser)
	sc.Step(`^I create an account from the `+app+` as "([^"]*)"$`, w.iCreateAnAccountFromTheApp)
	sc.Step(`^the `+app+` is waiting for me to sign in in the browser$`, w.theAppIsWaitingForTheBrowser)
	sc.Step(`^I cancel signing in from the `+app+`$`, w.iCancelSigningInFromTheApp)
	sc.Step(`^the `+app+` is not signed in$`, w.theAppIsNotSignedIn)
	sc.Step(`^I sign out from the `+app+`$`, w.iSignOutFromTheApp)
	sc.Step(`^the `+app+` keeps no session for the server$`, w.theAppKeepsNoSessionForTheServer)
	// sample:start
	sc.Step(`^I create an item named "([^"]*)" from the `+app+`$`, w.iCreateAnItemFromTheApp)
	// sample:end
	sc.Step(`^the `+app+` tells me to sign in$`, w.theAppTellsMeToSignIn)
	sc.Step(`^I have signed in from the command line on this computer as "([^"]*)"$`, w.iSignInFromTheCommandLineOnThisComputer)
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
		"STARTER_CREDENTIAL_STORE": "file",
	}}
	if !w.isBuiltForServer {
		local.env["STARTER_SERVER_URL"] = w.apiURL
	}
	// The opener is the system browser, which the scenario's sign-in drives.
	remote := &remoteApp{
		dir:           filepath.Clean(dir),
		browserErrors: make(chan error, 8),
		opened:        make(chan struct{}, 8),
		isMobile:      name == "mobile app",
	}
	openInBrowser := func(address string) error { return w.openInTheBrowser(remote, address) }
	if remote.isMobile {
		// The phone's private storage is its only folder.
		app, err := mobile.Open(local.getenv, dir, nil, openInBrowser)
		if err != nil {
			_ = os.RemoveAll(dir)
			return nil, fmt.Errorf("open the %s on the server: %w", name, err)
		}
		remote.handler, remote.closeApp = app.API, app.Close
		remote.tokenFolder = filepath.Join(dir, "starter")
	} else {
		app, err := desktop.Open(local.getenv, local.userConfigDir, nil, openInBrowser)
		if err != nil {
			_ = os.RemoveAll(dir)
			return nil, fmt.Errorf("open the %s on the server: %w", name, err)
		}
		remote.handler, remote.closeApp = app.Handler, app.Close
		remote.tokenFolder, err = config.AppFolder(local.configEnv())
		if err != nil {
			return nil, err
		}
	}
	if w.remoteApps == nil {
		w.remoteApps = map[string]*remoteApp{}
	}
	w.remoteApps[name] = remote
	return remote, nil
}

// openInTheBrowser is the app's opener: the person named by the step signs
// in at address a moment later, and the browser returns to the app's
// loopback port.
func (w *world) openInTheBrowser(a *remoteApp, address string) error {
	browser, ok := w.signIn.(nativeSignInBrowser)
	if !ok {
		return errors.New("this driver's sign-in can't finish an app's: tag the scenario @remote-mode")
	}
	a.opened <- struct{}{}
	if a.waits {
		return nil
	}
	name := a.signingInAs
	go func() { a.browserErrors <- finishInTheBrowser(browser, name, address) }()
	return nil
}

// request sends the app a request and returns the answer's status and body.
// The mobile app's requests go over its bridge, as the phone's do.
func (a *remoteApp) request(ctx context.Context, method, path, body string) (int, string) {
	status, reply := a.send(ctx, method, path, body)
	a.status = status
	return status, reply
}

// send is request without noting the status, for a request made while the
// scenario goes on.
func (a *remoteApp) send(ctx context.Context, method, path, body string) (int, string) {
	if a.isMobile {
		reply := mobile.Serve(a.handler, method, path, body)
		return reply.Status, reply.Body
	}
	req := httptest.NewRequestWithContext(ctx, method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	a.handler.ServeHTTP(rec, req)
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
	return w.signInFromTheApp(ctx, name, person, false)
}

func (w *world) iCreateAnAccountFromTheApp(ctx context.Context, name, person string) error {
	return w.signInFromTheApp(ctx, name, person, true)
}

func (w *world) signInFromTheApp(ctx context.Context, name, person string, isSignUp bool) error {
	a, err := w.theApp(name)
	if err != nil {
		return err
	}
	a.signingInAs = person
	ctx, cancel := context.WithTimeout(ctx, signInTimeout)
	defer cancel()
	status, body := a.request(ctx, http.MethodPost, "/api/v1/browser-sign-in", fmt.Sprintf(`{"isSignUp":%t}`, isSignUp))
	if status != http.StatusOK {
		return fmt.Errorf("signing in from the %s answered status %d: %.120s", name, status, body)
	}
	return nil
}

// theAppIsWaitingForTheBrowser starts a sign-in whose browser never comes
// back, and returns once the app has opened it.
func (w *world) theAppIsWaitingForTheBrowser(name string) error {
	a, err := w.theApp(name)
	if err != nil {
		return err
	}
	a.waits = true
	a.pending = make(chan int, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), signInTimeout)
		defer cancel()
		status, _ := a.send(ctx, http.MethodPost, "/api/v1/browser-sign-in", `{"isSignUp":false}`)
		a.pending <- status
	}()
	select {
	case <-a.opened:
		return nil
	case <-time.After(signInTimeout):
		return fmt.Errorf("the %s never opened the browser", name)
	}
}

func (w *world) iCancelSigningInFromTheApp(ctx context.Context, name string) error {
	a, err := w.theApp(name)
	if err != nil {
		return err
	}
	if a.pending == nil {
		return fmt.Errorf("the %s is not waiting for the browser", name)
	}
	if status, body := a.request(ctx, http.MethodDelete, "/api/v1/browser-sign-in", ""); status >= 300 {
		return fmt.Errorf("cancelling the sign-in from the %s answered status %d: %.120s", name, status, body)
	}
	select {
	case status := <-a.pending:
		if status == http.StatusOK {
			return fmt.Errorf("the %s's sign-in succeeded after it was cancelled", name)
		}
	case <-time.After(signInTimeout):
		return fmt.Errorf("the %s's sign-in kept waiting after it was cancelled", name)
	}
	return nil
}

func (w *world) theAppIsNotSignedIn(ctx context.Context, name string) error {
	a, err := w.theApp(name)
	if err != nil {
		return err
	}
	got, err := a.sessionName(ctx)
	if err != nil {
		return err
	}
	if got != "" {
		return fmt.Errorf("the %s is signed in as %q, want no one", name, got)
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

// theAppWasBuiltForTheServer gives the build the server's address, as
// -ldflags does, and leaves the app's settings and environment without one.
// The build's address is restored when the scenario ends.
func (w *world) theAppWasBuiltForTheServer(name string) error {
	if w.remoteApps[name] != nil {
		return fmt.Errorf("the %s was opened before it was built for the server", name)
	}
	if !w.isBuiltForServer {
		w.isBuiltForServer = true
		w.previousBuiltInServerURL = config.BuiltInServerURL
	}
	config.BuiltInServerURL = w.apiURL
	return nil
}

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

// iSignInFromTheCommandLineOnThisComputer signs in from a command line that
// keeps its token in the desktop app's settings folder, as the real one does
// on a computer that has the app. Unlike the other command line steps' memory
// store, this shows whether the two programs' tokens are kept apart.
func (w *world) iSignInFromTheCommandLineOnThisComputer(ctx context.Context, name string) error {
	if w.signIn == nil {
		return errNoSignIn
	}
	a, err := w.theApp("desktop app")
	if err != nil {
		return err
	}
	cfg := config.Config{ServerURL: w.apiURL, CredentialStore: config.CredentialStoreFile}
	client, _, err := bootstrap.NewClientSignInFor(cfg, a.tokenFolder, bootstrap.CommandLine)
	if err != nil {
		return err
	}
	w.commandLine = &commandLine{client: client}
	return w.signInCommandLine(ctx, w.commandLine, name)
}
