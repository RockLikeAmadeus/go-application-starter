package plans_test

import (
	"reflect"
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
