package mobile

import (
	"io/fs"
	"net/http"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/desktop"
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
// server's HTTP handler, as desktop.Open does. The folder is
// <storagePath>/starter unless STARTER_DATA_DIR is set. storagePath is the
// app's private storage (application.Android.StoragePath() in production),
// getenv reads the environment (os.Getenv in production) and webAssets is the
// built web app (web.Assets() in production). As on the desktop, /sw.js
// answers 404: the web app's files are built in, so it needs no service
// worker.
func Open(getenv func(string) string, storagePath string, webAssets fs.FS) (*App, error) {
	d, err := desktop.Open(getenv, func() (string, error) { return storagePath, nil }, webAssets)
	if err != nil {
		return nil, err
	}
	return &App{
		Handler: AssetHandler(d.Handler),
		API:     d.Handler,
		desktop: d,
	}, nil
}

// Close releases the data folder.
func (a *App) Close() error {
	return a.desktop.Close()
}
