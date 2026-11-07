package tui

import (
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// A gated key's help shows the key by its feature's state on the plan:
// plainly when on, marked locked (with the upgrade hint, if there is one)
// when locked, and not at all when hidden.
func TestGatedKeyHelpShowsTheKeyByItsFeaturesState(t *testing.T) {
	cases := map[string]struct {
		feature domain.Feature
		want    string
	}{
		"on": {
			domain.Feature{Key: "item-export", Name: "Item export", IsOn: true, WhenOff: domain.ShowLocked, AvailableOn: []string{"pro"}},
			"x export",
		},
		"locked, with one plan to name": {
			domain.Feature{Key: "item-export", Name: "Item export", WhenOff: domain.ShowLocked, AvailableOn: []string{"pro"}},
			"x export (locked: available on the pro plan)",
		},
		"locked, with several plans to name": {
			domain.Feature{Key: "item-export", Name: "Item export", WhenOff: domain.ShowLocked, AvailableOn: []string{"pro", "team"}},
			"x export (locked: available on the following plans: pro or team)",
		},
		"locked, with no plan to name": {
			domain.Feature{Key: "item-export", Name: "Item export", WhenOff: domain.ShowLocked},
			"x export (locked)",
		},
		"hidden": {
			domain.Feature{Key: "item-export", Name: "Item export", WhenOff: domain.Hide, AvailableOn: []string{"pro"}},
			"",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := gatedKeyHelp("x export", tc.feature); got != tc.want {
				t.Errorf("gatedKeyHelp = %q, want %q", got, tc.want)
			}
		})
	}
}
