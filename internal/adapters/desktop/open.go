// Package desktop wires the desktop app: it picks the data folder and opens
// the core behind the same HTTP handler the server uses. It has no Wails
// dependency, so it is unit-testable; cmd/desktop only adds the window.
package desktop

import (
	"errors"
	"io/fs"
	"net/http"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// App is the opened desktop app: the handler the window loads the SPA and
// its API from, the app's icon, and the wired core whose storage is closed
// on quit.
type App struct {
	Handler http.Handler
	// Icon is the product's icon as a PNG, or nil when the web app has none.
	Icon []byte
	core *bootstrap.App
}

// Open creates the default settings file when there is none
// (config.CreateDefaultSettingsFile), loads the configuration from the local
// settings file and the environment (config.Load). With a server set
// (server_url, or one built in) it opens remote mode, a window on that server
// with no local data. Otherwise it uses SQLite storage when the settings choose it
// and jsonfile otherwise (never memory: the desktop app keeps its data; and
// it refuses Postgres, because hosted data is reached through a server),
// opens the desktop app's data folder (config.LocalDataDir:
// STARTER_DATA_DIR, else data_dir in the settings file, else the user's
// settings folder) and wires the core behind the server's HTTP handler.
// getenv reads the environment (os.Getenv in production) and userConfigDir
// finds the user's settings folder (os.UserConfigDir on the desktop).
// webAssets is the built web app (web.Assets() in production). openBrowser
// shows an address in the system browser, for signing in to the server in
// remote mode (the Wails app's Browser.OpenURL in production).
func Open(getenv func(string) string, userConfigDir func() (string, error), webAssets fs.FS, openBrowser func(address string) error) (*App, error) {
	config.CreateDefaultSettingsFile(config.Env{Getenv: getenv, UserConfigDir: userConfigDir})
	return OpenWithoutDefaultSettingsFile(getenv, userConfigDir, webAssets, openBrowser)
}

// OpenWithoutDefaultSettingsFile opens the app as Open does, but creates no
// settings file when there is none.
func OpenWithoutDefaultSettingsFile(getenv func(string) string, userConfigDir func() (string, error), webAssets fs.FS, openBrowser func(address string) error) (*App, error) {
	env := config.Env{Getenv: getenv, UserConfigDir: userConfigDir}
	settingsFile, err := config.LocalConfigFile(env)
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(env, settingsFile)
	if err != nil {
		return nil, err
	}
	if cfg.ServerURL != "" {
		return openRemote(env, cfg, webAssets, openBrowser)
	}
	if cfg.Storage == config.StoragePostgres {
		return nil, errors.New(`The storage setting is "postgres", but this application keeps its data on this device. Please use "jsonfile" or "sqlite".`)
	}
	if cfg.Storage != config.StorageSQLite {
		cfg.Storage = config.StorageJSONFile
	}
	cfg.DataDir, err = config.LocalDataDir(env, settingsFile)
	if err != nil {
		return nil, err
	}
	core, err := bootstrap.NewEmbedded(cfg)
	if err != nil {
		return nil, err
	}
	api := httpapi.NewHandler(httpapi.Options{
		// sample:start
		Items: core.Items,
		// sample:end
		Entitlements: core.Entitlements,
		Ready:        core.Ready,
		Warnings:     core.Warnings,
		IsDataLocal:  true,
		WebAssets:    webAssets,
	})
	return &App{
		Handler: withoutServiceWorker(api),
		Icon:    iconOf(webAssets),
		core:    core,
	}, nil
}

// openRemote opens the app as a window on the server in cfg.ServerURL: no data
// folder is opened, so nothing is locked, and IsDataLocal stays unset so the
// web app waits for a connection.
func openRemote(env config.Env, cfg config.Config, webAssets fs.FS, openBrowser func(address string) error) (*App, error) {
	folder, err := config.AppFolder(env)
	if err != nil {
		return nil, err
	}
	core, err := bootstrap.NewRemote(cfg, folder, bootstrap.DesktopOrMobileApp)
	if err != nil {
		return nil, err
	}
	signIn, err := bootstrap.NewAppClientSignIn(cfg, folder)
	if err != nil {
		return nil, err
	}
	api := httpapi.NewHandler(httpapi.Options{
		// sample:start
		Items: core.Items,
		// sample:end
		Entitlements: core.Entitlements,
		WebAssets:    webAssets,
		ClientSignIn: signIn,
		OpenBrowser:  openBrowser,
	})
	return &App{
		Handler: withoutServiceWorker(api),
		Icon:    iconOf(webAssets),
		core:    core,
	}, nil
}

// iconFile is the web app's largest icon, which `pnpm icons` makes from
// web/public/icon.svg.
const iconFile = "pwa-512x512.png"

// iconOf returns the web app's icon, so the desktop app and the installed web
// app share one. Without a web app, or an icon in it, there is none: the
// window then shows the system's default icon, which is no reason to fail.
func iconOf(webAssets fs.FS) []byte {
	if webAssets == nil {
		return nil
	}
	icon, err := fs.ReadFile(webAssets, iconFile)
	if err != nil {
		return nil
	}
	return icon
}

// withoutServiceWorker answers the web app's service worker script with 404.
// The desktop app's files are built into it, so precaching adds nothing, and
// "update on next open" would show the old version once after each upgrade.
// With no script the web app can't register one, and a browser drops one
// already registered when its update check gets a 404.
func withoutServiceWorker(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/sw.js" {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Close releases the data folder.
func (a *App) Close() error {
	return a.core.Close()
}
