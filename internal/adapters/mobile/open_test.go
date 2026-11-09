package mobile_test

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/mobile"
)

// The phone's data is on the phone, so the web app must not wait for a
// connection. It asks over the bridge (API); the asset path (Handler) must not
// answer API calls at all.
func TestOpenAnswersTheAPIOverTheBridgeOnly(t *testing.T) {
	app, err := mobile.Open(func(string) string { return "" }, t.TempDir(), nil)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })
	if app.API == nil || app.Handler == nil {
		t.Fatalf("Open returned API %v and Handler %v, want both", app.API, app.Handler)
	}

	reply := mobile.Serve(app.API, http.MethodGet, "/api/v1/connection", "")
	var body struct {
		IsRequired *bool `json:"isRequired"`
	}
	if err := json.Unmarshal([]byte(reply.Body), &body); reply.Status != http.StatusOK || err != nil || body.IsRequired == nil || *body.IsRequired {
		t.Errorf("connection over the bridge: status %d, body %q; want 200 and required false", reply.Status, reply.Body)
	}

	if reply := mobile.Serve(app.Handler, http.MethodGet, "/api/v1/connection", ""); reply.Status != http.StatusNotFound {
		t.Errorf("connection on the asset path: status %d, want 404", reply.Status)
	}
}

// The phone has no settings folder the user can reach, so a settings file
// there would be one nobody can edit: Open creates none.
func TestOpenCreatesNoSettingsFile(t *testing.T) {
	storage := t.TempDir()

	app, err := mobile.Open(func(string) string { return "" }, storage, nil)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	settingsFile := filepath.Join(storage, "starter", "config.toml")
	if _, err := os.Stat(settingsFile); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Open left a settings file at %s (stat: %v), want none", settingsFile, err)
	}
}
