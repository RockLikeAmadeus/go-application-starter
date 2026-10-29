package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/memory"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/app"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/testsupport"
)

func newService() (*app.ItemService, *testsupport.FixedClock) {
	clock := testsupport.NewFixedClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	return app.NewItemService(memory.NewItemRepository(), clock, testsupport.NewSequentialIDs("item")), clock
}

func TestCreateItem_PersistsAndReturnsItem(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()

	created, err := svc.CreateItem(ctx, "Widget")
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}
	if created.ID != "item-1" || created.Name != "Widget" {
		t.Errorf("unexpected item: %+v", created)
	}

	got, err := svc.GetItem(ctx, created.ID)
	if err != nil || got != created {
		t.Errorf("GetItem = %+v, %v; want %+v", got, err, created)
	}
}

func TestCreateItem_ValidationErrorIsNotPersisted(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()

	if _, err := svc.CreateItem(ctx, ""); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
	items, _ := svc.ListItems(ctx)
	if len(items) != 0 {
		t.Errorf("expected no items, got %d", len(items))
	}
}

func TestRenameItem_UpdatesTimestamp(t *testing.T) {
	svc, clock := newService()
	ctx := context.Background()
	created, _ := svc.CreateItem(ctx, "Old")
	clock.Advance(time.Minute)

	renamed, err := svc.RenameItem(ctx, created.ID, "New")
	if err != nil {
		t.Fatalf("RenameItem: %v", err)
	}
	if renamed.Name != "New" || !renamed.UpdatedAt.After(created.UpdatedAt) {
		t.Errorf("unexpected item after rename: %+v", renamed)
	}
}

func TestRenameItem_NotFound(t *testing.T) {
	svc, _ := newService()
	if _, err := svc.RenameItem(context.Background(), "missing", "x"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestDeleteItem(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()
	created, _ := svc.CreateItem(ctx, "Doomed")

	if err := svc.DeleteItem(ctx, created.ID); err != nil {
		t.Fatalf("DeleteItem: %v", err)
	}
	if _, err := svc.GetItem(ctx, created.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetItem after delete err = %v, want ErrNotFound", err)
	}
}
