package domain_test

import (
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

func TestSession_IsExpiredAtOrAfterItsExpiry(t *testing.T) {
	expiresAt := time.Date(2026, 12, 10, 9, 0, 0, 0, time.UTC)
	session := domain.Session{ExpiresAt: expiresAt}

	cases := map[string]struct {
		at   time.Time
		want bool
	}{
		"before it expires": {expiresAt.Add(-time.Nanosecond), false},
		"as it expires":     {expiresAt, true},
		"after it expires":  {expiresAt.Add(time.Hour), true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := session.IsExpiredAt(tc.at); got != tc.want {
				t.Errorf("IsExpiredAt(%v) = %v, want %v", tc.at, got, tc.want)
			}
		})
	}
}
