package bootstrap_test

import (
	"context"
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
