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
	ExportItems(ctx context.Context) ([]byte, error)
}

// ItemRepository is the driven port for persisting Items. FindAll orders them
// by CreatedAt, then ID. Every implementation must pass
// storagetest.RunItemRepositoryContract, which includes the workspace
// isolation suite.
type ItemRepository interface {
	WorkspaceRepository[domain.Item]
}

// sample:end

// WorkspaceRepository is the shape of a driven port for an entity whose records
// belong to a workspace (ADR 5). Declare an entity's repository by embedding
// it, and add any methods the entity needs:
//
//	type WidgetRepository interface {
//		WorkspaceRepository[domain.Widget]
//	}
//
// Every implementation must pass storagetest.RunWorkspaceIsolation. See
// docs/workspaces.md.
type WorkspaceRepository[T any] interface {
	// Save inserts or replaces the record with the same ID. The record keeps
	// the workspace it carries.
	Save(ctx context.Context, record T) error
	// FindByID returns domain.ErrNotFound when the workspace has no record with
	// the given ID, including when the record belongs to another workspace.
	FindByID(ctx context.Context, workspace domain.WorkspaceID, id string) (T, error)
	// FindAll returns the workspace's records, and none from other workspaces.
	FindAll(ctx context.Context, workspace domain.WorkspaceID) ([]T, error)
	// Delete returns domain.ErrNotFound, and deletes nothing, when the
	// workspace has no record with the given ID, including when the record
	// belongs to another workspace.
	Delete(ctx context.Context, workspace domain.WorkspaceID, id string) error
}

// EntitlementsService is the driving port for what the current workspace's
// plan allows it.
type EntitlementsService interface {
	CurrentEntitlements(ctx context.Context) (domain.Entitlements, error)
}

// EntitlementsProvider tells a use case what a workspace's plan allows it.
type EntitlementsProvider interface {
	For(ctx context.Context, workspace domain.WorkspaceID) (domain.Entitlements, error)
}

// BillingService is the driving port for changing which plan a workspace is
// on. It takes the workspace as an argument, because an admin acts on someone
// else's workspace, so nothing resolves a current one.
type BillingService interface {
	ChangePlan(ctx context.Context, workspace domain.WorkspaceID, plan string) error
}

// WorkspacePlanRepository is the driven port for storing which plan each
// workspace is on. A plan is one value per workspace, not a list of records, so
// it doesn't embed WorkspaceRepository.
type WorkspacePlanRepository interface {
	// PlanOf returns domain.ErrNotFound for a workspace with no stored plan.
	PlanOf(ctx context.Context, workspace domain.WorkspaceID) (string, error)
	// SavePlan inserts or replaces the workspace's plan.
	SavePlan(ctx context.Context, workspace domain.WorkspaceID, plan string) error
}

// PlanCatalog tells a use case which plans exist.
type PlanCatalog interface {
	HasPlan(ctx context.Context, plan string) (bool, error)
}

// WorkspaceResolver tells a use case which workspace the current request
// works in.
type WorkspaceResolver interface {
	CurrentWorkspace(ctx context.Context) (domain.WorkspaceID, error)
}

// Clock abstracts time so use cases are deterministic under test.
type Clock interface {
	Now() time.Time
}

// IDGenerator abstracts identifier creation.
type IDGenerator interface {
	NewID() string
}
