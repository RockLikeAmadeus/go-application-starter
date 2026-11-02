package acceptance

import (
	"context"
	"errors"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/workspace"
)

// registerEntitlementsSteps binds the steps about which plan a workspace is
// on. "My workspace" is the one the scenario works in.
func (w *world) registerEntitlementsSteps(sc *godog.ScenarioContext) {
	sc.Step(`^my workspace is on the "([^"]*)" plan$`, w.myWorkspaceIsOnThePlan)
	sc.Step(`^my workspace moves to the "([^"]*)" plan$`, w.myWorkspaceIsOnThePlan)
	sc.Step(`^my workspace's plan is "([^"]*)"$`, w.myWorkspacesPlanIs)
}

// errFixedPlans is returned by the steps that set plans on a driver whose
// plans the tests don't set.
var errFixedPlans = errors.New("this driver's plans are fixed: tag the scenario @plans")

func (w *world) myWorkspaceIsOnThePlan(ctx context.Context, plan string) error {
	if w.plans == nil {
		return errFixedPlans
	}
	ws, err := workspace.FromContext{}.CurrentWorkspace(ctx)
	if err != nil {
		return err
	}
	w.plans.assign(ws, plan)
	return nil
}

func (w *world) myWorkspacesPlanIs(ctx context.Context, plan string) error {
	got, err := w.entitlements.CurrentEntitlements(ctx)
	if err != nil {
		return err
	}
	if got.Plan != plan {
		return fmt.Errorf("plan = %q, want %q", got.Plan, plan)
	}
	return nil
}
