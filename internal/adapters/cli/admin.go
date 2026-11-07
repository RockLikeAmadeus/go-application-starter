package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// newAdminCmd groups the commands for someone running the product.
func newAdminCmd(loadConfig func() (config.Config, error), newBilling BillingFactory) *cobra.Command {
	admin := &cobra.Command{
		Use:   "admin",
		Short: "Commands for product administrators",
	}
	admin.AddCommand(&cobra.Command{
		Use:   "set-plan WORKSPACE PLAN",
		Short: "Move a workspace to another plan",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			billing, err := newBilling(cfg)
			if err != nil {
				return err
			}
			workspace, plan := args[0], args[1]
			if err := billing.ChangePlan(cmd.Context(), domain.WorkspaceID(workspace), plan); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s is now on the %s plan\n", workspace, plan)
			return nil
		},
	})
	return admin
}
