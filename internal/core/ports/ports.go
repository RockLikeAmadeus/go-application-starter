// Package ports declares the boundaries of the hexagon.
//
// Driving (primary) ports are implemented by the core and called by UIs:
// HTTP, CLI, TUI. Driven (secondary) ports are implemented by adapters and
// called by the core: storage, clocks, ID generation.
package ports

import (
	"context"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// sample:start

// ItemService is the driving port for the Item use cases. Every UI talks to
// the core exclusively through this interface, which is also implemented by
// the HTTP client adapter so CLI/TUI can run against a remote server.
type ItemService interface {
	CreateItem(ctx context.Context, name string) (domain.Item, error)
	GetItem(ctx context.Context, id string) (domain.Item, error)
	ListItems(ctx context.Context) ([]domain.Item, error)
	RenameItem(ctx context.Context, id, name string) (domain.Item, error)
	DeleteItem(ctx context.Context, id string) error
}

// ItemRepository is the driven port for persisting Items. Every
// implementation must pass the contract suite in adapters/storage/storagetest.
type ItemRepository interface {
	// Save inserts or replaces the item with the same ID.
	Save(ctx context.Context, item domain.Item) error
	// FindByID returns domain.ErrNotFound when no item has the given ID.
	FindByID(ctx context.Context, id string) (domain.Item, error)
	// FindAll returns all items ordered by CreatedAt, then ID.
	FindAll(ctx context.Context) ([]domain.Item, error)
	// Delete returns domain.ErrNotFound when no item has the given ID.
	Delete(ctx context.Context, id string) error
}

// sample:end

// Clock abstracts time so use cases are deterministic under test.
type Clock interface {
	Now() time.Time
}

// IDGenerator abstracts identifier creation.
type IDGenerator interface {
	NewID() string
}
