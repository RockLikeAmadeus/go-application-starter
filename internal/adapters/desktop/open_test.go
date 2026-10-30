package desktop_test

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/desktop"
)

func TestOpenServesAnEmptyItemListFromTheDataFolder(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	app := openAt(t, dataDir)

	if names := itemNames(t, app.Handler); len(names) != 0 {
		t.Errorf("items = %q, want none", names)
	}
	if _, err := os.Stat(dataDir); err != nil {
		t.Errorf("data folder not created: %v", err)
	}
}

func TestOpenKeepsItemsAfterClosingAndReopening(t *testing.T) {
	env := envOf(map[string]string{"STARTER_DATA_DIR": t.TempDir()})

	first, err := desktop.Open(env, nil)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	createItem(t, first.Handler, "Widget")
	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	second, err := desktop.Open(env, nil)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	t.Cleanup(func() { _ = second.Close() })

	if names := itemNames(t, second.Handler); len(names) != 1 || names[0] != "Widget" {
		t.Errorf("items after reopening = %q, want just Widget", names)
	}
}

// createItem creates an item through the REST API, as the SPA does.
func createItem(t *testing.T, h http.Handler, name string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"name": name})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/items", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create %q: status = %d, body %s", name, rec.Code, rec.Body)
	}
}

// itemNames lists the items through the REST API, as the SPA does.
func itemNames(t *testing.T, h http.Handler) []string {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/items", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("list: status = %d, want %d", rec.Code, http.StatusOK)
	}
	var list struct {
		Items []struct {
			Name string `json:"name"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	names := []string{}
	for _, item := range list.Items {
		names = append(names, item.Name)
	}
	return names
}

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
	app, err := desktop.Open(envOf(map[string]string{"STARTER_DATA_DIR": dataDir}), webApp)
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
