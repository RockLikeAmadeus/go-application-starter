package sqlstore

import (
	"context"
	"database/sql"
	"errors"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

const accountColumns = `id, issuer, subject, name, workspace_id`

// AccountRepository stores accounts, in the accounts table.
type AccountRepository struct {
	db *sql.DB
}

var _ ports.AccountRepository = (*AccountRepository)(nil)

// NewAccountRepository returns a repository built on db, which Migrate has
// brought up to date.
func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// FindByIdentity returns the account with the identity's issuer and subject.
func (r *AccountRepository) FindByIdentity(ctx context.Context, identity domain.Identity) (domain.Account, error) {
	return r.findOne(ctx, `WHERE issuer = $1 AND subject = $2`, identity.Issuer, identity.Subject)
}

// Find returns the account with the ID.
func (r *AccountRepository) Find(ctx context.Context, id string) (domain.Account, error) {
	return r.findOne(ctx, `WHERE id = $1`, id)
}

// Save stores the account.
func (r *AccountRepository) Save(ctx context.Context, account domain.Account) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO accounts (`+accountColumns+`) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET issuer = excluded.issuer, subject = excluded.subject,
			name = excluded.name, workspace_id = excluded.workspace_id`,
		account.ID, account.Identity.Issuer, account.Identity.Subject, account.Identity.Name, account.WorkspaceID)
	return err
}

// findOne returns the one account that where (a WHERE clause) matches, or
// domain.ErrNotFound.
func (r *AccountRepository) findOne(ctx context.Context, where string, args ...any) (domain.Account, error) {
	var account domain.Account
	err := r.db.QueryRowContext(ctx, `SELECT `+accountColumns+` FROM accounts `+where, args...).
		Scan(&account.ID, &account.Identity.Issuer, &account.Identity.Subject, &account.Identity.Name, &account.WorkspaceID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Account{}, domain.ErrNotFound
	}
	return account, err
}
