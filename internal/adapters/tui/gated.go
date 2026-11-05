package tui

import "github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"

// gatedKeyHelp shows binding ("x export") in a screen's help line by its
// feature's state on the plan: plainly when on, marked locked (with the
// upgrade hint, when there is one) when locked, and empty when hidden, so the
// caller leaves it out. Gating is only shown here: pressing the key when the
// feature is off fails in the core, with a *domain.FeatureError.
func gatedKeyHelp(binding string, feature domain.Feature) string {
	switch {
	case feature.IsOn:
		return binding
	case feature.WhenOff == domain.Hide:
		return ""
	case len(feature.AvailableOn) > 0:
		return binding + " (locked: available on " + domain.PlansPhrase(feature.AvailableOn) + ")"
	default:
		return binding + " (locked)"
	}
}
