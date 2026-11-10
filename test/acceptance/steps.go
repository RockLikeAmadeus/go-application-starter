package acceptance

import (
	"context"
	"errors"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/workspace"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// world holds the state of a single scenario.
type world struct {
	*harness
	lastErr error
	// cliOut is what the command line last printed, and cliErrOut what it
	// last printed apart from that (its warnings).
	cliOut    string
	cliErrOut string
	// settingsFileLeft is the text the user last wrote to their settings
	// file.
	settingsFileLeft string
	// commandLineOptions are given to every command the command line runs,
	// before the command's own arguments.
	commandLineOptions []string
	// sample:start
	// exportedCSV is the items' CSV that the last export gave.
	exportedCSV []byte
	// secondServer shares the server's Postgres database.
	secondServer *testServer
	// sample:end
}

func initializeScenario(d driver) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := &world{}

		sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
			h, err := d.start()
			w.harness = h
			// Scenarios that don't name a workspace work in the local edition's
			// one. Drivers wired with workspace.FromContext read it from here.
			return workspace.WithID(ctx, domain.LocalWorkspace), err
		})
		sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
			if w.harness != nil {
				w.cleanup()
			}
			return ctx, err
		})

		sc.Step(`^I am told the "([^"]*)" is invalid$`, w.iAmToldTheFieldIsInvalid)
		w.registerDesktopSteps(sc)
		w.registerLocalDataSteps(sc)
		w.registerSettingsFileSteps(sc)
		w.registerDefaultSettingsFileSteps(sc)
		w.registerMobileSteps(sc)
		w.registerEntitlementsSteps(sc)
		w.registerReadinessSteps(sc)
		w.registerSQLiteSteps(sc)
		w.registerPostgresSteps(sc)
		w.registerSyncedFolderSteps(sc)
		w.registerSyncedPlainTextSteps(sc)
		// sample:start
		w.registerItemSteps(sc)
		w.registerPlanSteps(sc)
		w.registerPlanFeatureSteps(sc)
		w.registerSharedDatabaseSteps(sc)
		w.registerWorkspaceSteps(sc)
		// sample:end
	}
}

func (w *world) iAmToldTheFieldIsInvalid(field string) error {
	var verr *domain.ValidationError
	if !errors.As(w.lastErr, &verr) {
		return fmt.Errorf("expected a validation error, got %w", w.lastErr)
	}
	if verr.Field != field {
		return fmt.Errorf("validation error on %q, want %q", verr.Field, field)
	}
	return nil
}

// in returns a context that works in the named workspace.
func in(name string) context.Context {
	return workspace.WithID(context.Background(), domain.WorkspaceID(name))
}
