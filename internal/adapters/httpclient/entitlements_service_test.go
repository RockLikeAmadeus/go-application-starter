package httpclient_test

import (
	"context"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/app"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/testsupport"
)

// The client reads the server's answer back into the same entitlements the
// core gave the server.
func TestCurrentEntitlements_RoundTrip(t *testing.T) {
	free := domain.Entitlements{
		Plan:   "free",
		Limits: map[string]domain.Limit{"items": {Name: "items", Max: 3, WhenOver: domain.ReadOnly}},
	}
	core := app.NewEntitlementsService(testsupport.FixedEntitlements(free), testsupport.FixedWorkspace(domain.LocalWorkspace))
	srv := httptest.NewServer(httpapi.NewHandler(httpapi.Options{Entitlements: core}))
	t.Cleanup(srv.Close)
	client, err := httpclient.NewEntitlementsService(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	got, err := client.CurrentEntitlements(context.Background())
	if err != nil {
		t.Fatalf("CurrentEntitlements: %v", err)
	}
	if !reflect.DeepEqual(got, free) {
		t.Errorf("CurrentEntitlements = %+v, want %+v", got, free)
	}
}
