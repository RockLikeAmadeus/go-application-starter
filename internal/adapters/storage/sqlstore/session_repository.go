package sqlstore

import (
	"context"
	"database/sql"
	"errors"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// SessionRepository stores sessions, in the sessions table.
type SessionRepository struct {
	db *sql.DB
}

var _ ports.SessionRepository = (*SessionRepository)(nil)

// NewSessionRepository returns a repository built on db, which Migrate has
// brought up to date.
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// FindByHash returns the session kept by the token hash.
func (r *SessionRepository) FindByHash(ctx context.Context, tokenHash string) (domain.Session, error) {
	var session domain.Session
	var expiresAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT token_hash, account_id, expires_at FROM sessions WHERE token_hash = $1`, tokenHash).
		Scan(&session.TokenHash, &session.AccountID, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Session{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Session{}, err
	}
	if session.ExpiresAt, err = parseTime(expiresAt); err != nil {
		return domain.Session{}, err
	}
	return session, nil
}

// Save stores the session.
func (r *SessionRepository) Save(ctx context.Context, session domain.Session) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sessions (token_hash, account_id, expires_at) VALUES ($1, $2, $3)
		ON CONFLICT (token_hash) DO UPDATE SET account_id = excluded.account_id, expires_at = excluded.expires_at`,
		session.TokenHash, session.AccountID, formatTime(session.ExpiresAt))
	return err
}

// Delete removes the session kept by the token hash.
func (r *SessionRepository) Delete(ctx context.Context, tokenHash string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash)
	if err != nil {
		return err
	}
	deletedCount, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if deletedCount == 0 {
		return domain.ErrNotFound
	}
	return nil
}
