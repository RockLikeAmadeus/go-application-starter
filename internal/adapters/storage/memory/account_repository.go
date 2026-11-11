package memory

import (
	"context"
	"sync"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// AccountRepository stores accounts.
type AccountRepository struct {
	mu       sync.RWMutex
	accounts map[string]domain.Account
}

var _ ports.AccountRepository = (*AccountRepository)(nil)

// NewAccountRepository returns an empty repository.
func NewAccountRepository() *AccountRepository {
	return &AccountRepository{accounts: map[string]domain.Account{}}
}

// FindByIdentity returns the account with the identity's issuer and subject.
func (r *AccountRepository) FindByIdentity(_ context.Context, identity domain.Identity) (domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, account := range r.accounts {
		if account.Identity.Issuer == identity.Issuer && account.Identity.Subject == identity.Subject {
			return account, nil
		}
	}
	return domain.Account{}, domain.ErrNotFound
}

// Find returns the account with the ID.
func (r *AccountRepository) Find(_ context.Context, _ string) (domain.Account, error) {
	return domain.Account{}, nil
}

// Save stores the account.
func (r *AccountRepository) Save(_ context.Context, account domain.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.accounts[account.ID] = account
	return nil
}
