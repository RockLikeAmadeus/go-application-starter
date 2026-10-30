package mobile_test

import (
	"encoding/json"
	"net/http"
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
		Required *bool `json:"required"`
	}
	if err := json.Unmarshal([]byte(reply.Body), &body); reply.Status != http.StatusOK || err != nil || body.Required == nil || *body.Required {
		t.Errorf("connection over the bridge: status %d, body %q; want 200 and required false", reply.Status, reply.Body)
	}

	if reply := mobile.Serve(app.Handler, http.MethodGet, "/api/v1/connection", ""); reply.Status != http.StatusNotFound {
		t.Errorf("connection on the asset path: status %d, want 404", reply.Status)
	}
}
