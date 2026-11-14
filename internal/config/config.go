// Package config loads typed runtime configuration from an optional settings
// file (TOML) and the environment, which overrides it. Flags (CLI) may
// override these values after loading.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type StorageBackend string

const (
	StorageMemory   StorageBackend = "memory"
	StorageJSONFile StorageBackend = "jsonfile"
	StorageSQLite   StorageBackend = "sqlite"
	StoragePostgres StorageBackend = "postgres"
)

// BuiltInServerURL is the server address a build sets, as server_url's lowest
// layer: a settings file or STARTER_SERVER_URL still wins. Set it with
//
//	go build -ldflags "-X github.com/RockLikeAmadeus/go-application-starter/internal/config.BuiltInServerURL=<address>"
//
// A program built with one has no local mode, as an empty setting doesn't
// override it.
var BuiltInServerURL = ""

// CredentialStore says where the command line and the terminal application
// keep the session token they receive when you sign in to a server.
type CredentialStore string

const (
	// CredentialStoreSystem is the operating system's credential store, with
	// a file as the fallback.
	CredentialStoreSystem CredentialStore = "system"
	// CredentialStoreFile is a file in the settings folder only.
	CredentialStoreFile CredentialStore = "file"
)

// SignInMode says how people sign in to a hosted server, or that they don't.
type SignInMode string

const (
	// SignInOff is the local edition: no sign-in, and one workspace.
	SignInOff SignInMode = "off"
	// SignInDev signs people in by name alone, for development and tests
	// (builds with the devsignin tag).
	SignInDev SignInMode = "dev"
	// SignInOIDC signs people in through an OpenID Connect provider.
	SignInOIDC SignInMode = "oidc"
)

// IsOn says whether people sign in, which makes this the hosted edition. An
// unset mode is off, as the setting's default is.
func (m SignInMode) IsOn() bool {
	return m != "" && m != SignInOff
}

type Config struct {
	// Storage selects the storage adapter (storage; STARTER_STORAGE).
	Storage StorageBackend
	// DataDir is where the jsonfile adapter keeps its files (data_dir;
	// STARTER_DATA_DIR).
	DataDir string
	// DatabaseURL is where the postgres adapter connects (database_url;
	// STARTER_DATABASE_URL). It may hold a password.
	DatabaseURL string
	// Port is the HTTP listen port (port; PORT, the de-facto cloud convention).
	Port int
	// ServerURL, when set, makes CLI/TUI use a remote server (server_url;
	// STARTER_SERVER_URL).
	ServerURL string
	// CredentialStore says where the CLI and TUI keep their session tokens
	// (credential_store; STARTER_CREDENTIAL_STORE).
	CredentialStore CredentialStore
	// LogFormat is "json" or "text" (log_format; STARTER_LOG_FORMAT).
	LogFormat string
	// PlansFile replaces the plans file built into the binary (plans_file;
	// STARTER_PLANS_FILE). Empty means the built-in one.
	PlansFile string
	// SignIn says how people sign in to this server, which makes it the
	// hosted edition, or that they don't (sign_in; STARTER_SIGN_IN).
	SignIn SignInMode
	// PublicURL is the address people reach the server at, such as
	// https://app.example.com (public_url; STARTER_PUBLIC_URL).
	PublicURL string
	// OIDCIssuer is the sign-in provider's address (oidc_issuer;
	// STARTER_OIDC_ISSUER).
	OIDCIssuer string
	// OIDCClientID identifies this server to the provider (oidc_client_id;
	// STARTER_OIDC_CLIENT_ID).
	OIDCClientID string
	// OIDCClientSecret proves it to the provider (oidc_client_secret;
	// STARTER_OIDC_CLIENT_SECRET).
	OIDCClientSecret string
	// ShouldIgnoreSyncedFolder silences the warning that SQLite's data folder
	// looks synced (ignore_synced_folder; STARTER_IGNORE_SYNCED_FOLDER).
	ShouldIgnoreSyncedFolder bool
	// SettingsFile is the settings file Load read, or "" for none.
	SettingsFile string
	// Sources says where each setting came from, by its key in the settings
	// file.
	Sources map[string]string
}

// Env is the computer the app runs on, as configuration sees it. Getenv
// reads the environment (os.Getenv in production) and UserConfigDir finds
// the user's settings folder (os.UserConfigDir in production).
type Env struct {
	Getenv        func(string) string
	UserConfigDir func() (string, error)
}

// Load reads the configuration and validates it. Each setting comes from
// env's STARTER_* variable (or PORT) when set, else from its key in the TOML
// settingsFile, else from its default, and Sources records which.
// settingsFile "" means none.
func Load(env Env, settingsFile string) (Config, error) {
	fileSettings, err := readSettingsFile(settingsFile)
	if err != nil {
		return Config{}, err
	}
	sources := map[string]string{}
	setting := func(key, fallback string) string {
		if value := env.Getenv(envVars[key]); value != "" {
			sources[key] = "environment (" + envVars[key] + ")"
			return value
		}
		if value := fileSettings.String(key); value != "" {
			sources[key] = "settings file"
			return value
		}
		sources[key] = "built in"
		return fallback
	}
	cfg := Config{
		SettingsFile:     settingsFile,
		Sources:          sources,
		Storage:          StorageBackend(setting("storage", string(StorageJSONFile))),
		DataDir:          setting("data_dir", "./data"),
		DatabaseURL:      setting("database_url", ""),
		ServerURL:        setting("server_url", BuiltInServerURL),
		CredentialStore:  CredentialStore(setting("credential_store", string(CredentialStoreSystem))),
		LogFormat:        setting("log_format", "text"),
		PlansFile:        setting("plans_file", ""),
		SignIn:           SignInMode(setting("sign_in", string(SignInOff))),
		PublicURL:        setting("public_url", ""),
		OIDCIssuer:       setting("oidc_issuer", ""),
		OIDCClientID:     setting("oidc_client_id", ""),
		OIDCClientSecret: setting("oidc_client_secret", ""),
	}
	port, err := strconv.Atoi(setting("port", "8080"))
	if err != nil {
		return Config{}, fmt.Errorf("PORT: %w", err)
	}
	cfg.Port = port
	ignore, err := strconv.ParseBool(setting("ignore_synced_folder", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("STARTER_IGNORE_SYNCED_FOLDER: %w", err)
	}
	cfg.ShouldIgnoreSyncedFolder = ignore
	return cfg, cfg.Validate()
}

// Setting is one setting's key in the settings file, its value (a string, an
// int or a bool) and where it came from.
type Setting struct {
	Key    string
	Value  any
	Source string
}

// Settings lists c's settings, in the order a settings file would.
func (c Config) Settings() []Setting {
	values := []struct {
		key   string
		value any
	}{
		{"storage", string(c.Storage)},
		{"data_dir", c.DataDir},
		{"database_url", hiddenIfSet(c.DatabaseURL)},
		{"port", c.Port},
		{"server_url", c.ServerURL},
		{"credential_store", string(c.CredentialStore)},
		{"log_format", c.LogFormat},
		{"plans_file", c.PlansFile},
		{"sign_in", string(c.SignIn)},
		{"public_url", c.PublicURL},
		{"oidc_issuer", c.OIDCIssuer},
		{"oidc_client_id", c.OIDCClientID},
		{"oidc_client_secret", hiddenIfSet(c.OIDCClientSecret)},
		{"ignore_synced_folder", c.ShouldIgnoreSyncedFolder},
	}
	settings := make([]Setting, len(values))
	for i, v := range values {
		settings[i] = Setting{Key: v.key, Value: v.value, Source: c.Sources[v.key]}
	}
	return settings
}

// hiddenIfSet keeps a secret out of what Settings lists, while still showing
// whether it is set. A database URL usually holds a password.
func hiddenIfSet(secret string) string {
	if secret == "" {
		return ""
	}
	return "(hidden)"
}

func (c Config) Validate() error {
	switch c.Storage {
	case StorageMemory, StorageJSONFile, StorageSQLite, StoragePostgres:
	default:
		return UnrecognizedStorageError(c.Storage)
	}
	if c.CredentialStore != CredentialStoreSystem && c.CredentialStore != CredentialStoreFile {
		return fmt.Errorf("The credential store setting %q is not recognized. Please use %q or %q.", c.CredentialStore, CredentialStoreSystem, CredentialStoreFile)
	}
	if c.LogFormat != "json" && c.LogFormat != "text" {
		return fmt.Errorf("The log format setting %q is not recognized. Please use \"text\" or \"json\".", c.LogFormat)
	}
	if c.SignIn != SignInOff && c.SignIn != SignInDev && c.SignIn != SignInOIDC {
		return fmt.Errorf("The sign-in setting %q is not recognized. Please use %q, %q, or %q.", c.SignIn, SignInOff, SignInDev, SignInOIDC)
	}
	if c.SignIn == SignInOIDC {
		required := []struct{ key, value, hint string }{
			{"public_url", c.PublicURL, "the address people reach this server at, such as https://app.example.com"},
			{"oidc_issuer", c.OIDCIssuer, "the address of the sign-in provider"},
			{"oidc_client_id", c.OIDCClientID, "the client ID the sign-in provider gave this server"},
			{"oidc_client_secret", c.OIDCClientSecret, "the client secret the sign-in provider gave this server"},
		}
		for _, setting := range required {
			if setting.value == "" {
				return fmt.Errorf("The %s setting is required when the sign-in setting is %q. Please set it to %s.", setting.key, SignInOIDC, setting.hint)
			}
		}
	}
	if c.PublicURL != "" {
		u, err := url.Parse(c.PublicURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return errors.New("The public_url setting must be a full address starting with http:// or https://. Please set it to the address people reach this server at, such as https://app.example.com.")
		}
	}
	return nil
}

// UnrecognizedStorageError is the error for a storage setting this program
// doesn't have.
func UnrecognizedStorageError(storage StorageBackend) error {
	return fmt.Errorf("The storage setting %q is not recognized. Please use %q, %q, %q, or %q.", storage, StorageMemory, StorageJSONFile, StorageSQLite, StoragePostgres)
}

// envVars are the settings, by their key in a settings file, each with the
// environment variable that overrides it.
var envVars = map[string]string{
	"storage":              "STARTER_STORAGE",
	"data_dir":             "STARTER_DATA_DIR",
	"database_url":         "STARTER_DATABASE_URL",
	"port":                 "PORT",
	"server_url":           "STARTER_SERVER_URL",
	"credential_store":     "STARTER_CREDENTIAL_STORE",
	"log_format":           "STARTER_LOG_FORMAT",
	"plans_file":           "STARTER_PLANS_FILE",
	"sign_in":              "STARTER_SIGN_IN",
	"public_url":           "STARTER_PUBLIC_URL",
	"oidc_issuer":          "STARTER_OIDC_ISSUER",
	"oidc_client_id":       "STARTER_OIDC_CLIENT_ID",
	"oidc_client_secret":   "STARTER_OIDC_CLIENT_SECRET",
	"ignore_synced_folder": "STARTER_IGNORE_SYNCED_FOLDER",
}

// readSettingsFile reads the TOML settings file at path, refusing any key
// that isn't a setting. An empty path reads nothing.
func readSettingsFile(path string) (*koanf.Koanf, error) {
	fileSettings := koanf.New(".")
	if path == "" {
		return fileSettings, nil
	}
	if err := fileSettings.Load(file.Provider(path), toml.Parser()); err != nil {
		return nil, fmt.Errorf("The settings file %s could not be read: %w", path, err)
	}
	for _, key := range fileSettings.Keys() {
		if _, ok := envVars[key]; !ok {
			return nil, fmt.Errorf("The settings file %s contains an unknown setting: %q.", path, key)
		}
	}
	return fileSettings, nil
}
