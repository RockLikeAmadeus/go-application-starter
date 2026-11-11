package acceptance

import (
	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/postgres/postgrestest"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// registerPostgresSteps binds the steps of features/postgres-storage.feature,
// which only the server driver runs.
func (w *world) registerPostgresSteps(sc *godog.ScenarioContext) {
	sc.Step(`^the server keeps its data in Postgres$`, w.theServerKeepsItsDataInPostgres)
	sc.Step(`^the server's database can't be reached$`, w.theServersDatabaseCannotBeReached)
	sc.Step(`^the server's database can be reached again$`, w.theServersDatabaseCanBeReachedAgain)
}

// theServerKeepsItsDataInPostgres restarts the server on a database of its
// own, dropped when the scenario ends.
func (w *world) theServerKeepsItsDataInPostgres() error {
	if w.server == nil {
		return errNoServer
	}
	w.server.databaseURL = postgrestest.NewDatabase(w.server.scenario)
	return w.server.restart(config.StoragePostgres)
}

func (w *world) theServersDatabaseCannotBeReached() error {
	if w.server == nil {
		return errNoServer
	}
	postgrestest.BlockConnections(w.server.scenario, w.server.databaseURL)
	return nil
}

func (w *world) theServersDatabaseCanBeReachedAgain() error {
	if w.server == nil {
		return errNoServer
	}
	postgrestest.AllowConnections(w.server.scenario, w.server.databaseURL)
	return nil
}
