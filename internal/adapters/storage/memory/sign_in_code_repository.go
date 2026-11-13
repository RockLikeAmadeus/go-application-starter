package memory

import (
	"context"
	"sync"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// SignInCodeRepository stores sign-in codes by their hash.
type SignInCodeRepository struct {
	mu    sync.Mutex
	codes map[string]domain.SignInCode
}

var _ ports.SignInCodeRepository = (*SignInCodeRepository)(nil)

// NewSignInCodeRepository returns an empty repository.
func NewSignInCodeRepository() *SignInCodeRepository {
	return &SignInCodeRepository{codes: map[string]domain.SignInCode{}}
}

// Save stores the code.
func (r *SignInCodeRepository) Save(_ context.Context, code domain.SignInCode) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.codes[code.CodeHash] = code
	return nil
}

// Take returns the code kept by the hash and deletes it under one lock, so
// concurrent takes can't both get it.
func (r *SignInCodeRepository) Take(_ context.Context, codeHash string) (domain.SignInCode, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	code, ok := r.codes[codeHash]
	if !ok {
		return domain.SignInCode{}, domain.ErrNotFound
	}
	delete(r.codes, codeHash)
	return code, nil
}
