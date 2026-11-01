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
	return newServiceWith(memory.NewItemRepository(), testsupport.FixedWorkspace(workspace), testsupport.FixedEntitlements{})
}

// newServiceWith builds the service on the given repository, resolver and
// entitlements, with a fixed clock and sequential IDs.
func newServiceWith(
	repo ports.ItemRepository, workspaces ports.WorkspaceResolver, entitlements ports.Entitlements,
) (*app.ItemService, *testsupport.FixedClock) {
	clock := testsupport.NewFixedClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	return app.NewItemService(repo, workspaces, entitlements, clock, testsupport.NewSequentialIDs("item")), clock
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

// freePlan is a plan allowing 3 items, with whenOver past the limit.
func freePlan(whenOver domain.OverLimit) testsupport.FixedEntitlements {
	return testsupport.FixedEntitlements{
		Plan:   "free",
		Limits: map[string]domain.Limit{"items": {Name: "items", Max: 3, WhenOver: whenOver}},
	}
}

// serviceWithItems returns a service on entitlements whose workspace already
// has the named items, made on a plan with no limits, and those items.
func serviceWithItems(t *testing.T, entitlements ports.Entitlements, names ...string) (*app.ItemService, []domain.Item) {
	t.Helper()
	repo := memory.NewItemRepository()
	unlimited, _ := newServiceWith(repo, testsupport.FixedWorkspace(workspace), testsupport.FixedEntitlements{})
	var items []domain.Item
	for _, name := range names {
		item, err := unlimited.CreateItem(context.Background(), name)
		if err != nil {
			t.Fatalf("CreateItem(%q): %v", name, err)
		}
		items = append(items, item)
	}
	svc, _ := newServiceWith(repo, testsupport.FixedWorkspace(workspace), entitlements)
	return svc, items
}

// overTheFreePlan returns a service on freePlan(whenOver) whose workspace
// already has four items, and those items.
func overTheFreePlan(t *testing.T, whenOver domain.OverLimit) (*app.ItemService, []domain.Item) {
	t.Helper()
	return serviceWithItems(t, freePlan(whenOver), "First", "Second", "Third", "Fourth")
}

func TestCreateItem_AtTheLimitIsRefusedAndSavesNothing(t *testing.T) {
	svc, _ := newServiceWith(memory.NewItemRepository(), testsupport.FixedWorkspace(workspace), freePlan(domain.BlockNew))
	ctx := context.Background()
	for _, name := range []string{"First", "Second", "Third"} {
		if _, err := svc.CreateItem(ctx, name); err != nil {
			t.Fatalf("CreateItem(%q): %v", name, err)
		}
	}

	_, err := svc.CreateItem(ctx, "Fourth")
	var limitErr *domain.LimitError
	if !errors.As(err, &limitErr) || limitErr.Limit != "items" || limitErr.Max != 3 || limitErr.ReadOnly {
		t.Fatalf("err = %v, want the limit of 3 items reached", err)
	}
	if items, _ := svc.ListItems(ctx); len(items) != 3 {
		t.Errorf("got %d items, want 3", len(items))
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

func TestRenameItem_OverALimit(t *testing.T) {
	cases := map[string]struct {
		whenOver domain.OverLimit
		refused  bool
	}{
		"that makes them read only is refused and saves nothing": {domain.ReadOnly, true},
		"that blocks new ones works":                             {domain.BlockNew, false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			svc, items := overTheFreePlan(t, tc.whenOver)
			ctx := context.Background()
			first := items[0]

			_, err := svc.RenameItem(ctx, first.ID, "Renamed")
			got, _ := svc.GetItem(ctx, first.ID)
			if !tc.refused {
				if err != nil || got.Name != "Renamed" {
					t.Fatalf("RenameItem = %v, name %q; want it renamed", err, got.Name)
				}
				return
			}
			var limitErr *domain.LimitError
			if !errors.As(err, &limitErr) || limitErr.Limit != "items" || limitErr.Max != 3 || !limitErr.ReadOnly {
				t.Fatalf("err = %v, want items read only while over the limit of 3", err)
			}
			if got.Name != "First" {
				t.Errorf("name = %q, want it unchanged", got.Name)
			}
		})
	}
}

func TestDeleteItem_OverAReadOnlyLimitWorks(t *testing.T) {
	svc, items := overTheFreePlan(t, domain.ReadOnly)
	ctx := context.Background()
	last := items[len(items)-1]

	if err := svc.DeleteItem(ctx, last.ID); err != nil {
		t.Fatalf("DeleteItem: %v", err)
	}
	if _, err := svc.GetItem(ctx, last.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetItem after delete err = %v, want ErrNotFound", err)
	}
}

// errPlans is the error unavailablePlans fails with.
var errPlans = errors.New("plans unavailable")

// unavailablePlans is a ports.Entitlements whose lookup always fails.
type unavailablePlans struct{}

func (unavailablePlans) For(context.Context, domain.WorkspaceID) (domain.Entitlements, error) {
	return domain.Entitlements{}, errPlans
}

func TestUseCases_WhenThePlanCantBeLookedUpAreRefusedAndSaveNothing(t *testing.T) {
	cases := map[string]func(context.Context, *app.ItemService, domain.Item) error{
		"CreateItem": func(ctx context.Context, s *app.ItemService, _ domain.Item) error {
			_, err := s.CreateItem(ctx, "Second")
			return err
		},
		"RenameItem": func(ctx context.Context, s *app.ItemService, existing domain.Item) error {
			_, err := s.RenameItem(ctx, existing.ID, "Renamed")
			return err
		},
	}
	for name, call := range cases {
		t.Run(name, func(t *testing.T) {
			svc, made := serviceWithItems(t, unavailablePlans{}, "First")
			ctx := context.Background()
			existing := made[0]

			if err := call(ctx, svc, existing); !errors.Is(err, errPlans) {
				t.Fatalf("err = %v, want %v", err, errPlans)
			}
			items, _ := svc.ListItems(ctx)
			if len(items) != 1 || items[0] != existing {
				t.Errorf("items = %+v, want only %+v", items, existing)
			}
		})
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
			svc, _ := newServiceWith(repo, testsupport.NoWorkspace{}, testsupport.FixedEntitlements{})

			if err := call(context.Background(), svc); !errors.Is(err, domain.ErrNoWorkspace) {
				t.Errorf("err = %v, want ErrNoWorkspace", err)
			}
			if repo.calls != 0 {
				t.Errorf("repository called %d times, want 0", repo.calls)
			}
		})
	}
}
