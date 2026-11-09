// Package postgrestest gives tests a database of their own on the Postgres
// server that STARTER_TEST_DATABASE_URL names (task db:up starts one).
package postgrestest

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/url"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" driver
)

// serverURLVariable names the setting that points the tests at a server.
const serverURLVariable = "STARTER_TEST_DATABASE_URL"

// NewDatabase creates an empty database with a random name, drops it when the
// test ends, and returns its URL. It skips the test if the server URL is not
// set, and fails it if the server cannot be reached.
func NewDatabase(t testing.TB) string {
	t.Helper()
	serverURL := os.Getenv(serverURLVariable)
	if serverURL == "" {
		t.Skipf("%s is not set: skipping this Postgres test (task db:up starts a server)", serverURLVariable)
	}
	server, err := sql.Open("pgx", serverURL)
	if err != nil {
		t.Fatalf("open the test server: %v", err)
	}
	name := randomName(t)
	if _, err := server.ExecContext(t.Context(), "CREATE DATABASE "+name); err != nil {
		server.Close()
		t.Fatalf("create test database: %v", err)
	}
	t.Cleanup(func() {
		defer server.Close()
		// t.Context is already cancelled here. FORCE ends any connection the
		// test left open.
		if _, err := server.ExecContext(context.Background(), "DROP DATABASE "+name+" WITH (FORCE)"); err != nil {
			t.Errorf("drop test database: %v", err)
		}
	})
	return databaseURL(t, serverURL, name)
}

// randomName is a database name no other test run is using. It needs no
// quoting in SQL.
func randomName(t testing.TB) string {
	t.Helper()
	suffix := make([]byte, 8)
	if _, err := rand.Read(suffix); err != nil {
		t.Fatalf("random database name: %v", err)
	}
	return "test_" + hex.EncodeToString(suffix)
}

// databaseURL is serverURL pointed at the named database.
func databaseURL(t testing.TB, serverURL, name string) string {
	t.Helper()
	parsed, err := url.Parse(serverURL)
	if err != nil {
		t.Fatalf("%s is not a URL: %v", serverURLVariable, err)
	}
	parsed.Path = "/" + name
	return parsed.String()
}
