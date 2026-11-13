package domain_test

import (
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

func TestSignInCode_VerifiedByTheVerifierBehindItsChallenge(t *testing.T) {
	// The example verifier and S256 challenge from RFC 7636, appendix B.
	const verifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	const challenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"

	cases := map[string]struct {
		challenge string
		verifier  string
		want      bool
	}{
		"the matching verifier":     {challenge, verifier, true},
		"a different verifier":      {challenge, verifier + "x", false},
		"an empty verifier":         {challenge, "", false},
		"an empty challenge":        {"", verifier, false},
		"an empty pair":             {"", "", false},
		"the challenge as verifier": {challenge, challenge, false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			code := domain.SignInCode{Challenge: tc.challenge}
			if got := code.VerifiedBy(tc.verifier); got != tc.want {
				t.Errorf("VerifiedBy(%q) = %v, want %v", tc.verifier, got, tc.want)
			}
		})
	}
}

func TestSignInCode_ExpiresItsLifetimeAfterItsIssued(t *testing.T) {
	issued := time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC)
	code := domain.NewSignInCode("account-1", "challenge", issued)

	cases := map[string]struct {
		at   time.Time
		want bool
	}{
		"when just issued":           {issued, false},
		"just before its lifetime":   {issued.Add(domain.SignInCodeLifetime - time.Nanosecond), false},
		"at the end of its lifetime": {issued.Add(domain.SignInCodeLifetime), true},
		"after its lifetime":         {issued.Add(domain.SignInCodeLifetime + time.Second), true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := code.IsExpiredAt(tc.at); got != tc.want {
				t.Errorf("IsExpiredAt(%v) = %v, want %v", tc.at, got, tc.want)
			}
		})
	}
	if domain.SignInCodeLifetime != 5*time.Minute {
		t.Errorf("SignInCodeLifetime = %v, want 5m", domain.SignInCodeLifetime)
	}
}
