package config_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

func TestLocalDataDirUsesStarterDataDirWhenSet(t *testing.T) {
	env := envOf(map[string]string{"STARTER_DATA_DIR": "/srv/starter-data"})

	got, err := config.LocalDataDir(env, configDirAt("/home/someone/.config"))
	if err != nil {
		t.Fatalf("LocalDataDir: %v", err)
	}
	if want := "/srv/starter-data"; got != want {
		t.Errorf("LocalDataDir = %q, want %q", got, want)
	}
}

func TestLocalDataDirFallsBackToUserConfigDir(t *testing.T) {
	got, err := config.LocalDataDir(envOf(nil), configDirAt("/home/someone/.config"))
	if err != nil {
		t.Fatalf("LocalDataDir: %v", err)
	}
	if want := filepath.Join("/home/someone/.config", "starter"); got != want {
		t.Errorf("LocalDataDir = %q, want %q", got, want)
	}
}

func TestLocalDataDirReturnsTheUserConfigDirError(t *testing.T) {
	errNoHome := errors.New("neither $XDG_CONFIG_HOME nor $HOME are defined")
	userConfigDir := func() (string, error) { return "", errNoHome }

	_, err := config.LocalDataDir(envOf(nil), userConfigDir)
	if !errors.Is(err, errNoHome) {
		t.Errorf("LocalDataDir error = %v, want %v", err, errNoHome)
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
