package memory

import (
	"context"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// SessionRepository stores sessions by their token's hash.
type SessionRepository struct{}

var _ ports.SessionRepository = (*SessionRepository)(nil)

// NewSessionRepository returns an empty repository.
func NewSessionRepository() *SessionRepository {
	return &SessionRepository{}
}

// FindByHash returns the session kept by the token hash.
func (r *SessionRepository) FindByHash(_ context.Context, _ string) (domain.Session, error) {
	return domain.Session{}, domain.ErrNotFound
}

// Save stores the session.
func (r *SessionRepository) Save(_ context.Context, _ domain.Session) error {
	return nil
}

// Delete removes the session kept by the token hash.
func (r *SessionRepository) Delete(_ context.Context, _ string) error {
	return nil
}
