package config

import "fmt"

// DefaultSettingsFileText is the text of the settings file the desktop app,
// the CLI and the TUI create when there is none, for the data folder dataDir.
// Each setting it sets has the value it would have without the file.
func DefaultSettingsFileText(dataDir string) string {
	return fmt.Sprintf(`# The settings for this application. The application created this file with the
# values it uses when no settings file exists.

# How the application keeps its data: "jsonfile" (plain-text storage, which
# one program on this computer can use at a time), "sqlite" (a database,
# which several programs can use at once), or "memory" (kept only until the
# program ends).
storage = %q

# The folder in which the application keeps its data. The desktop
# application, the command line, and the terminal application share it.
data_dir = %q

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
`, StorageJSONFile, dataDir)
}
