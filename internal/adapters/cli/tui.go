package cli

import (
	"github.com/spf13/cobra"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/tui"
)

// runTUI starts the terminal UI. Tests substitute their own, so they can see
// what the command hands it without drawing on a terminal.
var runTUI = tui.Run

func newTUICmd(open appOpener, newSignIn SignInFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive terminal UI",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			app, cfg, err := open()
			if err != nil {
				return err
			}
			defer app.Close()
			ports := tui.Ports{
				// sample:start
				Items: app.Items,
				// sample:end
			}
			if cfg.ServerURL != "" {
				client, openBrowser, err := newSignIn(cfg)
				if err != nil {
					return err
				}
				ports.SignIn, ports.ServerURL, ports.OpenBrowser = client, cfg.ServerURL, openBrowser
			}
			return runTUI(ports, app.Warnings)
		},
	}
}
