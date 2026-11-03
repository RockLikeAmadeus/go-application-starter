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

type Storage string

const (
	StorageMemory   Storage = "memory"
	StorageJSONFile Storage = "jsonfile"
	StorageSQLite   Storage = "sqlite"
)

type Config struct {
	// Storage selects the storage adapter (storage; STARTER_STORAGE).
	Storage Storage
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
	k, err := readSettingsFile(settingsFile)
	if err != nil {
		return Config{}, err
	}
	sources := map[string]string{}
	setting := func(key, fallback string) string {
		if v := env.Getenv(envVars[key]); v != "" {
			sources[key] = "environment (" + envVars[key] + ")"
			return v
		}
		if v := k.String(key); v != "" {
			sources[key] = "settings file"
			return v
		}
		sources[key] = "built in"
		return fallback
	}
	cfg := Config{
		SettingsFile: settingsFile,
		Sources:      sources,
		Storage:      Storage(setting("storage", string(StorageJSONFile))),
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
	return cfg, cfg.Validate()
}

// Setting is one setting's key in the settings file, its value (a string or
// an int) and where it came from.
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
		return fmt.Errorf("unknown storage %q (want %q, %q or %q)", c.Storage, StorageMemory, StorageJSONFile, StorageSQLite)
	}
	if c.LogFormat != "json" && c.LogFormat != "text" {
		return fmt.Errorf("unknown log format %q", c.LogFormat)
	}
	return nil
}

// envVars are the settings, by their key in a settings file, each with the
// environment variable that overrides it.
var envVars = map[string]string{
	"storage":    "STARTER_STORAGE",
	"data_dir":   "STARTER_DATA_DIR",
	"port":       "PORT",
	"server_url": "STARTER_SERVER_URL",
	"log_format": "STARTER_LOG_FORMAT",
	"plans_file": "STARTER_PLANS_FILE",
}

// readSettingsFile reads the TOML settings file at path, refusing any key
// that isn't a setting. An empty path reads nothing.
func readSettingsFile(path string) (*koanf.Koanf, error) {
	k := koanf.New(".")
	if path == "" {
		return k, nil
	}
	if err := k.Load(file.Provider(path), toml.Parser()); err != nil {
		return nil, fmt.Errorf("read the settings file %s: %w", path, err)
	}
	for _, key := range k.Keys() {
		if _, ok := envVars[key]; !ok {
			return nil, fmt.Errorf("settings file %s: unknown setting %q", path, key)
		}
	}
	return k, nil
}
