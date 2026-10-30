// Package desktop wires the desktop app: it picks the data folder and opens
// the core behind the same HTTP handler the server uses. It has no Wails
// dependency, so it is unit-testable; cmd/desktop only adds the window.
package desktop

import (
	"fmt"
	"path/filepath"
)

// DataDir returns the folder where the desktop app keeps its data.
// getenv reads the environment (os.Getenv in production) and userConfigDir
// finds the user's settings folder (os.UserConfigDir in production).
func DataDir(getenv func(string) string, userConfigDir func() (string, error)) (string, error) {
	if dir := getenv("STARTER_DATA_DIR"); dir != "" {
		return dir, nil
	}
	configDir, err := userConfigDir()
	if err != nil {
		return "", fmt.Errorf("find the data folder: %w", err)
	}
	return filepath.Join(configDir, "starter"), nil
}
