package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// SessionLifetime is how long a session lasts from sign-in.
const SessionLifetime = 30 * 24 * time.Hour

// Session is a signed-in account's session, which lasts until ExpiresAt. It's
// kept by its token's hash, never the token itself, so stored sessions can't
// be used to sign in.
type Session struct {
	TokenHash string
	AccountID string
	ExpiresAt time.Time
}

// NewSession starts the account's session for the token at now.
func NewSession(token, accountID string, now time.Time) Session {
	return Session{TokenHash: HashSessionToken(token), AccountID: accountID, ExpiresAt: now.Add(SessionLifetime)}
}

// HashSessionToken is the hash a session is kept by: the token's SHA-256, in
// hex.
func HashSessionToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// IsExpiredAt reports whether the session has expired at now.
func (s Session) IsExpiredAt(now time.Time) bool {
	return !now.Before(s.ExpiresAt)
}
