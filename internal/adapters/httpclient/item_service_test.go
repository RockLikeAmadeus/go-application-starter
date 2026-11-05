package httpclient_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/memory"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/app"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/testsupport"
)

// TestRoundTrip exercises the HTTP server and client adapters together:
// client -> HTTP -> server -> core -> memory storage, and errors back.
func TestRoundTrip(t *testing.T) {
	client := newClient(t, testsupport.FixedEntitlements{})
	ctx := context.Background()

	created, err := client.CreateItem(ctx, "Widget")
	if err != nil || created.Name != "Widget" {
		t.Fatalf("CreateItem = %+v, %v", created, err)
	}

	renamed, err := client.RenameItem(ctx, created.ID, "Gadget")
	if err != nil || renamed.Name != "Gadget" {
		t.Fatalf("RenameItem = %+v, %v", renamed, err)
	}

	list, err := client.ListItems(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("ListItems = %+v, %v", list, err)
	}

	var verr *domain.ValidationError
	if _, err := client.CreateItem(ctx, " "); !errors.As(err, &verr) || verr.Field != "name" {
		t.Errorf("CreateItem(blank) err = %v, want ValidationError on name", err)
	}

	if err := client.DeleteItem(ctx, created.ID); err != nil {
		t.Fatalf("DeleteItem: %v", err)
	}
	if _, err := client.GetItem(ctx, created.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetItem after delete err = %v, want ErrNotFound", err)
	}
}

// TestExportItems_ReadsTheCSV checks that the client hands back the server's
// CSV export as it was sent.
func TestExportItems_ReadsTheCSV(t *testing.T) {
	plan := testsupport.FixedEntitlements{Features: map[string]domain.Feature{
		"item-export": {Key: "item-export", Name: "Item export", IsOn: true, WhenOff: domain.ShowLocked},
	}}
	client := newClient(t, plan)
	ctx := context.Background()
	if _, err := client.CreateItem(ctx, "Widget, large"); err != nil {
		t.Fatal(err)
	}

	got, err := client.ExportItems(ctx)

	want := "name,createdAt\n\"Widget, large\",1970-01-01T00:00:00Z\n"
	if err != nil || string(got) != want {
		t.Errorf("ExportItems = %q, %v; want %q", got, err, want)
	}
}

// newClient serves a core on memory storage, with the clock at the Unix
// epoch and the given plan, and returns a client for it.
func newClient(t *testing.T, plan testsupport.FixedEntitlements) *httpclient.ItemService {
	t.Helper()
	core := app.NewItemService(memory.NewItemRepository(), testsupport.FixedWorkspace(domain.LocalWorkspace), plan,
		testsupport.NewFixedClock(time.Unix(0, 0).UTC()), testsupport.NewSequentialIDs("item"))
	srv := httptest.NewServer(httpapi.NewHandler(httpapi.Options{Items: core}))
	t.Cleanup(srv.Close)
	client, err := httpclient.NewItemService(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
