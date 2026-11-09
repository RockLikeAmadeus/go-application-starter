package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// appFolder is the app's folder in the user's settings folder.
const appFolder = "starter"

// LocalDataDir returns the local edition's data folder on this computer,
// which the desktop app, the CLI and the TUI share: STARTER_DATA_DIR when set,
// else data_dir in the TOML settingsFile ("" for none), else the app's folder
// in the user's settings folder.
func LocalDataDir(env Env, settingsFile string) (string, error) {
	if dir := env.Getenv("STARTER_DATA_DIR"); dir != "" {
		return dir, nil
	}
	fileSettings, err := readSettingsFile(settingsFile)
	if err != nil {
		return "", err
	}
	if dir := fileSettings.String("data_dir"); dir != "" {
		return dir, nil
	}
	configDir, err := env.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find the data folder: %w", err)
	}
	return filepath.Join(configDir, appFolder), nil
}

// LocalConfigFile returns the local edition's settings file on this
// computer, which the desktop app, the CLI and the TUI share:
// STARTER_CONFIG_FILE when set, else config.toml in the app's folder in the
// user's settings folder. Load fails if a file named this way is missing,
// but the default one is optional: without it, there is none ("").
func LocalConfigFile(env Env) (string, error) {
	if file := env.Getenv("STARTER_CONFIG_FILE"); file != "" {
		return file, nil
	}
	file, err := defaultSettingsFile(env)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(file); errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	return file, nil
}

// CreateDefaultSettingsFile writes DefaultSettingsFileText, for the default
// data folder (the app's folder in the user's settings folder), to the
// default settings file, making the app's folder. It never overwrites a
// settings file that's already there.
func CreateDefaultSettingsFile(env Env) error {
	file, err := defaultSettingsFile(env)
	if err != nil {
		return err
	}
	appDir := filepath.Dir(file)
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return err
	}
	return writeNewFile(file, DefaultSettingsFileText(appDir))
}

// writeNewFile writes text to file only if there is no file there yet, so it
// never overwrites one, even another program's written at the same moment.
func writeNewFile(file, text string) error {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(text); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// defaultSettingsFile is config.toml in the app's folder in the user's
// settings folder.
func defaultSettingsFile(env Env) (string, error) {
	configDir, err := env.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, appFolder, "config.toml"), nil
}
