// Package system provides production implementations of the Clock and
// IDGenerator driven ports.
package system

import (
	"time"

	"github.com/google/uuid"
)

// Clock is a ports.Clock backed by the wall clock, in UTC.
type Clock struct{}

func (Clock) Now() time.Time { return time.Now().UTC() }

// UUIDs is a ports.IDGenerator producing time-ordered UUIDv7 strings.
type UUIDs struct{}

func (UUIDs) NewID() string { return uuid.Must(uuid.NewV7()).String() }
