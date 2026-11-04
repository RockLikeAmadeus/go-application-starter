package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// exportingItems is a ports.ItemService that exports csv and finds no item
// by ID. Its other methods aren't implemented.
type exportingItems struct {
	ports.ItemService
	csv []byte
}

func (s exportingItems) ExportItems(context.Context) ([]byte, error) { return s.csv, nil }

func (exportingItems) GetItem(context.Context, string) (domain.Item, error) {
	return domain.Item{}, domain.ErrNotFound
}

// The export is served as the core encodes it, as CSV, and "export" is never
// read as an item's ID.
func TestExportServesTheItemsAsCSV(t *testing.T) {
	csv := "name,createdAt\nfirst,2026-11-04T09:00:00Z\n"
	rec := httptest.NewRecorder()
	httpapi.NewHandler(httpapi.Options{Items: exportingItems{csv: []byte(csv)}}).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/items/export", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body %s)", rec.Code, http.StatusOK, rec.Body)
	}
	if got, want := rec.Header().Get("Content-Type"), "text/csv"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}
	if got := rec.Body.String(); got != csv {
		t.Errorf("body = %q, want %q", got, csv)
	}
}
