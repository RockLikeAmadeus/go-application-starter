package cli

import (
	"github.com/spf13/cobra"

	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// newAdminCmd groups the commands for someone running the product.
func newAdminCmd(loadConfig func() (config.Config, error), newBilling BillingFactory) *cobra.Command {
	admin := &cobra.Command{
		Use:   "admin",
		Short: "Commands for whoever runs the product",
	}
	admin.AddCommand(&cobra.Command{
		Use:   "set-plan WORKSPACE PLAN",
		Short: "Move a workspace to another plan",
		RunE: func(*cobra.Command, []string) error {
			cfg, _ := loadConfig()
			_, err := newBilling(cfg)
			return err
		},
	})
	return admin
}
