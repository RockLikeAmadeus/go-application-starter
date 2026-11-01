// Package memory is an in-process ItemRepository. It is used by unit and
// acceptance tests and by `STORAGE=memory`.
package memory

import (
	"context"
	"sync"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/storageutil"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

type ItemRepository struct {
	mu    sync.RWMutex
	items map[string]domain.Item
}

var _ ports.ItemRepository = (*ItemRepository)(nil)

func NewItemRepository() *ItemRepository {
	return &ItemRepository{items: map[string]domain.Item{}}
}

func (r *ItemRepository) Save(_ context.Context, item domain.Item) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[item.ID] = item
	return nil
}

func (r *ItemRepository) FindByID(_ context.Context, workspace domain.WorkspaceID, id string) (domain.Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.lookup(workspace, id)
	if !ok {
		return domain.Item{}, domain.ErrNotFound
	}
	return item, nil
}

func (r *ItemRepository) FindAll(_ context.Context, workspace domain.WorkspaceID) ([]domain.Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return storageutil.WorkspaceItems(r.items, workspace), nil
}

func (r *ItemRepository) Delete(_ context.Context, workspace domain.WorkspaceID, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.lookup(workspace, id); !ok {
		return domain.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

// lookup returns the workspace's item with the given ID. An item in another
// workspace is treated as missing. The caller must hold r.mu.
func (r *ItemRepository) lookup(workspace domain.WorkspaceID, id string) (domain.Item, bool) {
	item, ok := r.items[id]
	if !ok || item.WorkspaceID != workspace {
		return domain.Item{}, false
	}
	return item, true
}
