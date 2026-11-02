package desktop_test

import (
	"os"
	"path/filepath"
	"testing"
)

// envOf stands in for os.Getenv, reading from vars.
func envOf(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

// configDirAt stands in for os.UserConfigDir, always finding dir.
func configDirAt(dir string) func() (string, error) {
	return func() (string, error) { return dir, nil }
}

// writeSettings writes the local settings file, config.toml in the app's
// folder in the settings folder settings, with content.
func writeSettings(t *testing.T, settings, content string) {
	t.Helper()
	file := filepath.Join(settings, "starter", "config.toml")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(content+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}
