package acceptance

import "testing"

func TestServersSharingAPostgresDatabaseSeeEachOthersItems(t *testing.T) {
	h, err := startServer(t)
	if err != nil {
		t.Fatalf("start the server: %v", err)
	}
	t.Cleanup(h.cleanup)
	w := &world{harness: h}
	ctx := t.Context()

	if err := w.twoServersKeepTheirDataInTheSamePostgresDatabase(); err != nil {
		t.Fatalf("start two servers on one database: %v", err)
	}
	if err := w.anItemIsAddedThroughTheFirstServer(ctx, "Widget"); err != nil {
		t.Fatalf("add an item through the first server: %v", err)
	}
	if err := w.theSecondServerLists(ctx, "Widget"); err != nil {
		t.Errorf("with the item added through the first server: %v", err)
	}
	if err := w.theSecondServerLists(ctx, "Gadget"); err == nil {
		t.Error(`the second server lists "Gadget" = nil error, want an error for an item never added`)
	}
}
