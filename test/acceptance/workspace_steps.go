package acceptance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/workspace"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// registerWorkspaceSteps binds the steps of features/workspaces.feature, which
// only the drivers wired with workspace.FromContext run.
func (w *world) registerWorkspaceSteps(sc *godog.ScenarioContext) {
	sc.Step(`^workspace "([^"]*)" has an item named "([^"]*)"$`, w.workspaceHasAnItemNamed)
	sc.Step(`^workspace "([^"]*)" has the items:$`, w.workspaceHasTheItems)
	sc.Step(`^someone in workspace "([^"]*)" tries to (view|rename|delete) the item "([^"]*)" from workspace "([^"]*)"$`,
		w.someoneTriesToReachAnotherWorkspacesItem)
	sc.Step(`^someone with no workspace tries to list the items$`, w.someoneWithNoWorkspaceTriesToList)
	sc.Step(`^someone with no workspace tries to create an item named "([^"]*)"$`, w.someoneWithNoWorkspaceTriesToCreate)
	sc.Step(`^they are told they have no workspace$`, w.theyAreToldTheyHaveNoWorkspace)
}

// in returns a context that works in the named workspace.
func in(name string) context.Context {
	return workspace.WithID(context.Background(), domain.WorkspaceID(name))
}

func (w *world) workspaceHasAnItemNamed(name, item string) error {
	return w.anItemNamed(in(name), item)
}

func (w *world) workspaceHasTheItems(name string, table *godog.Table) error {
	return w.theItemsAre(in(name), table)
}

func (w *world) someoneTriesToReachAnotherWorkspacesItem(requester, action, item, owner string) error {
	target, err := w.findByName(in(owner), item)
	if err != nil {
		return err
	}
	ctx := in(requester)
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

// The no-workspace steps use a fresh context, not the scenario's, which the
// Before hook put in the local workspace.

func (w *world) someoneWithNoWorkspaceTriesToList() error {
	_, w.lastErr = w.items.ListItems(context.Background())
	return nil
}

func (w *world) someoneWithNoWorkspaceTriesToCreate(name string) error {
	_, w.lastErr = w.items.CreateItem(context.Background(), name)
	return nil
}

func (w *world) theyAreToldTheyHaveNoWorkspace() error {
	if !errors.Is(w.lastErr, domain.ErrNoWorkspace) {
		return fmt.Errorf("expected no workspace, got %w", w.lastErr)
	}
	return nil
}
