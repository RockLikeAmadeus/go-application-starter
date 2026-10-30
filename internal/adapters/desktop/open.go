package desktop

import (
	"io/fs"
	"net/http"
	"os"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// App is the opened desktop app: the handler the window loads the SPA and
// its API from, and the wired core whose storage is closed on quit.
type App struct {
	Handler http.Handler
	core    *bootstrap.App
}

// Open opens the desktop app's data folder (see DataDir: STARTER_DATA_DIR,
// else the user's settings folder) and wires the core behind the server's
// HTTP handler. getenv reads the environment (os.Getenv in production).
// webAssets is the built web app (web.Assets() in production).
func Open(getenv func(string) string, webAssets fs.FS) (*App, error) {
	dir, err := DataDir(getenv, os.UserConfigDir)
	if err != nil {
		return nil, err
	}
	core, err := bootstrap.NewEmbedded(config.Config{Storage: config.StorageJSONFile, DataDir: dir})
	if err != nil {
		return nil, err
	}
	api := httpapi.NewHandler(httpapi.Options{
		// sample:start
		Items: core.Items,
		// sample:end
		LocalData: true,
		WebAssets: webAssets,
	})
	return &App{
		Handler: withoutServiceWorker(api),
		core:    core,
	}, nil
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
