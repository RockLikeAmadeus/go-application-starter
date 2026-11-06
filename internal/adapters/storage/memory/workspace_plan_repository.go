// Package memory is in-process storage. It is used by unit and acceptance tests
// and by `STORAGE=memory`.
package memory

import (
	"context"
	"sync"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// WorkspacePlanRepository stores which plan each workspace is on.
type WorkspacePlanRepository struct {
	mu    sync.RWMutex
	plans map[domain.WorkspaceID]string
}

var _ ports.WorkspacePlanRepository = (*WorkspacePlanRepository)(nil)

// NewWorkspacePlanRepository returns an empty repository.
func NewWorkspacePlanRepository() *WorkspacePlanRepository {
	return &WorkspacePlanRepository{plans: map[domain.WorkspaceID]string{}}
}

// PlanOf returns the workspace's plan.
func (r *WorkspacePlanRepository) PlanOf(_ context.Context, workspace domain.WorkspaceID) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	plan, ok := r.plans[workspace]
	if !ok {
		return "", domain.ErrNotFound
	}
	return plan, nil
}

// SavePlan stores the workspace's plan.
func (r *WorkspacePlanRepository) SavePlan(_ context.Context, workspace domain.WorkspaceID, plan string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.plans[workspace] = plan
	return nil
}
