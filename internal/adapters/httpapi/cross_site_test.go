package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
)

func TestASignOutFromAnotherSiteIsRefused(t *testing.T) {
	rec, signIn := signOutWithHeader("Sec-Fetch-Site", "cross-site")

	got := wantCrossSiteProblem(t, rec)
	wantDetail := "The server refused this request because it did not come from the application's own pages. Please try again from the application."
	if got.Detail != wantDetail {
		t.Errorf("detail = %q, want %q", got.Detail, wantDetail)
	}
	if len(signIn.signedOut) != 0 {
		t.Errorf("signed out %q, want none", signIn.signedOut)
	}
}

func TestASignOutWithAnOriginFromAnotherHostIsRefused(t *testing.T) {
	rec, signIn := signOutWithHeader("Origin", "http://elsewhere.example")

	wantCrossSiteProblem(t, rec)
	if len(signIn.signedOut) != 0 {
		t.Errorf("signed out %q, want none", signIn.signedOut)
	}
}

func TestASameOriginSignOutStillGoesThrough(t *testing.T) {
	rec, signIn := signOutWithHeader("Sec-Fetch-Site", "same-origin")

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body %q", rec.Code, http.StatusSeeOther, rec.Body.String())
	}
	if len(signIn.signedOut) != 1 || signIn.signedOut[0] != "token-ada" {
		t.Errorf("signed out %q, want [token-ada]", signIn.signedOut)
	}
}

// signOutWithHeader sends a sign-out, with a session cookie and the given
// header, to a handler with sign-in on, served at http://starter.example.
func signOutWithHeader(name, value string) (*httptest.ResponseRecorder, *signInRecorder) {
	signIn := &signInRecorder{}
	handler := httpapi.NewHandler(httpapi.Options{SignIn: signIn})
	req := httptest.NewRequest(http.MethodPost, "http://starter.example/auth/sign-out", nil)
	req.AddCookie(&http.Cookie{Name: httpapi.SessionCookie, Value: "token-ada"})
	req.Header.Set(name, value)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec, signIn
}

type crossSiteProblem struct {
	Type   string `json:"type"`
	Detail string `json:"detail"`
}

// wantCrossSiteProblem fails the test unless rec is a 403 cross_site
// problem, and returns the problem.
func wantCrossSiteProblem(t *testing.T, rec *httptest.ResponseRecorder) crossSiteProblem {
	t.Helper()
	if rec.Code != http.StatusForbidden || rec.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("status = %d, Content-Type = %q; want %d, %q", rec.Code, rec.Header().Get("Content-Type"), http.StatusForbidden, "application/problem+json")
	}
	var got crossSiteProblem
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	if got.Type != httpapi.ProblemCrossSite {
		t.Errorf("type = %q, want %q", got.Type, httpapi.ProblemCrossSite)
	}
	return got
}
