package acceptance

import (
	"context"
	"errors"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/workspace"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// registerEntitlementsSteps binds the steps about plans and which one a
// workspace is on. "My workspace" is the one the scenario works in.
func (w *world) registerEntitlementsSteps(sc *godog.ScenarioContext) {
	sc.Step(`^the "([^"]*)" plan has no limits$`, w.thePlanHasNoLimits)
	sc.Step(`^my workspace is on the "([^"]*)" plan$`, w.moveMyWorkspaceToThePlan)
	sc.Step(`^my workspace moves to the "([^"]*)" plan$`, w.moveMyWorkspaceToThePlan)
	sc.Step(`^an admin moves my workspace to the "([^"]*)" plan$`, w.moveMyWorkspaceToThePlan)
	sc.Step(`^my workspace's plan is "([^"]*)"$`, w.myWorkspacesPlanIs)
}

// errFixedPlans is returned by the steps that set plans on a driver whose
// plans the tests don't set.
var errFixedPlans = errors.New("this driver's plans are fixed: tag the scenario @plans")

func (w *world) thePlanHasNoLimits(plan string) error {
	if w.plans == nil {
		return errFixedPlans
	}
	w.plans.define(domain.Entitlements{Plan: plan, Limits: map[string]domain.Limit{}})
	return nil
}

func (w *world) moveMyWorkspaceToThePlan(ctx context.Context, plan string) error {
	if w.plans == nil {
		return errFixedPlans
	}
	ws, err := workspace.FromContext{}.CurrentWorkspace(ctx)
	if err != nil {
		return err
	}
	return w.billing.ChangePlan(ctx, ws, plan)
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
