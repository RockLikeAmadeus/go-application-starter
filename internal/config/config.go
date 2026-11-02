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
}

// Env is the computer the app runs on, as configuration sees it. Getenv
// reads the environment (os.Getenv in production).
type Env struct {
	Getenv func(string) string
}

// Load reads the configuration and validates it. Each setting comes from
// env's STARTER_* variable (or PORT) when set, else from its key in the TOML
// settingsFile, else from its default. settingsFile "" means none.
func Load(env Env, settingsFile string) (Config, error) {
	k := koanf.New(".")
	if settingsFile != "" {
		if err := k.Load(file.Provider(settingsFile), toml.Parser()); err != nil {
			return Config{}, fmt.Errorf("read the settings file %s: %w", settingsFile, err)
		}
	}
	setting := func(envKey, key, fallback string) string {
		return getenv(env.Getenv, envKey, getenv(k.String, key, fallback))
	}
	cfg := Config{
		Storage:   Storage(setting("STARTER_STORAGE", "storage", string(StorageJSONFile))),
		DataDir:   setting("STARTER_DATA_DIR", "data_dir", "./data"),
		ServerURL: setting("STARTER_SERVER_URL", "server_url", ""),
		LogFormat: setting("STARTER_LOG_FORMAT", "log_format", "text"),
		PlansFile: setting("STARTER_PLANS_FILE", "plans_file", ""),
	}
	port, err := strconv.Atoi(setting("PORT", "port", "8080"))
	if err != nil {
		return Config{}, fmt.Errorf("PORT: %w", err)
	}
	cfg.Port = port
	return cfg, cfg.Validate()
}

func (c Config) Validate() error {
	switch c.Storage {
	case StorageMemory, StorageJSONFile:
	default:
		return fmt.Errorf("unknown storage %q (want %q or %q)", c.Storage, StorageMemory, StorageJSONFile)
	}
	if c.LogFormat != "json" && c.LogFormat != "text" {
		return fmt.Errorf("unknown log format %q", c.LogFormat)
	}
	return nil
}

// getenv returns lookup(key), or fallback when that's empty.
func getenv(lookup func(string) string, key, fallback string) string {
	if v := lookup(key); v != "" {
		return v
	}
	return fallback
}
