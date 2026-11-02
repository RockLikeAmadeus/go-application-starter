package cli_test

import (
	"bytes"
	"context"
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/cli"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/jsonfile"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// runCLI executes the CLI with args, as cmd/starter wires it, on a computer
// with an empty environment and settings folder.
func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	return runCLIWith(t, computer(t.TempDir(), nil), args...)
}

// runCLIWith executes the CLI with args on the computer env.
func runCLIWith(t *testing.T, env config.Env, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := cli.NewRootCmd(bootstrap.New, env, &out)
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(context.Background())
	return out.String(), err
}

func TestInfoNamesTheDataFolder(t *testing.T) {
	dir := t.TempDir()

	out, err := runCLI(t, "--data-dir", dir, "--storage", "jsonfile", "info")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if !strings.Contains(out, dir) {
		t.Errorf("info output doesn't name the data folder %s:\n%s", dir, out)
	}
}

func TestInfoSaysTheDataIsInUseWhileAnotherProgramHasItOpen(t *testing.T) {
	dir := t.TempDir()
	other, err := jsonfile.Open(dir)
	if err != nil {
		t.Fatalf("open the data folder first: %v", err)
	}
	defer other.Close()

	_, err = runCLI(t, "--data-dir", dir, "--storage", "jsonfile", "info")
	if !errors.Is(err, jsonfile.ErrLocked) {
		t.Errorf("info error = %v, want %v", err, jsonfile.ErrLocked)
	}
}

func TestInfoNamesTheWorkspacesPlan(t *testing.T) {
	out, err := runCLI(t, "--data-dir", t.TempDir(), "--storage", "jsonfile", "info")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if !strings.Contains(out, "Plan: "+bootstrap.LocalPlan) {
		t.Errorf("info output doesn't name the plan %q:\n%s", bootstrap.LocalPlan, out)
	}
}

// computer stands in for the user's computer: an environment holding vars,
// and the settings folder settings.
func computer(settings string, vars map[string]string) config.Env {
	return config.Env{
		Getenv:        func(key string) string { return vars[key] },
		UserConfigDir: func() (string, error) { return settings, nil },
	}
}

func TestWithoutDataDirCommandsUseTheLocalDataFolder(t *testing.T) {
	settings := t.TempDir()
	local := filepath.Join(settings, "starter")
	var opened string
	newApp := func(cfg config.Config) (*bootstrap.App, error) {
		opened = cfg.DataDir
		return bootstrap.New(cfg)
	}
	cmd := cli.NewRootCmd(newApp, computer(settings, nil), &bytes.Buffer{})
	cmd.SetArgs([]string{"--storage", "jsonfile", "info"})

	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("info: %v", err)
	}
	if opened != local {
		t.Errorf("opened the data folder %q, want the local one %q", opened, local)
	}
}

func TestWithoutTheLocalDataFolderCommandsFailSayingWhy(t *testing.T) {
	errNoHome := errors.New("find the data folder: neither $XDG_CONFIG_HOME nor $HOME are defined")
	env := computer("", nil)
	env.UserConfigDir = func() (string, error) { return "", errNoHome }
	_, err := runCLIWith(t, env, "--storage", "jsonfile", "info")
	if !errors.Is(err, errNoHome) {
		t.Errorf("info error = %v, want %v", err, errNoHome)
	}
}

func TestInfoWithAServerNamesTheServer(t *testing.T) {
	core, err := bootstrap.NewEmbedded(config.Config{Storage: config.StorageMemory})
	if err != nil {
		t.Fatalf("wire the server's core: %v", err)
	}
	srv := httptest.NewServer(httpapi.NewHandler(httpapi.Options{Entitlements: core.Entitlements}))
	defer srv.Close()

	out, err := runCLI(t, "--server", srv.URL, "info")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if !strings.Contains(out, "Server: "+srv.URL) {
		t.Errorf("info output doesn't name the server %s:\n%s", srv.URL, out)
	}
}

func TestInfoWithMemoryStorageSaysTheDataIsNotSaved(t *testing.T) {
	out, err := runCLI(t, "--storage", "memory", "info")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if !strings.Contains(out, "Data: kept in memory, not saved") || strings.Contains(out, "Data folder") {
		t.Errorf("info output doesn't say the data is kept in memory:\n%s", out)
	}
}

func TestCommandsUseTheLocalSettingsFile(t *testing.T) {
	settings := t.TempDir()
	data := filepath.Join(t.TempDir(), "work-data")
	writeSettings(t, filepath.Join(settings, "starter", "config.toml"), "data_dir = "+strconv.Quote(data))

	out, err := runCLIWith(t, computer(settings, nil), "info")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if !strings.Contains(out, "Data folder: "+data) {
		t.Errorf("info output doesn't name the settings file's data folder %s:\n%s", data, out)
	}
}

func TestConfigNamesTheSettingsFileForOneCommand(t *testing.T) {
	data := filepath.Join(t.TempDir(), "work-data")
	file := filepath.Join(t.TempDir(), "work.toml")
	writeSettings(t, file, "data_dir = "+strconv.Quote(data))

	out, err := runCLI(t, "--config", file, "info")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if !strings.Contains(out, "Data folder: "+data) {
		t.Errorf("info output doesn't name the --config file's data folder %s:\n%s", data, out)
	}
}

// writeSettings writes a settings file at path with content, creating its
// folder.
func writeSettings(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestInfoNamesTheSettingsFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "work.toml")
	writeSettings(t, file, `storage = "memory"`)

	tests := []struct {
		name string
		args []string
		want string
	}{
		{"named with --config", []string{"--config", file}, "Settings file: " + file + "\n"},
		{"none", nil, "Settings file: none\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runCLI(t, append(tt.args, "--storage", "memory", "info")...)
			if err != nil {
				t.Fatalf("info: %v", err)
			}
			if !strings.Contains(out, tt.want) {
				t.Errorf("info output doesn't say %q:\n%s", tt.want, out)
			}
		})
	}
}
