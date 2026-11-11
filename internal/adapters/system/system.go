// Package system provides production implementations of the Clock,
// IDGenerator, and TokenGenerator driven ports.
package system

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
)

// Clock is a ports.Clock backed by the wall clock, in UTC.
type Clock struct{}

func (Clock) Now() time.Time { return time.Now().UTC() }

// UUIDs is a ports.IDGenerator producing time-ordered UUIDv7 strings.
type UUIDs struct{}

func (UUIDs) NewID() string { return uuid.Must(uuid.NewV7()).String() }

// Tokens is a ports.TokenGenerator making 32 bytes from crypto/rand, written
// as unpadded base64url so a token is safe in a cookie or a header.
type Tokens struct{}

func (Tokens) NewToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
