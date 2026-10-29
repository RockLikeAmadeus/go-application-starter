// Package testsupport provides deterministic fakes for driven ports, shared
// by unit tests and acceptance tests.
package testsupport

import (
	"fmt"
	"sync"
	"time"
)

// FixedClock is a ports.Clock that only moves when told to.
type FixedClock struct {
	mu  sync.Mutex
	now time.Time
}

func NewFixedClock(t time.Time) *FixedClock { return &FixedClock{now: t} }

func (c *FixedClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *FixedClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// SequentialIDs is a ports.IDGenerator producing "<prefix>-1", "<prefix>-2", ...
type SequentialIDs struct {
	mu     sync.Mutex
	prefix string
	n      int
}

func NewSequentialIDs(prefix string) *SequentialIDs { return &SequentialIDs{prefix: prefix} }

func (g *SequentialIDs) NewID() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.n++
	return fmt.Sprintf("%s-%d", g.prefix, g.n)
}
