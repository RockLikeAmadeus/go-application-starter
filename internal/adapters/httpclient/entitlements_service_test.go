package httpclient_test

import (
	"context"
	"errors"
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
		Plan:     "free",
		Limits:   map[string]domain.Limit{"items": {Name: "items", Max: 3, WhenOver: domain.ReadOnly}},
		Features: map[string]domain.Feature{},
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

// The client reads every feature back by key, on or off, with its upgrade
// hint; a feature with no hint comes back with none, as the core sent it.
func TestCurrentEntitlements_ReadsTheFeatures(t *testing.T) {
	sent := domain.Entitlements{
		Plan:   "free",
		Limits: map[string]domain.Limit{},
		Features: map[string]domain.Feature{
			"export":  {Key: "export", Name: "Export", On: true, WhenOff: domain.ShowLocked, AvailableOn: []string{"pro"}},
			"reports": {Key: "reports", Name: "Reports", WhenOff: domain.ShowLocked, AvailableOn: []string{"pro", "team"}},
			"beta":    {Key: "beta", Name: "Beta", WhenOff: domain.Hide},
		},
	}
	core := app.NewEntitlementsService(testsupport.FixedEntitlements(sent), testsupport.FixedWorkspace(domain.LocalWorkspace))
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
	if !reflect.DeepEqual(got.Features, sent.Features) {
		t.Errorf("Features = %+v, want %+v", got.Features, sent.Features)
	}
}

// failingEntitlements is a ports.EntitlementsService that always fails with
// err.
type failingEntitlements struct{ err error }

func (e failingEntitlements) CurrentEntitlements(context.Context) (domain.Entitlements, error) {
	return domain.Entitlements{}, e.err
}

// A limit error on the server comes back as the same limit error, so callers
// can use errors.As whichever way they reach the core.
func TestALimitErrorComesBackAsALimitError(t *testing.T) {
	sent := &domain.LimitError{Limit: "items", Max: 3, ReadOnly: true}
	srv := httptest.NewServer(httpapi.NewHandler(httpapi.Options{Entitlements: failingEntitlements{sent}}))
	t.Cleanup(srv.Close)
	client, err := httpclient.NewEntitlementsService(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.CurrentEntitlements(context.Background())
	var got *domain.LimitError
	if !errors.As(err, &got) || *got != *sent {
		t.Fatalf("err = %v, want %+v", err, sent)
	}
}
