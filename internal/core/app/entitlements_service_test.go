package app_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/app"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/testsupport"
)

// plansByWorkspace is a ports.Entitlements that answers each workspace's
// entitlements from a map.
type plansByWorkspace map[domain.WorkspaceID]domain.Entitlements

func (p plansByWorkspace) For(_ context.Context, workspace domain.WorkspaceID) (domain.Entitlements, error) {
	return p[workspace], nil
}

func TestCurrentEntitlements_AreTheCurrentWorkspacesPlan(t *testing.T) {
	free := domain.Entitlements{
		Plan:   "free",
		Limits: map[string]domain.Limit{"items": {Name: "items", Max: 3, WhenOver: domain.BlockNew}},
	}
	plans := plansByWorkspace{"ws-a": free, "ws-b": {Plan: "pro"}}
	svc := app.NewEntitlementsService(plans, testsupport.FixedWorkspace("ws-a"))

	got, err := svc.CurrentEntitlements(context.Background())
	if err != nil {
		t.Fatalf("CurrentEntitlements: %v", err)
	}
	if !reflect.DeepEqual(got, free) {
		t.Errorf("CurrentEntitlements = %+v, want %+v", got, free)
	}
}
