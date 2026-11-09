// Package cli is the Cobra driving adapter. Commands depend on the driving
// ports only; the composition root decides whether those are the embedded
// core or a remote server.
package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// AppFactory builds the application for a command. Tests substitute their own.
type AppFactory func(cfg config.Config) (*bootstrap.App, error)

// BillingFactory builds the billing service for `starter admin set-plan`.
// Tests substitute their own.
type BillingFactory func(cfg config.Config) (ports.BillingService, error)

// appOpener opens the application, and says which configuration it opened
// it with. Commands call it only when they run, so --help and --version work
// without wiring anything.
type appOpener func() (*bootstrap.App, config.Config, error)

// Version is set at build time via -ldflags "-X .../cli.Version=v1.2.3".
var Version = "dev"

// NewRootCmd builds the command tree. env is the computer the CLI runs on
// (config.Env, with os.Getenv and os.UserConfigDir in production): its
// environment, and its local data folder, which commands use unless
// --data-dir names another.
func NewRootCmd(newApp AppFactory, newBilling BillingFactory, env config.Env, out io.Writer) *cobra.Command {
	var (
		serverURL  string
		dataDir    string
		storage    string
		configFile string
	)

	root := &cobra.Command{
		Use:           "starter",
		Short:         "Command-line interface for Starter",
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(out)
	root.PersistentFlags().StringVar(&serverURL, "server", "", "use a remote server at this URL instead of the embedded core (env STARTER_SERVER_URL)")
	root.PersistentFlags().StringVar(&dataDir, "data-dir", "", "data folder for embedded mode (default: the local data folder the desktop app uses; env STARTER_DATA_DIR)")
	root.PersistentFlags().StringVar(&configFile, "config", "", "settings file to read (default: config.toml in the app's settings folder; env STARTER_CONFIG_FILE)")
	root.PersistentFlags().StringVar(&storage, "storage", "", "storage adapter for embedded mode: jsonfile|sqlite|memory (env STARTER_STORAGE)")

	// loadConfig loads config (the local settings file, then the
	// environment) and applies flag overrides.
	loadConfig := func() (config.Config, error) {
		settingsFile, err := settingsFileFor(configFile, env)
		if err != nil {
			return config.Config{}, err
		}
		cfg, err := config.Load(env, settingsFile)
		if err != nil {
			return cfg, err
		}
		if serverURL != "" {
			cfg.ServerURL = serverURL
			chosenByOption(cfg, "server_url", "server")
		}
		if storage != "" {
			cfg.Storage = config.StorageBackend(storage)
			chosenByOption(cfg, "storage", "storage")
		}
		if dataDir != "" {
			cfg.DataDir = dataDir
			chosenByOption(cfg, "data_dir", "data-dir")
		} else if keepsDataInAFolder(cfg) {
			cfg.DataDir, err = config.LocalDataDir(env, settingsFile)
			if err != nil {
				return cfg, err
			}
		}
		if err := cfg.Validate(); err != nil {
			return cfg, err
		}
		return cfg, nil
	}

	// openApp loads config, wires the app and prints its warnings.
	openApp := func() (*bootstrap.App, config.Config, error) {
		cfg, err := loadConfig()
		if err != nil {
			return nil, cfg, err
		}
		app, err := newApp(cfg)
		if locked := bootstrap.LockedFolderError(err, "connect to a running server with --server"); locked != nil {
			return nil, cfg, locked
		}
		if err != nil {
			return nil, cfg, err
		}
		printWarnings(root.ErrOrStderr(), app.Warnings)
		return app, cfg, nil
	}

	noteGatedCommands(root, loadConfig)

	root.AddCommand(newTUICmd(openApp))
	root.AddCommand(newInfoCmd(openApp))
	root.AddCommand(newAdminCmd(loadConfig, newBilling))
	// sample:start
	root.AddCommand(newItemCmd(openApp))
	// sample:end
	return root
}

// chosenByOption records that the option --flag chose cfg's setting key.
func chosenByOption(cfg config.Config, key, flag string) {
	cfg.Sources[key] = "option (--" + flag + ")"
}

// settingsFileFor returns the settings file a command reads: the one
// --config names, else the local one (config.LocalConfigFile), which it
// first creates when it's missing (config.CreateDefaultSettingsFile).
func settingsFileFor(configFile string, env config.Env) (string, error) {
	if configFile != "" {
		return configFile, nil
	}
	config.CreateDefaultSettingsFile(env)
	return config.LocalConfigFile(env)
}

// keepsDataInAFolder reports whether cfg keeps the data in a data folder on
// this computer: not on a server, and not only in memory.
func keepsDataInAFolder(cfg config.Config) bool {
	return cfg.ServerURL == "" && (cfg.Storage == config.StorageJSONFile || cfg.Storage == config.StorageSQLite)
}

// Run runs the command tree and returns the program's exit status. A failed
// command's error goes on root's stderr, so a command's output, which scripts
// read, stays as it is.
func Run(ctx context.Context, root *cobra.Command) int {
	if err := root.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(root.ErrOrStderr(), "Error:", err)
		return 1
	}
	return 0
}

// printWarnings prints each of the app's warnings as it is given, on errOut
// (stderr), so a command's output, which scripts read, stays as it is.
func printWarnings(errOut io.Writer, warnings []string) {
	for _, w := range warnings {
		fmt.Fprintln(errOut, "Warning:", w)
	}
}

// withApp opens the app for the duration of a command.
func withApp(open appOpener, fn func(cmd *cobra.Command, args []string, app *bootstrap.App) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		app, _, err := open()
		if err != nil {
			return err
		}
		defer app.Close()
		return fn(cmd, args, app)
	}
}
