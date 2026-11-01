package bootstrap_test

import (
	"context"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// The local edition's one workspace is on the local plan.
func TestNewEmbedded_IsOnTheLocalPlan(t *testing.T) {
	a, err := bootstrap.NewEmbedded(config.Config{Storage: config.StorageMemory})
	if err != nil {
		t.Fatalf("NewEmbedded: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })
	if a.Entitlements == nil {
		t.Fatal("the App has no Entitlements")
	}

	got, err := a.Entitlements.CurrentEntitlements(context.Background())
	if err != nil {
		t.Fatalf("CurrentEntitlements: %v", err)
	}
	if got.Plan != "local" || len(got.Limits) != 0 {
		t.Errorf("CurrentEntitlements = %+v, want the local plan with no limits", got)
	}
}

// A build whose local plan isn't in the plans file fails at start, naming
// the plan, rather than on the first request.
func TestNewEmbedded_RefusesALocalPlanThePlansFileDoesntHave(t *testing.T) {
	defer func(plan string) { bootstrap.LocalPlan = plan }(bootstrap.LocalPlan)
	bootstrap.LocalPlan = "gold"

	a, err := bootstrap.NewEmbedded(config.Config{Storage: config.StorageMemory})
	if err == nil {
		_ = a.Close()
		t.Fatal("NewEmbedded succeeded, want an error naming the gold plan")
	}
	if !strings.Contains(err.Error(), `"gold"`) {
		t.Errorf("err = %v, want it to name the gold plan", err)
	}
}
