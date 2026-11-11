package acceptance

import (
	"context"
	"errors"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// registerSignInSteps binds the steps of features/sign-in.feature, which
// only the drivers running the hosted edition run.
func (w *world) registerSignInSteps(sc *godog.ScenarioContext) {
	// "I am signed in as" sets up a session in a Given, and checks it in a
	// Then, so the step needs to know which it is.
	sc.StepContext().Before(func(ctx context.Context, step *godog.Step) (context.Context, error) {
		w.isSettingUp = step.Type == "Context"
		return ctx, nil
	})
	sc.Step(`^I sign in as "([^"]*)"$`, w.iSignInAs)
	sc.Step(`^"([^"]*)" has signed in before$`, w.iSignInAs)
	sc.Step(`^I am signed in as "([^"]*)"$`, w.iAmSignedInAs)
	sc.Step(`^an admin has moved "([^"]*)"'s workspace to the "([^"]*)" plan$`, w.anAdminHasMovedTheAccountsWorkspace)
	sc.Step(`^someone who hasn't signed in tries to see their workspace's plan$`, w.someoneWhoHasntSignedInTriesToSeeTheirPlan)
	sc.Step(`^I sign out$`, w.iSignOut)
	sc.Step(`^I try to see my workspace's plan$`, w.iTryToSeeMyWorkspacesPlan)
	sc.Step(`^(?:I am|they are) told to sign in$`, w.iAmToldToSignIn)
	// sample:start
	sc.Step(`^"([^"]*)" has signed in before and added an item named "([^"]*)"$`, w.hasSignedInBeforeAndAddedAnItem)
	// sample:end
}

// errNoSignIn is returned by the sign-in steps on a driver running the
// local edition.
var errNoSignIn = errors.New("this driver has no sign-in: tag the scenario @sign-in")

func (w *world) iSignInAs(ctx context.Context, name string) (context.Context, error) {
	if w.signIn == nil {
		return ctx, errNoSignIn
	}
	return w.signIn.signIn(ctx, name)
}

func (w *world) iAmSignedInAs(ctx context.Context, name string) (context.Context, error) {
	if w.isSettingUp {
		return w.iSignInAs(ctx, name)
	}
	if w.signIn == nil {
		return ctx, errNoSignIn
	}
	got, err := w.signIn.signedInName(ctx)
	if err != nil {
		return ctx, err
	}
	if got != name {
		return ctx, fmt.Errorf("signed in as %q, want %q", got, name)
	}
	return ctx, nil
}

func (w *world) anAdminHasMovedTheAccountsWorkspace(ctx context.Context, name, plan string) error {
	if w.accounts == nil {
		return errNoSignIn
	}
	identity, err := devIdentity(name)
	if err != nil {
		return err
	}
	account, err := w.accounts.FindByIdentity(ctx, identity)
	if err != nil {
		return err
	}
	return w.moveWorkspaceToThePlan(string(account.WorkspaceID), plan)
}

func (w *world) someoneWhoHasntSignedInTriesToSeeTheirPlan(ctx context.Context) (context.Context, error) {
	if w.signIn == nil {
		return ctx, errNoSignIn
	}
	ctx, err := w.signIn.signedOut(ctx)
	if err != nil {
		return ctx, err
	}
	return ctx, w.iTryToSeeMyWorkspacesPlan(ctx)
}

func (w *world) iSignOut(ctx context.Context) (context.Context, error) {
	if w.signIn == nil {
		return ctx, errNoSignIn
	}
	return w.signIn.signOut(ctx)
}

func (w *world) iTryToSeeMyWorkspacesPlan(ctx context.Context) error {
	_, w.lastErr = w.entitlements.CurrentEntitlements(ctx)
	return nil
}

func (w *world) iAmToldToSignIn() error {
	if !errors.Is(w.lastErr, domain.ErrNoWorkspace) {
		return fmt.Errorf("expected to be told to sign in, got %w", w.lastErr)
	}
	return nil
}

// sample:start

func (w *world) hasSignedInBeforeAndAddedAnItem(ctx context.Context, name, item string) (context.Context, error) {
	ctx, err := w.iSignInAs(ctx, name)
	if err != nil {
		return ctx, err
	}
	return ctx, w.anItemNamed(ctx, item)
}

// sample:end
