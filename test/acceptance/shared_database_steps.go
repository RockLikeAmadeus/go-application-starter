package acceptance

import (
	"context"
	"errors"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// registerSharedDatabaseSteps binds the steps of the sample scenario in
// features/postgres-storage.feature, which only the server driver runs.
func (w *world) registerSharedDatabaseSteps(sc *godog.ScenarioContext) {
	sc.Step(`^two servers keep their data in the same Postgres database$`, w.twoServersKeepTheirDataInTheSamePostgresDatabase)
	sc.Step(`^an item named "([^"]*)" is added through the first server$`, w.anItemIsAddedThroughTheFirstServer)
	sc.Step(`^the second server lists "([^"]*)"$`, w.theSecondServerLists)
}

// twoServersKeepTheirDataInTheSamePostgresDatabase moves the server to a
// database of its own and starts a second server on the same database. Both
// stop when the scenario ends.
func (w *world) twoServersKeepTheirDataInTheSamePostgresDatabase() error {
	if err := w.theServerKeepsItsDataInPostgres(); err != nil {
		return err
	}
	second := &testServer{databaseURL: w.server.databaseURL, scenario: w.server.scenario}
	if err := second.start(config.StoragePostgres); err != nil {
		return err
	}
	w.server.scenario.Cleanup(second.stop)
	w.secondServer = second
	return nil
}

func (w *world) anItemIsAddedThroughTheFirstServer(ctx context.Context, name string) error {
	items, err := httpclient.NewItemService(w.server.url)
	if err != nil {
		return err
	}
	_, err = items.CreateItem(ctx, name)
	return err
}

func (w *world) theSecondServerLists(ctx context.Context, name string) error {
	if w.secondServer == nil {
		return errors.New("there is no second server")
	}
	items, err := httpclient.NewItemService(w.secondServer.url)
	if err != nil {
		return err
	}
	listed, err := items.ListItems(ctx)
	if err != nil {
		return err
	}
	for _, item := range listed {
		if item.Name == name {
			return nil
		}
	}
	return fmt.Errorf("the second server lists %d items, none named %q", len(listed), name)
}
