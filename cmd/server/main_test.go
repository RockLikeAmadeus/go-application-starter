package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
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

func TestTheServerReadsTheSettingsFileStarterConfigFileNames(t *testing.T) {
	file := filepath.Join(t.TempDir(), "server.toml")
	if err := os.WriteFile(file, []byte("port = 9090\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{"STARTER_CONFIG_FILE": file}

	cfg, err := loadConfig(config.Env{Getenv: func(key string) string { return env[key] }})
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.Port != 9090 {
		t.Errorf("Port = %d, want the settings file's 9090", cfg.Port)
	}
}

// The server's web app shows the app's warnings, so the handler serves them.
// A server that can't start says why on stderr, and ends with a failing exit
// status.
func TestAServerThatCannotStartPrintsItsErrorOnStderr(t *testing.T) {
	env := map[string]string{"STARTER_STORAGE": "disk"}
	var errOut bytes.Buffer

	status := runAndReport(config.Env{Getenv: func(key string) string { return env[key] }}, &errOut)

	if status != 1 {
		t.Errorf("exit status = %d, want 1", status)
	}
	want := `Error: The storage setting "disk" is not recognized. Please use "memory", "jsonfile", or "sqlite".` + "\n"
	if got := errOut.String(); got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestTheServerServesTheAppsWarnings(t *testing.T) {
	want := []string{"first warning", "second warning"}
	handler := httpapi.NewHandler(handlerOptions(&bootstrap.App{Warnings: want}, nil))

	got := servedWarnings(t, handler)

	if !slices.Equal(got, want) {
		t.Errorf("warnings = %q, want %q", got, want)
	}
}

// Whoever runs the server reads its log, so it logs each warning at start.
func TestTheServerLogsEachWarning(t *testing.T) {
	var out bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&out, nil))

	logWarnings(log, []string{"first warning", "second warning"})

	var got []string
	dec := json.NewDecoder(&out)
	for dec.More() {
		var record struct {
			Level string `json:"level"`
			Msg   string `json:"msg"`
		}
		if err := dec.Decode(&record); err != nil {
			t.Fatalf("decode: %v", err)
		}
		got = append(got, record.Level+" "+record.Msg)
	}
	want := []string{"WARN first warning", "WARN second warning"}
	if !slices.Equal(got, want) {
		t.Errorf("logged %q, want %q", got, want)
	}
}

// servedWarnings asks the REST API for the warnings' messages, as the web
// app does.
func servedWarnings(t *testing.T, h http.Handler) []string {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/warnings", nil))
	var body struct {
		Warnings []struct {
			Message string `json:"message"`
		} `json:"warnings"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	messages := make([]string, 0, len(body.Warnings))
	for _, w := range body.Warnings {
		messages = append(messages, w.Message)
	}
	return messages
}
