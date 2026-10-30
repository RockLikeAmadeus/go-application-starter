package desktop_test

import (
	"path/filepath"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/desktop"
)

func TestDataDirUsesStarterDataDirWhenSet(t *testing.T) {
	env := map[string]string{"STARTER_DATA_DIR": "/srv/starter-data"}
	userConfigDir := func() (string, error) { return "/home/someone/.config", nil }

	got, err := desktop.DataDir(func(k string) string { return env[k] }, userConfigDir)
	if err != nil {
		t.Fatalf("DataDir: %v", err)
	}
	if want := "/srv/starter-data"; got != want {
		t.Errorf("DataDir = %q, want %q", got, want)
	}
}

func TestDataDirFallsBackToUserConfigDir(t *testing.T) {
	noEnv := func(string) string { return "" }
	userConfigDir := func() (string, error) { return "/home/someone/.config", nil }

	got, err := desktop.DataDir(noEnv, userConfigDir)
	if err != nil {
		t.Fatalf("DataDir: %v", err)
	}
	if want := filepath.Join("/home/someone/.config", "starter"); got != want {
		t.Errorf("DataDir = %q, want %q", got, want)
	}
}
