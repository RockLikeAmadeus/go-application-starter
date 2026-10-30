package desktop_test

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/desktop"
)

// The desktop app's data is on this device, so the web app must not wait for
// a connection or say it is offline.
func TestOpenReportsNoConnectionIsRequired(t *testing.T) {
	app := openAt(t, t.TempDir())

	if connectionRequired(t, app.Handler) {
		t.Error("required = true, want false")
	}
}

func TestOpenServesTheWebApp(t *testing.T) {
	webApp := fstest.MapFS{"index.html": {Data: []byte("<title>Starter</title>")}}
	app := openWith(t, t.TempDir(), webApp)

	rec := httptest.NewRecorder()
	app.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Body.String(); got != "<title>Starter</title>" {
		t.Errorf("GET /: body = %q, want the web app's index.html", got)
	}
}

// The desktop app's files are built into it, so it needs no service worker
// (see withoutServiceWorker for why it would do harm).
func TestOpenServesNoServiceWorker(t *testing.T) {
	app := openAt(t, t.TempDir())

	rec := httptest.NewRecorder()
	app.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/sw.js", nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /sw.js: status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// The phone has no os.UserConfigDir, so Open takes the settings-folder lookup
// and keeps its data in a "starter" folder there.
func TestOpenCreatesTheDataFolderUnderTheGivenSettingsFolder(t *testing.T) {
	settings := t.TempDir()

	app, err := desktop.Open(envOf(nil), configDirAt(settings), nil)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	if _, err := os.Stat(filepath.Join(settings, "starter")); err != nil {
		t.Errorf("data folder not created under the settings folder: %v", err)
	}
}

// openAt opens the desktop app on dataDir, without the web app, and closes it
// when the test ends.
func openAt(t *testing.T, dataDir string) *desktop.App {
	t.Helper()
	return openWith(t, dataDir, nil)
}

// openWith opens the desktop app on dataDir, serving webApp, and closes it
// when the test ends.
func openWith(t *testing.T, dataDir string, webApp fs.FS) *desktop.App {
	t.Helper()
	app, err := desktop.Open(envOf(map[string]string{"STARTER_DATA_DIR": dataDir}), configDirAt(t.TempDir()), webApp)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })
	return app
}

// connectionRequired asks the REST API whether the web app needs a
// connection, as the SPA does.
func connectionRequired(t *testing.T, h http.Handler) bool {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/connection", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("connection: status = %d, want %d", rec.Code, http.StatusOK)
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
