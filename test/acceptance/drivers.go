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
}

type driver struct {
	name  string
	start func() (*harness, error)
}

var drivers = []driver{
	{name: "memory", start: startMemory},
	{name: "jsonfile", start: startJSONFile},
	{name: "http", start: startHTTP},
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
	// sample:start
	client, err := httpclient.New(srv.URL)
	if err != nil {
		srv.Close()
		return nil, err
	}
	h.items = client
	// sample:end
	return h, nil
}
