package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// SignInService implements ports.SignInService.
type SignInService struct {
	accounts ports.AccountRepository
	sessions ports.SessionRepository
	codes    ports.SignInCodeRepository
	tokens   ports.TokenGenerator
	ids      ports.IDGenerator
	clock    ports.Clock
}

var _ ports.SignInService = (*SignInService)(nil)

// NewSignInService wires the use case to its driven ports.
func NewSignInService(
	accounts ports.AccountRepository,
	sessions ports.SessionRepository,
	codes ports.SignInCodeRepository,
	tokens ports.TokenGenerator,
	ids ports.IDGenerator,
	clock ports.Clock,
) *SignInService {
	return &SignInService{accounts: accounts, sessions: sessions, codes: codes, tokens: tokens, ids: ids, clock: clock}
}

// SignIn returns a new session's token for the identity's account.
func (s *SignInService) SignIn(ctx context.Context, identity domain.Identity) (string, error) {
	account, err := s.accountFor(ctx, identity)
	if err != nil {
		return "", err
	}
	return s.startSession(ctx, account.ID)
}

// startSession saves a new session for the account and returns its token.
func (s *SignInService) startSession(ctx context.Context, accountID string) (string, error) {
	token := s.tokens.NewToken()
	if err := s.sessions.Save(ctx, domain.NewSession(token, accountID, s.clock.Now())); err != nil {
		return "", fmt.Errorf("save session: %w", err)
	}
	return token, nil
}

// CurrentAccount returns the account whose session the token is. A token with
// no session, or an expired one, is signed out, which is domain.ErrNoWorkspace.
// An expired session is deleted when found.
func (s *SignInService) CurrentAccount(ctx context.Context, token string) (domain.Account, error) {
	session, err := s.sessions.FindByHash(ctx, domain.HashSessionToken(token))
	if errors.Is(err, domain.ErrNotFound) {
		return domain.Account{}, domain.ErrNoWorkspace
	}
	if err != nil {
		return domain.Account{}, fmt.Errorf("find session: %w", err)
	}
	if session.IsExpiredAt(s.clock.Now()) {
		// A failed delete only leaves an expired session behind, which is
		// still refused next time, so it doesn't change the answer.
		_ = s.sessions.Delete(ctx, session.TokenHash)
		return domain.Account{}, domain.ErrNoWorkspace
	}
	account, err := s.accounts.Find(ctx, session.AccountID)
	if err != nil {
		return domain.Account{}, fmt.Errorf("find session's account: %w", err)
	}
	return account, nil
}

// SignOut ends the token's session. A token with no session is already signed
// out, so it isn't an error.
func (s *SignInService) SignOut(ctx context.Context, token string) error {
	err := s.sessions.Delete(ctx, domain.HashSessionToken(token))
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// accountFor returns the identity's account, creating it and its personal
// workspace the first time the identity signs in. The name isn't used to find
// it, so a changed name still finds the same account, and the provider's
// current name is saved on it.
func (s *SignInService) accountFor(ctx context.Context, identity domain.Identity) (domain.Account, error) {
	account, err := s.accounts.FindByIdentity(ctx, identity)
	if err == nil {
		if account.Identity.Name != identity.Name {
			account.Identity.Name = identity.Name
			if err := s.accounts.Save(ctx, account); err != nil {
				return domain.Account{}, fmt.Errorf("save account's new name: %w", err)
			}
		}
		return account, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return domain.Account{}, fmt.Errorf("find account: %w", err)
	}
	account = domain.Account{ID: s.ids.NewID(), Identity: identity}
	account.WorkspaceID = domain.WorkspaceID(s.ids.NewID())
	if err := s.accounts.Save(ctx, account); err != nil {
		// Another sign-in of the same identity may have saved its account
		// since the lookup above, and Save refuses a second one, so share it.
		if existing, findErr := s.accounts.FindByIdentity(ctx, identity); findErr == nil {
			return existing, nil
		}
		return domain.Account{}, fmt.Errorf("save account: %w", err)
	}
	return account, nil
}

// IssueSignInCode returns a one-time code for the account of the session the
// token is, bound to the challenge.
func (s *SignInService) IssueSignInCode(ctx context.Context, token, challenge string) (string, error) {
	account, err := s.CurrentAccount(ctx, token)
	if err != nil {
		return "", err
	}
	code := s.tokens.NewToken()
	signInCode := domain.NewSignInCode(account.ID, challenge, s.clock.Now())
	signInCode.CodeHash = domain.HashSignInCode(code)
	if err := s.codes.Save(ctx, signInCode); err != nil {
		return "", fmt.Errorf("save sign-in code: %w", err)
	}
	return code, nil
}

// RedeemSignInCode swaps a code and its PKCE verifier for a new session token
// for the code's account.
func (s *SignInService) RedeemSignInCode(ctx context.Context, code, verifier string) (string, error) {
	signInCode, err := s.codes.Take(ctx, domain.HashSignInCode(code))
	if err != nil {
		return "", fmt.Errorf("take sign-in code: %w", err)
	}
	return s.startSession(ctx, signInCode.AccountID)
}
