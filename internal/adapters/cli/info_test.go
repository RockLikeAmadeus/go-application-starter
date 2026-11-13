package cli_test

import (
	"bytes"
	"context"
	"errors"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/cli"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/jsonfile"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/postgres/postgrestest"
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
	cmd := cli.NewRootCmd(bootstrap.New, bootstrap.NewBilling, bootstrap.NewClientSignIn, env, &out)
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

func TestInfoHelpSaysWhatItShows(t *testing.T) {
	out, err := runCLI(t, "info", "--help")
	if err != nil {
		t.Fatalf("info --help: %v", err)
	}
	if want := "Show the data location and the current plan"; !strings.Contains(out, want) {
		t.Errorf("info --help doesn't say %q:\n%s", want, out)
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

func TestALockedDataFolderErrorOffersTheServerOption(t *testing.T) {
	dir := t.TempDir()
	other, err := jsonfile.Open(dir)
	if err != nil {
		t.Fatalf("open the data folder first: %v", err)
	}
	defer other.Close()

	want := "Your data folder is in use by another program on this computer (" + dir + "). Please close the other program and try again, or connect to a running server with --server."
	for _, command := range []string{"info", "tui"} {
		_, err = runCLI(t, "--data-dir", dir, "--storage", "jsonfile", command)
		if err == nil || err.Error() != want {
			t.Errorf("%s error = %v, want %q", command, err, want)
		}
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

// Both storages that keep data in a folder share the desktop app's.
func TestWithoutDataDirCommandsUseTheLocalDataFolder(t *testing.T) {
	for _, storage := range []string{"jsonfile", "sqlite"} {
		t.Run(storage, func(t *testing.T) {
			settings := t.TempDir()
			local := filepath.Join(settings, "starter")
			var opened string
			newApp := func(cfg config.Config) (*bootstrap.App, error) {
				opened = cfg.DataDir
				return bootstrap.New(cfg)
			}
			cmd := cli.NewRootCmd(newApp, bootstrap.NewBilling, bootstrap.NewClientSignIn, computer(settings, nil), &bytes.Buffer{})
			cmd.SetArgs([]string{"--storage", storage, "info"})

			if err := cmd.ExecuteContext(context.Background()); err != nil {
				t.Fatalf("info: %v", err)
			}
			if opened != local {
				t.Errorf("opened the data folder %q, want the local one %q", opened, local)
			}
		})
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
	srv := testServer(t)

	out, err := runCLI(t, "--server", srv, "info")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if !strings.Contains(out, "Server: "+srv) {
		t.Errorf("info output doesn't name the server %s:\n%s", srv, out)
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

func TestInfoWithPostgresStorageNamesTheSettingAndNotTheURL(t *testing.T) {
	databaseURL := postgrestest.NewDatabase(t)
	env := computer(t.TempDir(), map[string]string{"STARTER_DATABASE_URL": databaseURL})

	out, err := runCLIWith(t, env, "--storage", "postgres", "info")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if !strings.Contains(out, "Data: kept in the Postgres database at database_url\n") || strings.Contains(out, "Data folder") {
		t.Errorf("info output doesn't say the data is in Postgres:\n%s", out)
	}
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("database URL: %v", err)
	}
	// The whole user:password pair: the test server's password is the word
	// "postgres", which the output contains anyway as the storage setting.
	if strings.Contains(out, databaseURL) || (parsed.User != nil && strings.Contains(out, parsed.User.String())) {
		t.Errorf("info output shows the database URL or its password:\n%s", out)
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

func TestCommandsCreateTheDefaultSettingsFileWhenThereIsNone(t *testing.T) {
	settings := t.TempDir()
	local := filepath.Join(settings, "starter")

	if _, err := runCLIWith(t, computer(settings, nil), "--storage", "memory", "info"); err != nil {
		t.Fatalf("info: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(local, "config.toml"))
	if err != nil {
		t.Fatalf("read the settings file: %v", err)
	}
	if want := config.DefaultSettingsFileText(local); string(got) != want {
		t.Errorf("settings file =\n%s\nwant the default one:\n%s", got, want)
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

	// A settings folder the default settings file can't be created in.
	readOnly := t.TempDir()
	if err := os.Chmod(readOnly, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(readOnly, 0o700) })

	tests := []struct {
		name     string
		settings string
		args     []string
		want     string
	}{
		{"named with --config", t.TempDir(), []string{"--config", file}, "Settings file: " + file + "\n"},
		{"none, when the default one can't be created", readOnly, nil, "Settings file: none\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runCLIWith(t, computer(tt.settings, nil), append(tt.args, "--storage", "memory", "info")...)
			if err != nil {
				t.Fatalf("info: %v", err)
			}
			if !strings.Contains(out, tt.want) {
				t.Errorf("info output doesn't say %q:\n%s", tt.want, out)
			}
		})
	}
}

func TestInfoListsEachSettingWithItsSource(t *testing.T) {
	settings := t.TempDir()
	writeSettings(t, filepath.Join(settings, "starter", "config.toml"), `log_format = "json"`)
	data := t.TempDir()

	out, err := runCLIWith(t, computer(settings, map[string]string{"STARTER_DATA_DIR": data}), "info")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	for _, want := range []string{
		`  log_format = "json"  # settings file` + "\n",
		`  data_dir = ` + strconv.Quote(data) + `  # environment (STARTER_DATA_DIR)` + "\n",
		`  port = 8080  # built in` + "\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("info output doesn't list %q:\n%s", want, out)
		}
	}
}

func TestInfoSaysWhichSettingsAnOptionChose(t *testing.T) {
	srv := testServer(t)
	data := t.TempDir()

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{"--storage and --data-dir", []string{"--storage", "jsonfile", "--data-dir", data}, []string{
			`  storage = "jsonfile"  # option (--storage)` + "\n",
			`  data_dir = ` + strconv.Quote(data) + `  # option (--data-dir)` + "\n",
		}},
		{"--server", []string{"--server", srv}, []string{
			`  server_url = ` + strconv.Quote(srv) + `  # option (--server)` + "\n",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runCLI(t, append(tt.args, "info")...)
			if err != nil {
				t.Fatalf("info: %v", err)
			}
			for _, want := range tt.want {
				if !strings.Contains(out, want) {
					t.Errorf("info output doesn't list %q:\n%s", want, out)
				}
			}
		})
	}
}

// testServer starts a server, with only the entitlements behind it, for the
// test, and returns its URL.
func testServer(t *testing.T) string {
	t.Helper()
	core, err := bootstrap.NewEmbedded(config.Config{Storage: config.StorageMemory})
	if err != nil {
		t.Fatalf("wire the server's core: %v", err)
	}
	srv := httptest.NewServer(httpapi.NewHandler(httpapi.Options{Entitlements: core.Entitlements}))
	t.Cleanup(srv.Close)
	return srv.URL
}
