package mobile

import (
	"io/fs"
	"net/http"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/desktop"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// App is the opened mobile app: the desktop app's wiring, with one handler for
// each way Wails' Android window reaches Go.
type App struct {
	// Handler is for Wails' asset path: it serves the web app and answers API
	// calls with 404 (see AssetHandler).
	Handler http.Handler
	// API is for the bridge: Serve runs the web app's API requests through it.
	API     http.Handler
	desktop *desktop.App
}

// Open opens the mobile app's data folder and wires the core behind the
// server's HTTP handler, as desktop.Open does, but creates no settings file
// (desktop.OpenWithoutDefaultSettingsFile): the user can't reach the app's
// private storage to edit one. The folder is
// <storagePath>/starter unless STARTER_DATA_DIR is set. storagePath is the
// app's private storage (application.Android.StoragePath() in production),
// getenv reads the environment (os.Getenv in production) and webAssets is the
// built web app (web.Assets() in production). As on the desktop, /sw.js
// answers 404: the web app's files are built in, so it needs no service
// worker. With a server set, the app is a window on it, as on the desktop;
// openBrowser shows the sign-in address in the system browser (Wails'
// Browser.OpenURL in production), and the token is kept in a file in the
// private storage.
func Open(getenv func(string) string, storagePath string, webAssets fs.FS, openBrowser func(address string) error) (*App, error) {
	privateStorage := func() (string, error) { return storagePath, nil }
	desktopApp, err := desktop.OpenWithoutDefaultSettingsFile(withFileTokenStore(getenv), privateStorage, webAssets, openBrowser)
	if err != nil {
		return nil, err
	}
	return &App{
		Handler: AssetHandler(desktopApp.Handler),
		API:     desktopApp.Handler,
		desktop: desktopApp,
	}, nil
}

// withFileTokenStore makes the app keep its session token in a file in its
// private storage. Android has no Secret Service, and go-keyring's Linux code
// would try D-Bus first; only the app can read its private storage.
func withFileTokenStore(getenv func(string) string) func(string) string {
	return func(key string) string {
		if key == "STARTER_CREDENTIAL_STORE" {
			return string(config.CredentialStoreFile)
		}
		return getenv(key)
	}
}

// Close releases the data folder.
func (a *App) Close() error {
	return a.desktop.Close()
}
