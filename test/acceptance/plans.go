package acceptance

import (
	"context"
	"sync"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/plans"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// testPlans is a ports.Entitlements whose plans the scenarios set: a catalog,
// starting from the shipped plans file, and the plan each workspace is on. A
// workspace no scenario puts on a plan is on the local plan, as in the local
// edition. It answers through the plans adapter, which it locks around,
// because the http driver's server reads it from other goroutines.
type testPlans struct {
	mu       sync.Mutex
	catalog  plans.Catalog
	assigned map[domain.WorkspaceID]string
}

var _ ports.Entitlements = (*testPlans)(nil)

func newTestPlans() (*testPlans, error) {
	catalog, err := plans.Default()
	if err != nil {
		return nil, err
	}
	return &testPlans{catalog: catalog, assigned: map[domain.WorkspaceID]string{}}, nil
}

func (p *testPlans) For(ctx context.Context, workspace domain.WorkspaceID) (domain.Entitlements, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	adapter := plans.Entitlements{
		Catalog: p.catalog,
		PlanOf: func(context.Context, domain.WorkspaceID) (string, error) {
			if plan, ok := p.assigned[workspace]; ok {
				return plan, nil
			}
			return bootstrap.LocalPlan, nil
		},
		Hints:     true,
		LocalPlan: bootstrap.LocalPlan,
	}
	return adapter.For(ctx, workspace)
}

// define adds the plan to the catalog, or replaces it.
func (p *testPlans) define(entitlements domain.Entitlements) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.catalog[entitlements.Plan] = entitlements
}

// assign puts the workspace on the plan.
func (p *testPlans) assign(workspace domain.WorkspaceID, plan string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.assigned[workspace] = plan
}
