package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
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
