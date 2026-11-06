package jsonfile

import (
	"context"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

const workspacePlansTable = "workspace_plans"

// workspacePlanRecord is the on-disk shape of a workspace's plan. The workspace
// ID is its key, so it isn't in the record.
type workspacePlanRecord struct {
	Plan string `json:"plan"`
}

// WorkspacePlanRepository stores which plan each workspace is on, each in
// <dir>/workspace_plans/<workspace>.json.
type WorkspacePlanRepository struct {
	store *Store
}

var _ ports.WorkspacePlanRepository = (*WorkspacePlanRepository)(nil)

// NewWorkspacePlanRepository returns a repository built on store.
func NewWorkspacePlanRepository(store *Store) *WorkspacePlanRepository {
	return &WorkspacePlanRepository{store: store}
}

// PlanOf returns the workspace's plan.
func (r *WorkspacePlanRepository) PlanOf(_ context.Context, workspace domain.WorkspaceID) (string, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	records, err := readRecords[workspacePlanRecord](r.store, workspacePlansTable)
	if err != nil {
		return "", err
	}
	rec, ok := records[string(workspace)]
	if !ok {
		return "", domain.ErrNotFound
	}
	return rec.Plan, nil
}

// SavePlan stores the workspace's plan.
func (r *WorkspacePlanRepository) SavePlan(_ context.Context, workspace domain.WorkspaceID, plan string) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	return writeRecord(r.store, workspacePlansTable, string(workspace), workspacePlanRecord{Plan: plan})
}
