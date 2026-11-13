package domain_test

import (
	"testing"

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
