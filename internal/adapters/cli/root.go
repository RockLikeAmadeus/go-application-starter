// Package cli is the Cobra driving adapter. Commands depend on
// ports.ItemService only; the composition root decides whether that is the
// embedded core or a remote server.
package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// AppFactory builds the application for a command. Tests substitute their own.
type AppFactory func(cfg config.Config) (*bootstrap.App, error)

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
func NewRootCmd(newApp AppFactory, env config.Env, out io.Writer) *cobra.Command {
	var (
		serverURL string
		dataDir   string
		storage   string
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
	root.PersistentFlags().StringVar(&storage, "storage", "", "storage adapter for embedded mode: jsonfile|memory (env STARTER_STORAGE)")

	// openApp loads config (the local settings file, then the environment),
	// applies flag overrides and wires the app.
	openApp := func() (*bootstrap.App, config.Config, error) {
		settingsFile, err := config.LocalConfigFile(env)
		if err != nil {
			return nil, config.Config{}, err
		}
		cfg, err := config.Load(env, settingsFile)
		if err != nil {
			return nil, cfg, err
		}
		if serverURL != "" {
			cfg.ServerURL = serverURL
		}
		if storage != "" {
			cfg.Storage = config.Storage(storage)
		}
		if dataDir != "" {
			cfg.DataDir = dataDir
		} else if keepsDataInAFolder(cfg) {
			cfg.DataDir, err = config.LocalDataDir(env, settingsFile)
			if err != nil {
				return nil, cfg, err
			}
		}
		if err := cfg.Validate(); err != nil {
			return nil, cfg, err
		}
		app, err := newApp(cfg)
		if err != nil {
			return nil, cfg, fmt.Errorf("start application: %w", err)
		}
		return app, cfg, nil
	}

	root.AddCommand(newTUICmd(openApp))
	root.AddCommand(newInfoCmd(openApp))
	// sample:start
	root.AddCommand(newItemCmd(openApp))
	// sample:end
	return root
}

// keepsDataInAFolder reports whether cfg keeps the data in a data folder on
// this computer: not on a server, and not only in memory.
func keepsDataInAFolder(cfg config.Config) bool {
	return cfg.ServerURL == "" && cfg.Storage == config.StorageJSONFile
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
