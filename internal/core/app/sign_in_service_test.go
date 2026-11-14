package app_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/memory"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/app"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
	"github.com/RockLikeAmadeus/go-application-starter/internal/testsupport"
)

func TestSignIn_WithANewIdentityCreatesAnAccountWorkspaceAndThirtyDaySession(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	accounts := testsupport.StoredAccounts{}
	sessions := testsupport.StoredSessions{}
	svc := app.NewSignInService(accounts, sessions, nil, testsupport.FixedToken("token-1"),
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
	svc := app.NewSignInService(accounts, sessions, nil, testsupport.FixedToken("token-1"),
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

func TestSignIn_WithAKnownIdentityWhoseNameChangedSavesTheNewName(t *testing.T) {
	known := domain.Account{ID: "account-1", Identity: domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada"}, WorkspaceID: "workspace-1"}
	accounts := testsupport.StoredAccounts{known.ID: known}
	svc := app.NewSignInService(accounts, testsupport.StoredSessions{}, nil, testsupport.FixedToken("token-1"),
		testsupport.NewSequentialIDs("id"), testsupport.NewFixedClock(time.Now()))

	token, err := svc.SignIn(context.Background(), domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada Lovelace"})
	if err != nil {
		t.Fatalf("SignIn: %v", err)
	}

	account, err := svc.CurrentAccount(context.Background(), token)
	if err != nil {
		t.Fatalf("CurrentAccount: %v", err)
	}
	if account.Identity.Name != "Ada Lovelace" {
		t.Errorf("name = %q, want %q", account.Identity.Name, "Ada Lovelace")
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
	svc := app.NewSignInService(unfindableAccounts{accounts}, sessions, nil, testsupport.FixedToken("token-1"),
		testsupport.NewSequentialIDs("id"), testsupport.NewFixedClock(time.Now()))

	_, err := svc.SignIn(context.Background(), domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada"})

	if !errors.Is(err, errAccounts) {
		t.Fatalf("err = %v, want %v", err, errAccounts)
	}
	if len(accounts) != 0 || len(sessions) != 0 {
		t.Errorf("accounts = %v, sessions = %v, want none saved", accounts, sessions)
	}
}

// raceLostAccounts is a StoredAccounts where another sign-in of the same
// identity saves its account, the rival, just after this one's first
// FindByIdentity finds none. Like the real repositories, its Save refuses
// another account with the same issuer and subject.
type raceLostAccounts struct {
	testsupport.StoredAccounts
	rival         domain.Account
	hasRivalSaved bool
}

func (a *raceLostAccounts) FindByIdentity(ctx context.Context, identity domain.Identity) (domain.Account, error) {
	account, err := a.StoredAccounts.FindByIdentity(ctx, identity)
	if !a.hasRivalSaved {
		a.hasRivalSaved = true
		a.StoredAccounts[a.rival.ID] = a.rival
	}
	return account, err
}

func (a *raceLostAccounts) Save(ctx context.Context, account domain.Account) error {
	if stored, err := a.StoredAccounts.FindByIdentity(ctx, account.Identity); err == nil && stored.ID != account.ID {
		return errors.New("another account has the same issuer and subject")
	}
	return a.StoredAccounts.Save(ctx, account)
}

func TestSignIn_WhenAnotherSignInSavesTheIdentitysAccountFirstSharesThatAccount(t *testing.T) {
	identity := domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada"}
	rival := domain.Account{ID: "account-rival", Identity: identity, WorkspaceID: "workspace-rival"}
	accounts := &raceLostAccounts{StoredAccounts: testsupport.StoredAccounts{}, rival: rival}
	sessions := testsupport.StoredSessions{}
	svc := app.NewSignInService(accounts, sessions, nil, testsupport.FixedToken("token-1"),
		testsupport.NewSequentialIDs("id"), testsupport.NewFixedClock(time.Now()))

	token, err := svc.SignIn(context.Background(), identity)
	if err != nil {
		t.Fatalf("SignIn: %v", err)
	}

	if want := (testsupport.StoredAccounts{rival.ID: rival}); !reflect.DeepEqual(accounts.StoredAccounts, want) {
		t.Errorf("accounts = %v, want only the rival %v", accounts.StoredAccounts, want)
	}
	if got := sessions[domain.HashSessionToken(token)].AccountID; got != rival.ID {
		t.Errorf("session's account = %q, want the rival %q", got, rival.ID)
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
	svc := app.NewSignInService(accounts, sessions, nil, testsupport.FixedToken("unused"),
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
	svc := app.NewSignInService(testsupport.StoredAccounts{}, testsupport.StoredSessions{}, nil, testsupport.FixedToken("unused"),
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
	svc := app.NewSignInService(testsupport.StoredAccounts{ada.ID: ada}, sessions, nil, testsupport.FixedToken("unused"),
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
	svc := app.NewSignInService(testsupport.StoredAccounts{ada.ID: ada}, sessions, nil, testsupport.FixedToken("unused"),
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

func TestSignOut_OfAnUnknownSessionIsNotAnError(t *testing.T) {
	svc := app.NewSignInService(testsupport.StoredAccounts{}, testsupport.StoredSessions{}, nil, testsupport.FixedToken("unused"),
		testsupport.NewSequentialIDs("id"), testsupport.NewFixedClock(time.Now()))

	if err := svc.SignOut(context.Background(), "token-unknown"); err != nil {
		t.Errorf("SignOut: %v, want no error", err)
	}
}

func TestIssueSignInCode_KeepsTheCodesHashForTheSessionsAccountAndReturnsTheCode(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	ada := domain.Account{
		ID:          "account-1",
		Identity:    domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada"},
		WorkspaceID: "workspace-1",
	}
	session := domain.NewSession("token-ada", ada.ID, now)
	sessions := testsupport.StoredSessions{session.TokenHash: session}
	codes := memory.NewSignInCodeRepository()
	svc := app.NewSignInService(testsupport.StoredAccounts{ada.ID: ada}, sessions, codes, testsupport.FixedToken("code-1"),
		testsupport.NewSequentialIDs("id"), testsupport.NewFixedClock(now))

	code, err := svc.IssueSignInCode(context.Background(), "token-ada", "challenge-1")
	if err != nil {
		t.Fatalf("IssueSignInCode: %v", err)
	}

	if code != "code-1" {
		t.Errorf("code = %q, want %q", code, "code-1")
	}
	got, err := codes.Take(context.Background(), domain.HashSignInCode("code-1"))
	if err != nil {
		t.Fatalf("Take: %v", err)
	}
	want := domain.SignInCode{
		CodeHash:  domain.HashSignInCode("code-1"),
		AccountID: ada.ID,
		Challenge: "challenge-1",
		ExpiresAt: now.Add(domain.SignInCodeLifetime),
	}
	if got != want {
		t.Errorf("kept code = %v, want %v", got, want)
	}
}

func TestIssueSignInCode_RefusesAnUnknownOrExpiredSessionAndKeepsNoCode(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	ada := domain.Account{
		ID:          "account-1",
		Identity:    domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada"},
		WorkspaceID: "workspace-1",
	}
	expired := domain.NewSession("token-ada", ada.ID, now.Add(-domain.SessionLifetime))
	tests := []struct {
		name     string
		sessions testsupport.StoredSessions
	}{
		{"unknown session", testsupport.StoredSessions{}},
		{"expired session", testsupport.StoredSessions{expired.TokenHash: expired}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codes := memory.NewSignInCodeRepository()
			svc := app.NewSignInService(testsupport.StoredAccounts{ada.ID: ada}, tt.sessions, codes, testsupport.FixedToken("code-1"),
				testsupport.NewSequentialIDs("id"), testsupport.NewFixedClock(now))

			code, err := svc.IssueSignInCode(context.Background(), "token-ada", "challenge-1")

			if !errors.Is(err, domain.ErrNoWorkspace) {
				t.Errorf("err = %v, want %v", err, domain.ErrNoWorkspace)
			}
			if code != "" {
				t.Errorf("code = %q, want none", code)
			}
			if _, err := codes.Take(context.Background(), domain.HashSignInCode("code-1")); !errors.Is(err, domain.ErrNotFound) {
				t.Errorf("Take: err = %v, want %v (no code kept)", err, domain.ErrNotFound)
			}
		})
	}
}

func TestRedeemSignInCode_GivesANewSessionTokenForTheCodesAccount(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	ada := domain.Account{
		ID:          "account-1",
		Identity:    domain.Identity{Issuer: "dev", Subject: "ada", Name: "Ada"},
		WorkspaceID: "workspace-1",
	}
	// The RFC 7636 appendix B pair.
	const verifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	const challenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	issued := domain.NewSignInCode(ada.ID, challenge, now)
	issued.CodeHash = domain.HashSignInCode("code-1")
	codes := memory.NewSignInCodeRepository()
	if err := codes.Save(context.Background(), issued); err != nil {
		t.Fatalf("Save: %v", err)
	}
	sessions := testsupport.StoredSessions{}
	svc := app.NewSignInService(testsupport.StoredAccounts{ada.ID: ada}, sessions, codes, testsupport.FixedToken("token-new"),
		testsupport.NewSequentialIDs("id"), testsupport.NewFixedClock(now))

	token, err := svc.RedeemSignInCode(context.Background(), "code-1", verifier)
	if err != nil {
		t.Fatalf("RedeemSignInCode: %v", err)
	}

	if token != "token-new" {
		t.Errorf("token = %q, want %q", token, "token-new")
	}
	session := domain.NewSession("token-new", ada.ID, now)
	wantSessions := testsupport.StoredSessions{session.TokenHash: session}
	if !reflect.DeepEqual(sessions, wantSessions) {
		t.Errorf("sessions = %v, want %v", sessions, wantSessions)
	}
}

func TestRedeemSignInCode_RefusesACodeUsedBeforeAWrongVerifierAndAnExpiredCode(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	// The RFC 7636 appendix B pair.
	const verifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	const challenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	tests := []struct {
		name     string
		issuedAt time.Time
		verifier string
		redeems  int
	}{
		{"a code used before", now, verifier, 2},
		{"a wrong verifier", now, "another-verifier", 1},
		{"an expired code", now.Add(-domain.SignInCodeLifetime), verifier, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issued := domain.NewSignInCode("account-1", challenge, tt.issuedAt)
			issued.CodeHash = domain.HashSignInCode("code-1")
			codes := memory.NewSignInCodeRepository()
			if err := codes.Save(context.Background(), issued); err != nil {
				t.Fatalf("Save: %v", err)
			}
			sessions := testsupport.StoredSessions{}
			svc := app.NewSignInService(testsupport.StoredAccounts{}, sessions, codes, testsupport.FixedToken("token-new"),
				testsupport.NewSequentialIDs("id"), testsupport.NewFixedClock(now))

			var token string
			var err error
			for range tt.redeems {
				token, err = svc.RedeemSignInCode(context.Background(), "code-1", tt.verifier)
			}

			if err == nil {
				t.Errorf("err = nil, want a refusal")
			}
			if token != "" {
				t.Errorf("token = %q, want none", token)
			}
			if tt.redeems == 2 {
				// The first redeem started the one session.
				if len(sessions) != 1 {
					t.Errorf("sessions = %v, want only the first redeem's", sessions)
				}
			} else if len(sessions) != 0 {
				t.Errorf("sessions = %v, want none", sessions)
			}
		})
	}
}
