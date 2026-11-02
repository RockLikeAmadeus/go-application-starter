package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// newInfoCmd shows where the data is: it opens the app, as any command that
// uses the data does, and names the data folder it opened (or the server it
// uses, or says the data is only in memory) and the workspace's plan.
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
			where, place := "Data folder", cfg.DataDir
			switch {
			case cfg.ServerURL != "":
				where, place = "Server", cfg.ServerURL
			case cfg.Storage == config.StorageMemory:
				where, place = "Data", "kept in memory, not saved"
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\nPlan: %s\n", where, place, entitlements.Plan)
			return err
		},
	}
}
