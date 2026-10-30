package cli

import (
	"github.com/spf13/cobra"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/tui"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
)

func newTUICmd(open appOpener) *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive terminal UI",
		Args:  cobra.NoArgs,
		RunE: withApp(open, func(_ *cobra.Command, _ []string, app *bootstrap.App) error {
			return tui.Run(tui.Ports{
				// sample:start
				Items: app.Items,
				// sample:end
			})
		}),
	}
}
