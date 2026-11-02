package config

import (
	"fmt"
	"path/filepath"
)

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
	return filepath.Join(configDir, "starter"), nil
}
