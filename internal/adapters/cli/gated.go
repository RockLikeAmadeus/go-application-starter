package cli

import (
	"github.com/spf13/cobra"

	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// featureAnnotation is the Cobra annotation that holds a gated command's
// feature key.
const featureAnnotation = "feature"

// Gated marks cmd as gated by the plan's feature key, and returns it. The
// help lists it by the feature's state: plainly when on, with a note when
// locked, and not at all when hidden. Gating is only shown here: running cmd
// when the feature is off fails in the core, with a *domain.FeatureError.
func Gated(key string, cmd *cobra.Command) *cobra.Command {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[featureAnnotation] = key
	return cmd
}

// noteGatedCommands makes the help, for root and every command under it,
// list each gated subcommand by its feature's state. It reads only the
// entitlements, never opening the app, so help works with a locked or
// missing data folder. When they can't be read, gated commands are listed
// plainly. It changes the commands it notes, which is fine because the
// program ends after the help.
func noteGatedCommands(root *cobra.Command, loadConfig func() (config.Config, error)) {
	help := root.HelpFunc()
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if e, ok := readEntitlements(cmd, loadConfig); ok {
			for _, child := range cmd.Commands() {
				if key, gated := child.Annotations[featureAnnotation]; gated {
					note(child, e.Features[key])
				}
			}
		}
		help(cmd, args)
	})
}

// note shows cmd's feature f in the help: nothing when it's on, hidden when
// f hides it, and otherwise noted as disabled, with the upgrade hint when
// there is one.
func note(cmd *cobra.Command, f domain.Feature) {
	switch {
	case f.On:
	case f.WhenOff == domain.Hide:
		cmd.Hidden = true
	case len(f.AvailableOn) > 0:
		cmd.Short += " (disabled on your plan: available on " + domain.PlansPhrase(f.AvailableOn) + ")"
	default:
		cmd.Short += " (disabled on your plan)"
	}
}

// readEntitlements reads the plan's entitlements through bootstrap's light
// path (the plans file, or the server's), and reports whether it could.
func readEntitlements(cmd *cobra.Command, loadConfig func() (config.Config, error)) (domain.Entitlements, bool) {
	cfg, err := loadConfig()
	if err != nil {
		return domain.Entitlements{}, false
	}
	service, err := bootstrap.NewEntitlements(cfg)
	if err != nil {
		return domain.Entitlements{}, false
	}
	e, err := service.CurrentEntitlements(cmd.Context())
	if err != nil {
		return domain.Entitlements{}, false
	}
	return e, true
}
