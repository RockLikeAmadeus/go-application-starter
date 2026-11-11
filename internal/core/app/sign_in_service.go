package app

import (
	"context"
	"fmt"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// SignInService implements ports.SignInService.
type SignInService struct {
	accounts ports.AccountRepository
	sessions ports.SessionRepository
	tokens   ports.TokenGenerator
	ids      ports.IDGenerator
	clock    ports.Clock
}

var _ ports.SignInService = (*SignInService)(nil)

// NewSignInService wires the use case to its driven ports.
func NewSignInService(accounts ports.AccountRepository, sessions ports.SessionRepository, tokens ports.TokenGenerator, ids ports.IDGenerator, clock ports.Clock) *SignInService {
	return &SignInService{accounts: accounts, sessions: sessions, tokens: tokens, ids: ids, clock: clock}
}

// SignIn returns a new session's token for the identity's account.
func (s *SignInService) SignIn(ctx context.Context, identity domain.Identity) (string, error) {
	account := domain.Account{ID: s.ids.NewID(), Identity: identity}
	account.WorkspaceID = domain.WorkspaceID(s.ids.NewID())
	if err := s.accounts.Save(ctx, account); err != nil {
		return "", fmt.Errorf("save account: %w", err)
	}
	token := s.tokens.NewToken()
	if err := s.sessions.Save(ctx, domain.NewSession(token, account.ID, s.clock.Now())); err != nil {
		return "", fmt.Errorf("save session: %w", err)
	}
	return token, nil
}
