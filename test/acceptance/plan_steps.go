package acceptance

import (
	"context"
	"errors"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// registerPlanSteps binds the item limit's steps in features/plans.feature,
// which only the drivers whose plans the tests set run.
func (w *world) registerPlanSteps(sc *godog.ScenarioContext) {
	sc.Step(`^the "([^"]*)" plan allows (\d+) items, and (blocks new ones|makes them read only) when over$`, w.thePlanAllowsItems)
	sc.Step(`^its item limit is (\d+), which (blocks new ones|makes them read only) when over$`, w.itsItemLimitIs)
	sc.Step(`^I am told the limit of (\d+) items on the "([^"]*)" plan is reached$`, w.iAmToldTheLimitIsReached)
	sc.Step(`^I am told items are read only while over the limit of (\d+) on the "([^"]*)" plan$`, w.iAmToldItemsAreReadOnly)
}

// whenOver is what the scenarios call each over-limit behaviour.
var whenOver = map[string]domain.OverLimitBehavior{
	"blocks new ones":      domain.BlockNew,
	"makes them read only": domain.ReadOnly,
}

func (w *world) thePlanAllowsItems(plan string, max int, behaviour string) error {
	if w.plans == nil {
		return errFixedPlans
	}
	limit, err := domain.NewLimit("items", max, whenOver[behaviour])
	if err != nil {
		return err
	}
	w.plans.define(domain.Entitlements{Plan: plan, Limits: map[string]domain.Limit{"items": limit}})
	return nil
}

func (w *world) itsItemLimitIs(ctx context.Context, max int, behaviour string) error {
	got, err := w.entitlements.CurrentEntitlements(ctx)
	if err != nil {
		return err
	}
	want := domain.Limit{Name: "items", Max: max, WhenOver: whenOver[behaviour]}
	if got.Limits["items"] != want {
		return fmt.Errorf("item limit = %+v, want %+v", got.Limits["items"], want)
	}
	return nil
}

func (w *world) iAmToldTheLimitIsReached(max int, plan string) error {
	return w.expectLimitError(domain.LimitError{Limit: "items", Max: max, Plan: plan})
}

func (w *world) iAmToldItemsAreReadOnly(max int, plan string) error {
	return w.expectLimitError(domain.LimitError{Limit: "items", Max: max, Plan: plan, IsReadOnly: true})
}

func (w *world) expectLimitError(want domain.LimitError) error {
	var got *domain.LimitError
	if !errors.As(w.lastErr, &got) {
		return fmt.Errorf("expected %q, got %w", want.Error(), w.lastErr)
	}
	if *got != want {
		return fmt.Errorf("limit error = %+v, want %+v", *got, want)
	}
	return nil
}
