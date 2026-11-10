package config_test

import (
	"errors"
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

func TestAnUnknownSettingErrorReadsAsASentence(t *testing.T) {
	file := writeSettings(t, `colour = "blue"`)

	_, err := config.Load(config.Env{Getenv: envOf(nil)}, file)

	want := "The settings file " + file + ` contains an unknown setting: "colour".`
	if err == nil || err.Error() != want {
		t.Errorf("error = %v, want %q", err, want)
	}
}

func TestAnUnreadableSettingsFileErrorReadsAsASentence(t *testing.T) {
	file := filepath.Join(t.TempDir(), "missing.toml")

	_, err := config.Load(config.Env{Getenv: envOf(nil)}, file)

	cause := errors.Unwrap(err)
	if cause == nil {
		t.Fatalf("error = %v, want one wrapping why the file could not be read", err)
	}
	want := "The settings file " + file + " could not be read: " + cause.Error()
	if err.Error() != want {
		t.Errorf("error = %v, want %q", err, want)
	}
}

func TestASettingsFileThatIsntTOMLFailsNamingTheFile(t *testing.T) {
	file := writeSettings(t, `data_dir = "/srv/work-data`)

	_, err := config.Load(config.Env{Getenv: envOf(nil)}, file)
	if err == nil || !strings.Contains(err.Error(), file) {
		t.Errorf("Load error = %v, want one naming %s", err, file)
	}
}

func TestLoadSaysWhereEachSettingCameFrom(t *testing.T) {
	file := writeSettings(t, `
storage = "memory"
data_dir = "/srv/work-data"
`)
	env := envOf(map[string]string{"STARTER_DATA_DIR": "/tmp/scratch"})

	cfg, err := config.Load(config.Env{Getenv: env}, file)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := map[string]string{
		"storage":    "settings file",
		"data_dir":   "environment (STARTER_DATA_DIR)",
		"log_format": "built in",
		"port":       "built in",
	}
	for key, source := range want {
		if got := cfg.Sources[key]; got != source {
			t.Errorf("Sources[%q] = %q, want %q", key, got, source)
		}
	}
}

func TestStorageAcceptsSQLite(t *testing.T) {
	cfg, err := config.Load(config.Env{Getenv: envOf(map[string]string{"STARTER_STORAGE": "sqlite"})}, "")
	if err != nil {
		t.Fatalf("Load with sqlite: %v", err)
	}
	if cfg.Storage != config.StorageSQLite {
		t.Errorf("Storage = %q, want %q", cfg.Storage, config.StorageSQLite)
	}
}

func TestStorageAcceptsPostgres(t *testing.T) {
	cfg, err := config.Load(config.Env{Getenv: envOf(map[string]string{"STARTER_STORAGE": "postgres"})}, "")
	if err != nil {
		t.Fatalf("Load with postgres: %v", err)
	}
	if cfg.Storage != config.StoragePostgres {
		t.Errorf("Storage = %q, want %q", cfg.Storage, config.StoragePostgres)
	}
}

func TestDatabaseURLIsASettingReadFromTheEnvironment(t *testing.T) {
	const databaseURL = "postgres://starter@db.example.com:5432/starter"
	env := envOf(map[string]string{"STARTER_DATABASE_URL": databaseURL})

	cfg, err := config.Load(config.Env{Getenv: env}, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.DatabaseURL != databaseURL {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, databaseURL)
	}
	const wantSource = "environment (STARTER_DATABASE_URL)"
	isListed := false
	for _, setting := range cfg.Settings() {
		if setting.Key == "database_url" && setting.Source == wantSource {
			isListed = true
		}
	}
	if !isListed {
		t.Errorf("Settings() = %v, want database_url from %q", cfg.Settings(), wantSource)
	}
}

func TestAnUnknownStorageErrorReadsAsASentence(t *testing.T) {
	_, err := config.Load(config.Env{Getenv: envOf(map[string]string{"STARTER_STORAGE": "disk"})}, "")

	want := `The storage setting "disk" is not recognized. Please use "memory", "jsonfile", or "sqlite".`
	if err == nil || err.Error() != want {
		t.Errorf("error = %v, want %q", err, want)
	}
}

func TestAnUnknownLogFormatErrorReadsAsASentence(t *testing.T) {
	_, err := config.Load(config.Env{Getenv: envOf(map[string]string{"STARTER_LOG_FORMAT": "xml"})}, "")

	want := `The log format setting "xml" is not recognized. Please use "text" or "json".`
	if err == nil || err.Error() != want {
		t.Errorf("error = %v, want %q", err, want)
	}
}

func TestShouldIgnoreSyncedFolderIsABoolFromTheSettingsFileOrEnvironmentDefaultingToFalse(t *testing.T) {
	cases := []struct {
		name     string
		file     string
		env      map[string]string
		want     bool
		wantFrom string
	}{
		{"built in", "", nil, false, "built in"},
		{"settings file", "ignore_synced_folder = true", nil, true, "settings file"},
		{"environment", "", map[string]string{"STARTER_IGNORE_SYNCED_FOLDER": "true"}, true, "environment (STARTER_IGNORE_SYNCED_FOLDER)"},
		{"environment over file", "ignore_synced_folder = true", map[string]string{"STARTER_IGNORE_SYNCED_FOLDER": "false"}, false, "environment (STARTER_IGNORE_SYNCED_FOLDER)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			file := ""
			if c.file != "" {
				file = writeSettings(t, c.file)
			}

			cfg, err := config.Load(config.Env{Getenv: envOf(c.env)}, file)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if cfg.ShouldIgnoreSyncedFolder != c.want {
				t.Errorf("ShouldIgnoreSyncedFolder = %v, want %v", cfg.ShouldIgnoreSyncedFolder, c.want)
			}
			if !listsSetting(cfg.Settings(), config.Setting{Key: "ignore_synced_folder", Value: c.want, Source: c.wantFrom}) {
				t.Errorf("Settings() = %v, want ignore_synced_folder = %v from %q", cfg.Settings(), c.want, c.wantFrom)
			}
		})
	}

	t.Run("not a bool", func(t *testing.T) {
		_, err := config.Load(config.Env{Getenv: envOf(map[string]string{"STARTER_IGNORE_SYNCED_FOLDER": "sometimes"})}, "")
		if err == nil || !strings.Contains(err.Error(), "STARTER_IGNORE_SYNCED_FOLDER") {
			t.Errorf("Load error = %v, want one naming %s", err, "STARTER_IGNORE_SYNCED_FOLDER")
		}
	})
}

// listsSetting reports whether settings holds want.
func listsSetting(settings []config.Setting, want config.Setting) bool {
	for _, s := range settings {
		if s == want {
			return true
		}
	}
	return false
}
