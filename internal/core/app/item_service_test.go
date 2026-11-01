package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/memory"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/app"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
	"github.com/RockLikeAmadeus/go-application-starter/internal/testsupport"
)

// workspace is the workspace every newService request works in.
const workspace domain.WorkspaceID = "ws-a"

func newService() (*app.ItemService, *testsupport.FixedClock) {
	return newServiceWith(memory.NewItemRepository(), testsupport.FixedWorkspace(workspace))
}

// newServiceWith builds the service on the given repository and resolver, with
// a fixed clock and sequential IDs.
func newServiceWith(repo ports.ItemRepository, workspaces ports.WorkspaceResolver) (*app.ItemService, *testsupport.FixedClock) {
	clock := testsupport.NewFixedClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	return app.NewItemService(repo, workspaces, clock, testsupport.NewSequentialIDs("item")), clock
}

func TestCreateItem_PutsTheItemInTheCurrentWorkspace(t *testing.T) {
	svc, _ := newService()

	created, err := svc.CreateItem(context.Background(), "Widget")
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}
	if created.WorkspaceID != workspace {
		t.Errorf("WorkspaceID = %q, want %q", created.WorkspaceID, workspace)
	}
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

// untouchedRepo is an ItemRepository that counts every call, for tests that
// must not reach storage.
type untouchedRepo struct{ calls int }

func (r *untouchedRepo) Save(context.Context, domain.Item) error {
	r.calls++
	return nil
}

func (r *untouchedRepo) FindByID(context.Context, domain.WorkspaceID, string) (domain.Item, error) {
	r.calls++
	return domain.Item{}, nil
}

func (r *untouchedRepo) FindAll(context.Context, domain.WorkspaceID) ([]domain.Item, error) {
	r.calls++
	return nil, nil
}

func (r *untouchedRepo) Delete(context.Context, domain.WorkspaceID, string) error {
	r.calls++
	return nil
}

func TestUseCases_WithoutAWorkspaceAreRefusedBeforeStorage(t *testing.T) {
	cases := map[string]func(context.Context, *app.ItemService) error{
		"CreateItem": func(ctx context.Context, s *app.ItemService) error {
			_, err := s.CreateItem(ctx, "Widget")
			return err
		},
		"GetItem": func(ctx context.Context, s *app.ItemService) error {
			_, err := s.GetItem(ctx, "item-1")
			return err
		},
		"ListItems": func(ctx context.Context, s *app.ItemService) error {
			_, err := s.ListItems(ctx)
			return err
		},
		"RenameItem": func(ctx context.Context, s *app.ItemService) error {
			_, err := s.RenameItem(ctx, "item-1", "New")
			return err
		},
		"DeleteItem": func(ctx context.Context, s *app.ItemService) error {
			return s.DeleteItem(ctx, "item-1")
		},
	}
	for name, call := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &untouchedRepo{}
			svc, _ := newServiceWith(repo, testsupport.NoWorkspace{})

			if err := call(context.Background(), svc); !errors.Is(err, domain.ErrNoWorkspace) {
				t.Errorf("err = %v, want ErrNoWorkspace", err)
			}
			if repo.calls != 0 {
				t.Errorf("repository called %d times, want 0", repo.calls)
			}
		})
	}
}
