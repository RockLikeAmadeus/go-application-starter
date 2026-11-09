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

	config.CreateDefaultSettingsFile(env)

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

	config.CreateDefaultSettingsFile(env)

	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read the settings file: %v", err)
	}
	if string(got) != mine {
		t.Errorf("settings file =\n%s\nwant it left as\n%s", got, mine)
	}
}

func TestNoDefaultSettingsFileIsCreatedWhenStarterConfigFileNamesOne(t *testing.T) {
	settings := t.TempDir()
	named := filepath.Join(t.TempDir(), "settings.toml")
	env := config.Env{
		Getenv:        envOf(map[string]string{"STARTER_CONFIG_FILE": named}),
		UserConfigDir: configDirAt(settings),
	}

	config.CreateDefaultSettingsFile(env)

	for _, file := range []string{filepath.Join(settings, "starter", "config.toml"), named} {
		if _, err := os.Stat(file); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("%s was created (Stat: %v), want nothing created", file, err)
		}
	}
}

func TestTheDefaultSettingsFileCarriesOnWhenItCantWrite(t *testing.T) {
	settings := t.TempDir()
	// A file where the app's folder should be, so nothing can be written there.
	blocker := filepath.Join(settings, "starter")
	const mine = "not a folder\n"
	if err := os.WriteFile(blocker, []byte(mine), 0o600); err != nil {
		t.Fatal(err)
	}
	env := config.Env{Getenv: envOf(nil), UserConfigDir: configDirAt(settings)}

	config.CreateDefaultSettingsFile(env)

	got, err := os.ReadFile(blocker)
	if err != nil {
		t.Fatalf("read the file in the app's folder's place: %v", err)
	}
	if string(got) != mine {
		t.Errorf("file in the app's folder's place =\n%s\nwant it left as\n%s", got, mine)
	}
}

func TestTheDefaultSettingsFileCarriesOnWithoutASettingsFolder(t *testing.T) {
	// Without a settings folder, a relative path like starter/config.toml
	// would land in the working directory, so watch an empty one.
	workDir := t.TempDir()
	t.Chdir(workDir)
	errNoHome := errors.New("neither $XDG_CONFIG_HOME nor $HOME are defined")
	env := config.Env{Getenv: envOf(nil), UserConfigDir: func() (string, error) { return "", errNoHome }}

	config.CreateDefaultSettingsFile(env)

	entries, err := os.ReadDir(workDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		t.Errorf("%s was created in the working directory, want nothing created", entry.Name())
	}
}
