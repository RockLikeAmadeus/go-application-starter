package sqlstore

import (
	"context"
	"database/sql"
	"errors"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// SignInCodeRepository stores sign-in codes, in the sign_in_codes table.
type SignInCodeRepository struct {
	db *sql.DB
}

var _ ports.SignInCodeRepository = (*SignInCodeRepository)(nil)

// NewSignInCodeRepository returns a repository built on db, which Migrate has
// brought up to date.
func NewSignInCodeRepository(db *sql.DB) *SignInCodeRepository {
	return &SignInCodeRepository{db: db}
}

// Save stores the code.
func (r *SignInCodeRepository) Save(ctx context.Context, code domain.SignInCode) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sign_in_codes (code_hash, account_id, challenge, expires_at) VALUES ($1, $2, $3, $4)
		ON CONFLICT (code_hash) DO UPDATE SET account_id = excluded.account_id,
			challenge = excluded.challenge, expires_at = excluded.expires_at`,
		code.CodeHash, code.AccountID, code.Challenge, formatTime(code.ExpiresAt))
	return err
}

// Take returns the code kept by the hash and deletes it, in one statement so
// that two servers can't both redeem it.
func (r *SignInCodeRepository) Take(ctx context.Context, codeHash string) (domain.SignInCode, error) {
	var code domain.SignInCode
	var expiresAt string
	err := r.db.QueryRowContext(ctx,
		`DELETE FROM sign_in_codes WHERE code_hash = $1 RETURNING code_hash, account_id, challenge, expires_at`, codeHash).
		Scan(&code.CodeHash, &code.AccountID, &code.Challenge, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.SignInCode{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.SignInCode{}, err
	}
	if code.ExpiresAt, err = parseTime(expiresAt); err != nil {
		return domain.SignInCode{}, err
	}
	return code, nil
}
