// Package acceptance runs the Gherkin features in /features with Godog.
//
// Each scenario runs once per driver. A driver is a way of reaching the core
// through its driving ports, so the same business scenarios prove the core,
// each storage adapter, and the HTTP server+client adapters.
package acceptance

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/desktop"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient/apiclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/mobile"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/jsonfile"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/memory"
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
}

type driver struct {
	name  string
	start func() (*harness, error)
	// tags limits the scenarios this driver runs, as a Godog tag expression
	// combined with --godog.tags. Empty runs every scenario.
	tags string
}

// Desktop-only scenarios (@desktop) run only on the desktop driver, and
// mobile-only ones (@mobile) only on the mobile driver. Each of those also
// runs every feature that isn't the other's. Scenarios that switch between
// workspaces (@workspaces) run only where the core is wired directly: nothing
// over HTTP can pick a workspace until sign-in exists, and the desktop and
// mobile apps are the local edition. Scenarios that choose each workspace's
// plan (@plans) run only where the tests wire the plans: the desktop and
// mobile apps are on the local plan. Scenarios about running the server
// (@server) run only on the server driver, which runs nothing else: they
// need its data folder, and the http driver covers the API.
var drivers = []driver{
	{name: "memory", start: startMemory, tags: "~@desktop && ~@mobile && ~@server"},
	{name: "jsonfile", start: startJSONFile, tags: "~@desktop && ~@mobile && ~@server"},
	{name: "sqlite", start: startSQLite, tags: "~@desktop && ~@mobile && ~@server"},
	{name: "http", start: startHTTP, tags: "~@desktop && ~@mobile && ~@workspaces && ~@server"},
	{name: "desktop", start: startDesktop, tags: "~@mobile && ~@workspaces && ~@plans && ~@server"},
	{name: "mobile", start: startMobile, tags: "~@desktop && ~@workspaces && ~@plans && ~@server"},
	{name: "server", start: startServer, tags: "@server"},
}

func newHarness() *harness {
	return &harness{
		clock:   testsupport.NewFixedClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
		cleanup: func() {},
	}
}

// usePlans gives the harness its own plans, with each workspace's plan in
// stored, and the entitlements use case on them, in the workspace that
// workspaces resolves. An admin moves workspaces on stored directly, as
// `starter admin set-plan` would.
func (h *harness) usePlans(stored ports.WorkspacePlanRepository, workspaces ports.WorkspaceResolver) error {
	p, err := newTestPlans(stored)
	if err != nil {
		return err
	}
	h.plans = p
	h.entitlements = app.NewEntitlementsService(p, workspaces)
	h.billing = app.NewBillingService(stored, p)
	return nil
}

// sample:start

// newItemService builds the Item use cases on repo, driven by the harness's clock,
// in the workspace that workspaces resolves.
func (h *harness) newItemService(repo ports.ItemRepository, workspaces ports.WorkspaceResolver) *app.ItemService {
	return app.NewItemService(repo, workspaces, h.plans, h.clock, testsupport.NewSequentialIDs("item"))
}

// sample:end

func startMemory() (*harness, error) {
	h := newHarness()
	if err := h.usePlans(memory.NewWorkspacePlanRepository(), workspace.FromContext{}); err != nil {
		return nil, err
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
	if err := h.usePlans(jsonfile.NewWorkspacePlanRepository(store), workspace.FromContext{}); err != nil {
		h.cleanup()
		return nil, err
	}
	// sample:start
	h.items = h.newItemService(jsonfile.NewItemRepository(store), workspace.FromContext{})
	// sample:end
	return h, nil
}

func startSQLite() (*harness, error) {
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
	if err := h.usePlans(sqlstore.NewWorkspacePlanRepository(store.DB()), workspace.FromContext{}); err != nil {
		h.cleanup()
		return nil, err
	}
	// sample:start
	h.items = h.newItemService(sqlstore.NewItemRepository(store.DB()), workspace.FromContext{})
	// sample:end
	return h, nil
}

func startHTTP() (*harness, error) {
	h := newHarness()
	// Nothing over HTTP carries a workspace yet, so the server is the local
	// edition, as cmd/server is.
	if err := h.usePlans(memory.NewWorkspacePlanRepository(), workspace.Local{}); err != nil {
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

// driveOverHTTP points the harness's driving ports at the server at url,
// through the HTTP client adapter.
func (h *harness) driveOverHTTP(url string) error {
	h.apiURL = url
	entitlements, err := httpclient.NewEntitlementsService(url)
	if err != nil {
		return err
	}
	h.entitlements = entitlements
	// sample:start
	client, err := httpclient.NewItemService(url)
	if err != nil {
		return err
	}
	h.items = client
	// sample:end
	return nil
}

// startServer wires the core as cmd/server does, on jsonfile storage in a
// temporary data folder, and serves it over HTTP.
func startServer() (*harness, error) {
	dir, err := os.MkdirTemp("", "acceptance-server-*")
	if err != nil {
		return nil, err
	}
	h := newHarness()
	h.server = &testServer{dataDir: dir}
	h.cleanup = func() {
		h.server.stop()
		os.RemoveAll(dir)
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
	close   func()
}

// start wires the core as cmd/server does, on storage in the server's data
// folder, and serves it over HTTP.
func (s *testServer) start(storage config.StorageBackend) error {
	core, err := bootstrap.NewEmbedded(config.Config{Storage: storage, DataDir: s.dataDir})
	if err != nil {
		return err
	}
	srv := httptest.NewServer(httpapi.NewHandler(httpapi.Options{
		Entitlements: core.Entitlements,
		Ready:        core.Ready,
		Warnings:     core.Warnings,
	}))
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
		second, err := desktop.Open(app.getenv, app.userConfigDir, nil)
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
	app, err := desktop.Open(d.getenv, d.userConfigDir, nil)
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
