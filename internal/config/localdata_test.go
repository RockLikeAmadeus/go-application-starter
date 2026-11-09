package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

func TestLocalDataDirUsesStarterDataDirWhenSet(t *testing.T) {
	env := envOf(map[string]string{"STARTER_DATA_DIR": "/srv/starter-data"})

	got, err := config.LocalDataDir(config.Env{Getenv: env, UserConfigDir: configDirAt("/home/someone/.config")}, "")
	if err != nil {
		t.Fatalf("LocalDataDir: %v", err)
	}
	if want := "/srv/starter-data"; got != want {
		t.Errorf("LocalDataDir = %q, want %q", got, want)
	}
}

func TestLocalDataDirFallsBackToUserConfigDir(t *testing.T) {
	got, err := config.LocalDataDir(config.Env{Getenv: envOf(nil), UserConfigDir: configDirAt("/home/someone/.config")}, "")
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

	_, err := config.LocalDataDir(config.Env{Getenv: envOf(nil), UserConfigDir: userConfigDir}, "")
	if !errors.Is(err, errNoHome) {
		t.Errorf("LocalDataDir error = %v, want %v", err, errNoHome)
	}
}

func TestLocalDataDirUsesTheSettingsFilesDataDir(t *testing.T) {
	file := writeSettings(t, `data_dir = "/srv/work-data"`)
	env := config.Env{Getenv: envOf(nil), UserConfigDir: configDirAt("/home/someone/.config")}

	got, err := config.LocalDataDir(env, file)
	if err != nil {
		t.Fatalf("LocalDataDir: %v", err)
	}
	if want := "/srv/work-data"; got != want {
		t.Errorf("LocalDataDir = %q, want %q", got, want)
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

func TestLocalConfigFileIsConfigTomlInTheAppsSettingsFolder(t *testing.T) {
	settings := t.TempDir()
	want := filepath.Join(settings, "starter", "config.toml")
	if err := os.MkdirAll(filepath.Dir(want), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(want, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := config.LocalConfigFile(config.Env{Getenv: envOf(nil), UserConfigDir: configDirAt(settings)})
	if err != nil {
		t.Fatalf("LocalConfigFile: %v", err)
	}
	if got != want {
		t.Errorf("LocalConfigFile = %q, want %q", got, want)
	}
}

func TestStarterConfigFileNamesAnotherSettingsFile(t *testing.T) {
	env := config.Env{
		Getenv:        envOf(map[string]string{"STARTER_CONFIG_FILE": "/srv/starter/settings.toml"}),
		UserConfigDir: configDirAt(t.TempDir()),
	}

	got, err := config.LocalConfigFile(env)
	if err != nil {
		t.Fatalf("LocalConfigFile: %v", err)
	}
	if want := "/srv/starter/settings.toml"; got != want {
		t.Errorf("LocalConfigFile = %q, want %q", got, want)
	}
}

func TestWithoutTheDefaultSettingsFileThereIsNone(t *testing.T) {
	got, err := config.LocalConfigFile(config.Env{Getenv: envOf(nil), UserConfigDir: configDirAt(t.TempDir())})
	if err != nil {
		t.Fatalf("LocalConfigFile: %v", err)
	}
	if got != "" {
		t.Errorf("LocalConfigFile = %q, want none", got)
	}
}

func TestTheDefaultSettingsFileIsCreatedWhenMissing(t *testing.T) {
	settings := t.TempDir()
	env := config.Env{Getenv: envOf(nil), UserConfigDir: configDirAt(settings)}

	if err := config.CreateDefaultSettingsFile(env); err != nil {
		t.Fatalf("CreateDefaultSettingsFile: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(settings, "starter", "config.toml"))
	if err != nil {
		t.Fatalf("read the settings file: %v", err)
	}
	if want := config.DefaultSettingsFileText(filepath.Join(settings, "starter")); string(got) != want {
		t.Errorf("settings file =\n%s\nwant\n%s", got, want)
	}
}

func TestTheDefaultSettingsFileIsNeverOverwritten(t *testing.T) {
	settings := t.TempDir()
	file := filepath.Join(settings, "starter", "config.toml")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	const mine = "storage = \"memory\"\n"
	if err := os.WriteFile(file, []byte(mine), 0o600); err != nil {
		t.Fatal(err)
	}
	env := config.Env{Getenv: envOf(nil), UserConfigDir: configDirAt(settings)}

	_ = config.CreateDefaultSettingsFile(env)

	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read the settings file: %v", err)
	}
	if string(got) != mine {
		t.Errorf("settings file =\n%s\nwant it left as\n%s", got, mine)
	}
}
