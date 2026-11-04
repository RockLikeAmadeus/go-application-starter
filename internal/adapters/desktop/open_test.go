package desktop_test

import (
	"bytes"
	"encoding/json"
	"image/png"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// The desktop app serves the server's readiness probe too, and it asks the
// data folder, so it isn't ready once that folder has gone.
func TestOpenIsNotReadyOnceTheDataFolderIsRemoved(t *testing.T) {
	dir := t.TempDir()
	app := openAt(t, dir)
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	app.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("GET /readyz: status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

// The desktop app's icon is the web app's, which `pnpm icons` makes from
// web/public/icon.svg. Opening on web/public itself checks that the icon Open
// looks for is one that command makes.
func TestOpenReturnsTheProductsIcon(t *testing.T) {
	app := openWith(t, t.TempDir(), os.DirFS("../../../web/public"))

	icon, err := png.Decode(bytes.NewReader(app.Icon))
	if err != nil {
		t.Fatalf("Icon is not a PNG: %v", err)
	}
	if size := icon.Bounds().Size(); size.X != 512 || size.Y != 512 {
		t.Errorf("Icon is %v, want 512×512", size)
	}
}

// A missing icon only costs the window its icon, so the app still opens.
func TestOpenWithoutAnIconStillOpens(t *testing.T) {
	tests := map[string]fs.FS{
		"no web app":       nil,
		"web app, no icon": fstest.MapFS{"index.html": {Data: []byte("<title>Starter</title>")}},
	}
	for name, webApp := range tests {
		t.Run(name, func(t *testing.T) {
			app := openWith(t, t.TempDir(), webApp)

			if app.Icon != nil {
				t.Errorf("Icon = %d bytes, want none", len(app.Icon))
			}
		})
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

func TestOpenUsesTheSettingsFilesDataFolder(t *testing.T) {
	settings := t.TempDir()
	data := filepath.Join(t.TempDir(), "work-data")
	writeSettings(t, settings, "data_dir = "+strconv.Quote(data))

	app, err := desktop.Open(envOf(nil), configDirAt(settings), nil)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	if _, err := os.Stat(filepath.Join(data, ".lock")); err != nil {
		t.Errorf("the settings file's data folder %s isn't in use: %v", data, err)
	}
}

// SQLite has no lock, so with it the desktop app can be opened while another
// program (here, a second copy) has the same data folder open. On plain text
// the second Open fails with "data directory is locked".
func TestOpenOnTheSettingsFilesSQLiteCanBeOpenedTwiceOnOneDataFolder(t *testing.T) {
	settings := t.TempDir()
	writeSettings(t, settings, `storage = "sqlite"`)
	env := envOf(map[string]string{"STARTER_DATA_DIR": t.TempDir()})

	first, err := desktop.Open(env, configDirAt(settings), nil)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	t.Cleanup(func() { _ = first.Close() })
	second, err := desktop.Open(env, configDirAt(settings), nil)
	if err != nil {
		t.Fatalf("second Open on the same data folder: %v", err)
	}
	t.Cleanup(func() { _ = second.Close() })
}

func TestOpenUsesTheSettingsFilesPlansFile(t *testing.T) {
	settings := t.TempDir()
	plans := filepath.Join(t.TempDir(), "plans.jsonc")
	if err := os.WriteFile(plans, []byte(`{"plans": `), 0o600); err != nil {
		t.Fatal(err)
	}
	writeSettings(t, settings, "plans_file = "+strconv.Quote(plans))

	app, err := desktop.Open(envOf(nil), configDirAt(settings), nil)
	if err == nil {
		_ = app.Close()
		t.Fatal("Open succeeded, want it to refuse the settings file's broken plans file")
	}
	if !strings.Contains(err.Error(), "plans file") {
		t.Errorf("Open error = %v, want it to be about the plans file", err)
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
	var body struct {
		Required *bool `json:"required"`
	}
	getJSON(t, h, "/api/v1/connection", &body)
	if body.Required == nil {
		t.Fatal("connection: the answer has no required field")
	}
	return *body.Required
}

// getJSON asks the REST API for path, as the SPA does, and decodes its
// answer into body, failing the test unless it is a 200.
func getJSON(t *testing.T, h http.Handler, path string, body any) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: status = %d, want %d", path, rec.Code, http.StatusOK)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), body); err != nil {
		t.Fatalf("GET %s: decode %q: %v", path, rec.Body.String(), err)
	}
}

// The desktop app is the local edition, so its web app sees the local plan.
func TestOpenServesTheLocalPlan(t *testing.T) {
	app := openAt(t, t.TempDir())

	var body struct {
		Plan string `json:"plan"`
	}
	getJSON(t, app.Handler, "/api/v1/entitlements", &body)
	if body.Plan != "local" {
		t.Errorf("plan = %q, want %q", body.Plan, "local")
	}
}

// The core's warnings reach the desktop app's window through the REST API,
// as the server's do.
func TestOpenServesTheCoresWarnings(t *testing.T) {
	synced := t.TempDir()
	if err := os.WriteFile(filepath.Join(synced, ".dropbox"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	settings := t.TempDir()
	writeSettings(t, settings, `storage = "sqlite"`)
	env := envOf(map[string]string{"STARTER_DATA_DIR": filepath.Join(synced, "data")})
	app, err := desktop.Open(env, configDirAt(settings), nil)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	got := servedWarnings(t, app.Handler)
	if len(got) != 1 || !strings.Contains(got[0], "Dropbox") {
		t.Errorf("warnings = %q, want one naming Dropbox", got)
	}
}

// servedWarnings asks the REST API for the warnings' messages, as the web
// app does.
func servedWarnings(t *testing.T, h http.Handler) []string {
	t.Helper()
	var body struct {
		Warnings []struct {
			Message string `json:"message"`
		} `json:"warnings"`
	}
	getJSON(t, h, "/api/v1/warnings", &body)
	messages := make([]string, 0, len(body.Warnings))
	for _, w := range body.Warnings {
		messages = append(messages, w.Message)
	}
	return messages
}
