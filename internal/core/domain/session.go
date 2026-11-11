package domain

import "time"

// Session is a signed-in account's session, which lasts until ExpiresAt.
type Session struct {
	ExpiresAt time.Time
}

// IsExpiredAt reports whether the session has expired at now.
func (s Session) IsExpiredAt(now time.Time) bool {
	return !now.Before(s.ExpiresAt)
}
