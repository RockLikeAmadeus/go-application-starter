package acceptance

import "testing"

func TestServerReadinessFollowsItsPostgresDatabase(t *testing.T) {
	h, err := startServer(t)
	if err != nil {
		t.Fatalf("start the server: %v", err)
	}
	t.Cleanup(h.cleanup)
	w := &world{harness: h}

	if err := w.theServerKeepsItsDataInPostgres(); err != nil {
		t.Fatalf("keep the data in Postgres: %v", err)
	}
	if err := w.theServerSaysItIsReady(); err != nil {
		t.Errorf("with the database reachable: %v", err)
	}

	if err := w.theServersDatabaseCannotBeReached(); err != nil {
		t.Fatalf("make the database unreachable: %v", err)
	}
	if err := w.theServerSaysItIsNotReady(); err != nil {
		t.Errorf("with the database unreachable: %v", err)
	}

	if err := w.theServersDatabaseCanBeReachedAgain(); err != nil {
		t.Fatalf("make the database reachable again: %v", err)
	}
	if err := w.theServerSaysItIsReady(); err != nil {
		t.Errorf("with the database reachable again: %v", err)
	}
}
