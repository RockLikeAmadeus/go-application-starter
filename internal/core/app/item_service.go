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
	repo  ports.ItemRepository
	clock ports.Clock
	ids   ports.IDGenerator
}

var _ ports.ItemService = (*ItemService)(nil)

// NewItemService wires the use cases to their driven ports.
func NewItemService(repo ports.ItemRepository, clock ports.Clock, ids ports.IDGenerator) *ItemService {
	return &ItemService{repo: repo, clock: clock, ids: ids}
}

func (s *ItemService) CreateItem(ctx context.Context, name string) (domain.Item, error) {
	item, err := domain.NewItem(s.ids.NewID(), name, s.clock.Now())
	if err != nil {
		return domain.Item{}, err
	}
	if err := s.repo.Save(ctx, item); err != nil {
		return domain.Item{}, fmt.Errorf("save item: %w", err)
	}
	return item, nil
}

func (s *ItemService) GetItem(ctx context.Context, id string) (domain.Item, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ItemService) ListItems(ctx context.Context) ([]domain.Item, error) {
	return s.repo.FindAll(ctx)
}

func (s *ItemService) RenameItem(ctx context.Context, id, name string) (domain.Item, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
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
	return s.repo.Delete(ctx, id)
}
