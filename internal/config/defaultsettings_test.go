package config_test

import (
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

func TestTheDefaultSettingsFileSetsTheStorageDataFolderAndSyncedFolderSettings(t *testing.T) {
	dataDir := `C:\Users\Ann\AppData\Roaming\starter`
	file := writeSettings(t, config.DefaultSettingsFileText(dataDir))

	cfg, err := config.Load(config.Env{Getenv: envOf(nil)}, file)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	for _, want := range []config.Setting{
		{Key: "storage", Value: string(config.StorageJSONFile), Source: "settings file"},
		{Key: "data_dir", Value: dataDir, Source: "settings file"},
		{Key: "ignore_synced_folder", Value: false, Source: "settings file"},
	} {
		if !listsSetting(cfg.Settings(), want) {
			t.Errorf("Settings() = %v, want it to list %v", cfg.Settings(), want)
		}
	}
}
