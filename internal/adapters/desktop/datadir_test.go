package desktop_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/desktop"
)

func TestDataDirUsesStarterDataDirWhenSet(t *testing.T) {
	env := envOf(map[string]string{"STARTER_DATA_DIR": "/srv/starter-data"})

	got, err := desktop.DataDir(env, configDirAt("/home/someone/.config"))
	if err != nil {
		t.Fatalf("DataDir: %v", err)
	}
	if want := "/srv/starter-data"; got != want {
		t.Errorf("DataDir = %q, want %q", got, want)
	}
}

func TestDataDirFallsBackToUserConfigDir(t *testing.T) {
	got, err := desktop.DataDir(envOf(nil), configDirAt("/home/someone/.config"))
	if err != nil {
		t.Fatalf("DataDir: %v", err)
	}
	if want := filepath.Join("/home/someone/.config", "starter"); got != want {
		t.Errorf("DataDir = %q, want %q", got, want)
	}
}

func TestDataDirReturnsTheUserConfigDirError(t *testing.T) {
	errNoHome := errors.New("neither $XDG_CONFIG_HOME nor $HOME are defined")
	userConfigDir := func() (string, error) { return "", errNoHome }

	_, err := desktop.DataDir(envOf(nil), userConfigDir)
	if !errors.Is(err, errNoHome) {
		t.Errorf("DataDir error = %v, want %v", err, errNoHome)
	}
}

// envOf stands in for os.Getenv, reading from vars.
func envOf(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

// configDirAt stands in for os.UserConfigDir, always finding dir.
func configDirAt(dir string) func() (string, error) {
	return func() (string, error) { return dir, nil }
}
