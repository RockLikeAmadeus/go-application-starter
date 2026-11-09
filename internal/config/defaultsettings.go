package config

import "fmt"

// DefaultSettingsFileText is the text of the settings file the desktop app,
// the CLI and the TUI create when there is none, for the data folder dataDir.
// Each setting it sets has the value it would have without the file.
func DefaultSettingsFileText(dataDir string) string {
	return fmt.Sprintf("storage = %q\ndata_dir = %q\nignore_synced_folder = false\n", StorageJSONFile, dataDir)
}
