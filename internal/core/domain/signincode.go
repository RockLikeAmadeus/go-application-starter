package domain

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"time"
)

// SignInCodeLifetime is how long a sign-in code can be swapped for a session
// after it's issued.
const SignInCodeLifetime = 5 * time.Minute

// SignInCode is a one-time code that a browser sign-in hands to the command
// line, to be swapped for a session. It's bound to the PKCE challenge of the
// client that asked for it, and to the account that signed in.
type SignInCode struct {
	AccountID string
	Challenge string
	ExpiresAt time.Time
}

// NewSignInCode issues a code for the account at now, bound to the challenge.
func NewSignInCode(accountID, challenge string, now time.Time) SignInCode {
	return SignInCode{AccountID: accountID, Challenge: challenge, ExpiresAt: now.Add(SignInCodeLifetime)}
}

// IsExpiredAt reports whether the code has expired at now.
func (c SignInCode) IsExpiredAt(now time.Time) bool {
	return !now.Before(c.ExpiresAt)
}

// VerifiedBy reports whether the verifier matches the code's challenge, as
// PKCE's S256 method (RFC 7636) defines: the challenge is the unpadded
// base64url of the verifier's SHA-256. An empty verifier or challenge never
// matches.
func (c SignInCode) VerifiedBy(verifier string) bool {
	if verifier == "" || c.Challenge == "" {
		return false
	}
	hash := sha256.Sum256([]byte(verifier))
	want := base64.RawURLEncoding.EncodeToString(hash[:])
	return subtle.ConstantTimeCompare([]byte(want), []byte(c.Challenge)) == 1
}
