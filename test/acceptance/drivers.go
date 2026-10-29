// Package acceptance runs the Gherkin features in /features with Godog.
//
// Each scenario runs once per driver. A driver is a way of reaching the core
// through its driving port, so the same business scenarios prove the core,
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

// harness is everything a scenario needs: the service under test, the clock
// controlling it, and a cleanup function.
type harness struct {
	items   ports.ItemService
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

func newCore(repo ports.ItemRepository) (*app.ItemService, *testsupport.FixedClock) {
	clock := testsupport.NewFixedClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	return app.NewItemService(repo, clock, testsupport.NewSequentialIDs("item")), clock
}

func startMemory() (*harness, error) {
	svc, clock := newCore(memory.NewItemRepository())
	return &harness{items: svc, clock: clock, cleanup: func() {}}, nil
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
	svc, clock := newCore(jsonfile.NewItemRepository(store))
	return &harness{items: svc, clock: clock, cleanup: func() {
		store.Close()
		os.RemoveAll(dir)
	}}, nil
}

func startHTTP() (*harness, error) {
	svc, clock := newCore(memory.NewItemRepository())
	srv := httptest.NewServer(httpapi.NewHandler(httpapi.Options{Items: svc}))
	client, err := httpclient.New(srv.URL)
	if err != nil {
		srv.Close()
		return nil, err
	}
	return &harness{items: client, clock: clock, cleanup: srv.Close}, nil
}
