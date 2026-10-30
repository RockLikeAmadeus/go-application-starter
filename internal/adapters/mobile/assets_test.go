package mobile_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/mobile"
)

// Android asks for /index.html, which Go's file server answers with a
// redirect to "./" instead of the page.
func TestAssetHandlerServesTheWebAppForIndexHTML(t *testing.T) {
	webApp := fstest.MapFS{"index.html": {Data: []byte("<title>Starter</title>")}}
	h := mobile.AssetHandler(http.FileServerFS(webApp))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/index.html", nil))

	if rec.Code != http.StatusOK || rec.Body.String() != "<title>Starter</title>" {
		t.Errorf("GET /index.html: status %d, body %q; want 200 and the web app's index.html", rec.Code, rec.Body)
	}
}

// Wails' asset path turns every request into a GET with no body, so an API
// call that reaches it instead of the bridge must fail loudly. The spike saw a
// POST answered as a GET with 200: the web app showed success, but nothing was
// saved.
func TestAssetHandlerAnswersAPIRequestsWithNotFound(t *testing.T) {
	h := mobile.AssetHandler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/items", nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /api/v1/items: status %d, body %q; want 404", rec.Code, rec.Body)
	}
}

func TestAssetHandlerPassesOtherPathsThrough(t *testing.T) {
	var gotPath string
	h := mobile.AssetHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte("console.log('Starter')"))
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))

	if gotPath != "/assets/app.js" {
		t.Errorf("next got path %q, want /assets/app.js", gotPath)
	}
	if rec.Code != http.StatusOK || rec.Body.String() != "console.log('Starter')" {
		t.Errorf("GET /assets/app.js: status %d, body %q; want next's reply", rec.Code, rec.Body)
	}
}
