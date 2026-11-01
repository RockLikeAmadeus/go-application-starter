// Package config loads typed runtime configuration from the environment.
// Flags (CLI) may override these values after loading.
package config

import (
	"fmt"
	"os"
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

func Load() (Config, error) {
	cfg := Config{
		Storage:   Storage(getenv("STARTER_STORAGE", string(StorageJSONFile))),
		DataDir:   getenv("STARTER_DATA_DIR", "./data"),
		ServerURL: os.Getenv("STARTER_SERVER_URL"),
		LogFormat: getenv("STARTER_LOG_FORMAT", "text"),
		PlansFile: os.Getenv("STARTER_PLANS_FILE"),
	}
	port, err := strconv.Atoi(getenv("PORT", "8080"))
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

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
