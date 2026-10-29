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

// Version is set at build time via -ldflags "-X .../cli.Version=v1.2.3".
var Version = "dev"

// NewRootCmd builds the command tree.
func NewRootCmd(newApp AppFactory, out io.Writer) *cobra.Command {
	var (
		serverURL string
		dataDir   string
		storage   string
	)

	root := &cobra.Command{
		Use:           "starter",
		Short:         "Command-line interface for the starter application",
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(out)
	root.PersistentFlags().StringVar(&serverURL, "server", "", "use a remote server at this URL instead of the embedded core (env STARTER_SERVER_URL)")
	root.PersistentFlags().StringVar(&dataDir, "data-dir", "", "data directory for embedded mode (env STARTER_DATA_DIR)")
	root.PersistentFlags().StringVar(&storage, "storage", "", "storage adapter for embedded mode: jsonfile|memory (env STARTER_STORAGE)")

	// openApp loads config, applies flag overrides and wires the app.
	openApp := func() (*bootstrap.App, error) {
		cfg, err := config.Load()
		if err != nil {
			return nil, err
		}
		if serverURL != "" {
			cfg.ServerURL = serverURL
		}
		if dataDir != "" {
			cfg.DataDir = dataDir
		}
		if storage != "" {
			cfg.Storage = config.Storage(storage)
		}
		if err := cfg.Validate(); err != nil {
			return nil, err
		}
		app, err := newApp(cfg)
		if err != nil {
			return nil, fmt.Errorf("start application: %w", err)
		}
		return app, nil
	}

	root.AddCommand(newItemCmd(openApp), newTUICmd(openApp))
	return root
}
