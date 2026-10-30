package desktop_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

	if names := itemNames(t, app.Handler); len(names) != 0 {
		t.Errorf("items = %q, want none", names)
	}
	if _, err := os.Stat(dataDir); err != nil {
		t.Errorf("data folder not created: %v", err)
	}
}

func TestOpenKeepsItemsAfterClosingAndReopening(t *testing.T) {
	env := envOf(map[string]string{"STARTER_DATA_DIR": t.TempDir()})

	first, err := desktop.Open(env)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	createItem(t, first.Handler, "Widget")
	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	second, err := desktop.Open(env)
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
