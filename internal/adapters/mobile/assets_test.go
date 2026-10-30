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
