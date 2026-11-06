// Package memory is in-process storage. It is used by unit and acceptance tests
// and by `STORAGE=memory`.
package memory

import (
	"context"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// WorkspacePlanRepository stores which plan each workspace is on.
type WorkspacePlanRepository struct {
	plan string
}

var _ ports.WorkspacePlanRepository = (*WorkspacePlanRepository)(nil)

// NewWorkspacePlanRepository returns an empty repository.
func NewWorkspacePlanRepository() *WorkspacePlanRepository {
	return &WorkspacePlanRepository{}
}

// PlanOf returns the workspace's plan.
func (r *WorkspacePlanRepository) PlanOf(_ context.Context, _ domain.WorkspaceID) (string, error) {
	if r.plan == "" {
		return "", domain.ErrNotFound
	}
	return r.plan, nil
}

// SavePlan stores the workspace's plan.
func (r *WorkspacePlanRepository) SavePlan(_ context.Context, _ domain.WorkspaceID, plan string) error {
	r.plan = plan
	return nil
}
