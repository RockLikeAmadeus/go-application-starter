package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

func TestLoadReadsTheEnvironmentItIsGiven(t *testing.T) {
	cfg, err := config.Load(config.Env{Getenv: envOf(map[string]string{"STARTER_STORAGE": "memory"})}, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Storage != config.StorageMemory {
		t.Errorf("Storage = %q, want %q", cfg.Storage, config.StorageMemory)
	}
}

func TestLoadUsesTheSettingsFile(t *testing.T) {
	file := writeSettings(t, `
storage = "memory"
data_dir = "/srv/work-data"
`)

	cfg, err := config.Load(config.Env{Getenv: envOf(nil)}, file)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Storage != config.StorageMemory || cfg.DataDir != "/srv/work-data" {
		t.Errorf("Storage, DataDir = %q, %q, want %q, %q", cfg.Storage, cfg.DataDir, config.StorageMemory, "/srv/work-data")
	}
}

// writeSettings writes a settings file with content, and returns its path.
func writeSettings(t *testing.T, content string) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return file
}
