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
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/oidc"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/plans"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/jsonfile"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/memory"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/postgres"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlite"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlstore"
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
	// SignIn signs people in to the hosted edition. It's nil with sign-in
	// off, as the local edition has no accounts.
	SignIn ports.SignInService
	// IsDevSignInOn says whether the server serves the dev sign-in under
	// /auth/dev: sign_in is "dev", which only a devsignin build accepts.
	IsDevSignInOn bool
	// SignInProvider is the OIDC sign-in provider, discovered at start. It's
	// nil unless sign_in is "oidc".
	SignInProvider httpapi.SignInProvider
	// Warnings are messages every front end shows the user as given, such
	// as SQLite's data folder looking synced, or a conflict copy in plain text.
	Warnings []string
	close    func() error
	ready    func(context.Context) error
	// billing changes the stored plans, for NewBilling. It's nil with sign-in
	// off, as the local edition's plan is fixed.
	billing ports.BillingService
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
	// Plain-text storage has no account or session repositories.
	if cfg.SignIn.IsOn() && cfg.Storage == config.StorageJSONFile {
		return nil, fmt.Errorf(
			"The sign-in setting is %q, but the storage setting is %q. Sign-in needs %q, %q, or %q storage. Please change the storage setting, or set the sign-in setting to %q.",
			cfg.SignIn, cfg.Storage, config.StorageMemory, config.StorageSQLite, config.StoragePostgres, config.SignInOff)
	}
	if cfg.SignIn == config.SignInDev && !isDevSignInBuilt {
		return nil, fmt.Errorf(
			"The sign-in setting is %q, but this program was built without the dev sign-in, which is only for development and tests. Please set the sign-in setting to %q, or build the program with the %q build tag.",
			cfg.SignIn, config.SignInOff, devSignInBuildTag)
	}
	a := &App{IsDevSignInOn: cfg.SignIn == config.SignInDev}
	if cfg.SignIn == config.SignInOIDC {
		provider, err := oidc.New(context.Background(), oidc.Settings{
			Issuer:       cfg.OIDCIssuer,
			ClientID:     cfg.OIDCClientID,
			ClientSecret: cfg.OIDCClientSecret,
			RedirectURL:  strings.TrimRight(cfg.PublicURL, "/") + "/auth/oidc/callback",
		})
		if err != nil {
			return nil, err
		}
		a.SignInProvider = provider
	}
	// sample:start
	var itemRepo ports.ItemRepository
	// sample:end
	var accountRepo ports.AccountRepository
	var sessionRepo ports.SessionRepository
	var workspacePlanRepo ports.WorkspacePlanRepository
	switch cfg.Storage {
	case config.StorageMemory:
		accountRepo = memory.NewAccountRepository()
		sessionRepo = memory.NewSessionRepository()
		workspacePlanRepo = memory.NewWorkspacePlanRepository()
		// sample:start
		itemRepo = memory.NewItemRepository()
		// sample:end
	case config.StorageJSONFile:
		store, err := jsonfile.Open(cfg.DataDir)
		if err != nil {
			return nil, err
		}
		a.ready = store.Ready
		if warning, ok := conflictCopyWarning(store); ok {
			a.Warnings = append(a.Warnings, warning)
		}
		if warning, ok := otherDeviceWarning(cfg); ok {
			a.Warnings = append(a.Warnings, warning)
		}
		stopMarking := markFolder(cfg)
		a.close = func() error {
			stopMarking()
			return store.Close()
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
		accountRepo = sqlstore.NewAccountRepository(store.DB())
		sessionRepo = sqlstore.NewSessionRepository(store.DB())
		workspacePlanRepo = sqlstore.NewWorkspacePlanRepository(store.DB())
		// sample:start
		itemRepo = sqlstore.NewItemRepository(store.DB())
		// sample:end
	case config.StoragePostgres:
		if cfg.DatabaseURL == "" {
			return nil, errors.New(`The storage setting is "postgres", but the database_url setting is missing. Please set database_url in the settings file or the STARTER_DATABASE_URL environment variable.`)
		}
		store, err := postgres.Open(context.Background(), cfg.DatabaseURL)
		if err != nil {
			return nil, err
		}
		a.close = store.Close
		a.ready = store.Ready
		accountRepo = sqlstore.NewAccountRepository(store.DB())
		sessionRepo = sqlstore.NewSessionRepository(store.DB())
		workspacePlanRepo = sqlstore.NewWorkspacePlanRepository(store.DB())
		// sample:start
		itemRepo = sqlstore.NewItemRepository(store.DB())
		// sample:end
	default:
		return nil, config.UnrecognizedStorageError(cfg.Storage)
	}
	entitlementsProvider, err := entitlementsProviderFor(cfg, workspacePlanRepo)
	if err != nil {
		return nil, err
	}
	a.Entitlements = app.NewEntitlementsService(entitlementsProvider, workspacesFor(cfg))
	if cfg.SignIn.IsOn() {
		a.SignIn = app.NewSignInService(accountRepo, sessionRepo, system.Tokens{}, system.UUIDs{}, system.Clock{})
		a.billing = app.NewBillingService(workspacePlanRepo, entitlementsProvider)
	}
	// Build each use case on its repository and the driven ports:
	// workspacesFor(cfg) (the session's workspace with sign-in on, else the
	// one local workspace, ADR 5), entitlementsProviderFor(cfg),
	// system.Clock{}, and system.UUIDs{}. See docs/workspaces.md.
	// sample:start
	a.Items = app.NewItemService(itemRepo, workspacesFor(cfg), entitlementsProvider, system.Clock{}, system.UUIDs{})
	// sample:end
	return a, nil
}

// workspacesFor picks how use cases find the current workspace: with sign-in
// on (the hosted edition), the one the session middleware put in the request's
// context, so a request signed out has none; otherwise the local edition's one
// workspace (ADR 5).
func workspacesFor(cfg config.Config) ports.WorkspaceResolver {
	if cfg.SignIn.IsOn() {
		return workspace.FromContext{}
	}
	return workspace.Local{}
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
	file, err := loadPlansFile(cfg.PlansFile)
	if err != nil {
		return nil, err
	}
	return app.NewEntitlementsService(localEntitlementsProvider(file), workspace.Local{}), nil
}

// ErrLocalPlanIsFixed is the refusal to change a plan in the local edition,
// whose plan is LocalPlan.
var ErrLocalPlanIsFixed = errors.New("The plan for the local edition is set when the application is built and cannot be changed.")

// ErrNoPlanChangesOverAServer is the refusal to change a plan with --server.
var ErrNoPlanChangesOverAServer = errors.New("A plan cannot be changed through a server connection.")

// NewBilling returns the billing service, for `starter admin set-plan`. With
// sign-in on (the hosted edition), it opens the configured storage as
// NewEmbedded does and changes the plans stored there, which the hosted
// edition's entitlements read. The storage stays open until the command's
// program ends. The local edition's plan is fixed, so otherwise it refuses with
// ErrLocalPlanIsFixed, or ErrNoPlanChangesOverAServer with cfg.ServerURL set,
// opening no storage.
func NewBilling(cfg config.Config) (ports.BillingService, error) {
	if cfg.ServerURL != "" {
		return nil, ErrNoPlanChangesOverAServer
	}
	if cfg.SignIn.IsOn() {
		a, err := NewEmbedded(cfg)
		if err != nil {
			return nil, err
		}
		return a.billing, nil
	}
	return nil, ErrLocalPlanIsFixed
}

// LockedFolderError returns the error for a locked data folder that also
// offers the user alternative, a phrase that follows "or", or nil if err isn't
// that error.
func LockedFolderError(err error, alternative string) error {
	var locked *jsonfile.LockedError
	if !errors.As(err, &locked) {
		return nil
	}
	return &jsonfile.LockedError{Path: locked.Path, Alternative: alternative}
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

// markFolder leaves this device's mark in the data folder and renews it every
// minute, until stop ends the renewing and removes the mark. It marks only a
// folder that looks synced, when the setting doesn't say to ignore that, so an
// unsynced folder gets no background writes. If this computer's name can't be
// read, it leaves none. With no mark, stop does nothing.
func markFolder(cfg config.Config) (stop func()) {
	noMark := func() {}
	if cfg.ShouldIgnoreSyncedFolder {
		return noMark
	}
	if _, ok := syncedfolder.Detect(cfg.DataDir); !ok {
		return noMark
	}
	device, err := os.Hostname()
	if err != nil {
		return noMark
	}
	_ = syncedfolder.WriteMark(cfg.DataDir, device, system.Clock{}.Now())
	ticker := time.NewTicker(time.Minute)
	stopKeeping := syncedfolder.KeepMark(cfg.DataDir, device, ticker.C)
	return func() {
		ticker.Stop()
		stopKeeping()
	}
}

// deviceList names two or more devices: "a and b", or "a, b, and c".
func deviceList(devices []string) string {
	last := len(devices) - 1
	if last == 1 {
		return devices[0] + " and " + devices[1]
	}
	return strings.Join(devices[:last], ", ") + ", and " + devices[last]
}

// entitlementsProviderFor picks the plans each workspace is on: with sign-in on
// (the hosted edition), hostedEntitlementsProvider's, read from workspacePlans;
// otherwise the local edition's one plan, localEntitlementsProvider's.
func entitlementsProviderFor(cfg config.Config, workspacePlans ports.WorkspacePlanRepository) (plans.Provider, error) {
	file, err := loadPlansFile(cfg.PlansFile)
	if err != nil {
		return plans.Provider{}, err
	}
	if cfg.SignIn.IsOn() {
		return hostedEntitlementsProvider(file, workspacePlans), nil
	}
	return localEntitlementsProvider(file), nil
}

// loadPlansFile returns the plans file at plansFile, or else the one built into
// the binary. It refuses a file without LocalPlan.
func loadPlansFile(plansFile string) (plans.File, error) {
	file, err := plans.Load(plansFile)
	if err != nil {
		return plans.File{}, err
	}
	if _, ok := file.Catalog[LocalPlan]; !ok {
		return plans.File{}, fmt.Errorf("local plan %q isn't in the plans file", LocalPlan)
	}
	return file, nil
}

// localEntitlementsProvider answers LocalPlan for every workspace: the local
// edition has only one. It gives no upgrade hints, since a local copy can't
// change plan.
func localEntitlementsProvider(file plans.File) plans.Provider {
	return plans.Provider{
		Catalog: file.Catalog,
		PlanOf: func(context.Context, domain.WorkspaceID) (string, error) {
			return LocalPlan, nil
		},
	}
}

// hostedEntitlementsProvider answers each workspace's stored plan, or the
// file's NewWorkspacePlan for a workspace with none, with upgrade hints that
// never name LocalPlan, which the hosted edition doesn't offer.
func hostedEntitlementsProvider(file plans.File, workspacePlans ports.WorkspacePlanRepository) plans.Provider {
	return plans.Provider{
		Catalog:         file.Catalog,
		PlanOf:          plans.StoredPlanOf(workspacePlans, file.NewWorkspacePlan),
		ShouldGiveHints: true,
		LocalPlan:       LocalPlan,
	}
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
