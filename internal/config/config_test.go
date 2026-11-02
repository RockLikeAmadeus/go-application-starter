package config_test

import (
	"os"
	"path/filepath"
	"strings"
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

func TestTheEnvironmentOverridesTheSettingsFile(t *testing.T) {
	file := writeSettings(t, `data_dir = "/srv/work-data"`)
	env := envOf(map[string]string{"STARTER_DATA_DIR": "/tmp/scratch"})

	cfg, err := config.Load(config.Env{Getenv: env}, file)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DataDir != "/tmp/scratch" {
		t.Errorf("DataDir = %q, want the environment's %q", cfg.DataDir, "/tmp/scratch")
	}
}

func TestASettingTheAppDoesntKnowFailsNamingTheFileAndSetting(t *testing.T) {
	file := writeSettings(t, `colour = "blue"`)

	_, err := config.Load(config.Env{Getenv: envOf(nil)}, file)
	if err == nil || !strings.Contains(err.Error(), file) || !strings.Contains(err.Error(), "colour") {
		t.Errorf("Load error = %v, want one naming %s and %q", err, file, "colour")
	}
}

func TestASettingsFileThatIsntTOMLFailsNamingTheFile(t *testing.T) {
	file := writeSettings(t, `data_dir = "/srv/work-data`)

	_, err := config.Load(config.Env{Getenv: envOf(nil)}, file)
	if err == nil || !strings.Contains(err.Error(), file) {
		t.Errorf("Load error = %v, want one naming %s", err, file)
	}
}
