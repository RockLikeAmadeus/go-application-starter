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
	"strings"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/desktop"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient/apiclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/mobile"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/jsonfile"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/memory"
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
	clock   *testsupport.FixedClock
	cleanup func()
	// reopen closes the desktop or mobile app and opens it again on the same
	// data folder. Only the desktop and mobile drivers set it.
	reopen func() error
	// openCopy tries to open a second copy of the desktop app on the same
	// data folder, closing it again if it opens. Only the desktop driver sets it.
	openCopy func() error
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
// mobile apps are the local edition.
var drivers = []driver{
	{name: "memory", start: startMemory, tags: "~@desktop && ~@mobile"},
	{name: "jsonfile", start: startJSONFile, tags: "~@desktop && ~@mobile"},
	{name: "http", start: startHTTP, tags: "~@desktop && ~@mobile && ~@workspaces"},
	{name: "desktop", start: startDesktop, tags: "~@mobile && ~@workspaces"},
	{name: "mobile", start: startMobile, tags: "~@desktop && ~@workspaces"},
}

func newHarness() *harness {
	return &harness{
		clock:   testsupport.NewFixedClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
		cleanup: func() {},
	}
}

// sample:start

// newItems builds the Item use cases on repo, driven by the harness's clock.
func (h *harness) newItems(repo ports.ItemRepository) *app.ItemService {
	return app.NewItemService(repo, h.clock, testsupport.NewSequentialIDs("item"))
}

// sample:end

func startMemory() (*harness, error) {
	h := newHarness()
	// sample:start
	h.items = h.newItems(memory.NewItemRepository())
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
		return nil, err
	}
	h := newHarness()
	h.cleanup = func() {
		store.Close()
		os.RemoveAll(dir)
	}
	// sample:start
	h.items = h.newItems(jsonfile.NewItemRepository(store))
	// sample:end
	return h, nil
}

func startHTTP() (*harness, error) {
	h := newHarness()
	var opts httpapi.Options
	// sample:start
	opts.Items = h.newItems(memory.NewItemRepository())
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
	// sample:start
	client, err := httpclient.New(url)
	if err != nil {
		return err
	}
	h.items = client
	// sample:end
	return nil
}

// startDesktop opens the desktop app's wiring on a temporary data folder and
// drives it over HTTP, as the desktop window's SPA does.
func startDesktop() (*harness, error) {
	dir, err := os.MkdirTemp("", "acceptance-desktop-*")
	if err != nil {
		return nil, err
	}
	env := map[string]string{"STARTER_DATA_DIR": dir}
	app := &desktopApp{getenv: func(key string) string { return env[key] }}

	h := newHarness()
	h.cleanup = func() {
		app.close()
		os.RemoveAll(dir)
	}
	h.reopen = func() error {
		app.close()
		return app.open(h)
	}
	h.openCopy = func() error {
		second, err := desktop.Open(app.getenv, os.UserConfigDir, nil)
		if err != nil {
			return err
		}
		return second.Close()
	}
	if err := app.open(h); err != nil {
		h.cleanup()
		return nil, err
	}
	return h, nil
}

// desktopApp is one copy of the desktop app, opened on the data folder that
// getenv names and served over HTTP.
type desktopApp struct {
	getenv func(string) string
	app    *desktop.App
	srv    *httptest.Server
}

// open opens the app and points h's driving ports at it. The scenarios drive
// the API only, so the app is opened without the web app.
func (d *desktopApp) open(h *harness) error {
	app, err := desktop.Open(d.getenv, os.UserConfigDir, nil)
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
	// sample:start
	// The bridge ignores the host: only the path and query reach mobile.Serve.
	client, err := httpclient.New("http://mobile.invalid", apiclient.WithHTTPClient(bridge{app.API}))
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
