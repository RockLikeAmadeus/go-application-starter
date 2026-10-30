// Package bootstrap is the composition root: the only place that knows about
// every adapter and decides which ones to plug into the core.
package bootstrap

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/jsonfile"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/memory"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/system"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/app"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// App is the wired application handed to a driving adapter: one driving port
// per capability.
type App struct {
	// sample:start
	Items ports.ItemService
	// sample:end
	close func() error
}

// Close releases resources such as storage locks.
func (a *App) Close() error {
	if a.close == nil {
		return nil
	}
	return a.close()
}

// NewEmbedded wires the core in-process with the configured storage adapter.
func NewEmbedded(cfg config.Config) (*App, error) {
	a := &App{}
	// sample:start
	var items ports.ItemRepository
	// sample:end
	switch cfg.Storage {
	case config.StorageMemory:
		// sample:start
		items = memory.NewItemRepository()
		// sample:end
	case config.StorageJSONFile:
		store, err := jsonfile.Open(cfg.DataDir)
		if err != nil {
			return nil, err
		}
		a.close = store.Close
		// sample:start
		items = jsonfile.NewItemRepository(store)
		// sample:end
	default:
		return nil, fmt.Errorf("unsupported storage %q", cfg.Storage)
	}
	// sample:start
	a.Items = app.NewItemService(items, system.Clock{}, system.UUIDs{})
	// sample:end
	return a, nil
}

// NewRemote returns an App whose driving ports talk to a server over HTTP.
func NewRemote(serverURL string) (*App, error) {
	a := &App{}
	// sample:start
	svc, err := httpclient.New(serverURL)
	if err != nil {
		return nil, fmt.Errorf("create API client: %w", err)
	}
	a.Items = svc
	// sample:end
	return a, nil
}

// New picks remote mode when cfg.ServerURL is set, embedded otherwise.
func New(cfg config.Config) (*App, error) {
	if cfg.ServerURL != "" {
		return NewRemote(cfg.ServerURL)
	}
	return NewEmbedded(cfg)
}

// NewLogger builds the process-wide structured logger.
func NewLogger(cfg config.Config, w io.Writer) *slog.Logger {
	if cfg.LogFormat == "json" {
		return slog.New(slog.NewJSONHandler(w, nil))
	}
	return slog.New(slog.NewTextHandler(w, nil))
}
