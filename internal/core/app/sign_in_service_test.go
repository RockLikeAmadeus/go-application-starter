package app_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/app"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
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
