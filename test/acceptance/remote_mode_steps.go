package acceptance

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/desktop"
)

// remoteApp is the desktop app opened on the driver's server, with its own
// settings folder, as features/remote-mode.feature uses it.
type remoteApp struct {
	app *desktop.App
	dir string
	// status is the answer to the latest request the app made.
	status int
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
	sc.Step(`^the desktop app tries to see my workspace's plan without signing in$`, w.theDesktopAppTriesToSeeThePlan)
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
	app, err := desktop.Open(local.getenv, local.userConfigDir, nil)
	if err != nil {
		_ = os.RemoveAll(dir)
		return nil, fmt.Errorf("open the desktop app on the server: %w", err)
	}
	w.desktopApp = &remoteApp{app: app, dir: filepath.Clean(dir)}
	return w.desktopApp, nil
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
