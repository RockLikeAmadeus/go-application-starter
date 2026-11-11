package acceptance

import (
	"context"
	"sync"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/plans"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// testPlans is a ports.EntitlementsProvider and ports.PlanCatalog whose plans the
// scenarios set: a catalog, starting from the shipped plans file, and the
// repository holding the plan each workspace is on. A workspace no scenario puts
// on a plan is on the local plan, as in the local edition. It answers through
// the plans adapter, which it locks around, because the http driver's server
// reads it from other goroutines.
type testPlans struct {
	mu      sync.Mutex
	catalog plans.Catalog
	stored  ports.WorkspacePlanRepository
}

var (
	_ ports.EntitlementsProvider = (*testPlans)(nil)
	_ ports.PlanCatalog          = (*testPlans)(nil)
)

func newTestPlans(stored ports.WorkspacePlanRepository) (*testPlans, error) {
	file, err := plans.Default()
	if err != nil {
		return nil, err
	}
	return &testPlans{catalog: file.Catalog, stored: stored}, nil
}

func (p *testPlans) For(ctx context.Context, workspace domain.WorkspaceID) (domain.Entitlements, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	adapter := plans.Provider{
		Catalog:         p.catalog,
		PlanOf:          plans.StoredPlanOf(p.stored, bootstrap.LocalPlan),
		ShouldGiveHints: true,
		LocalPlan:       bootstrap.LocalPlan,
	}
	return adapter.For(ctx, workspace)
}

func (p *testPlans) HasPlan(ctx context.Context, plan string) (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return plans.Provider{Catalog: p.catalog}.HasPlan(ctx, plan)
}

// define adds the plan to the catalog, or replaces it.
func (p *testPlans) define(entitlements domain.Entitlements) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.catalog[entitlements.Plan] = entitlements
}
