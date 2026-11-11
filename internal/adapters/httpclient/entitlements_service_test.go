package httpclient_test

import (
	"context"
	"errors"
	"net/http"
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
	client := clientOf(t, httpapi.NewHandler(httpapi.Options{Entitlements: core}))

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
			"export":  {Key: "export", Name: "Export", IsOn: true, WhenOff: domain.ShowLocked, AvailableOn: []string{"pro"}},
			"reports": {Key: "reports", Name: "Reports", WhenOff: domain.ShowLocked, AvailableOn: []string{"pro", "team"}},
			"beta":    {Key: "beta", Name: "Beta", WhenOff: domain.Hide},
		},
	}
	core := app.NewEntitlementsService(testsupport.FixedEntitlements(sent), testsupport.FixedWorkspace(domain.LocalWorkspace))
	client := clientOf(t, httpapi.NewHandler(httpapi.Options{Entitlements: core}))

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

// clientOf returns a client of a test server that answers with handler.
func clientOf(t *testing.T, handler http.Handler) *httpclient.EntitlementsService {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client, err := httpclient.NewEntitlementsService(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

// errorOverHTTP returns what the client reports when the core fails with sent
// on the server.
func errorOverHTTP(t *testing.T, sent error) error {
	t.Helper()
	client := clientOf(t, httpapi.NewHandler(httpapi.Options{Entitlements: failingEntitlements{sent}}))
	_, err := client.CurrentEntitlements(context.Background())
	return err
}

// A validation error on the server comes back as the same validation error,
// with its sentence whole even when it starts with the field's name, so
// callers show what the server said.
func TestAValidationErrorComesBackAsAValidationError(t *testing.T) {
	sent := &domain.ValidationError{Field: "name", Message: "name is the field this sentence starts with."}
	err := errorOverHTTP(t, sent)
	var got *domain.ValidationError
	if !errors.As(err, &got) || *got != *sent {
		t.Fatalf("err = %v (%+v), want %+v", err, got, sent)
	}
}

// A limit error on the server comes back as the same limit error, so callers
// can use errors.As whichever way they reach the core.
func TestALimitErrorComesBackAsALimitError(t *testing.T) {
	sent := &domain.LimitError{Limit: "items", Max: 3, Plan: "free", IsReadOnly: true}
	err := errorOverHTTP(t, sent)
	var got *domain.LimitError
	if !errors.As(err, &got) || *got != *sent {
		t.Fatalf("err = %v (%+v), want %+v", err, got, sent)
	}
}

// A feature error on the server comes back as the same feature error, with
// its upgrade hint or without one, so callers can use errors.As whichever way
// they reach the core.
func TestAFeatureErrorComesBackAsAFeatureError(t *testing.T) {
	tests := map[string]*domain.FeatureError{
		"with a hint":    {Feature: "export", Name: "Export", AvailableOn: []string{"pro", "team"}},
		"without a hint": {Feature: "export", Name: "Export"},
	}
	for name, sent := range tests {
		t.Run(name, func(t *testing.T) {
			err := errorOverHTTP(t, sent)
			var got *domain.FeatureError
			if !errors.As(err, &got) || !reflect.DeepEqual(got, sent) {
				t.Fatalf("err = %#v, want %#v", err, sent)
			}
		})
	}
}

// No workspace on the server, as when signed out, comes back as
// domain.ErrNoWorkspace, so callers can tell the person to sign in whichever
// way they reach the core.
func TestNoWorkspaceComesBackAsErrNoWorkspace(t *testing.T) {
	err := errorOverHTTP(t, domain.ErrNoWorkspace)
	if !errors.Is(err, domain.ErrNoWorkspace) {
		t.Fatalf("err = %v, want %v", err, domain.ErrNoWorkspace)
	}
}

// An error the client has no domain error for reads as a sentence, with the
// status and what the server said.
func TestAnyOtherServerErrorReadsAsASentence(t *testing.T) {
	err := errorOverHTTP(t, errors.New("boom"))
	want := "The server returned an error (500): An unexpected error occurred on the server. Please try again later."
	if err == nil || err.Error() != want {
		t.Fatalf("err = %v, want %q", err, want)
	}
}

// An answer that isn't a problem document, such as a proxy's error page, reads
// as a sentence with the status, as the client has no detail to pass on.
func TestAServerErrorWithNoProblemDocumentReadsAsASentence(t *testing.T) {
	client := clientOf(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))

	_, err := client.CurrentEntitlements(context.Background())

	want := "The server returned an unexpected response (502)."
	if err == nil || err.Error() != want {
		t.Fatalf("err = %v, want %q", err, want)
	}
}
