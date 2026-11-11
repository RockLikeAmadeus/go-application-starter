package memory

import (
	"context"
	"errors"
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
	account, ok := r.accountWith(identity)
	if !ok {
		return domain.Account{}, domain.ErrNotFound
	}
	return account, nil
}

// Find returns the account with the ID.
func (r *AccountRepository) Find(_ context.Context, id string) (domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	account, ok := r.accounts[id]
	if !ok {
		return domain.Account{}, domain.ErrNotFound
	}
	return account, nil
}

// Save stores the account, refusing another account with the same issuer and
// subject.
func (r *AccountRepository) Save(_ context.Context, account domain.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if stored, ok := r.accountWith(account.Identity); ok && stored.ID != account.ID {
		return errors.New("another account has the same issuer and subject")
	}
	r.accounts[account.ID] = account
	return nil
}

// accountWith returns the account with the identity's issuer and subject, if
// there is one. The caller holds the lock.
func (r *AccountRepository) accountWith(identity domain.Identity) (domain.Account, bool) {
	for _, account := range r.accounts {
		if account.Identity.Issuer == identity.Issuer && account.Identity.Subject == identity.Subject {
			return account, true
		}
	}
	return domain.Account{}, false
}
