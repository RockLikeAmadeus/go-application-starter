// Package bootstrap is the composition root: the only place that knows about
// every adapter and decides which ones to plug into the core.
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/plans"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/jsonfile"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/memory"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlite"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/syncedfolder"
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
	// Warnings are messages every front end shows the user as given, such
	// as SQLite's data folder looking synced, or a conflict copy in plain text.
	Warnings []string
	close    func() error
	ready    func(context.Context) error
}

// Close releases resources such as storage locks.
func (a *App) Close() error {
	if a.close == nil {
		return nil
	}
	return a.close()
}

// Ready reports whether storage can be reached, for the server's /readyz.
// Storage with nothing to check (memory) and a remote App are always ready.
func (a *App) Ready(ctx context.Context) error {
	if a.ready == nil {
		return nil
	}
	return a.ready(ctx)
}

// NewEmbedded wires the core in-process with the configured storage adapter.
func NewEmbedded(cfg config.Config) (*App, error) {
	a := &App{}
	// sample:start
	var itemRepo ports.ItemRepository
	// sample:end
	switch cfg.Storage {
	case config.StorageMemory:
		// sample:start
		itemRepo = memory.NewItemRepository()
		// sample:end
	case config.StorageJSONFile:
		store, err := jsonfile.Open(cfg.DataDir)
		if err != nil {
			return nil, err
		}
		a.close = store.Close
		a.ready = store.Ready
		if warning, ok := conflictCopyWarning(store); ok {
			a.Warnings = append(a.Warnings, warning)
		}
		if warning, ok := otherDeviceWarning(cfg); ok {
			a.Warnings = append(a.Warnings, warning)
		}
		// sample:start
		itemRepo = jsonfile.NewItemRepository(store)
		// sample:end
	case config.StorageSQLite:
		store, err := sqlite.Open(cfg.DataDir)
		if err != nil {
			return nil, err
		}
		a.close = store.Close
		a.ready = store.Ready
		if warning, ok := syncedFolderWarning(cfg); ok {
			a.Warnings = append(a.Warnings, warning)
		}
		// sample:start
		itemRepo = sqlite.NewItemRepository(store)
		// sample:end
	default:
		return nil, fmt.Errorf("unsupported storage %q", cfg.Storage)
	}
	entitlementsProvider, err := localEntitlementsProvider(cfg.PlansFile)
	if err != nil {
		return nil, err
	}
	a.Entitlements = app.NewEntitlementsService(entitlementsProvider, workspace.Local{})
	// Build each use case on its repository and the local edition's driven
	// ports: workspace.Local{} (the one local workspace, ADR 5), the local
	// plan's entitlements, system.Clock{} and system.UUIDs{}. See
	// docs/workspaces.md.
	// sample:start
	a.Items = app.NewItemService(itemRepo, workspace.Local{}, entitlementsProvider, system.Clock{}, system.UUIDs{})
	// sample:end
	return a, nil
}

// NewEntitlementsService returns the entitlements service alone: the server's when
// cfg.ServerURL is set, as NewRemote wires it, else the local edition's, as
// NewEmbedded wires it, but opening no storage: it creates no data folder,
// runs no migrations and isn't stopped by a locked folder. The CLI's help
// reads the plan through it.
func NewEntitlementsService(cfg config.Config) (ports.EntitlementsService, error) {
	if cfg.ServerURL != "" {
		return httpclient.NewEntitlementsService(cfg.ServerURL)
	}
	entitlementsProvider, err := localEntitlementsProvider(cfg.PlansFile)
	if err != nil {
		return nil, err
	}
	return app.NewEntitlementsService(entitlementsProvider, workspace.Local{}), nil
}

// ErrLocalPlanIsFixed is the refusal to change a plan in the local edition,
// whose plan is LocalPlan.
var ErrLocalPlanIsFixed = errors.New("the local edition's plan is set when the app is built")

// ErrNoPlanChangesOverAServer is the refusal to change a plan with --server.
var ErrNoPlanChangesOverAServer = errors.New("a plan can't be changed over a server yet")

// NewBilling returns the billing service, for `starter admin set-plan`. Every
// copy is the local edition until sign-in adds the hosted one, so it always
// refuses with ErrLocalPlanIsFixed, or ErrNoPlanChangesOverAServer with
// cfg.ServerURL set, opening no storage. It will wire the stored plans for
// the hosted edition.
func NewBilling(cfg config.Config) (ports.BillingService, error) {
	if cfg.ServerURL != "" {
		return nil, ErrNoPlanChangesOverAServer
	}
	return nil, ErrLocalPlanIsFixed
}

// syncedFolderWarning warns when SQLite's data folder looks kept in sync by
// a file-sync service, which can damage the database, unless the setting
// turns it off. Plain text is safe there (one device at a time), so only
// SQLite asks.
func syncedFolderWarning(cfg config.Config) (string, bool) {
	if cfg.ShouldIgnoreSyncedFolder {
		return "", false
	}
	service, ok := syncedfolder.Detect(cfg.DataDir)
	if !ok {
		return "", false
	}
	return "Your data folder appears to be synced by " + service + ". A file sync service can damage SQLite data: " +
		"please move your data folder out of the synced location before continuing. Alternatively, you can " +
		"switch to plain-text storage, but your existing data will not be transferred automatically.", true
}

// conflictCopyWarning warns about the conflict copies in a plain-text data
// folder, naming each one, or that it couldn't look for them: a failed scan
// doesn't stop the app, but it mustn't read as there being none.
func conflictCopyWarning(store *jsonfile.Store) (string, bool) {
	copies, err := store.ConflictCopies()
	if err != nil {
		return "Your data folder could not be checked for conflicted copies, so a change may be missing from your data. " +
			"The following error occurred: " + err.Error(), true
	}
	if len(copies) == 0 {
		return "", false
	}
	names := strings.Join(copies, ", ")
	if len(copies) == 1 {
		return "Your data folder contains 1 conflicted copy, " + names + ", which appears to have been created by " +
			"a file sync service. A change may be missing from your data. Please compare it with the original file, " +
			"then either delete it or use it to replace the original before continuing.", true
	}
	return fmt.Sprintf("Your data folder contains %d conflicted copies that appear to have been created by a file "+
		"sync service: %s. Changes may be missing from your data. Please compare each one with its original file, "+
		"then either delete it or use it to replace the original before continuing.", len(copies), names), true
}

// otherDeviceWarning warns when another device's mark says it has a plain-text
// data folder in use, naming every such device. This computer's name tells its
// own mark apart; if it can't be read, every mark counts as another device's.
func otherDeviceWarning(cfg config.Config) (string, bool) {
	thisDevice, _ := os.Hostname()
	devices := syncedfolder.OtherDevices(cfg.DataDir, thisDevice, system.Clock{}.Now())
	if len(devices) == 0 {
		return "", false
	}
	if len(devices) == 1 {
		return "Your data folder appears to be in use on " + devices[0] + ". Running the application from two " +
			"devices at once while plain-text storage is enabled can result in lost changes: please close the " +
			"application on your other device before continuing.", true
	}
	return "Your data folder appears to be in use on " + deviceList(devices) + ". Running the " +
		"application from multiple devices at once while plain-text storage is enabled can result in lost " +
		"changes: please close the application on your other devices before continuing.", true
}

// deviceList names two or more devices: "a and b", or "a, b, and c".
func deviceList(devices []string) string {
	last := len(devices) - 1
	if last == 1 {
		return devices[0] + " and " + devices[1]
	}
	return strings.Join(devices[:last], ", ") + ", and " + devices[last]
}

// localEntitlementsProvider answers LocalPlan, from plansFile or else the plans file
// built into the binary, for every workspace: the local edition has only
// one. It gives no upgrade hints, since a local copy can't change plan. It
// refuses a LocalPlan the file doesn't have.
func localEntitlementsProvider(plansFile string) (plans.Provider, error) {
	catalog, err := plans.Load(plansFile)
	if err != nil {
		return plans.Provider{}, err
	}
	if _, ok := catalog[LocalPlan]; !ok {
		return plans.Provider{}, fmt.Errorf("local plan %q isn't in the plans file", LocalPlan)
	}
	return plans.Provider{
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
	svc, err := httpclient.NewItemService(serverURL)
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
