// Package config loads typed runtime configuration from the environment.
// Flags (CLI) may override these values after loading.
package config

import (
	"fmt"
	"strconv"
)

type Storage string

const (
	StorageMemory   Storage = "memory"
	StorageJSONFile Storage = "jsonfile"
)

type Config struct {
	// Storage selects the ItemRepository adapter (STARTER_STORAGE).
	Storage Storage
	// DataDir is where the jsonfile adapter keeps its files (STARTER_DATA_DIR).
	DataDir string
	// Port is the HTTP listen port (PORT, the de-facto cloud convention).
	Port int
	// ServerURL, when set, makes CLI/TUI use a remote server (STARTER_SERVER_URL).
	ServerURL string
	// LogFormat is "json" or "text" (STARTER_LOG_FORMAT).
	LogFormat string
	// PlansFile replaces the plans file built into the binary
	// (STARTER_PLANS_FILE). Empty means the built-in one.
	PlansFile string
}

// Env is the computer the app runs on, as configuration sees it. Getenv
// reads the environment (os.Getenv in production).
type Env struct {
	Getenv func(string) string
}

// Load reads the configuration from env's STARTER_* variables and PORT,
// using the defaults for any that aren't set, and validates it.
func Load(env Env) (Config, error) {
	cfg := Config{
		Storage:   Storage(getenv(env.Getenv, "STARTER_STORAGE", string(StorageJSONFile))),
		DataDir:   getenv(env.Getenv, "STARTER_DATA_DIR", "./data"),
		ServerURL: env.Getenv("STARTER_SERVER_URL"),
		LogFormat: getenv(env.Getenv, "STARTER_LOG_FORMAT", "text"),
		PlansFile: env.Getenv("STARTER_PLANS_FILE"),
	}
	port, err := strconv.Atoi(getenv(env.Getenv, "PORT", "8080"))
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
