package plans_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/plans"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/testsupport"
)

func TestParse_GivesEachPlansEntitlements(t *testing.T) {
	doc := `{
		"plans": {
			"free": {"limits": {"items": {"max": 3, "whenOver": "read-only"}}},
			"pro": {"limits": {}}
		}
	}`

	got, err := plans.Parse([]byte(doc))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := plans.Catalog{
		"free": {
			Plan:     "free",
			Limits:   map[string]domain.Limit{"items": {Name: "items", Max: 3, WhenOver: domain.ReadOnly}},
			Features: map[string]domain.Feature{},
		},
		"pro": {Plan: "pro", Limits: map[string]domain.Limit{}, Features: map[string]domain.Feature{}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse =\n%+v\nwant\n%+v", got, want)
	}
}

func TestParse_IgnoresFullLineComments(t *testing.T) {
	doc := `// The product's plans.
	{
		"plans": {
			// The free plan.
			"free": {
				"limits": {
					// The items limit.
					"items": {"max": 3, "whenOver": "block-new"}
					// The end of the limits.
				}
			}
		}
	}`

	got, err := plans.Parse([]byte(doc))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := plans.Catalog{"free": {
		Plan:     "free",
		Limits:   map[string]domain.Limit{"items": {Name: "items", Max: 3, WhenOver: domain.BlockNew}},
		Features: map[string]domain.Feature{},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse =\n%+v\nwant\n%+v", got, want)
	}
}

func TestParse_ReadsFeatures(t *testing.T) {
	features := `"features": {
		"export": {"name": "Item export", "whenOff": "show-locked"},
		"audit": {"name": "Audit log", "whenOff": "hide"}
	}`
	t.Run("gives every plan every feature, on when it lists it", func(t *testing.T) {
		doc := `{` + features + `, "plans": {
			"free": {"limits": {}},
			"pro": {"limits": {}, "features": ["export", "audit"]}
		}}`

		got, err := plans.Parse([]byte(doc))
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		catalogFeatures := func(isOn bool) map[string]domain.Feature {
			return map[string]domain.Feature{
				"export": {Key: "export", Name: "Item export", IsOn: isOn, WhenOff: domain.ShowLocked},
				"audit":  {Key: "audit", Name: "Audit log", IsOn: isOn, WhenOff: domain.Hide},
			}
		}
		want := plans.Catalog{
			"free": {Plan: "free", Limits: map[string]domain.Limit{}, Features: catalogFeatures(false)},
			"pro":  {Plan: "pro", Limits: map[string]domain.Limit{}, Features: catalogFeatures(true)},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Parse =\n%+v\nwant\n%+v", got, want)
		}
	})

	refused := map[string]struct {
		doc  string
		want []string // each must appear in the error
	}{
		"a plan listing an unknown feature": {
			`{` + features + `, "plans": {"pro": {"limits": {}, "features": ["export", "sso"]}}}`,
			[]string{`plan "pro"`, `feature "sso"`},
		},
		"a feature with no name": {
			`{"features": {"export": {"whenOff": "hide"}}, "plans": {}}`,
			[]string{`feature "export"`, "name"},
		},
		"a feature with a bad whenOff": {
			`{"features": {"export": {"name": "Item export", "whenOff": "grey-out"}}, "plans": {}}`,
			[]string{`feature "export"`, "whenOff"},
		},
	}
	for name, tc := range refused {
		t.Run("refuses "+name, func(t *testing.T) {
			_, err := plans.Parse([]byte(tc.doc))
			if err == nil {
				t.Fatal("Parse succeeded, want an error")
			}
			for _, want := range tc.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q doesn't mention %q", err, want)
				}
			}
		})
	}
}

func TestParse_RefusesABadPlansFile(t *testing.T) {
	cases := map[string]struct {
		doc  string
		want []string // each must appear in the error
	}{
		"invalid JSON": {`{"plans": {`, []string{"plans file"}},
		"a negative max": {
			`{"plans": {"free": {"limits": {"items": {"max": -1, "whenOver": "block-new"}}}}}`,
			[]string{`plan "free"`, `limit "items"`, "max"},
		},
		"an unknown over-limit behaviour": {
			`{"plans": {"free": {"limits": {"items": {"max": 3, "whenOver": "delete-all"}}}}}`,
			[]string{`plan "free"`, `limit "items"`, "whenOver"},
		},
		"an unknown field": {
			`{"plans": {"free": {"limits": {"items": {"maximum": 3, "whenOver": "block-new"}}}}}`,
			[]string{`unknown field "maximum"`},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := plans.Parse([]byte(tc.doc))
			if err == nil {
				t.Fatal("Parse succeeded, want an error")
			}
			for _, want := range tc.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q doesn't mention %q", err, want)
				}
			}
		})
	}
}

func TestEntitlements_AnswersThePlanTheWorkspaceIsOn(t *testing.T) {
	free := domain.Entitlements{
		Plan:   "free",
		Limits: map[string]domain.Limit{"items": {Name: "items", Max: 3, WhenOver: domain.BlockNew}},
	}
	pro := domain.Entitlements{Plan: "pro", Limits: map[string]domain.Limit{}}
	onPlan := map[domain.WorkspaceID]string{"ws-a": "free", "ws-b": "pro"}
	adapter := plans.Provider{
		Catalog: plans.Catalog{"free": free, "pro": pro},
		PlanOf: func(_ context.Context, ws domain.WorkspaceID) (string, error) {
			return onPlan[ws], nil
		},
	}

	got, err := adapter.For(context.Background(), "ws-a")
	if err != nil {
		t.Fatalf("For: %v", err)
	}
	if !reflect.DeepEqual(got, free) {
		t.Errorf("For(ws-a) = %+v, want %+v", got, free)
	}
}

func TestEntitlements_FillsTheUpgradeHintWhenHintsAreOn(t *testing.T) {
	feature := func(isOn bool) map[string]domain.Feature {
		return map[string]domain.Feature{
			"export": {Key: "export", Name: "Item export", IsOn: isOn, WhenOff: domain.ShowLocked},
		}
	}
	catalog := plans.Catalog{
		"free":  {Plan: "free", Features: feature(false)},
		"team":  {Plan: "team", Features: feature(true)},
		"pro":   {Plan: "pro", Features: feature(true)},
		"local": {Plan: "local", Features: feature(true)},
	}
	onFree := func(context.Context, domain.WorkspaceID) (string, error) { return "free", nil }
	cases := map[string]struct {
		shouldGiveHints bool
		want            []string
	}{
		"on: the sorted plans with it, without the local plan": {shouldGiveHints: true, want: []string{"pro", "team"}},
		"off: no hint": {shouldGiveHints: false, want: nil},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			adapter := plans.Provider{Catalog: catalog, PlanOf: onFree, ShouldGiveHints: tc.shouldGiveHints, LocalPlan: "local"}

			got, err := adapter.For(context.Background(), "ws-a")
			if err != nil {
				t.Fatalf("For: %v", err)
			}
			if availableOn := got.Features["export"].AvailableOn; !reflect.DeepEqual(availableOn, tc.want) {
				t.Errorf("AvailableOn = %#v, want %#v", availableOn, tc.want)
			}
		})
	}
}

func TestEntitlements_RefusesAPlanItCantAnswer(t *testing.T) {
	lookupFailed := errors.New("lookup failed")
	cases := map[string]struct {
		planOf func(context.Context, domain.WorkspaceID) (string, error)
		check  func(t *testing.T, err error)
	}{
		"a plan the catalog doesn't have": {
			planOf: func(context.Context, domain.WorkspaceID) (string, error) { return "gold", nil },
			check: func(t *testing.T, err error) {
				if err == nil || !strings.Contains(err.Error(), `plan "gold"`) {
					t.Errorf("err = %v, want one naming plan \"gold\"", err)
				}
			},
		},
		"PlanOf failing": {
			planOf: func(context.Context, domain.WorkspaceID) (string, error) { return "", lookupFailed },
			check: func(t *testing.T, err error) {
				if !errors.Is(err, lookupFailed) {
					t.Errorf("err = %v, want %v", err, lookupFailed)
				}
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			adapter := plans.Provider{Catalog: plans.Catalog{"free": {Plan: "free"}}, PlanOf: tc.planOf}
			_, err := adapter.For(context.Background(), "ws-a")
			tc.check(t, err)
		})
	}
}

func TestHasPlan_AnswersWhetherTheCatalogHasThePlan(t *testing.T) {
	adapter := plans.Provider{Catalog: plans.Catalog{"free": {Plan: "free"}}}

	for plan, want := range map[string]bool{"free": true, "platinum": false} {
		got, err := adapter.HasPlan(context.Background(), plan)
		if err != nil {
			t.Fatalf("HasPlan(%q): %v", plan, err)
		}
		if got != want {
			t.Errorf("HasPlan(%q) = %v, want %v", plan, got, want)
		}
	}
}

func TestStoredPlanOf_AnswersTheWorkspacesStoredPlan(t *testing.T) {
	stored := testsupport.StoredPlans{"ws-a": "pro"}

	got, err := plans.StoredPlanOf(stored, "free")(context.Background(), "ws-a")
	if err != nil {
		t.Fatalf("PlanOf: %v", err)
	}
	if got != "pro" {
		t.Errorf("PlanOf = %q, want %q", got, "pro")
	}
}

func TestStoredPlanOf_AnswersTheDefaultForAWorkspaceWithNoStoredPlan(t *testing.T) {
	got, err := plans.StoredPlanOf(testsupport.StoredPlans{}, "free")(context.Background(), "ws-a")
	if err != nil {
		t.Fatalf("PlanOf: %v", err)
	}
	if got != "free" {
		t.Errorf("PlanOf = %q, want %q", got, "free")
	}
}

// errPlans is the error unreadablePlans fails with.
var errPlans = errors.New("plans unavailable")

// unreadablePlans is a ports.WorkspacePlanRepository whose read always fails.
type unreadablePlans struct{ testsupport.StoredPlans }

func (unreadablePlans) PlanOf(context.Context, domain.WorkspaceID) (string, error) {
	return "", errPlans
}

func TestStoredPlanOf_PassesOnAnyOtherErrorInsteadOfTheDefault(t *testing.T) {
	got, err := plans.StoredPlanOf(unreadablePlans{}, "free")(context.Background(), "ws-a")

	if !errors.Is(err, errPlans) {
		t.Fatalf("PlanOf = %q, %v, want the repository's error", got, err)
	}
}

func TestDefault_HasALocalPlanWithNoLimits(t *testing.T) {
	catalog, err := plans.Default()
	if err != nil {
		t.Fatalf("Default: %v", err)
	}
	local, ok := catalog["local"]
	if !ok {
		t.Fatalf("the shipped plans file has no local plan; it has %v", catalog)
	}
	if len(local.Limits) != 0 {
		t.Errorf("local plan limits = %v, want none", local.Limits)
	}
}
