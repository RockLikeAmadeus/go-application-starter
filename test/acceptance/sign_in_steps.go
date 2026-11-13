package acceptance

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

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
	sc.Step(`^"([^"]*)" changes their name at the sign-in provider to "([^"]*)" and signs in again$`, w.changesTheirNameAtTheProvider)
	sc.Step(`^another person named "([^"]*)" signs in$`, w.anotherPersonNamedSignsIn)
	sc.Step(`^I cancel signing in at the sign-in provider$`, w.iCancelSigningIn)
	sc.Step(`^someone comes back from the sign-in provider without having started to sign in on this server$`, w.someoneComesBackWithoutStarting)
	sc.Step(`^the sign-in provider can't start at its sign-up page$`, w.theProviderCantStartAtSignUp)
	sc.Step(`^I create an account at the sign-in provider as "([^"]*)"$`, w.iCreateAnAccountAtTheProviderAs)
	sc.Step(`^the sign-in provider showed me its sign-up page first$`, w.theProviderShowedMeItsSignUpPageFirst)
	sc.Step(`^the sign-in provider showed me its sign-in page first$`, w.theProviderShowedMeItsSignInPageFirst)
	sc.Step(`^(?:I am|they are) told that signing in didn't work$`, w.iAmToldSigningInDidntWork)
	sc.Step(`^(?:I am|they are) not signed in$`, w.iAmNotSignedIn)
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

// errNoProvider is returned by the sign-in provider steps on a driver
// without a fake sign-in provider.
var errNoProvider = errors.New("this driver has no sign-in provider: tag the scenario @sign-in-provider")

func (w *world) oidcSignIn() (*oidcSignInOverHTTP, error) {
	o, ok := w.signIn.(*oidcSignInOverHTTP)
	if !ok {
		return nil, errNoProvider
	}
	return o, nil
}

func (w *world) changesTheirNameAtTheProvider(ctx context.Context, name, newName string) (context.Context, error) {
	o, err := w.oidcSignIn()
	if err != nil {
		return ctx, err
	}
	person := personNamed(name)
	person.Name = newName
	return o.signInAs(ctx, name, person)
}

func (w *world) anotherPersonNamedSignsIn(ctx context.Context, name string) (context.Context, error) {
	o, err := w.oidcSignIn()
	if err != nil {
		return ctx, err
	}
	person := personNamed(name)
	person.Subject += "-another"
	return o.signInAs(ctx, "another "+name, person)
}

func (w *world) iCancelSigningIn(ctx context.Context) (context.Context, error) {
	o, err := w.oidcSignIn()
	if err != nil {
		return ctx, err
	}
	return o.cancel(ctx)
}

func (w *world) someoneComesBackWithoutStarting(ctx context.Context) (context.Context, error) {
	o, err := w.oidcSignIn()
	if err != nil {
		return ctx, err
	}
	return o.comeBackWithoutStarting(ctx)
}

func (w *world) theProviderCantStartAtSignUp() error {
	o, err := w.oidcSignIn()
	if err != nil {
		return err
	}
	return o.cantStartAtSignUp()
}

func (w *world) iCreateAnAccountAtTheProviderAs(ctx context.Context, name string) (context.Context, error) {
	o, err := w.oidcSignIn()
	if err != nil {
		return ctx, err
	}
	return o.createAccount(ctx, name)
}

func (w *world) theProviderShowedMeItsSignUpPageFirst() error {
	o, err := w.oidcSignIn()
	if err != nil {
		return err
	}
	return o.showedFirst("create")
}

func (w *world) theProviderShowedMeItsSignInPageFirst() error {
	o, err := w.oidcSignIn()
	if err != nil {
		return err
	}
	return o.showedFirst("")
}

// signInFailedHeading is the heading of the page a refused sign-in answers.
const signInFailedHeading = "<h1>Sign-in failed</h1>"

func (w *world) iAmToldSigningInDidntWork() error {
	o, err := w.oidcSignIn()
	if err != nil {
		return err
	}
	if o.lastAnswer.status != http.StatusBadRequest || !strings.Contains(o.lastAnswer.body, signInFailedHeading) {
		return fmt.Errorf("expected the sign-in failed page, got status %d: %.80q", o.lastAnswer.status, o.lastAnswer.body)
	}
	return nil
}

func (w *world) iAmNotSignedIn(ctx context.Context) error {
	if w.signIn == nil {
		return errNoSignIn
	}
	name, err := w.signIn.signedInName(ctx)
	if err != nil {
		return nil
	}
	if name != "" {
		return fmt.Errorf("signed in as %q, want no one", name)
	}
	return nil
}

func (w *world) anAdminHasMovedTheAccountsWorkspace(ctx context.Context, name, plan string) error {
	if w.accounts == nil {
		return errNoSignIn
	}
	identity, err := w.signIn.identity(name)
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
