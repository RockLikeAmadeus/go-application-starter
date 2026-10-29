package jsonfile

import (
	"context"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/storageutil"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

const itemsTable = "items"

// itemRecord is the on-disk shape of an Item. Keeping it separate from the
// domain type lets the file format evolve independently.
type itemRecord struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toRecord(i domain.Item) itemRecord {
	return itemRecord{ID: i.ID, Name: i.Name, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt}
}

func (r itemRecord) toDomain() domain.Item {
	return domain.Item{ID: r.ID, Name: r.Name, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

// ItemRepository persists Items to <dir>/items.json. Reads go to disk each
// time so the file stays the single source of truth and can be hand-edited
// while the process is stopped.
type ItemRepository struct {
	store *Store
}

var _ ports.ItemRepository = (*ItemRepository)(nil)

func NewItemRepository(store *Store) *ItemRepository {
	return &ItemRepository{store: store}
}

func (r *ItemRepository) Save(_ context.Context, item domain.Item) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	records, err := readTable[itemRecord](r.store, itemsTable)
	if err != nil {
		return err
	}
	records[item.ID] = toRecord(item)
	return writeTable(r.store, itemsTable, records)
}

func (r *ItemRepository) FindByID(_ context.Context, id string) (domain.Item, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	records, err := readTable[itemRecord](r.store, itemsTable)
	if err != nil {
		return domain.Item{}, err
	}
	rec, ok := records[id]
	if !ok {
		return domain.Item{}, domain.ErrNotFound
	}
	return rec.toDomain(), nil
}

func (r *ItemRepository) FindAll(_ context.Context) ([]domain.Item, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	records, err := readTable[itemRecord](r.store, itemsTable)
	if err != nil {
		return nil, err
	}
	items := make(map[string]domain.Item, len(records))
	for id, rec := range records {
		items[id] = rec.toDomain()
	}
	return storageutil.SortedItems(items), nil
}

func (r *ItemRepository) Delete(_ context.Context, id string) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	records, err := readTable[itemRecord](r.store, itemsTable)
	if err != nil {
		return err
	}
	if _, ok := records[id]; !ok {
		return domain.ErrNotFound
	}
	delete(records, id)
	return writeTable(r.store, itemsTable, records)
}
