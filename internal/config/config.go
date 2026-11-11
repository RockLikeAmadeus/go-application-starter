// Package config loads typed runtime configuration from an optional settings
// file (TOML) and the environment, which overrides it. Flags (CLI) may
// override these values after loading.
package config

import (
	"fmt"
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

// SignInMode says how people sign in to a hosted server, or that they don't.
type SignInMode string

const (
	// SignInOff is the local edition: no sign-in, and one workspace.
	SignInOff SignInMode = "off"
	// SignInDev signs people in by name alone, for development and tests
	// (builds with the devsignin tag).
	SignInDev SignInMode = "dev"
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
	// LogFormat is "json" or "text" (log_format; STARTER_LOG_FORMAT).
	LogFormat string
	// PlansFile replaces the plans file built into the binary (plans_file;
	// STARTER_PLANS_FILE). Empty means the built-in one.
	PlansFile string
	// SignIn says how people sign in to this server, which makes it the
	// hosted edition, or that they don't (sign_in; STARTER_SIGN_IN).
	SignIn SignInMode
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
		SettingsFile: settingsFile,
		Sources:      sources,
		Storage:      StorageBackend(setting("storage", string(StorageJSONFile))),
		DataDir:      setting("data_dir", "./data"),
		DatabaseURL:  setting("database_url", ""),
		ServerURL:    setting("server_url", ""),
		LogFormat:    setting("log_format", "text"),
		PlansFile:    setting("plans_file", ""),
		SignIn:       SignInMode(setting("sign_in", string(SignInOff))),
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
		{"log_format", c.LogFormat},
		{"plans_file", c.PlansFile},
		{"sign_in", string(c.SignIn)},
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
	if c.LogFormat != "json" && c.LogFormat != "text" {
		return fmt.Errorf("The log format setting %q is not recognized. Please use \"text\" or \"json\".", c.LogFormat)
	}
	if c.SignIn != SignInOff && c.SignIn != SignInDev {
		return fmt.Errorf("The sign-in setting %q is not recognized. Please use %q or %q.", c.SignIn, SignInOff, SignInDev)
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
	"log_format":           "STARTER_LOG_FORMAT",
	"plans_file":           "STARTER_PLANS_FILE",
	"sign_in":              "STARTER_SIGN_IN",
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
