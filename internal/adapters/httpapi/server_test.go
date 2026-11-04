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

// Storage with nothing to check, such as memory, is always ready.
func TestReadyzIsReadyWithNoReadinessCheck(t *testing.T) {
	rec := httptest.NewRecorder()
	httpapi.NewHandler(httpapi.Options{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Body.String(); got != "ready" {
		t.Errorf("body = %q, want %q", got, "ready")
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
	_ = json.Unmarshal([]byte(`{"plan":"free","limits":{"items":{"max":3,"whenOver":"block-new"}},"features":{}}`), &want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("body = %s, want %v", rec.Body, want)
	}
}

// The entitlements list every feature the plans know, by key, whether it's on
// or off, so front ends can show it, lock it or hide it.
func TestEntitlementsIncludeEveryFeatureByKey(t *testing.T) {
	free := fixedEntitlements{
		Plan: "free",
		Features: map[string]domain.Feature{
			"item-export": {Key: "item-export", Name: "Item export", WhenOff: domain.ShowLocked, AvailableOn: []string{"pro"}},
			"sharing":     {Key: "sharing", Name: "Sharing", On: true, WhenOff: domain.Hide},
			"audit-log":   {Key: "audit-log", Name: "Audit log", WhenOff: domain.Hide},
		},
	}
	rec := httptest.NewRecorder()
	httpapi.NewHandler(httpapi.Options{Entitlements: free}).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/entitlements", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got struct {
		Features any `json:"features"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	var want any
	_ = json.Unmarshal([]byte(`{
		"item-export": {"name": "Item export", "on": false, "whenOff": "show-locked", "availableOn": ["pro"]},
		"sharing":     {"name": "Sharing", "on": true, "whenOff": "hide", "availableOn": []},
		"audit-log":   {"name": "Audit log", "on": false, "whenOff": "hide", "availableOn": []}
	}`), &want)
	if !reflect.DeepEqual(got.Features, want) {
		t.Errorf("features = %v, want %v", got.Features, want)
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

// Warnings come from the composition root (bootstrap.App.Warnings) and are
// served as given, for every front end to show. With none, the list is empty
// rather than missing, so the web app can always read it.
func TestWarningsAreServedAsGivenAndEmptyWhenUnset(t *testing.T) {
	tests := []struct {
		name     string
		warnings []string
		want     []string
	}{
		{name: "unset", warnings: nil, want: []string{}},
		{name: "given", warnings: []string{"first warning", "second warning"}, want: []string{"first warning", "second warning"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := servedWarnings(t, httpapi.Options{Warnings: tt.warnings})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("warnings = %q, want %q", got, tt.want)
			}
		})
	}
}

// servedWarnings asks a handler built from opts for its warnings' messages,
// failing the test if the list is missing.
func servedWarnings(t *testing.T, opts httpapi.Options) []string {
	t.Helper()
	rec := httptest.NewRecorder()
	httpapi.NewHandler(opts).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/warnings", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body struct {
		Warnings *[]struct {
			Message string `json:"message"`
		} `json:"warnings"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	if body.Warnings == nil {
		t.Fatalf("body = %s, want a warnings list", rec.Body)
	}
	messages := make([]string, len(*body.Warnings))
	for i, w := range *body.Warnings {
		messages[i] = w.Message
	}
	return messages
}
