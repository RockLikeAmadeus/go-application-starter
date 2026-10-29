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

// App is the wired application handed to a driving adapter.
type App struct {
	Items ports.ItemService
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
	repo, closer, err := newItemRepository(cfg)
	if err != nil {
		return nil, err
	}
	return &App{
		Items: app.NewItemService(repo, system.Clock{}, system.UUIDs{}),
		close: closer,
	}, nil
}

// NewRemote returns an App whose ItemService talks to a server over HTTP.
func NewRemote(serverURL string) (*App, error) {
	svc, err := httpclient.New(serverURL)
	if err != nil {
		return nil, fmt.Errorf("create API client: %w", err)
	}
	return &App{Items: svc}, nil
}

// New picks remote mode when cfg.ServerURL is set, embedded otherwise.
func New(cfg config.Config) (*App, error) {
	if cfg.ServerURL != "" {
		return NewRemote(cfg.ServerURL)
	}
	return NewEmbedded(cfg)
}

func newItemRepository(cfg config.Config) (ports.ItemRepository, func() error, error) {
	switch cfg.Storage {
	case config.StorageMemory:
		return memory.NewItemRepository(), nil, nil
	case config.StorageJSONFile:
		store, err := jsonfile.Open(cfg.DataDir)
		if err != nil {
			return nil, nil, err
		}
		return jsonfile.NewItemRepository(store), store.Close, nil
	default:
		return nil, nil, fmt.Errorf("unsupported storage %q", cfg.Storage)
	}
}

// NewLogger builds the process-wide structured logger.
func NewLogger(cfg config.Config, w io.Writer) *slog.Logger {
	if cfg.LogFormat == "json" {
		return slog.New(slog.NewJSONHandler(w, nil))
	}
	return slog.New(slog.NewTextHandler(w, nil))
}
