package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

func TestLoadReadsTheEnvironmentItIsGiven(t *testing.T) {
	cfg, err := config.Load(config.Env{Getenv: envOf(map[string]string{"STARTER_STORAGE": "memory"})}, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Storage != config.StorageMemory {
		t.Errorf("Storage = %q, want %q", cfg.Storage, config.StorageMemory)
	}
}

func TestLoadUsesTheSettingsFile(t *testing.T) {
	file := writeSettings(t, `
storage = "memory"
data_dir = "/srv/work-data"
`)

	cfg, err := config.Load(config.Env{Getenv: envOf(nil)}, file)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Storage != config.StorageMemory || cfg.DataDir != "/srv/work-data" {
		t.Errorf("Storage, DataDir = %q, %q, want %q, %q", cfg.Storage, cfg.DataDir, config.StorageMemory, "/srv/work-data")
	}
}

// writeSettings writes a settings file with content, and returns its path.
func writeSettings(t *testing.T, content string) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return file
}

func TestTheEnvironmentOverridesTheSettingsFile(t *testing.T) {
	file := writeSettings(t, `data_dir = "/srv/work-data"`)
	env := envOf(map[string]string{"STARTER_DATA_DIR": "/tmp/scratch"})

	cfg, err := config.Load(config.Env{Getenv: env}, file)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DataDir != "/tmp/scratch" {
		t.Errorf("DataDir = %q, want the environment's %q", cfg.DataDir, "/tmp/scratch")
	}
}

func TestAnUnknownSettingErrorReadsAsASentence(t *testing.T) {
	file := writeSettings(t, `colour = "blue"`)

	_, err := config.Load(config.Env{Getenv: envOf(nil)}, file)

	want := "The settings file " + file + ` contains an unknown setting: "colour".`
	if err == nil || err.Error() != want {
		t.Errorf("error = %v, want %q", err, want)
	}
}

func TestAnUnreadableSettingsFileErrorReadsAsASentence(t *testing.T) {
	file := filepath.Join(t.TempDir(), "missing.toml")

	_, err := config.Load(config.Env{Getenv: envOf(nil)}, file)

	cause := errors.Unwrap(err)
	if cause == nil {
		t.Fatalf("error = %v, want one wrapping why the file could not be read", err)
	}
	want := "The settings file " + file + " could not be read: " + cause.Error()
	if err.Error() != want {
		t.Errorf("error = %v, want %q", err, want)
	}
}

func TestASettingsFileThatIsntTOMLFailsNamingTheFile(t *testing.T) {
	file := writeSettings(t, `data_dir = "/srv/work-data`)

	_, err := config.Load(config.Env{Getenv: envOf(nil)}, file)
	if err == nil || !strings.Contains(err.Error(), file) {
		t.Errorf("Load error = %v, want one naming %s", err, file)
	}
}

func TestLoadSaysWhereEachSettingCameFrom(t *testing.T) {
	file := writeSettings(t, `
storage = "memory"
data_dir = "/srv/work-data"
`)
	env := envOf(map[string]string{"STARTER_DATA_DIR": "/tmp/scratch"})

	cfg, err := config.Load(config.Env{Getenv: env}, file)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := map[string]string{
		"storage":    "settings file",
		"data_dir":   "environment (STARTER_DATA_DIR)",
		"log_format": "built in",
		"port":       "built in",
	}
	for key, source := range want {
		if got := cfg.Sources[key]; got != source {
			t.Errorf("Sources[%q] = %q, want %q", key, got, source)
		}
	}
}

func TestStorageAcceptsSQLite(t *testing.T) {
	cfg, err := config.Load(config.Env{Getenv: envOf(map[string]string{"STARTER_STORAGE": "sqlite"})}, "")
	if err != nil {
		t.Fatalf("Load with sqlite: %v", err)
	}
	if cfg.Storage != config.StorageSQLite {
		t.Errorf("Storage = %q, want %q", cfg.Storage, config.StorageSQLite)
	}
}

func TestStorageAcceptsPostgres(t *testing.T) {
	cfg, err := config.Load(config.Env{Getenv: envOf(map[string]string{"STARTER_STORAGE": "postgres"})}, "")
	if err != nil {
		t.Fatalf("Load with postgres: %v", err)
	}
	if cfg.Storage != config.StoragePostgres {
		t.Errorf("Storage = %q, want %q", cfg.Storage, config.StoragePostgres)
	}
}

func TestDatabaseURLIsASettingReadFromTheEnvironment(t *testing.T) {
	const databaseURL = "postgres://starter@db.example.com:5432/starter"
	env := envOf(map[string]string{"STARTER_DATABASE_URL": databaseURL})

	cfg, err := config.Load(config.Env{Getenv: env}, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.DatabaseURL != databaseURL {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, databaseURL)
	}
	const wantSource = "environment (STARTER_DATABASE_URL)"
	isListed := false
	for _, setting := range cfg.Settings() {
		if setting.Key == "database_url" && setting.Source == wantSource {
			isListed = true
		}
	}
	if !isListed {
		t.Errorf("Settings() = %v, want database_url from %q", cfg.Settings(), wantSource)
	}
}

func TestSettingsHidesTheDatabaseURLOnlyWhenItIsSet(t *testing.T) {
	const databaseURL = "postgres://starter:secret@db.example.com:5432/starter"
	withURL, err := config.Load(config.Env{Getenv: envOf(map[string]string{"STARTER_DATABASE_URL": databaseURL})}, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	withoutURL, err := config.Load(config.Env{Getenv: envOf(nil)}, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	wantHidden := config.Setting{Key: "database_url", Value: "(hidden)", Source: "environment (STARTER_DATABASE_URL)"}
	if !listsSetting(withURL.Settings(), wantHidden) {
		t.Errorf("Settings() = %v, want it to list %v", withURL.Settings(), wantHidden)
	}
	wantEmpty := config.Setting{Key: "database_url", Value: "", Source: "built in"}
	if !listsSetting(withoutURL.Settings(), wantEmpty) {
		t.Errorf("Settings() = %v, want it to list %v", withoutURL.Settings(), wantEmpty)
	}
}

func TestAnUnknownStorageErrorReadsAsASentence(t *testing.T) {
	_, err := config.Load(config.Env{Getenv: envOf(map[string]string{"STARTER_STORAGE": "disk"})}, "")

	want := `The storage setting "disk" is not recognized. Please use "memory", "jsonfile", "sqlite", or "postgres".`
	if err == nil || err.Error() != want {
		t.Errorf("error = %v, want %q", err, want)
	}
}

func TestAnUnknownLogFormatErrorReadsAsASentence(t *testing.T) {
	_, err := config.Load(config.Env{Getenv: envOf(map[string]string{"STARTER_LOG_FORMAT": "xml"})}, "")

	want := `The log format setting "xml" is not recognized. Please use "text" or "json".`
	if err == nil || err.Error() != want {
		t.Errorf("error = %v, want %q", err, want)
	}
}

func TestAnUnknownSignInErrorReadsAsASentence(t *testing.T) {
	_, err := config.Load(config.Env{Getenv: envOf(map[string]string{"STARTER_SIGN_IN": "google"})}, "")

	want := `The sign-in setting "google" is not recognized. Please use "off", "dev", or "oidc".`
	if err == nil || err.Error() != want {
		t.Errorf("error = %v, want %q", err, want)
	}
}

func TestShouldIgnoreSyncedFolderIsABoolFromTheSettingsFileOrEnvironmentDefaultingToFalse(t *testing.T) {
	cases := []struct {
		name     string
		file     string
		env      map[string]string
		want     bool
		wantFrom string
	}{
		{"built in", "", nil, false, "built in"},
		{"settings file", "ignore_synced_folder = true", nil, true, "settings file"},
		{"environment", "", map[string]string{"STARTER_IGNORE_SYNCED_FOLDER": "true"}, true, "environment (STARTER_IGNORE_SYNCED_FOLDER)"},
		{"environment over file", "ignore_synced_folder = true", map[string]string{"STARTER_IGNORE_SYNCED_FOLDER": "false"}, false, "environment (STARTER_IGNORE_SYNCED_FOLDER)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			file := ""
			if c.file != "" {
				file = writeSettings(t, c.file)
			}

			cfg, err := config.Load(config.Env{Getenv: envOf(c.env)}, file)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if cfg.ShouldIgnoreSyncedFolder != c.want {
				t.Errorf("ShouldIgnoreSyncedFolder = %v, want %v", cfg.ShouldIgnoreSyncedFolder, c.want)
			}
			if !listsSetting(cfg.Settings(), config.Setting{Key: "ignore_synced_folder", Value: c.want, Source: c.wantFrom}) {
				t.Errorf("Settings() = %v, want ignore_synced_folder = %v from %q", cfg.Settings(), c.want, c.wantFrom)
			}
		})
	}

	t.Run("not a bool", func(t *testing.T) {
		_, err := config.Load(config.Env{Getenv: envOf(map[string]string{"STARTER_IGNORE_SYNCED_FOLDER": "sometimes"})}, "")
		if err == nil || !strings.Contains(err.Error(), "STARTER_IGNORE_SYNCED_FOLDER") {
			t.Errorf("Load error = %v, want one naming %s", err, "STARTER_IGNORE_SYNCED_FOLDER")
		}
	})
}

func TestTheCredentialStoreIsSystemByDefaultOrFileFromTheSettingsFileOrEnvironmentAndAnyOtherIsRefused(t *testing.T) {
	cases := []struct {
		name     string
		file     string
		env      map[string]string
		want     config.CredentialStore
		wantFrom string
	}{
		{"built in", "", nil, config.CredentialStoreSystem, "built in"},
		{"settings file", `credential_store = "file"`, nil, config.CredentialStoreFile, "settings file"},
		{"environment", "", map[string]string{"STARTER_CREDENTIAL_STORE": "file"}, config.CredentialStoreFile, "environment (STARTER_CREDENTIAL_STORE)"},
		{"environment over file", `credential_store = "file"`, map[string]string{"STARTER_CREDENTIAL_STORE": "system"}, config.CredentialStoreSystem, "environment (STARTER_CREDENTIAL_STORE)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			file := ""
			if c.file != "" {
				file = writeSettings(t, c.file)
			}

			cfg, err := config.Load(config.Env{Getenv: envOf(c.env)}, file)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if cfg.CredentialStore != c.want {
				t.Errorf("CredentialStore = %q, want %q", cfg.CredentialStore, c.want)
			}
			if !listsSetting(cfg.Settings(), config.Setting{Key: "credential_store", Value: string(c.want), Source: c.wantFrom}) {
				t.Errorf("Settings() = %v, want credential_store = %q from %q", cfg.Settings(), c.want, c.wantFrom)
			}
		})
	}

	t.Run("not recognized", func(t *testing.T) {
		_, err := config.Load(config.Env{Getenv: envOf(map[string]string{"STARTER_CREDENTIAL_STORE": "vault"})}, "")

		want := `The credential store setting "vault" is not recognized. Please use "system" or "file".`
		if err == nil || err.Error() != want {
			t.Errorf("error = %v, want %q", err, want)
		}
	})
}

func TestSignInIsOffByDefaultOrDevFromTheSettingsFileOrEnvironment(t *testing.T) {
	cases := []struct {
		name     string
		file     string
		env      map[string]string
		want     config.SignInMode
		wantFrom string
	}{
		{"built in", "", nil, config.SignInOff, "built in"},
		{"settings file", `sign_in = "dev"`, nil, config.SignInDev, "settings file"},
		{"environment", "", map[string]string{"STARTER_SIGN_IN": "dev"}, config.SignInDev, "environment (STARTER_SIGN_IN)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			file := ""
			if c.file != "" {
				file = writeSettings(t, c.file)
			}

			cfg, err := config.Load(config.Env{Getenv: envOf(c.env)}, file)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if cfg.SignIn != c.want {
				t.Errorf("SignIn = %q, want %q", cfg.SignIn, c.want)
			}
			if !listsSetting(cfg.Settings(), config.Setting{Key: "sign_in", Value: string(c.want), Source: c.wantFrom}) {
				t.Errorf("Settings() = %v, want sign_in = %q from %q", cfg.Settings(), c.want, c.wantFrom)
			}
		})
	}
}

func TestOIDCSignInAndItsSettingsComeFromTheEnvironment(t *testing.T) {
	cfg, err := config.Load(config.Env{Getenv: envOf(map[string]string{
		"STARTER_SIGN_IN":            "oidc",
		"STARTER_PUBLIC_URL":         "https://app.example.com",
		"STARTER_OIDC_ISSUER":        "https://accounts.example.com",
		"STARTER_OIDC_CLIENT_ID":     "client-1",
		"STARTER_OIDC_CLIENT_SECRET": "s3cret",
	})}, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.SignIn != config.SignInOIDC {
		t.Errorf("SignIn = %q, want %q", cfg.SignIn, config.SignInOIDC)
	}
	if cfg.PublicURL != "https://app.example.com" || cfg.OIDCIssuer != "https://accounts.example.com" ||
		cfg.OIDCClientID != "client-1" || cfg.OIDCClientSecret != "s3cret" {
		t.Errorf("settings = %q, %q, %q, %q", cfg.PublicURL, cfg.OIDCIssuer, cfg.OIDCClientID, cfg.OIDCClientSecret)
	}
	listed := map[string]any{}
	for _, s := range cfg.Settings() {
		listed[s.Key] = s.Value
	}
	want := map[string]any{
		"public_url":         "https://app.example.com",
		"oidc_issuer":        "https://accounts.example.com",
		"oidc_client_id":     "client-1",
		"oidc_client_secret": "(hidden)",
	}
	for key, value := range want {
		if listed[key] != value {
			t.Errorf("Settings() lists %s = %v, want %v", key, listed[key], value)
		}
	}
}

// listsSetting reports whether settings holds want.
func listsSetting(settings []config.Setting, want config.Setting) bool {
	for _, s := range settings {
		if s == want {
			return true
		}
	}
	return false
}

func TestOIDCSignInRefusesAMissingSettingNamingIt(t *testing.T) {
	complete := map[string]string{
		"STARTER_SIGN_IN":            "oidc",
		"STARTER_PUBLIC_URL":         "https://app.example.com",
		"STARTER_OIDC_ISSUER":        "https://accounts.example.com",
		"STARTER_OIDC_CLIENT_ID":     "client-1",
		"STARTER_OIDC_CLIENT_SECRET": "s3cret",
	}
	tests := []struct {
		missing string
		want    string
	}{
		{"STARTER_PUBLIC_URL", `The public_url setting is required when the sign-in setting is "oidc". Please set it to the address people reach this server at, such as https://app.example.com.`},
		{"STARTER_OIDC_ISSUER", `The oidc_issuer setting is required when the sign-in setting is "oidc". Please set it to the address of the sign-in provider.`},
		{"STARTER_OIDC_CLIENT_ID", `The oidc_client_id setting is required when the sign-in setting is "oidc". Please set it to the client ID the sign-in provider gave this server.`},
		{"STARTER_OIDC_CLIENT_SECRET", `The oidc_client_secret setting is required when the sign-in setting is "oidc". Please set it to the client secret the sign-in provider gave this server.`},
	}
	for _, tc := range tests {
		t.Run(tc.missing, func(t *testing.T) {
			env := map[string]string{}
			for k, v := range complete {
				if k != tc.missing {
					env[k] = v
				}
			}

			_, err := config.Load(config.Env{Getenv: envOf(env)}, "")

			if err == nil || err.Error() != tc.want {
				t.Errorf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestPublicURLThatIsNotAnHTTPAddressIsRefused(t *testing.T) {
	want := `The public_url setting must be a full address starting with http:// or https://. Please set it to the address people reach this server at, such as https://app.example.com.`
	for _, value := range []string{"app.example.com", "ftp://app.example.com", "/path", "https://", "http:///path"} {
		t.Run(value, func(t *testing.T) {
			env := map[string]string{"STARTER_PUBLIC_URL": value}

			_, err := config.Load(config.Env{Getenv: envOf(env)}, "")

			if err == nil || err.Error() != want {
				t.Errorf("error = %v, want %q", err, want)
			}
		})
	}
}

func TestTheBuiltInServerURLIsTheLowestLayerOfServerURL(t *testing.T) {
	previous := config.BuiltInServerURL
	config.BuiltInServerURL = "https://built-in.example.com"
	t.Cleanup(func() { config.BuiltInServerURL = previous })

	tests := []struct {
		name, settings string
		env            map[string]string
		wantURL        string
		wantSource     string
	}{
		{"nothing else set", "", nil, "https://built-in.example.com", "built in"},
		{"an empty setting", `server_url = ""`, nil, "https://built-in.example.com", "built in"},
		{"the settings file", `server_url = "https://file.example.com"`, nil, "https://file.example.com", "settings file"},
		{"the environment", `server_url = "https://file.example.com"`, map[string]string{"STARTER_SERVER_URL": "https://env.example.com"}, "https://env.example.com", "environment (STARTER_SERVER_URL)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := ""
			if tt.settings != "" {
				file = writeSettings(t, tt.settings)
			}
			cfg, err := config.Load(config.Env{Getenv: envOf(tt.env)}, file)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if cfg.ServerURL != tt.wantURL || cfg.Sources["server_url"] != tt.wantSource {
				t.Errorf("ServerURL, source = %q, %q, want %q, %q", cfg.ServerURL, cfg.Sources["server_url"], tt.wantURL, tt.wantSource)
			}
		})
	}
}
