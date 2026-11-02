package config

import (
	"fmt"
	"path/filepath"
)

// appFolder is the app's folder in the user's settings folder.
const appFolder = "starter"

// LocalDataDir returns the local edition's data folder on this computer,
// which the desktop app, the CLI and the TUI share: STARTER_DATA_DIR when set,
// else the app's folder in the user's settings folder. getenv reads the
// environment (os.Getenv in production) and userConfigDir finds the user's
// settings folder (os.UserConfigDir in production).
func LocalDataDir(getenv func(string) string, userConfigDir func() (string, error)) (string, error) {
	if dir := getenv("STARTER_DATA_DIR"); dir != "" {
		return dir, nil
	}
	configDir, err := userConfigDir()
	if err != nil {
		return "", fmt.Errorf("find the data folder: %w", err)
	}
	return filepath.Join(configDir, appFolder), nil
}

// LocalConfigFile returns the local edition's settings file on this
// computer, which the desktop app, the CLI and the TUI share:
// STARTER_CONFIG_FILE when set, else config.toml in the app's folder in the
// user's settings folder. Load fails if a file named this way is missing.
func LocalConfigFile(env Env) (string, error) {
	if file := env.Getenv("STARTER_CONFIG_FILE"); file != "" {
		return file, nil
	}
	configDir, err := env.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, appFolder, "config.toml"), nil
}
