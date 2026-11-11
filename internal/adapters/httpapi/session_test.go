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
	handler := httpapi.NewHandler(httpapi.Options{Entitlements: workspaceAsPlan{}, SignIn: signIn})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/entitlements", nil)
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
