// Package config loads typed runtime configuration from an optional settings
// file (TOML) and the environment, which overrides it. Flags (CLI) may
// override these values after loading.
package config

import (
	"fmt"
	"strconv"

	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type StorageBackend string

const (
	StorageMemory   StorageBackend = "memory"
	StorageJSONFile StorageBackend = "jsonfile"
	StorageSQLite   StorageBackend = "sqlite"
)

type Config struct {
	// Storage selects the storage adapter (storage; STARTER_STORAGE).
	Storage StorageBackend
	// DataDir is where the jsonfile adapter keeps its files (data_dir;
	// STARTER_DATA_DIR).
	DataDir string
	// Port is the HTTP listen port (port; PORT, the de-facto cloud convention).
	Port int
	// ServerURL, when set, makes CLI/TUI use a remote server (server_url;
	// STARTER_SERVER_URL).
	ServerURL string
	// LogFormat is "json" or "text" (log_format; STARTER_LOG_FORMAT).
	LogFormat string
	// PlansFile replaces the plans file built into the binary (plans_file;
	// STARTER_PLANS_FILE). Empty means the built-in one.
	PlansFile string
	// ShouldIgnoreSyncedFolder silences the warning that SQLite's data folder
	// looks synced (ignore_synced_folder; STARTER_IGNORE_SYNCED_FOLDER).
	ShouldIgnoreSyncedFolder bool
	// SettingsFile is the settings file Load read, or "" for none.
	SettingsFile string
	// Sources says where each setting came from, by its key in the settings
	// file.
	Sources map[string]string
}

// Env is the computer the app runs on, as configuration sees it. Getenv
// reads the environment (os.Getenv in production) and UserConfigDir finds
// the user's settings folder (os.UserConfigDir in production).
type Env struct {
	Getenv        func(string) string
	UserConfigDir func() (string, error)
}

// Load reads the configuration and validates it. Each setting comes from
// env's STARTER_* variable (or PORT) when set, else from its key in the TOML
// settingsFile, else from its default, and Sources records which.
// settingsFile "" means none.
func Load(env Env, settingsFile string) (Config, error) {
	fileSettings, err := readSettingsFile(settingsFile)
	if err != nil {
		return Config{}, err
	}
	sources := map[string]string{}
	setting := func(key, fallback string) string {
		if value := env.Getenv(envVars[key]); value != "" {
			sources[key] = "environment (" + envVars[key] + ")"
			return value
		}
		if value := fileSettings.String(key); value != "" {
			sources[key] = "settings file"
			return value
		}
		sources[key] = "built in"
		return fallback
	}
	cfg := Config{
		SettingsFile: settingsFile,
		Sources:      sources,
		Storage:      StorageBackend(setting("storage", string(StorageJSONFile))),
		DataDir:      setting("data_dir", "./data"),
		ServerURL:    setting("server_url", ""),
		LogFormat:    setting("log_format", "text"),
		PlansFile:    setting("plans_file", ""),
	}
	port, err := strconv.Atoi(setting("port", "8080"))
	if err != nil {
		return Config{}, fmt.Errorf("PORT: %w", err)
	}
	cfg.Port = port
	ignore, err := strconv.ParseBool(setting("ignore_synced_folder", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("STARTER_IGNORE_SYNCED_FOLDER: %w", err)
	}
	cfg.ShouldIgnoreSyncedFolder = ignore
	return cfg, cfg.Validate()
}

// Setting is one setting's key in the settings file, its value (a string, an
// int or a bool) and where it came from.
type Setting struct {
	Key    string
	Value  any
	Source string
}

// Settings lists c's settings, in the order a settings file would.
func (c Config) Settings() []Setting {
	values := []struct {
		key   string
		value any
	}{
		{"storage", string(c.Storage)},
		{"data_dir", c.DataDir},
		{"port", c.Port},
		{"server_url", c.ServerURL},
		{"log_format", c.LogFormat},
		{"plans_file", c.PlansFile},
		{"ignore_synced_folder", c.ShouldIgnoreSyncedFolder},
	}
	settings := make([]Setting, len(values))
	for i, v := range values {
		settings[i] = Setting{Key: v.key, Value: v.value, Source: c.Sources[v.key]}
	}
	return settings
}

func (c Config) Validate() error {
	switch c.Storage {
	case StorageMemory, StorageJSONFile, StorageSQLite:
	default:
		//nolint:staticcheck // ST1005: users read it as written, as a full sentence.
		return fmt.Errorf("The storage setting %q is not recognized. Please use %q, %q, or %q.", c.Storage, StorageMemory, StorageJSONFile, StorageSQLite)
	}
	if c.LogFormat != "json" && c.LogFormat != "text" {
		//nolint:staticcheck // ST1005: users read it as written, as a full sentence.
		return fmt.Errorf("The log format setting %q is not recognized. Please use \"text\" or \"json\".", c.LogFormat)
	}
	return nil
}

// envVars are the settings, by their key in a settings file, each with the
// environment variable that overrides it.
var envVars = map[string]string{
	"storage":              "STARTER_STORAGE",
	"data_dir":             "STARTER_DATA_DIR",
	"port":                 "PORT",
	"server_url":           "STARTER_SERVER_URL",
	"log_format":           "STARTER_LOG_FORMAT",
	"plans_file":           "STARTER_PLANS_FILE",
	"ignore_synced_folder": "STARTER_IGNORE_SYNCED_FOLDER",
}

// readSettingsFile reads the TOML settings file at path, refusing any key
// that isn't a setting. An empty path reads nothing.
func readSettingsFile(path string) (*koanf.Koanf, error) {
	fileSettings := koanf.New(".")
	if path == "" {
		return fileSettings, nil
	}
	if err := fileSettings.Load(file.Provider(path), toml.Parser()); err != nil {
		//nolint:staticcheck // ST1005: users read it as written, as a full sentence.
		return nil, fmt.Errorf("The settings file %s could not be read: %w", path, err)
	}
	for _, key := range fileSettings.Keys() {
		if _, ok := envVars[key]; !ok {
			//nolint:staticcheck // ST1005: users read it as written, as a full sentence.
			return nil, fmt.Errorf("The settings file %s contains an unknown setting: %q.", path, key)
		}
	}
	return fileSettings, nil
}
