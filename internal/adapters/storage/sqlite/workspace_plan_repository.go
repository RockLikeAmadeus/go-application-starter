package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// WorkspacePlanRepository stores which plan each workspace is on, in the
// store's workspace_plans table.
type WorkspacePlanRepository struct {
	store *Store
}

var _ ports.WorkspacePlanRepository = (*WorkspacePlanRepository)(nil)

// NewWorkspacePlanRepository returns a repository built on store.
func NewWorkspacePlanRepository(store *Store) *WorkspacePlanRepository {
	return &WorkspacePlanRepository{store: store}
}

// PlanOf returns the workspace's plan.
func (r *WorkspacePlanRepository) PlanOf(ctx context.Context, workspace domain.WorkspaceID) (string, error) {
	var plan string
	err := r.store.db.QueryRowContext(ctx,
		`SELECT plan FROM workspace_plans WHERE workspace_id = ?`, workspace).Scan(&plan)
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.ErrNotFound
	}
	return plan, err
}

// SavePlan stores the workspace's plan.
func (r *WorkspacePlanRepository) SavePlan(ctx context.Context, workspace domain.WorkspaceID, plan string) error {
	_, err := r.store.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO workspace_plans (workspace_id, plan) VALUES (?, ?)`, workspace, plan)
	return err
}
