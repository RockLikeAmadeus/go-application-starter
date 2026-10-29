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
	core := app.NewItemService(memory.NewItemRepository(),
		testsupport.NewFixedClock(time.Unix(0, 0).UTC()), testsupport.NewSequentialIDs("item"))
	srv := httptest.NewServer(httpapi.NewHandler(httpapi.Options{Items: core}))
	t.Cleanup(srv.Close)

	client, err := httpclient.New(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
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
