package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// newInfoCmd shows where the settings and the data are: it opens the app, as
// any command that uses the data does, and names the settings file it read
// (or none), the data folder it opened (or the server it uses, or says the
// data is only in memory) and the workspace's plan.
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
			where, place := dataLocation(cfg)
			settingsFile := cfg.SettingsFile
			if settingsFile == "" {
				settingsFile = "none"
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Settings file: %s\n%s: %s\nPlan: %s\n", settingsFile, where, place, entitlements.Plan)
			return err
		},
	}
}

// dataLocation says where cfg keeps the data, as a label and a place: a data
// folder, a server, or only memory.
func dataLocation(cfg config.Config) (where, place string) {
	switch {
	case cfg.ServerURL != "":
		return "Server", cfg.ServerURL
	case cfg.Storage == config.StorageMemory:
		return "Data", "kept in memory, not saved"
	}
	return "Data folder", cfg.DataDir
}
