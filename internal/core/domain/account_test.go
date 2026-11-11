package domain_test

import (
	"errors"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

func TestNewIdentity_RefusesABlankIssuerOrSubject(t *testing.T) {
	invalid := map[string]struct {
		issuer, subject string
		field           string
	}{
		"an empty issuer":      {"", "alpha", "issuer"},
		"a whitespace issuer":  {"  ", "alpha", "issuer"},
		"an empty subject":     {"dev", "", "subject"},
		"a whitespace subject": {"dev", "\t", "subject"},
	}
	for name, tc := range invalid {
		t.Run(name, func(t *testing.T) {
			_, err := domain.NewIdentity(tc.issuer, tc.subject, "Alpha")
			var verr *domain.ValidationError
			if !errors.As(err, &verr) {
				t.Fatalf("NewIdentity(%q, %q, Alpha) = %v, want a *ValidationError", tc.issuer, tc.subject, err)
			}
			if verr.Field != tc.field {
				t.Errorf("Field = %q, want %q", verr.Field, tc.field)
			}
		})
	}
}
