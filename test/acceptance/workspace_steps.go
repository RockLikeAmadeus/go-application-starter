package acceptance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// registerWorkspaceSteps binds the steps of features/workspaces.feature, which
// only the drivers that can switch between workspaces run: those wired with
// workspace.FromContext, and http, which signs in to each workspace's account.
func (w *world) registerWorkspaceSteps(sc *godog.ScenarioContext) {
	sc.Step(`^workspace "([^"]*)" has an item named "([^"]*)"$`, w.workspaceHasAnItemNamed)
	sc.Step(`^workspace "([^"]*)" has the items:$`, w.workspaceHasTheItems)
	sc.Step(`^someone in workspace "([^"]*)" tries to (view|rename|delete) the item "([^"]*)" from workspace "([^"]*)"$`,
		w.someoneTriesToReachAnotherWorkspacesItem)
	sc.Step(`^someone with no workspace tries to list the items$`, w.someoneWithNoWorkspaceTriesToList)
	sc.Step(`^someone with no workspace tries to create an item named "([^"]*)"$`, w.someoneWithNoWorkspaceTriesToCreate)
	sc.Step(`^they are told they have no workspace$`, w.theyAreToldTheyHaveNoWorkspace)
}

func (w *world) workspaceHasAnItemNamed(name, item string) error {
	ctx, err := w.inWorkspace(name)
	if err != nil {
		return err
	}
	return w.anItemNamed(ctx, item)
}

func (w *world) workspaceHasTheItems(name string, table *godog.Table) error {
	ctx, err := w.inWorkspace(name)
	if err != nil {
		return err
	}
	return w.theItemsAre(ctx, table)
}

func (w *world) someoneTriesToReachAnotherWorkspacesItem(requester, action, item, owner string) error {
	ownerCtx, err := w.inWorkspace(owner)
	if err != nil {
		return err
	}
	target, err := w.findByName(ownerCtx, item)
	if err != nil {
		return err
	}
	ctx, err := w.inWorkspace(requester)
	if err != nil {
		return err
	}
	w.clock.Advance(time.Second)
	switch action {
	case "view":
		_, w.lastErr = w.items.GetItem(ctx, target.ID)
	case "rename":
		_, w.lastErr = w.items.RenameItem(ctx, target.ID, "Renamed")
	case "delete":
		w.lastErr = w.items.DeleteItem(ctx, target.ID)
	}
	return nil
}

func (w *world) someoneWithNoWorkspaceTriesToList() error {
	ctx, err := w.inNoWorkspace()
	if err != nil {
		return err
	}
	_, w.lastErr = w.items.ListItems(ctx)
	return nil
}

func (w *world) someoneWithNoWorkspaceTriesToCreate(name string) error {
	ctx, err := w.inNoWorkspace()
	if err != nil {
		return err
	}
	_, w.lastErr = w.items.CreateItem(ctx, name)
	return nil
}

func (w *world) theyAreToldTheyHaveNoWorkspace() error {
	if !errors.Is(w.lastErr, domain.ErrNoWorkspace) {
		return fmt.Errorf("expected no workspace, got %w", w.lastErr)
	}
	return nil
}

// inNoWorkspace returns a context that works in no workspace: on a driver
// running the hosted edition, as someone who hasn't signed in.
func (w *world) inNoWorkspace() (context.Context, error) {
	// A fresh context, not the scenario's, which the Before hook may have put
	// in the local workspace.
	if w.signIn == nil {
		return context.Background(), nil
	}
	return w.signIn.signedOut(context.Background())
}
