package memory

import (
	"context"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// AccountRepository stores accounts.
type AccountRepository struct{}

var _ ports.AccountRepository = (*AccountRepository)(nil)

// NewAccountRepository returns an empty repository.
func NewAccountRepository() *AccountRepository {
	return &AccountRepository{}
}

// FindByIdentity returns the account with the identity's issuer and subject.
func (r *AccountRepository) FindByIdentity(_ context.Context, _ domain.Identity) (domain.Account, error) {
	return domain.Account{}, domain.ErrNotFound
}

// Find returns the account with the ID.
func (r *AccountRepository) Find(_ context.Context, _ string) (domain.Account, error) {
	return domain.Account{}, nil
}

// Save stores the account.
func (r *AccountRepository) Save(_ context.Context, _ domain.Account) error {
	return nil
}
