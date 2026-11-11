// Package testsupport provides deterministic fakes for driven ports, shared
// by unit tests and acceptance tests.
package testsupport

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// FixedClock is a ports.Clock that only moves when told to.
type FixedClock struct {
	mu  sync.Mutex
	now time.Time
}

func NewFixedClock(t time.Time) *FixedClock { return &FixedClock{now: t} }

func (c *FixedClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *FixedClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// SequentialIDs is a ports.IDGenerator producing "<prefix>-1", "<prefix>-2", ...
type SequentialIDs struct {
	mu     sync.Mutex
	prefix string
	n      int
}

func NewSequentialIDs(prefix string) *SequentialIDs { return &SequentialIDs{prefix: prefix} }

func (g *SequentialIDs) NewID() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.n++
	return fmt.Sprintf("%s-%d", g.prefix, g.n)
}

// FixedWorkspace is a ports.WorkspaceResolver that always resolves to itself.
type FixedWorkspace domain.WorkspaceID

func (w FixedWorkspace) CurrentWorkspace(context.Context) (domain.WorkspaceID, error) {
	return domain.WorkspaceID(w), nil
}

// NoWorkspace is a ports.WorkspaceResolver for a request with no workspace.
type NoWorkspace struct{}

func (NoWorkspace) CurrentWorkspace(context.Context) (domain.WorkspaceID, error) {
	return "", domain.ErrNoWorkspace
}

// FixedEntitlements is a ports.EntitlementsProvider that answers the same
// entitlements for every workspace. Its zero value is a plan with no limits.
type FixedEntitlements domain.Entitlements

func (e FixedEntitlements) For(context.Context, domain.WorkspaceID) (domain.Entitlements, error) {
	return domain.Entitlements(e), nil
}

// StoredPlans is a ports.WorkspacePlanRepository that holds each workspace's
// plan in a map.
type StoredPlans map[domain.WorkspaceID]string

func (p StoredPlans) PlanOf(_ context.Context, workspace domain.WorkspaceID) (string, error) {
	plan, ok := p[workspace]
	if !ok {
		return "", domain.ErrNotFound
	}
	return plan, nil
}

func (p StoredPlans) SavePlan(_ context.Context, workspace domain.WorkspaceID, plan string) error {
	p[workspace] = plan
	return nil
}

// KnownPlans is a ports.PlanCatalog that has only the plans it lists.
type KnownPlans []string

func (p KnownPlans) HasPlan(_ context.Context, plan string) (bool, error) {
	return slices.Contains(p, plan), nil
}
