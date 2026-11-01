// Package app implements the driving ports (use cases) on top of the domain
// and the driven ports. It contains orchestration only; business rules live
// in the domain package.
package app

import (
	"context"
	"fmt"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// ItemService implements ports.ItemService.
type ItemService struct {
	repo         ports.ItemRepository
	workspaces   ports.WorkspaceResolver
	entitlements ports.Entitlements
	clock        ports.Clock
	ids          ports.IDGenerator
}

var _ ports.ItemService = (*ItemService)(nil)

// NewItemService wires the use cases to their driven ports.
func NewItemService(
	repo ports.ItemRepository,
	workspaces ports.WorkspaceResolver,
	entitlements ports.Entitlements,
	clock ports.Clock,
	ids ports.IDGenerator,
) *ItemService {
	return &ItemService{repo: repo, workspaces: workspaces, entitlements: entitlements, clock: clock, ids: ids}
}

// itemsLimit is the name of the plan limit on how many items a workspace has.
const itemsLimit = "items"

// itemUsage returns the workspace's entitlements and how many items it has,
// for checking the items limit.
func (s *ItemService) itemUsage(ctx context.Context, workspace domain.WorkspaceID) (domain.Entitlements, int) {
	entitlements, _ := s.entitlements.For(ctx, workspace)
	existing, _ := s.repo.FindAll(ctx, workspace)
	return entitlements, len(existing)
}

func (s *ItemService) CreateItem(ctx context.Context, name string) (domain.Item, error) {
	workspace, err := s.workspaces.CurrentWorkspace(ctx)
	if err != nil {
		return domain.Item{}, err
	}
	entitlements, count := s.itemUsage(ctx, workspace)
	if err := entitlements.CheckCreate(itemsLimit, count); err != nil {
		return domain.Item{}, err
	}
	item, err := domain.NewItem(workspace, s.ids.NewID(), name, s.clock.Now())
	if err != nil {
		return domain.Item{}, err
	}
	if err := s.repo.Save(ctx, item); err != nil {
		return domain.Item{}, fmt.Errorf("save item: %w", err)
	}
	return item, nil
}

func (s *ItemService) GetItem(ctx context.Context, id string) (domain.Item, error) {
	workspace, err := s.workspaces.CurrentWorkspace(ctx)
	if err != nil {
		return domain.Item{}, err
	}
	return s.repo.FindByID(ctx, workspace, id)
}

func (s *ItemService) ListItems(ctx context.Context) ([]domain.Item, error) {
	workspace, err := s.workspaces.CurrentWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.FindAll(ctx, workspace)
}

func (s *ItemService) RenameItem(ctx context.Context, id, name string) (domain.Item, error) {
	workspace, err := s.workspaces.CurrentWorkspace(ctx)
	if err != nil {
		return domain.Item{}, err
	}
	item, err := s.repo.FindByID(ctx, workspace, id)
	if err != nil {
		return domain.Item{}, err
	}
	entitlements, count := s.itemUsage(ctx, workspace)
	if err := entitlements.CheckChange(itemsLimit, count); err != nil {
		return domain.Item{}, err
	}
	if err := item.Rename(name, s.clock.Now()); err != nil {
		return domain.Item{}, err
	}
	if err := s.repo.Save(ctx, item); err != nil {
		return domain.Item{}, fmt.Errorf("save item: %w", err)
	}
	return item, nil
}

func (s *ItemService) DeleteItem(ctx context.Context, id string) error {
	workspace, err := s.workspaces.CurrentWorkspace(ctx)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, workspace, id)
}
