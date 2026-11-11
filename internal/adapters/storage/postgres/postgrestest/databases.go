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
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" driver
)

// ServerURLVariable names the setting that points the tests at a server.
const ServerURLVariable = "STARTER_TEST_DATABASE_URL"

// NewDatabase creates an empty database with a random name, drops it when the
// test ends, and returns its URL. It skips the test if the server URL is not
// set, and fails it if the server cannot be reached.
func NewDatabase(t testing.TB) string {
	t.Helper()
	serverURL := os.Getenv(ServerURLVariable)
	if serverURL == "" {
		t.Skipf("%s is not set: skipping this Postgres test (task db:up starts a server)", ServerURLVariable)
	}
	server := openServer(t, serverURL)
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

// BlockConnections makes the database at databaseURL, from NewDatabase,
// unreachable without stopping the server: it refuses new connections and ends
// the ones already open.
func BlockConnections(t testing.TB, databaseURL string) {
	t.Helper()
	server, name := openServerOf(t, databaseURL)
	defer server.Close()
	if _, err := server.ExecContext(t.Context(), "ALTER DATABASE "+name+" ALLOW_CONNECTIONS false"); err != nil {
		t.Fatalf("block connections: %v", err)
	}
	// The timeout makes each call wait until the connection has really ended.
	const endConnections = "SELECT pg_terminate_backend(pid, 5000) FROM pg_stat_activity WHERE datname = $1"
	if _, err := server.ExecContext(t.Context(), endConnections, name); err != nil {
		t.Fatalf("end open connections: %v", err)
	}
}

// AllowConnections undoes BlockConnections.
func AllowConnections(t testing.TB, databaseURL string) {
	t.Helper()
	server, name := openServerOf(t, databaseURL)
	defer server.Close()
	if _, err := server.ExecContext(t.Context(), "ALTER DATABASE "+name+" ALLOW_CONNECTIONS true"); err != nil {
		t.Fatalf("allow connections: %v", err)
	}
}

// openServerOf connects to the test server, and returns the connection with
// the name of the database that databaseURL points at.
func openServerOf(t testing.TB, databaseURL string) (*sql.DB, string) {
	t.Helper()
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("database URL: %v", err)
	}
	return openServer(t, os.Getenv(ServerURLVariable)), strings.TrimPrefix(parsed.Path, "/")
}

// openServer connects to the test server at serverURL.
func openServer(t testing.TB, serverURL string) *sql.DB {
	t.Helper()
	server, err := sql.Open("pgx", serverURL)
	if err != nil {
		t.Fatalf("open the test server: %v", err)
	}
	return server
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
		t.Fatalf("%s is not a URL: %v", ServerURLVariable, err)
	}
	parsed.Path = "/" + name
	return parsed.String()
}
