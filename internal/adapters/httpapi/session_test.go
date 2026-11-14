package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/workspace"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// sessions is a ports.SignInService that knows each session's account by its
// token. Only CurrentAccount is used by the session middleware.
type sessions map[string]domain.Account

func (s sessions) SignIn(context.Context, domain.Identity) (string, error) {
	panic("not used by the session middleware")
}

func (s sessions) CurrentAccount(_ context.Context, token string) (domain.Account, error) {
	account, ok := s[token]
	if !ok {
		return domain.Account{}, domain.ErrNoWorkspace
	}
	return account, nil
}

func (s sessions) SignOut(context.Context, string) error {
	panic("not used by the session middleware")
}

// IssueSignInCode answers a code that spells out the session token and the
// challenge it was issued for, so a test can see both.
func (s sessions) IssueSignInCode(_ context.Context, token, challenge string) (string, error) {
	return "code-for-" + token + "-" + challenge, nil
}

func (s sessions) RedeemSignInCode(context.Context, string, string) (string, error) {
	panic("not used by these tests")
}

// workspaceAsPlan is a ports.EntitlementsService that answers the request's
// workspace (workspace.FromContext) as its plan, or "none" without one, so a
// test can see which workspace the server put in the request's context.
type workspaceAsPlan struct{}

func (workspaceAsPlan) CurrentEntitlements(ctx context.Context) (domain.Entitlements, error) {
	id, err := workspace.FromContext{}.CurrentWorkspace(ctx)
	if err != nil {
		return domain.Entitlements{Plan: "none"}, nil
	}
	return domain.Entitlements{Plan: string(id)}, nil
}

// requestWorkspace sends GET /api/v1/entitlements through a handler with the
// session middleware, with the given cookies, and returns the workspace the
// request ran in.
func requestWorkspace(t *testing.T, signIn sessions, cookies ...*http.Cookie) string {
	t.Helper()
	return requestWorkspaceWithHeaders(t, signIn, nil, cookies...)
}

// requestWorkspaceWithHeaders is requestWorkspace with extra request headers.
func requestWorkspaceWithHeaders(t *testing.T, signIn sessions, headers map[string]string, cookies ...*http.Cookie) string {
	t.Helper()
	handler := httpapi.NewHandler(httpapi.Options{Entitlements: workspaceAsPlan{}, SignIn: signIn})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/entitlements", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body %s", rec.Code, http.StatusOK, rec.Body)
	}
	var body struct{ Plan string }
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	return body.Plan
}

func TestASessionCookiePutsItsAccountsWorkspaceInTheRequest(t *testing.T) {
	signIn := sessions{"token-a": {ID: "account-a", WorkspaceID: "ws-a"}}

	got := requestWorkspace(t, signIn, &http.Cookie{Name: httpapi.SessionCookie, Value: "token-a"})

	if got != "ws-a" {
		t.Errorf("workspace = %q, want %q", got, "ws-a")
	}
}

func TestABearerTokenPutsItsAccountsWorkspaceInTheRequestAheadOfTheCookie(t *testing.T) {
	signIn := sessions{
		"token-a": {ID: "account-a", WorkspaceID: "ws-a"},
		"token-b": {ID: "account-b", WorkspaceID: "ws-b"},
	}

	got := requestWorkspaceWithHeaders(t, signIn,
		map[string]string{"Authorization": "Bearer token-a"},
		&http.Cookie{Name: httpapi.SessionCookie, Value: "token-b"})

	if got != "ws-a" {
		t.Errorf("workspace = %q, want %q", got, "ws-a")
	}
}

func TestNoSessionCookieLeavesNoWorkspaceInTheRequest(t *testing.T) {
	signIn := sessions{"token-a": {ID: "account-a", WorkspaceID: "ws-a"}}

	got := requestWorkspace(t, signIn)

	if got != "none" {
		t.Errorf("workspace = %q, want none", got)
	}
}

func TestAnUnknownSessionCookieLeavesNoWorkspaceInTheRequest(t *testing.T) {
	signIn := sessions{"token-a": {ID: "account-a", WorkspaceID: "ws-a"}}

	got := requestWorkspace(t, signIn, &http.Cookie{Name: httpapi.SessionCookie, Value: "token-unknown"})

	if got != "none" {
		t.Errorf("workspace = %q, want none", got)
	}
}

// session is the body of GET /api/v1/session.
type session struct {
	IsSignInOn   bool    `json:"isSignInOn"`
	Name         *string `json:"name"`
	SignInMethod string  `json:"signInMethod"`
}

// requestSession sends GET /api/v1/session through a handler with the given
// sign-in service (nil for sign-in off) and cookies, and returns the session it
// answers.
func requestSession(t *testing.T, signIn ports.SignInService, cookies ...*http.Cookie) session {
	t.Helper()
	handler := httpapi.NewHandler(httpapi.Options{SignIn: signIn})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/session", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body %s", rec.Code, http.StatusOK, rec.Body)
	}
	var body session
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	return body
}

// sessionsWithAda knows one session, token-a, for an account named Ada.
var sessionsWithAda = sessions{"token-a": {ID: "account-a", Identity: domain.Identity{Issuer: "dev", Subject: "Ada", Name: "Ada"}, WorkspaceID: "ws-a"}}

func TestTheSessionEndpointNamesTheSignedInAccount(t *testing.T) {
	got := requestSession(t, sessionsWithAda, &http.Cookie{Name: httpapi.SessionCookie, Value: "token-a"})

	if !got.IsSignInOn || got.Name == nil || *got.Name != "Ada" {
		t.Errorf("session = %+v, want sign-in on and the name Ada", got)
	}
}

func TestTheSessionEndpointSaysSignInIsOnWithNoNameWhenSignedOut(t *testing.T) {
	got := requestSession(t, sessionsWithAda)

	if !got.IsSignInOn || got.Name != nil {
		t.Errorf("session = %+v, want sign-in on and no name", got)
	}
}

func TestTheSessionEndpointSaysSignInIsOffWithNoSignInService(t *testing.T) {
	got := requestSession(t, nil)

	if got.IsSignInOn || got.Name != nil {
		t.Errorf("session = %+v, want sign-in off and no name", got)
	}
}

func TestTheSessionEndpointSaysToSignInOnAPage(t *testing.T) {
	got := requestSession(t, sessionsWithAda)

	if got.SignInMethod != "page" {
		t.Errorf("signInMethod = %q, want %q", got.SignInMethod, "page")
	}
}
