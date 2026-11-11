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
	// isSettingUp is whether the step running is a Given.
	isSettingUp bool
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

		sc.Before(func(ctx context.Context, scenario *godog.Scenario) (context.Context, error) {
			if isSignInScenario(scenario) {
				// Scenarios that sign in run on the hosted edition, and start
				// signed out, in no workspace.
				if d.startHosted == nil {
					return ctx, fmt.Errorf("the %s driver has no hosted edition: leave out @sign-in in its tags", d.name)
				}
				h, err := d.startHosted()
				w.harness = h
				return ctx, err
			}
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
		w.registerSignInSteps(sc)
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

// isSignInScenario is whether the scenario signs in (@sign-in).
func isSignInScenario(scenario *godog.Scenario) bool {
	for _, tag := range scenario.Tags {
		if tag.Name == "@sign-in" {
			return true
		}
	}
	return false
}

// in returns a context that works in the named workspace.
func in(name string) context.Context {
	return workspace.WithID(context.Background(), domain.WorkspaceID(name))
}
