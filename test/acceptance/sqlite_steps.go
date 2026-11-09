package acceptance

import (
	"errors"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// sqliteStorage is the storage setting that keeps data in SQLite.
const sqliteStorage = config.StorageBackend("sqlite")

// registerSQLiteSteps binds the steps of features/sqlite-storage.feature that
// choose SQLite storage: on the desktop app's computer (desktop driver) or
// for the server (server driver).
func (w *world) registerSQLiteSteps(sc *godog.ScenarioContext) {
	sc.Step(`^the desktop app keeps its data in SQLite$`, w.theDesktopAppKeepsItsDataInSQLite)
	sc.Step(`^the server keeps its data in SQLite$`, w.theServerKeepsItsDataInSQLite)
}

// theDesktopAppKeepsItsDataInSQLite chooses SQLite in the settings file and
// opens the desktop app again, so it reads it.
func (w *world) theDesktopAppKeepsItsDataInSQLite() error {
	if w.local == nil || w.reopen == nil {
		return errNoLocalComputer
	}
	if err := w.local.setSetting("storage", string(sqliteStorage)); err != nil {
		return err
	}
	return w.reopen()
}

var errNoServer = errors.New("only the server driver runs the server: tag the scenario @server")

func (w *world) theServerKeepsItsDataInSQLite() error {
	if w.server == nil {
		return errNoServer
	}
	return w.server.restart(sqliteStorage)
}
