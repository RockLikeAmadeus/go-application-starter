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
	// commandLine is the command line the scenario signs in from, made by
	// the first step that uses it.
	commandLine *commandLine
	// codeReplay is the answer to the latest second use of a sign-in code.
	codeReplay answer
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
			if isSignInScenario(scenario) || (d.needsSignInForWorkspaces && hasTag(scenario, "@workspaces")) {
				// Scenarios that sign in run on the hosted edition, and start
				// signed out, in no workspace. So do those that switch between
				// workspaces on a driver that can switch only by signing in.
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
		w.registerCommandLineSignInSteps(sc)
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

// isSignInScenario is whether the scenario signs in (@sign-in,
// @command-line-sign-in, which signs in from the command line, or
// @remote-mode, which signs in from the desktop and mobile apps).
func isSignInScenario(scenario *godog.Scenario) bool {
	return hasTag(scenario, "@sign-in") || hasTag(scenario, "@sign-in-provider") ||
		hasTag(scenario, "@command-line-sign-in") || hasTag(scenario, "@remote-mode")
}

// hasTag is whether the scenario, or its feature, has the tag.
func hasTag(scenario *godog.Scenario, name string) bool {
	for _, tag := range scenario.Tags {
		if tag.Name == name {
			return true
		}
	}
	return false
}

// inWorkspace returns a context that works in the named workspace. On a
// driver running the hosted edition, that is the personal workspace of the
// account of that name, so it signs in as it.
func (w *world) inWorkspace(name string) (context.Context, error) {
	if w.signIn == nil {
		return workspace.WithID(context.Background(), domain.WorkspaceID(name)), nil
	}
	return w.signIn.signIn(context.Background(), name)
}

// workspaceNamed is the ID of the named workspace. On a driver running the
// hosted edition, it signs in as the account of that name, making the
// account if there isn't one, and answers its personal workspace.
func (w *world) workspaceNamed(name string) (domain.WorkspaceID, error) {
	if w.signIn == nil {
		return domain.WorkspaceID(name), nil
	}
	if _, err := w.signIn.signIn(context.Background(), name); err != nil {
		return "", err
	}
	identity, err := w.signIn.identity(name)
	if err != nil {
		return "", err
	}
	account, err := w.accounts.FindByIdentity(context.Background(), identity)
	if err != nil {
		return "", err
	}
	return account.WorkspaceID, nil
}
