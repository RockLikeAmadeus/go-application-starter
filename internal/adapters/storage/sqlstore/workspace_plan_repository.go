package sqlstore

import (
	"context"
	"database/sql"
	"errors"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// WorkspacePlanRepository stores which plan each workspace is on, in the
// workspace_plans table.
type WorkspacePlanRepository struct {
	db *sql.DB
}

var _ ports.WorkspacePlanRepository = (*WorkspacePlanRepository)(nil)

// NewWorkspacePlanRepository returns a repository built on db, which
// Migrate has brought up to date.
func NewWorkspacePlanRepository(db *sql.DB) *WorkspacePlanRepository {
	return &WorkspacePlanRepository{db: db}
}

// PlanOf returns the workspace's plan.
func (r *WorkspacePlanRepository) PlanOf(ctx context.Context, workspace domain.WorkspaceID) (string, error) {
	var plan string
	err := r.db.QueryRowContext(ctx,
		`SELECT plan FROM workspace_plans WHERE workspace_id = $1`, workspace).Scan(&plan)
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.ErrNotFound
	}
	return plan, err
}

// SavePlan stores the workspace's plan.
func (r *WorkspacePlanRepository) SavePlan(ctx context.Context, workspace domain.WorkspaceID, plan string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO workspace_plans (workspace_id, plan) VALUES ($1, $2)
		ON CONFLICT (workspace_id) DO UPDATE SET plan = excluded.plan`, workspace, plan)
	return err
}
