package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// newInfoCmd shows where the settings and the data are: it opens the app, as
// any command that uses the data does, and names the settings file it read
// (or none), lists each setting with where it came from, says where the data
// is (see dataLocation) and names the workspace's plan.
func newInfoCmd(open appOpener) *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show the data location and the current plan",
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
			label, location := dataLocation(cfg)
			settingsFile := cfg.SettingsFile
			if settingsFile == "" {
				settingsFile = "none"
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Settings file: %s\n", settingsFile)
			for _, s := range cfg.Settings() {
				fmt.Fprintf(out, "  %s = %s  # %s\n", s.Key, tomlValue(s.Value), s.Source)
			}
			_, err = fmt.Fprintf(out, "%s: %s\nPlan: %s\n", label, location, entitlements.Plan)
			return err
		},
	}
}

// dataLocation says where cfg keeps the data, as a label and a location: a data
// folder, a server, only memory, or a Postgres database, named by its setting
// because its URL usually holds a password.
func dataLocation(cfg config.Config) (label, location string) {
	switch {
	case cfg.ServerURL != "":
		return "Server", cfg.ServerURL
	case cfg.Storage == config.StorageMemory:
		return "Data", "kept in memory, not saved"
	case cfg.Storage == config.StoragePostgres:
		return "Data", "kept in the Postgres database at database_url"
	}
	return "Data folder", cfg.DataDir
}

// tomlValue writes a setting's value as a settings file would: a string
// quoted, a number or a bool as it is.
func tomlValue(v any) string {
	if s, ok := v.(string); ok {
		return strconv.Quote(s)
	}
	return fmt.Sprint(v)
}
