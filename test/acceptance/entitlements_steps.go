package acceptance

import (
	"context"
	"errors"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/workspace"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// registerEntitlementsSteps binds the steps about plans and which one a
// workspace is on. "My workspace" is the one the scenario works in; a named
// workspace needs @workspaces, because only those drivers can switch to it.
func (w *world) registerEntitlementsSteps(sc *godog.ScenarioContext) {
	sc.Step(`^the "([^"]*)" plan has no limits$`, w.thePlanHasNoLimits)
	sc.Step(`^my workspace is on the "([^"]*)" plan$`, w.moveMyWorkspaceToThePlan)
	sc.Step(`^my workspace moves to the "([^"]*)" plan$`, w.moveMyWorkspaceToThePlan)
	sc.Step(`^an admin moves my workspace to the "([^"]*)" plan$`, w.moveMyWorkspaceToThePlan)
	sc.Step(`^an admin tries to move my workspace to the "([^"]*)" plan$`, w.anAdminTriesToMoveMyWorkspaceToThePlan)
	sc.Step(`^I try to move my workspace to the "([^"]*)" plan from the command line$`, w.iTryToMoveMyWorkspaceFromTheCommandLine)
	sc.Step(`^I am told the local edition's plan is set when the app is built$`, w.iAmToldTheLocalPlanIsFixed)
	sc.Step(`^my workspace's plan is "([^"]*)"$`, w.myWorkspacesPlanIs)
	sc.Step(`^workspace "([^"]*)" is on the "([^"]*)" plan$`, w.moveWorkspaceToThePlan)
	sc.Step(`^an admin moves workspace "([^"]*)" to the "([^"]*)" plan$`, w.moveWorkspaceToThePlan)
	sc.Step(`^workspace "([^"]*)"'s plan is "([^"]*)"$`, w.workspacesPlanIs)
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
	ws, err := workspace.FromContext{}.CurrentWorkspace(ctx)
	if err != nil {
		return err
	}
	return w.moveWorkspaceToThePlan(string(ws), plan)
}

func (w *world) anAdminTriesToMoveMyWorkspaceToThePlan(ctx context.Context, plan string) error {
	w.lastErr = w.moveMyWorkspaceToThePlan(ctx, plan)
	return nil
}

func (w *world) iTryToMoveMyWorkspaceFromTheCommandLine(ctx context.Context, plan string) error {
	ws, err := workspace.FromContext{}.CurrentWorkspace(ctx)
	if err != nil {
		return err
	}
	return w.runCommandLine(ctx, "admin", "set-plan", string(ws), plan)
}

func (w *world) iAmToldTheLocalPlanIsFixed() error {
	if !errors.Is(w.lastErr, bootstrap.ErrLocalPlanIsFixed) {
		return fmt.Errorf("expected %q, got %w", bootstrap.ErrLocalPlanIsFixed.Error(), w.lastErr)
	}
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

func (w *world) moveWorkspaceToThePlan(name, plan string) error {
	if w.plans == nil {
		return errFixedPlans
	}
	return w.billing.ChangePlan(context.Background(), domain.WorkspaceID(name), plan)
}

func (w *world) workspacesPlanIs(name, plan string) error {
	return w.myWorkspacesPlanIs(in(name), plan)
}
