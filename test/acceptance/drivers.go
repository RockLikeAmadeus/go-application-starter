// Package acceptance runs the Gherkin features in /features with Godog.
//
// Each scenario runs once per driver. A driver is a way of reaching the core
// through its driving ports, so the same business scenarios prove the core,
// each storage adapter, and the HTTP server+client adapters.
package acceptance

import (
	"net/http/httptest"
	"os"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/desktop"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient"
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
	// reopen closes the desktop app and opens it again on the same data
	// folder. Only the desktop driver sets it.
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

// Desktop-only scenarios (@desktop) run only on the desktop driver, which
// also runs every other feature.
var drivers = []driver{
	{name: "memory", start: startMemory, tags: "~@desktop"},
	{name: "jsonfile", start: startJSONFile, tags: "~@desktop"},
	{name: "http", start: startHTTP, tags: "~@desktop"},
	{name: "desktop", start: startDesktop},
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
		second, err := desktop.Open(app.getenv, nil)
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
	app, err := desktop.Open(d.getenv, nil)
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
