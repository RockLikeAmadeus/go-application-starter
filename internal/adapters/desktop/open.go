package desktop

import (
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
func Open(getenv func(string) string) (*App, error) {
	dir, err := DataDir(getenv, os.UserConfigDir)
	if err != nil {
		return nil, err
	}
	core, err := bootstrap.NewEmbedded(config.Config{Storage: config.StorageJSONFile, DataDir: dir})
	if err != nil {
		return nil, err
	}
	return &App{
		Handler: httpapi.NewHandler(httpapi.Options{
			// sample:start
			Items: core.Items,
			// sample:end
			LocalData: true,
		}),
		core: core,
	}, nil
}

// Close releases the data folder.
func (a *App) Close() error {
	return a.core.Close()
}
