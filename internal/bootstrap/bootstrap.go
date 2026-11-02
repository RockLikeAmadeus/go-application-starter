// Package bootstrap is the composition root: the only place that knows about
// every adapter and decides which ones to plug into the core.
package bootstrap

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/plans"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/jsonfile"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/memory"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/system"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/workspace"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/app"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// LocalPlan names the plan the local edition is on, from the plans file. A
// build picks another with
//
//	go build -ldflags "-X github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap.LocalPlan=<plan>"
//
// NewEmbedded fails if the plans file has no such plan.
var LocalPlan = "local"

// App is the wired application handed to a driving adapter: one driving port
// per capability.
type App struct {
	// sample:start
	Items ports.ItemService
	// sample:end
	Entitlements ports.EntitlementsService
	close        func() error
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
	entitlements, err := localEntitlements(cfg.PlansFile)
	if err != nil {
		return nil, err
	}
	a.Entitlements = app.NewEntitlementsService(entitlements, workspace.Local{})
	// Build each use case on its repository and the local edition's driven
	// ports: workspace.Local{} (the one local workspace, ADR 5), the local
	// plan's entitlements, system.Clock{} and system.UUIDs{}. See
	// docs/workspaces.md.
	// sample:start
	a.Items = app.NewItemService(items, workspace.Local{}, entitlements, system.Clock{}, system.UUIDs{})
	// sample:end
	return a, nil
}

// localEntitlements answers LocalPlan, from plansFile or else the plans file
// built into the binary, for every workspace: the local edition has only
// one. It refuses a LocalPlan the file doesn't have.
func localEntitlements(plansFile string) (plans.Entitlements, error) {
	catalog, err := plans.Load(plansFile)
	if err != nil {
		return plans.Entitlements{}, err
	}
	if _, ok := catalog[LocalPlan]; !ok {
		return plans.Entitlements{}, fmt.Errorf("local plan %q isn't in the plans file", LocalPlan)
	}
	return plans.Entitlements{
		Catalog: catalog,
		PlanOf: func(context.Context, domain.WorkspaceID) (string, error) {
			return LocalPlan, nil
		},
	}, nil
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
	entitlements, err := httpclient.NewEntitlementsService(serverURL)
	if err != nil {
		return nil, fmt.Errorf("create API client: %w", err)
	}
	a.Entitlements = entitlements
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
