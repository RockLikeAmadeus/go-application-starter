// Package acceptance runs the Gherkin features in /features with Godog.
//
// Each scenario runs once per driver. A driver is a way of reaching the core
// through its driving ports, so the same business scenarios prove the core,
// each storage adapter, and the HTTP server+client adapters.
package acceptance

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/desktop"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient/apiclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/mobile"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/oidc"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/oidc/oidctest"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/jsonfile"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/memory"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/postgres"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/postgres/postgrestest"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlite"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlstore"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/workspace"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/app"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
	"github.com/RockLikeAmadeus/go-application-starter/internal/testsupport"
)

// harness is everything a scenario needs: the driving ports under test, the
// clock controlling them, and a cleanup function.
type harness struct {
	// sample:start
	items ports.ItemService
	// sample:end
	entitlements ports.EntitlementsService
	// plans sets the plans the core answers, and billing moves workspaces
	// between them. Only the drivers that wire the core themselves set them:
	// the desktop and mobile apps are on the local plan.
	plans   *testPlans
	billing ports.BillingService
	clock   *testsupport.FixedClock
	cleanup func()
	// reopen closes the desktop or mobile app and opens it again on the same
	// data folder. Only the desktop and mobile drivers set it.
	reopen func() error
	// openCopy tries to open a second copy of the desktop app on the same
	// data folder, closing it again if it opens. Only the desktop driver sets it.
	openCopy func() error
	// closeApp closes the desktop app, releasing its data folder. Only the
	// desktop driver sets it.
	closeApp func()
	// local is the computer the desktop app runs on, where the command line
	// runs beside it. Only the desktop driver sets it.
	local *localComputer
	// phoneStorage is the mobile app's private storage on the phone. Only
	// the mobile driver sets it.
	phoneStorage string
	// server is the running server, as cmd/server wires it. Only the server
	// driver sets it.
	server *testServer
	// apiURL is the address of the API the driving ports are reached
	// through, when they're reached over HTTP.
	apiURL string
	// signIn signs scenarios in and out, and accounts holds the accounts
	// they sign in to. Only the drivers running the hosted edition set them.
	signIn   signInDriver
	accounts ports.AccountRepository
	// nativeSwaps keeps the latest sign-in code the command line swapped for
	// a session token. Only the drivers whose server signs people in set it.
	nativeSwaps *nativeSwapRecorder
}

type driver struct {
	name string
	// start starts the driver as the local edition.
	start func() (*harness, error)
	// startHosted starts the driver as the hosted edition, with sign-in on,
	// for the scenarios that sign in (@sign-in). Only the drivers whose
	// storage keeps accounts and sessions set it.
	startHosted func() (*harness, error)
	// tags limits the scenarios this driver runs, as a Godog tag expression
	// combined with --godog.tags. Empty runs every scenario.
	tags string
	// needsSignInForWorkspaces is whether the driver can work in a named
	// workspace only by signing in to the account of that name, whose
	// personal workspace it is. Then it runs @workspaces as the hosted
	// edition too.
	needsSignInForWorkspaces bool
	// skipReason, when set, is why this driver can't run here.
	skipReason string
}

// driversFor lists the drivers, for a test Postgres server at serverURL, which
// is empty when there isn't one. Then the Postgres driver is left out with a
// note, and the server driver leaves out the scenarios that need Postgres.
//
// Desktop-only scenarios (@desktop) run only on the desktop driver, and
// mobile-only ones (@mobile) only on the mobile driver. Each of those also
// runs every feature that isn't the other's. Scenarios that switch between
// workspaces (@workspaces) run where the core is wired directly, and over
// HTTP as the hosted edition, signing in to each workspace's account. The
// desktop and mobile apps are the local edition. Scenarios that choose each workspace's
// plan (@plans) run only where the tests wire the plans: the desktop and
// mobile apps are on the local plan. Scenarios about running the server
// (@server) run only on the server driver, which runs nothing else: they
// need its data folder, and the http driver covers the API. Scenarios that
// sign in (@sign-in) run only on the drivers whose storage keeps accounts and
// sessions, as the hosted edition: memory, sqlite, postgres, and http.
// Scenarios that open the desktop and mobile apps on a hosted server
// (@remote-mode) run only on the http and oidc drivers, whose servers sign
// people in, as the command line's sign-in (@command-line-sign-in) does.
func driversFor(t testing.TB, serverURL string) []driver {
	postgresDriver := driver{
		name:        "postgres",
		start:       func() (*harness, error) { return startPostgres(t, localEdition) },
		startHosted: func() (*harness, error) { return startPostgres(t, hostedEdition) },
		tags: "~@desktop && ~@mobile && ~@server && ~@sign-in-provider && ~@command-line-sign-in" +
			" && ~@remote-mode",
	}
	serverTags := "@server"
	if serverURL == "" {
		postgresDriver.skipReason = postgrestest.ServerURLVariable + " is not set: leaving out the Postgres driver (task db:up starts a server)"
		serverTags = "@server && ~@postgres"
	}
	return []driver{
		{
			name:        "memory",
			start:       func() (*harness, error) { return startMemory(localEdition) },
			startHosted: func() (*harness, error) { return startMemory(hostedEdition) },
			tags: "~@desktop && ~@mobile && ~@server && ~@sign-in-provider && ~@command-line-sign-in" +
				" && ~@remote-mode",
		},
		{
			name:  "jsonfile",
			start: startJSONFile,
			tags: "~@desktop && ~@mobile && ~@server && ~@sign-in && ~@sign-in-provider" +
				" && ~@command-line-sign-in && ~@remote-mode",
		},
		{
			name:        "sqlite",
			start:       func() (*harness, error) { return startSQLite(localEdition) },
			startHosted: func() (*harness, error) { return startSQLite(hostedEdition) },
			tags: "~@desktop && ~@mobile && ~@server && ~@sign-in-provider && ~@command-line-sign-in" +
				" && ~@remote-mode",
		},
		postgresDriver,
		{
			name:                     "http",
			start:                    startHTTP,
			startHosted:              startHostedHTTP,
			tags:                     "~@desktop && ~@mobile && ~@server && ~@sign-in-provider",
			needsSignInForWorkspaces: true,
		},
		{
			name:        "oidc",
			startHosted: func() (*harness, error) { return startHostedOIDC(t) },
			tags:        "@sign-in,@sign-in-provider,@command-line-sign-in,@remote-mode",
		},
		{
			name:  "desktop",
			start: startDesktop,
			tags: "~@mobile && ~@workspaces && ~@plans && ~@server && ~@sign-in && ~@sign-in-provider" +
				" && ~@command-line-sign-in && ~@remote-mode",
		},
		{
			name:  "mobile",
			start: startMobile,
			tags: "~@desktop && ~@workspaces && ~@plans && ~@server && ~@sign-in && ~@sign-in-provider" +
				" && ~@command-line-sign-in && ~@remote-mode",
		},
		{name: "server", start: func() (*harness, error) { return startServer(t) }, tags: serverTags},
	}
}

// edition is which of the two editions (ADR 5) a driver runs the core as.
type edition int

const (
	// localEdition is one workspace, on the local plan, with no sign-in.
	localEdition edition = iota
	// hostedEdition signs people in, each to their own workspace, which
	// starts on the plans file's newWorkspacePlan.
	hostedEdition
)

func newHarness() *harness {
	return &harness{
		clock:   testsupport.NewFixedClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
		cleanup: func() {},
	}
}

// usePlans gives the harness its own plans, with each workspace's plan in
// stored, and the entitlements use case on them, in the workspace that
// workspaces resolves. A workspace with no stored plan is on the edition's
// starting plan. An admin moves workspaces on stored directly, as
// `starter admin set-plan` would.
func (h *harness) usePlans(stored ports.WorkspacePlanRepository, workspaces ports.WorkspaceResolver, e edition) error {
	p, err := newTestPlans(stored, e)
	if err != nil {
		return err
	}
	h.plans = p
	h.entitlements = app.NewEntitlementsService(p, workspaces)
	h.billing = app.NewBillingService(stored, p)
	return nil
}

// useSQLStore keeps the scenario's plans and data in db, through the shared
// SQL repositories, in the workspace named by the context. As the hosted
// edition, it keeps accounts and sessions there too.
func (h *harness) useSQLStore(db *sql.DB, e edition) error {
	if err := h.usePlans(sqlstore.NewWorkspacePlanRepository(db), workspace.FromContext{}, e); err != nil {
		return err
	}
	if e == hostedEdition {
		h.useDirectSignIn(sqlstore.NewAccountRepository(db), sqlstore.NewSessionRepository(db))
	}
	// sample:start
	h.items = h.newItemService(sqlstore.NewItemRepository(db), workspace.FromContext{})
	// sample:end
	return nil
}

// sample:start

// newItemService builds the Item use cases on repo, driven by the harness's clock,
// in the workspace that workspaces resolves.
func (h *harness) newItemService(repo ports.ItemRepository, workspaces ports.WorkspaceResolver) *app.ItemService {
	return app.NewItemService(repo, workspaces, h.plans, h.clock, testsupport.NewSequentialIDs("item"))
}

// sample:end

func startMemory(e edition) (*harness, error) {
	h := newHarness()
	if err := h.usePlans(memory.NewWorkspacePlanRepository(), workspace.FromContext{}, e); err != nil {
		return nil, err
	}
	if e == hostedEdition {
		h.useDirectSignIn(memory.NewAccountRepository(), memory.NewSessionRepository())
	}
	// sample:start
	h.items = h.newItemService(memory.NewItemRepository(), workspace.FromContext{})
	// sample:end
	return h, nil
}

func startJSONFile() (*harness, error) {
	dir, err := os.MkdirTemp("", "acceptance-jsonfile-*")
	if err != nil {
		return nil, err
	}
	store, err := jsonfile.Open(dir)
	if err != nil {
		os.RemoveAll(dir)
		return nil, err
	}
	h := newHarness()
	h.cleanup = func() {
		store.Close()
		os.RemoveAll(dir)
	}
	if err := h.usePlans(jsonfile.NewWorkspacePlanRepository(store), workspace.FromContext{}, localEdition); err != nil {
		h.cleanup()
		return nil, err
	}
	// sample:start
	h.items = h.newItemService(jsonfile.NewItemRepository(store), workspace.FromContext{})
	// sample:end
	return h, nil
}

func startSQLite(e edition) (*harness, error) {
	dir, err := os.MkdirTemp("", "acceptance-sqlite-*")
	if err != nil {
		return nil, err
	}
	store, err := sqlite.Open(dir)
	if err != nil {
		os.RemoveAll(dir)
		return nil, err
	}
	h := newHarness()
	h.cleanup = func() {
		store.Close()
		os.RemoveAll(dir)
	}
	if err := h.useSQLStore(store.DB(), e); err != nil {
		h.cleanup()
		return nil, err
	}
	return h, nil
}

// startPostgres keeps the scenario's data in a database of its own, dropped
// when the scenario ends.
func startPostgres(t testing.TB, e edition) (*harness, error) {
	scenario := &scenarioTest{TB: t}
	store, err := postgres.Open(t.Context(), postgrestest.NewDatabase(scenario))
	if err != nil {
		scenario.runCleanups()
		return nil, err
	}
	h := newHarness()
	h.cleanup = func() {
		store.Close()
		scenario.runCleanups()
	}
	if err := h.useSQLStore(store.DB(), e); err != nil {
		h.cleanup()
		return nil, err
	}
	return h, nil
}

// scenarioTest is the driver's test, except that what is left to clean up
// waits for the end of the scenario rather than of the driver's whole run.
// Godog runs scenarios as subtests that a driver can't reach, and databases
// kept to the end of the run would use up the server's connections.
type scenarioTest struct {
	testing.TB
	cleanups []func()
}

// Cleanup registers f to run when the scenario ends.
func (s *scenarioTest) Cleanup(f func()) {
	s.cleanups = append(s.cleanups, f)
}

// runCleanups runs what was registered, last first.
func (s *scenarioTest) runCleanups() {
	for _, cleanup := range slices.Backward(s.cleanups) {
		cleanup()
	}
	s.cleanups = nil
}

func startHTTP() (*harness, error) {
	h := newHarness()
	// With sign-in off, nothing over HTTP carries a workspace, so the server
	// is the local edition, as cmd/server is by default.
	if err := h.usePlans(memory.NewWorkspacePlanRepository(), workspace.Local{}, localEdition); err != nil {
		return nil, err
	}
	opts := httpapi.Options{Entitlements: h.entitlements}
	// sample:start
	opts.Items = h.newItemService(memory.NewItemRepository(), workspace.Local{})
	// sample:end
	srv := httptest.NewServer(httpapi.NewHandler(opts))
	h.cleanup = srv.Close
	if err := h.driveOverHTTP(srv.URL); err != nil {
		srv.Close()
		return nil, err
	}
	return h, nil
}

// startHostedHTTP serves the hosted edition over HTTP, with the dev sign-in
// on, keeping its plans, data, accounts, and sessions in memory. Scenarios
// sign in through the dev sign-in, as a browser does.
func startHostedHTTP() (*harness, error) {
	h := newHarness()
	if err := h.usePlans(memory.NewWorkspacePlanRepository(), workspace.FromContext{}, hostedEdition); err != nil {
		return nil, err
	}
	h.accounts = memory.NewAccountRepository()
	opts := httpapi.Options{
		Entitlements:  h.entitlements,
		SignIn:        h.newSignInService(h.accounts, memory.NewSessionRepository()),
		IsDevSignInOn: true,
	}
	// sample:start
	opts.Items = h.newItemService(memory.NewItemRepository(), workspace.FromContext{})
	// sample:end
	h.nativeSwaps = &nativeSwapRecorder{}
	srv := httptest.NewServer(h.nativeSwaps.recording(httpapi.NewHandler(opts)))
	h.cleanup = srv.Close
	h.signIn = newDevSignInOverHTTP(h, srv.URL)
	if err := h.driveOverHTTP(srv.URL); err != nil {
		srv.Close()
		return nil, err
	}
	return h, nil
}

// startHostedOIDC serves the hosted edition over HTTP, with sign-in through
// a fake sign-in provider, keeping its plans, data, accounts, and sessions in
// memory. Scenarios sign in through the provider, as a browser does.
func startHostedOIDC(t testing.TB) (*harness, error) {
	h := newHarness()
	if err := h.usePlans(memory.NewWorkspacePlanRepository(), workspace.FromContext{}, hostedEdition); err != nil {
		return nil, err
	}
	h.accounts = memory.NewAccountRepository()
	// The provider is sent back to the server's address, which is known
	// before the server starts serving.
	srv := httptest.NewUnstartedServer(nil)
	publicURL := "http://" + srv.Listener.Addr().String()
	provider := oidctest.New(t)
	provider.SetPromptCreateSupported(true)
	sessions := memory.NewSessionRepository()
	opts := httpapi.Options{
		Entitlements: h.entitlements,
		SignIn:       h.newSignInService(h.accounts, sessions),
		PublicURL:    publicURL,
	}
	// sample:start
	opts.Items = h.newItemService(memory.NewItemRepository(), workspace.FromContext{})
	// sample:end
	// The server discovers the provider when it starts, so a scenario that
	// changes the provider starts the handler again.
	h.nativeSwaps = &nativeSwapRecorder{}
	var handler atomic.Pointer[http.Handler]
	useProvider := func() error {
		signInProvider, err := oidc.New(context.Background(), oidc.Settings{
			Issuer:       provider.Issuer(),
			ClientID:     oidctest.ClientID,
			ClientSecret: oidctest.ClientSecret,
			RedirectURL:  publicURL + "/auth/oidc/callback",
		})
		if err != nil {
			return err
		}
		opts.SignInProvider = signInProvider
		next := httpapi.NewHandler(opts)
		handler.Store(&next)
		return nil
	}
	if err := useProvider(); err != nil {
		srv.Close()
		return nil, err
	}
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.nativeSwaps.recording(*handler.Load()).ServeHTTP(w, r)
	})
	srv.Start()
	h.cleanup = func() {
		srv.Close()
		provider.Close()
	}
	signIn := newOIDCSignInOverHTTP(h, srv.URL, provider)
	signIn.useProvider = useProvider
	h.signIn = signIn
	if err := h.driveOverHTTP(srv.URL); err != nil {
		h.cleanup()
		return nil, err
	}
	return h, nil
}

// driveOverHTTP points the harness's driving ports at the server at url,
// through the HTTP client adapter built with opts.
func (h *harness) driveOverHTTP(url string, opts ...apiclient.ClientOption) error {
	h.apiURL = url
	entitlements, err := httpclient.NewEntitlementsService(url, opts...)
	if err != nil {
		return err
	}
	h.entitlements = entitlements
	// sample:start
	client, err := httpclient.NewItemService(url, opts...)
	if err != nil {
		return err
	}
	h.items = client
	// sample:end
	return nil
}

// startServer wires the core as cmd/server does, on jsonfile storage in a
// temporary data folder, and serves it over HTTP.
func startServer(t testing.TB) (*harness, error) {
	dir, err := os.MkdirTemp("", "acceptance-server-*")
	if err != nil {
		return nil, err
	}
	scenario := &scenarioTest{TB: t}
	h := newHarness()
	h.server = &testServer{dataDir: dir, scenario: scenario}
	h.cleanup = func() {
		h.server.stop()
		os.RemoveAll(dir)
		scenario.runCleanups()
	}
	if err := h.server.start(config.StorageJSONFile); err != nil {
		h.cleanup()
		return nil, err
	}
	return h, nil
}

// testServer is a running server and the data folder it keeps its data in.
type testServer struct {
	url     string
	dataDir string
	// databaseURL is the Postgres database the server keeps its data in, when
	// it does.
	databaseURL string
	// scenario holds what is left to clean up when the scenario ends.
	scenario *scenarioTest
	close    func()
}

// start wires the core as cmd/server does, on storage in the server's data
// folder, and serves it over HTTP.
func (s *testServer) start(storage config.StorageBackend) error {
	core, err := bootstrap.NewEmbedded(config.Config{Storage: storage, DataDir: s.dataDir, DatabaseURL: s.databaseURL})
	if err != nil {
		return err
	}
	opts := httpapi.Options{
		Entitlements: core.Entitlements,
		Ready:        core.Ready,
		Warnings:     core.Warnings,
	}
	// sample:start
	opts.Items = core.Items
	// sample:end
	srv := httptest.NewServer(httpapi.NewHandler(opts))
	s.url = srv.URL
	s.close = func() {
		srv.Close()
		core.Close()
	}
	return nil
}

// stop stops the server, if it is running.
func (s *testServer) stop() {
	if s.close != nil {
		s.close()
		s.close = nil
	}
}

// restart stops the server and starts it again on storage, in the same data
// folder.
func (s *testServer) restart(storage config.StorageBackend) error {
	s.stop()
	return s.start(storage)
}

// readiness asks the server whether it is ready, as a deployment's probe
// does, and returns the status and body of its answer.
func (s *testServer) readiness() (int, string, error) {
	resp, err := http.Get(s.url + "/readyz")
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body), err
}

// startDesktop opens the desktop app's wiring on a temporary settings
// folder, with no data folder chosen, and drives it over HTTP, as the desktop
// window's SPA does.
func startDesktop() (*harness, error) {
	dir, err := os.MkdirTemp("", "acceptance-desktop-*")
	if err != nil {
		return nil, err
	}
	local := &localComputer{dir: dir, env: map[string]string{}}
	app := &desktopApp{getenv: local.getenv, userConfigDir: local.userConfigDir}

	h := newHarness()
	h.local = local
	h.cleanup = func() {
		app.close()
		os.RemoveAll(dir)
	}
	h.reopen = func() error {
		app.close()
		return app.open(h)
	}
	h.openCopy = func() error {
		second, err := desktop.Open(app.getenv, app.userConfigDir, nil, nil)
		if err != nil {
			return err
		}
		return second.Close()
	}
	h.closeApp = app.close
	if err := app.open(h); err != nil {
		h.cleanup()
		return nil, err
	}
	return h, nil
}

// desktopApp is one copy of the desktop app, opened on the data folder that
// getenv and userConfigDir lead to, and served over HTTP.
type desktopApp struct {
	getenv        func(string) string
	userConfigDir func() (string, error)
	app           *desktop.App
	srv           *httptest.Server
}

// localComputer is the user's computer, as the desktop app and the command
// line see it: an environment and a settings folder, inside dir.
type localComputer struct {
	dir string
	env map[string]string
}

func (c *localComputer) getenv(key string) string { return c.env[key] }

func (c *localComputer) userConfigDir() (string, error) {
	return filepath.Join(c.dir, "settings"), nil
}

// configEnv is the computer as configuration sees it.
func (c *localComputer) configEnv() config.Env {
	return config.Env{Getenv: c.getenv, UserConfigDir: c.userConfigDir}
}

// open opens the app and points h's driving ports at it. The scenarios drive
// the API only, so the app is opened without the web app.
func (d *desktopApp) open(h *harness) error {
	app, err := desktop.Open(d.getenv, d.userConfigDir, nil, nil)
	if err != nil {
		return err
	}
	d.app, d.srv = app, httptest.NewServer(app.Handler)
	return h.driveOverHTTP(d.srv.URL)
}

// close stops serving the app and releases its data folder, if it is open.
func (d *desktopApp) close() {
	if d.app == nil {
		return
	}
	d.srv.Close()
	d.app.Close()
	d.app, d.srv = nil, nil
}

// startMobile opens the mobile app's wiring on a temporary storage folder, as
// the phone does, and drives it over the bridge, as the phone's web app does.
func startMobile() (*harness, error) {
	dir, err := os.MkdirTemp("", "acceptance-mobile-*")
	if err != nil {
		return nil, err
	}
	app := &mobileApp{storagePath: dir}

	h := newHarness()
	h.phoneStorage = dir
	h.cleanup = func() {
		app.close()
		os.RemoveAll(dir)
	}
	h.reopen = func() error {
		app.close()
		return app.open(h)
	}
	if err := app.open(h); err != nil {
		h.cleanup()
		return nil, err
	}
	return h, nil
}

// mobileApp is one copy of the mobile app, opened on storagePath.
type mobileApp struct {
	storagePath string
	app         *mobile.App
}

// open opens the app and points h's driving ports at its bridge. The
// scenarios drive the API only, so the app is opened without the web app.
func (m *mobileApp) open(h *harness) error {
	app, err := mobile.Open(func(string) string { return "" }, m.storagePath, nil)
	if err != nil {
		return err
	}
	m.app = app
	// The bridge ignores the host: only the path and query reach mobile.Serve.
	const host = "http://mobile.invalid"
	entitlements, err := httpclient.NewEntitlementsService(host, apiclient.WithHTTPClient(bridge{app.API}))
	if err != nil {
		return err
	}
	h.entitlements = entitlements
	// sample:start
	client, err := httpclient.NewItemService(host, apiclient.WithHTTPClient(bridge{app.API}))
	if err != nil {
		return err
	}
	h.items = client
	// sample:end
	return nil
}

// close releases the app's data folder, if it is open.
func (m *mobileApp) close() {
	if m.app == nil {
		return
	}
	m.app.Close()
	m.app = nil
}

// bridge sends the HTTP client's requests through mobile.Serve, as the
// phone's web app sends them over the Wails bridge.
type bridge struct {
	api http.Handler
}

func (b bridge) Do(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		var err error
		if body, err = io.ReadAll(req.Body); err != nil {
			return nil, err
		}
	}
	reply := mobile.Serve(b.api, req.Method, req.URL.RequestURI(), string(body))
	return &http.Response{
		StatusCode: reply.Status,
		Header:     http.Header{"Content-Type": {reply.ContentType}},
		Body:       io.NopCloser(strings.NewReader(reply.Body)),
		Request:    req,
	}, nil
}
