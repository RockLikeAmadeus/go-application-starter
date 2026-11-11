package app_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/app"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
	"github.com/RockLikeAmadeus/go-application-starter/internal/testsupport"
)

func TestSignIn_WithANewIdentityCreatesAnAccountWorkspaceAndThirtyDaySession(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	accounts := testsupport.StoredAccounts{}
	sessions := testsupport.StoredSessions{}
	svc := app.NewSignInService(accounts, sessions, testsupport.FixedToken("token-1"),
		testsupport.NewSequentialIDs("id"), testsupport.NewFixedClock(now))
	identity := domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada"}

	token, err := svc.SignIn(context.Background(), identity)
	if err != nil {
		t.Fatalf("SignIn: %v", err)
	}

	if token != "token-1" {
		t.Errorf("token = %q, want %q", token, "token-1")
	}
	wantAccounts := testsupport.StoredAccounts{
		"id-1": {ID: "id-1", Identity: identity, WorkspaceID: "id-2"},
	}
	if !reflect.DeepEqual(accounts, wantAccounts) {
		t.Errorf("accounts = %v, want %v", accounts, wantAccounts)
	}
	hash := sha256.Sum256([]byte("token-1"))
	tokenHash := hex.EncodeToString(hash[:])
	wantSessions := testsupport.StoredSessions{
		tokenHash: {TokenHash: tokenHash, AccountID: "id-1", ExpiresAt: now.Add(30 * 24 * time.Hour)},
	}
	if !reflect.DeepEqual(sessions, wantSessions) {
		t.Errorf("sessions = %v, want %v", sessions, wantSessions)
	}
}

func TestSignIn_WithAKnownIdentityReusesItsAccountEvenWhenTheNameHasChanged(t *testing.T) {
	known := domain.Account{ID: "account-1", Identity: domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada"}, WorkspaceID: "workspace-1"}
	accounts := testsupport.StoredAccounts{known.ID: known}
	sessions := testsupport.StoredSessions{}
	svc := app.NewSignInService(accounts, sessions, testsupport.FixedToken("token-1"),
		testsupport.NewSequentialIDs("id"), testsupport.NewFixedClock(time.Now()))

	_, err := svc.SignIn(context.Background(), domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada Lovelace"})
	if err != nil {
		t.Fatalf("SignIn: %v", err)
	}

	if len(accounts) != 1 || accounts[known.ID].WorkspaceID != known.WorkspaceID {
		t.Errorf("accounts = %v, want only %q with workspace %q", accounts, known.ID, known.WorkspaceID)
	}
	session := sessions[domain.HashSessionToken("token-1")]
	if session.AccountID != known.ID {
		t.Errorf("session's account = %q, want %q", session.AccountID, known.ID)
	}
}

// errAccounts is the error unfindableAccounts' FindByIdentity fails with.
var errAccounts = errors.New("accounts unavailable")

// unfindableAccounts is an AccountRepository whose FindByIdentity always
// fails; its other calls go to the repository it wraps.
type unfindableAccounts struct{ ports.AccountRepository }

func (unfindableAccounts) FindByIdentity(context.Context, domain.Identity) (domain.Account, error) {
	return domain.Account{}, errAccounts
}

func TestSignIn_WhenTheAccountCantBeLookedUpIsRefusedAndSavesNothing(t *testing.T) {
	accounts := testsupport.StoredAccounts{}
	sessions := testsupport.StoredSessions{}
	svc := app.NewSignInService(unfindableAccounts{accounts}, sessions, testsupport.FixedToken("token-1"),
		testsupport.NewSequentialIDs("id"), testsupport.NewFixedClock(time.Now()))

	_, err := svc.SignIn(context.Background(), domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada"})

	if !errors.Is(err, errAccounts) {
		t.Fatalf("err = %v, want %v", err, errAccounts)
	}
	if len(accounts) != 0 || len(sessions) != 0 {
		t.Errorf("accounts = %v, sessions = %v, want none saved", accounts, sessions)
	}
}

func TestCurrentAccount_WithASignedInTokenReturnsTheSessionsAccount(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	ada := domain.Account{ID: "account-1", Identity: domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada"}, WorkspaceID: "workspace-1"}
	grace := domain.Account{ID: "account-2", Identity: domain.Identity{Issuer: "dev", Subject: "grace", Name: "Grace"}, WorkspaceID: "workspace-2"}
	accounts := testsupport.StoredAccounts{ada.ID: ada, grace.ID: grace}
	sessions := testsupport.StoredSessions{}
	for token, account := range map[string]domain.Account{"token-ada": ada, "token-grace": grace} {
		session := domain.NewSession(token, account.ID, now)
		sessions[session.TokenHash] = session
	}
	svc := app.NewSignInService(accounts, sessions, testsupport.FixedToken("unused"),
		testsupport.NewSequentialIDs("id"), testsupport.NewFixedClock(now.Add(time.Hour)))

	got, err := svc.CurrentAccount(context.Background(), "token-grace")
	if err != nil {
		t.Fatalf("CurrentAccount: %v", err)
	}

	if !reflect.DeepEqual(got, grace) {
		t.Errorf("account = %v, want %v", got, grace)
	}
}

func TestCurrentAccount_WithAnUnknownTokenHasNoWorkspace(t *testing.T) {
	svc := app.NewSignInService(testsupport.StoredAccounts{}, testsupport.StoredSessions{}, testsupport.FixedToken("unused"),
		testsupport.NewSequentialIDs("id"), testsupport.NewFixedClock(time.Now()))

	_, err := svc.CurrentAccount(context.Background(), "token-unknown")

	if !errors.Is(err, domain.ErrNoWorkspace) {
		t.Fatalf("err = %v, want %v", err, domain.ErrNoWorkspace)
	}
}

func TestCurrentAccount_WithAnExpiredTokenHasNoWorkspaceAndDeletesTheSession(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	ada := domain.Account{ID: "account-1", Identity: domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada"}, WorkspaceID: "workspace-1"}
	session := domain.NewSession("token-ada", ada.ID, now)
	sessions := testsupport.StoredSessions{session.TokenHash: session}
	svc := app.NewSignInService(testsupport.StoredAccounts{ada.ID: ada}, sessions, testsupport.FixedToken("unused"),
		testsupport.NewSequentialIDs("id"), testsupport.NewFixedClock(session.ExpiresAt))

	_, err := svc.CurrentAccount(context.Background(), "token-ada")

	if !errors.Is(err, domain.ErrNoWorkspace) {
		t.Fatalf("err = %v, want %v", err, domain.ErrNoWorkspace)
	}
	if len(sessions) != 0 {
		t.Errorf("sessions = %v, want the expired one deleted", sessions)
	}
}

func TestSignOut_DeletesTheSessionSoItsTokenIsSignedOut(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	ada := domain.Account{ID: "account-1", Identity: domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada"}, WorkspaceID: "workspace-1"}
	session := domain.NewSession("token-ada", ada.ID, now)
	sessions := testsupport.StoredSessions{session.TokenHash: session}
	svc := app.NewSignInService(testsupport.StoredAccounts{ada.ID: ada}, sessions, testsupport.FixedToken("unused"),
		testsupport.NewSequentialIDs("id"), testsupport.NewFixedClock(now.Add(time.Hour)))

	if err := svc.SignOut(context.Background(), "token-ada"); err != nil {
		t.Fatalf("SignOut: %v", err)
	}

	if len(sessions) != 0 {
		t.Errorf("sessions = %v, want the session deleted", sessions)
	}
	if _, err := svc.CurrentAccount(context.Background(), "token-ada"); !errors.Is(err, domain.ErrNoWorkspace) {
		t.Errorf("CurrentAccount after SignOut: err = %v, want %v", err, domain.ErrNoWorkspace)
	}
}
