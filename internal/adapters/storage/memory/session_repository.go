package memory

import (
	"context"
	"sync"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// SessionRepository stores sessions by their token's hash.
type SessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]domain.Session
}

var _ ports.SessionRepository = (*SessionRepository)(nil)

// NewSessionRepository returns an empty repository.
func NewSessionRepository() *SessionRepository {
	return &SessionRepository{sessions: map[string]domain.Session{}}
}

// FindByHash returns the session kept by the token hash.
func (r *SessionRepository) FindByHash(_ context.Context, tokenHash string) (domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, ok := r.sessions[tokenHash]
	if !ok {
		return domain.Session{}, domain.ErrNotFound
	}
	return session, nil
}

// Save stores the session.
func (r *SessionRepository) Save(_ context.Context, session domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.TokenHash] = session
	return nil
}

// Delete removes the session kept by the token hash.
func (r *SessionRepository) Delete(_ context.Context, _ string) error {
	return nil
}
