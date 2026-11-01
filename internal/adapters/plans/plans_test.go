package plans_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/plans"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
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
			Plan:   "free",
			Limits: map[string]domain.Limit{"items": {Name: "items", Max: 3, WhenOver: domain.ReadOnly}},
		},
		"pro": {Plan: "pro", Limits: map[string]domain.Limit{}},
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
		Plan:   "free",
		Limits: map[string]domain.Limit{"items": {Name: "items", Max: 3, WhenOver: domain.BlockNew}},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse =\n%+v\nwant\n%+v", got, want)
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
	adapter := plans.Entitlements{
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
			adapter := plans.Entitlements{Catalog: plans.Catalog{"free": {Plan: "free"}}, PlanOf: tc.planOf}
			_, err := adapter.For(context.Background(), "ws-a")
			tc.check(t, err)
		})
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
