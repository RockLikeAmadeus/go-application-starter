package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// The web app manifest's type must not depend on the host's MIME tables:
// Go's own table, the freedesktop one and the distroless image all lack it.
func TestServesWebAppManifestAsManifestJSON(t *testing.T) {
	assets := fstest.MapFS{
		"index.html":           {Data: []byte("<!doctype html>")},
		"manifest.webmanifest": {Data: []byte(`{"name":"Starter"}`)},
	}
	handler := httpapi.NewHandler(httpapi.Options{WebAssets: assets})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/manifest.webmanifest", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got, want := rec.Header().Get("Content-Type"), "application/manifest+json"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}
}

// Readiness comes from storage, not from any one kind of data, so it works
// without the sample's items.
func TestReadyzIsUnavailableAndLogsWhyWhenStorageIsNotReady(t *testing.T) {
	var logs bytes.Buffer
	handler := httpapi.NewHandler(httpapi.Options{
		Logger: slog.New(slog.NewTextHandler(&logs, nil)),
		Ready:  func(context.Context) error { return errors.New("disk on fire") },
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	if got := rec.Body.String(); !strings.Contains(got, "storage unavailable") {
		t.Errorf("body = %q, want it to say storage unavailable", got)
	}
	if !strings.Contains(logs.String(), "disk on fire") {
		t.Errorf("logs = %q, want the readiness error", logs.String())
	}
}

// The server is reached over the network, so by default the web app must
// treat a lost connection as losing its data.
func TestConnectionIsRequiredByDefault(t *testing.T) {
	if !connectionRequired(t, httpapi.Options{}) {
		t.Error("required = false, want true")
	}
}

// In the desktop app's local-only mode the data is on this device, so a lost
// connection doesn't matter.
func TestConnectionIsNotRequiredForLocalData(t *testing.T) {
	if connectionRequired(t, httpapi.Options{LocalData: true}) {
		t.Error("required = true, want false")
	}
}

// connectionRequired asks a handler built from opts whether the web app needs
// a connection, failing the test if the answer is missing.
func connectionRequired(t *testing.T, opts httpapi.Options) bool {
	t.Helper()
	rec := httptest.NewRecorder()
	httpapi.NewHandler(opts).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/connection", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body struct {
		Required *bool `json:"required"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	if body.Required == nil {
		t.Fatalf("body = %s, want a required field", rec.Body)
	}
	return *body.Required
}

// fixedEntitlements is a ports.EntitlementsService that always answers the
// same entitlements.
type fixedEntitlements domain.Entitlements

func (e fixedEntitlements) CurrentEntitlements(context.Context) (domain.Entitlements, error) {
	return domain.Entitlements(e), nil
}

func TestEntitlementsAreTheCurrentWorkspacesPlanAndLimits(t *testing.T) {
	free := fixedEntitlements{
		Plan:   "free",
		Limits: map[string]domain.Limit{"items": {Name: "items", Max: 3, WhenOver: domain.BlockNew}},
	}
	rec := httptest.NewRecorder()
	httpapi.NewHandler(httpapi.Options{Entitlements: free}).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/entitlements", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got, want any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	_ = json.Unmarshal([]byte(`{"plan":"free","limits":{"items":{"max":3,"whenOver":"block-new"}}}`), &want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("body = %s, want %v", rec.Body, want)
	}
}

// failingEntitlements is a ports.EntitlementsService that always fails with
// err.
type failingEntitlements struct{ err error }

func (e failingEntitlements) CurrentEntitlements(context.Context) (domain.Entitlements, error) {
	return domain.Entitlements{}, e.err
}

// A limit error is forbidden, and its problem document says which limit, so
// clients can map it back.
func TestALimitErrorIsAForbiddenProblemNamingTheLimit(t *testing.T) {
	limitErr := &domain.LimitError{Limit: "items", Max: 3, ReadOnly: true}
	rec := httptest.NewRecorder()
	httpapi.NewHandler(httpapi.Options{Entitlements: failingEntitlements{limitErr}}).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/entitlements", nil))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if got, want := rec.Header().Get("Content-Type"), "application/problem+json"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	want := map[string]any{
		"type": "limit", "title": "Forbidden", "status": 403.0, "detail": limitErr.Error(),
		"limit": "items", "max": 3.0, "readOnly": true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("body = %v, want %v", got, want)
	}
}
