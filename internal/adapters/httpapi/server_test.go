package httpapi_test

import (
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
