package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// The server's readiness probe asks its storage, so a server whose data
// folder has gone isn't ready.
func TestServerIsNotReadyOnceItsDataFolderIsRemoved(t *testing.T) {
	dir := t.TempDir()
	app, err := bootstrap.NewEmbedded(config.Config{Storage: config.StorageJSONFile, DataDir: dir})
	if err != nil {
		t.Fatalf("NewEmbedded: %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })
	handler := httpapi.NewHandler(handlerOptions(app, nil))
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}
