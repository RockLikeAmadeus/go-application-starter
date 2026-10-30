package desktop_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/desktop"
)

func TestOpenServesAnEmptyItemListFromTheDataFolder(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	app, err := desktop.Open(envOf(map[string]string{"STARTER_DATA_DIR": dataDir}))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	rec := httptest.NewRecorder()
	app.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/items", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body struct {
		Items []json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	if len(body.Items) != 0 {
		t.Errorf("items = %s, want none", body.Items)
	}
	if _, err := os.Stat(dataDir); err != nil {
		t.Errorf("data folder not created: %v", err)
	}
}
