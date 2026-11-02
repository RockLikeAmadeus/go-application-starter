package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newInfoCmd shows where the data is: it opens the app, as any command that
// uses the data does, and names the data folder it opened and the
// workspace's plan.
func newInfoCmd(open appOpener) *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show where the data is, and the plan",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, cfg, err := open()
			if err != nil {
				return err
			}
			defer app.Close()
			entitlements, err := app.Entitlements.CurrentEntitlements(cmd.Context())
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Data folder: %s\nPlan: %s\n", cfg.DataDir, entitlements.Plan)
			return err
		},
	}
}
