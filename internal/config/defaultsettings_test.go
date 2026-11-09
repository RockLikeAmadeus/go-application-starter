package config_test

import (
	"regexp"
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

func TestTheDefaultSettingsFileDescribesEachSettingAndLeavesOutTheRest(t *testing.T) {
	text := config.DefaultSettingsFileText("/home/ann/.config/starter")

	want := `# The settings for this application. The application created this file with the
# values it uses when no settings file exists.

# How the application keeps its data: "jsonfile" (plain-text storage, which
# one program on this computer can use at a time), "sqlite" (a database,
# which several programs can use at once), or "memory" (kept only until the
# program ends).
storage = "jsonfile"

# The folder in which the application keeps its data. The desktop
# application, the command line, and the terminal application share it.
data_dir = "/home/ann/.config/starter"

# Set this to true if the data folder appears to be synced by a file sync
# service but is excluded from sync. The application then does not warn that
# SQLite data in the folder may be damaged, and plain-text storage does not
# mark the folder as in use on this computer.
ignore_synced_folder = false

# To make the command line and the terminal application use a server instead
# of the data on this computer, please remove the "#" below and give the
# server's address.
# server_url = "https://example.com"

# To use a plans file of your own instead of the one built into the
# application, please remove the "#" below and give the file's full path.
# plans_file = "/path/to/plans.jsonc"
`
	if text != want {
		t.Errorf("DefaultSettingsFileText() =\n%s\nwant\n%s", text, want)
	}

	cfg, err := config.Load(config.Env{Getenv: envOf(nil)}, writeSettings(t, text))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	for _, setting := range cfg.Settings() {
		switch setting.Key {
		case "server_url", "plans_file", "port", "log_format":
			if setting.Source != "built in" {
				t.Errorf("%s comes from %q, want it built in", setting.Key, setting.Source)
			}
		}
	}
}

func TestTheDefaultSettingsFileHasEverySettingTheDesktopAppAndTheCommandLineRead(t *testing.T) {
	text := config.DefaultSettingsFileText("/home/ann/.config/starter")
	// Only the server reads these, and it never reads the default settings
	// file.
	serverOnly := map[string]bool{"port": true, "log_format": true}

	cfg, err := config.Load(config.Env{Getenv: envOf(nil)}, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	for _, setting := range cfg.Settings() {
		if serverOnly[setting.Key] {
			continue
		}
		setOrCommentedOut := regexp.MustCompile(`(?m)^(# )?` + regexp.QuoteMeta(setting.Key) + ` = `)
		if !setOrCommentedOut.MatchString(text) {
			t.Errorf("DefaultSettingsFileText() has no line for %s, set or commented out", setting.Key)
		}
	}
}
