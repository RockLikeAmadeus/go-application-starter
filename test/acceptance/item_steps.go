package acceptance

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// registerItemSteps binds the steps of features/items.feature, and the item
// steps of features/local-data.feature.
func (w *world) registerItemSteps(sc *godog.ScenarioContext) {
	sc.Step(`^an item named "([^"]*)"$`, w.anItemNamed)
	sc.Step(`^I create an item named "([^"]*)"$`, w.anItemNamed)
	sc.Step(`^I try to create an item named "([^"]*)"$`, w.iTryToCreateAnItemNamed)
	sc.Step(`^I rename "([^"]*)" to "([^"]*)"$`, w.iRename)
	sc.Step(`^I try to rename "([^"]*)" to "([^"]*)"$`, w.iTryToRename)
	sc.Step(`^I delete "([^"]*)"$`, w.iDelete)
	sc.Step(`^I try to delete an item that does not exist$`, w.iTryToDeleteAMissingItem)
	sc.Step(`^the items are:$`, w.theItemsAre)
	sc.Step(`^there are no items$`, w.thereAreNoItems)
	sc.Step(`^I am told the item was not found$`, w.iAmToldNotFound)
	sc.Step(`^I list the items from the command line$`, w.iListTheItemsFromTheCommandLine)
	sc.Step(`^the command line shows "([^"]*)"$`, w.theCommandLineShows)
}

func (w *world) iListTheItemsFromTheCommandLine(ctx context.Context) error {
	return w.runCommandLine(ctx, "item", "list")
}

func (w *world) theCommandLineShows(name string) error {
	return w.commandLineSaid(name)
}

func (w *world) anItemNamed(ctx context.Context, name string) error {
	// Advance the clock so creation order is unambiguous.
	w.clock.Advance(time.Second)
	_, err := w.items.CreateItem(ctx, name)
	return err
}

func (w *world) iTryToCreateAnItemNamed(ctx context.Context, name string) error {
	_, w.lastErr = w.items.CreateItem(ctx, name)
	return nil
}

func (w *world) iRename(ctx context.Context, from, to string) error {
	item, err := w.findByName(ctx, from)
	if err != nil {
		return err
	}
	w.clock.Advance(time.Second)
	_, err = w.items.RenameItem(ctx, item.ID, to)
	return err
}

func (w *world) iTryToRename(ctx context.Context, from, to string) error {
	item, err := w.findByName(ctx, from)
	if err != nil {
		return err
	}
	w.clock.Advance(time.Second)
	_, w.lastErr = w.items.RenameItem(ctx, item.ID, to)
	return nil
}

func (w *world) iDelete(ctx context.Context, name string) error {
	item, err := w.findByName(ctx, name)
	if err != nil {
		return err
	}
	return w.items.DeleteItem(ctx, item.ID)
}

func (w *world) iTryToDeleteAMissingItem(ctx context.Context) error {
	w.lastErr = w.items.DeleteItem(ctx, "does-not-exist")
	return nil
}

func (w *world) theItemsAre(ctx context.Context, table *godog.Table) error {
	var want []string
	for _, row := range table.Rows {
		want = append(want, row.Cells[0].Value)
	}
	got, err := w.names(ctx)
	if err != nil {
		return err
	}
	if !slices.Equal(got, want) {
		return fmt.Errorf("items = %q, want %q", got, want)
	}
	return nil
}

func (w *world) thereAreNoItems(ctx context.Context) error {
	got, err := w.names(ctx)
	if err != nil {
		return err
	}
	if len(got) != 0 {
		return fmt.Errorf("expected no items, got %q", got)
	}
	return nil
}

func (w *world) iAmToldNotFound() error {
	if !errors.Is(w.lastErr, domain.ErrNotFound) {
		return fmt.Errorf("expected not found, got %w", w.lastErr)
	}
	return nil
}

func (w *world) names(ctx context.Context) ([]string, error) {
	items, err := w.items.ListItems(ctx)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(items))
	for _, it := range items {
		names = append(names, it.Name)
	}
	return names, nil
}

func (w *world) findByName(ctx context.Context, name string) (domain.Item, error) {
	items, err := w.items.ListItems(ctx)
	if err != nil {
		return domain.Item{}, err
	}
	for _, it := range items {
		if it.Name == name {
			return it, nil
		}
	}
	return domain.Item{}, fmt.Errorf("no item named %q", name)
}
