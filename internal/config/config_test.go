package config_test

import (
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

func TestLoadReadsTheEnvironmentItIsGiven(t *testing.T) {
	cfg, err := config.Load(config.Env{Getenv: envOf(map[string]string{"STARTER_STORAGE": "memory"})})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Storage != config.StorageMemory {
		t.Errorf("Storage = %q, want %q", cfg.Storage, config.StorageMemory)
	}
}
