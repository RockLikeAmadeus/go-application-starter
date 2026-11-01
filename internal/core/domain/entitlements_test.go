package domain_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

func entitlementsWithItemLimit(max int, whenOver domain.OverLimit) domain.Entitlements {
	return domain.Entitlements{
		Plan:   "free",
		Limits: map[string]domain.Limit{"items": {Name: "items", Max: max, WhenOver: whenOver}},
	}
}

// assertLimitError checks that err, returned by call, is want, or nil when
// want is nil.
func assertLimitError(t *testing.T, call string, err error, want *domain.LimitError) {
	t.Helper()
	if want == nil {
		if err != nil {
			t.Fatalf("%s = %v, want nil", call, err)
		}
		return
	}
	var got *domain.LimitError
	if !errors.As(err, &got) {
		t.Fatalf("%s = %v, want a *LimitError", call, err)
	}
	if *got != *want {
		t.Errorf("%s = %+v, want %+v", call, *got, *want)
	}
}

func TestEntitlements_CheckCreate(t *testing.T) {
	reached := &domain.LimitError{Limit: "items", Max: 3}
	cases := map[string]struct {
		entitlements domain.Entitlements
		count        int
		want         *domain.LimitError
	}{
		"below the limit":           {entitlementsWithItemLimit(3, domain.BlockNew), 2, nil},
		"at the limit":              {entitlementsWithItemLimit(3, domain.BlockNew), 3, reached},
		"over the limit":            {entitlementsWithItemLimit(3, domain.BlockNew), 4, reached},
		"at a read-only limit":      {entitlementsWithItemLimit(3, domain.ReadOnly), 3, reached},
		"a plan without that limit": {domain.Entitlements{Plan: "pro"}, 100, nil},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.entitlements.CheckCreate("items", tc.count)
			assertLimitError(t, fmt.Sprintf("CheckCreate(items, %d)", tc.count), err, tc.want)
		})
	}
}

func TestEntitlements_CheckChange(t *testing.T) {
	cases := map[string]struct {
		entitlements domain.Entitlements
		count        int
		want         *domain.LimitError
	}{
		"over a read-only limit":    {entitlementsWithItemLimit(3, domain.ReadOnly), 4, &domain.LimitError{Limit: "items", Max: 3, ReadOnly: true}},
		"at a read-only limit":      {entitlementsWithItemLimit(3, domain.ReadOnly), 3, nil},
		"over a block-new limit":    {entitlementsWithItemLimit(3, domain.BlockNew), 4, nil},
		"a plan without that limit": {domain.Entitlements{Plan: "pro"}, 100, nil},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.entitlements.CheckChange("items", tc.count)
			assertLimitError(t, fmt.Sprintf("CheckChange(items, %d)", tc.count), err, tc.want)
		})
	}
}

func TestLimitError_Message(t *testing.T) {
	cases := map[string]struct {
		err  *domain.LimitError
		want string
	}{
		"reached":   {&domain.LimitError{Limit: "items", Max: 3}, "the limit of 3 items is reached"},
		"read only": {&domain.LimitError{Limit: "items", Max: 3, ReadOnly: true}, "items are read only while over the limit of 3"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := tc.err.Error(); got != tc.want {
				t.Errorf("Error() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNewLimit(t *testing.T) {
	t.Run("a valid limit", func(t *testing.T) {
		got, err := domain.NewLimit("items", 3, domain.ReadOnly)
		if err != nil {
			t.Fatalf("NewLimit: %v", err)
		}
		want := domain.Limit{Name: "items", Max: 3, WhenOver: domain.ReadOnly}
		if got != want {
			t.Errorf("NewLimit = %+v, want %+v", got, want)
		}
	})

	invalid := map[string]struct {
		max      int
		whenOver domain.OverLimit
		field    string
	}{
		"a negative max":       {-1, domain.BlockNew, "max"},
		"an unknown behaviour": {3, "delete-all", "whenOver"},
		"no behaviour":         {3, "", "whenOver"},
	}
	for name, tc := range invalid {
		t.Run(name, func(t *testing.T) {
			_, err := domain.NewLimit("items", tc.max, tc.whenOver)
			var verr *domain.ValidationError
			if !errors.As(err, &verr) {
				t.Fatalf("NewLimit(items, %d, %q) = %v, want a *ValidationError", tc.max, tc.whenOver, err)
			}
			if verr.Field != tc.field {
				t.Errorf("Field = %q, want %q", verr.Field, tc.field)
			}
		})
	}
}
