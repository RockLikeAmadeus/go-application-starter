package httpapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// fakeSignInProvider is an httpapi.SignInProvider that starts every sign-in
// with the same URL, state, and attempt, and finishes one with identity when
// it gets that attempt and code "the-code".
type fakeSignInProvider struct {
	url, state, attempt string
	identity            domain.Identity
}

func (f fakeSignInProvider) Finish(_ context.Context, attempt, code string) (domain.Identity, error) {
	if attempt != f.attempt || code != "the-code" {
		return domain.Identity{}, errors.New("unexpected attempt or code")
	}
	return f.identity, nil
}

func (f fakeSignInProvider) Start() (string, string, string) { return f.url, f.state, f.attempt }

func TestStartingAnOIDCSignInKeepsTheAttemptInACookieAndGoesToTheProvider(t *testing.T) {
	provider := fakeSignInProvider{
		url:     "https://provider.example/authorize?state=s1",
		state:   "s1",
		attempt: "opaque-attempt",
	}
	handler := httpapi.NewHandler(httpapi.Options{SignIn: &signInRecorder{}, SignInProvider: provider})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/auth/oidc", nil))

	if rec.Code != http.StatusFound || rec.Header().Get("Location") != provider.url {
		t.Fatalf("status = %d, Location = %q; want %d to %q", rec.Code, rec.Header().Get("Location"), http.StatusFound, provider.url)
	}
	var cookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == httpapi.AttemptCookie {
			cookie = c
		}
	}
	if cookie == nil {
		t.Fatalf("no %q cookie in %v", httpapi.AttemptCookie, rec.Result().Header.Values("Set-Cookie"))
	}
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.Path != "/auth/oidc" ||
		cookie.MaxAge != int((10*time.Minute).Seconds()) || cookie.Value != "s1.opaque-attempt" {
		t.Errorf("cookie = %+v, want HttpOnly, SameSite=Lax, Path /auth/oidc, Max-Age 600, and the state and attempt as s1.opaque-attempt", cookie)
	}
}

func TestSigningInWithAnOIDCProviderGoesToTheOIDCStart(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{SignIn: &signInRecorder{}, SignInProvider: fakeSignInProvider{}})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/auth/sign-in", nil))

	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/auth/oidc" {
		t.Errorf("status = %d, Location = %q; want %d to /auth/oidc", rec.Code, rec.Header().Get("Location"), http.StatusFound)
	}
}

func TestTheProviderSendingTheBrowserBackSignsInSetsTheSessionCookieAndGoesToTheWebApp(t *testing.T) {
	identity, _ := domain.NewIdentity("https://provider.example", "sub-ada", "Ada")
	provider := fakeSignInProvider{state: "s1", attempt: "opaque-attempt", identity: identity}
	signIn := &signInRecorder{token: "token-ada"}
	handler := httpapi.NewHandler(httpapi.Options{SignIn: signIn, SignInProvider: provider})
	req := httptest.NewRequest(http.MethodGet, "/auth/oidc/callback?state=s1&code=the-code", nil)
	req.AddCookie(&http.Cookie{Name: httpapi.AttemptCookie, Value: "s1.opaque-attempt"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/" {
		t.Fatalf("status = %d, Location = %q; want %d to /", rec.Code, rec.Header().Get("Location"), http.StatusSeeOther)
	}
	if len(signIn.identities) != 1 || signIn.identities[0] != identity {
		t.Errorf("signed in %v, want [%v]", signIn.identities, identity)
	}
	if cookie := sessionCookieOf(t, rec); cookie.Value != "token-ada" {
		t.Errorf("session cookie = %+v, want token-ada", cookie)
	}
	var attempt *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == httpapi.AttemptCookie {
			attempt = c
		}
	}
	if attempt == nil || attempt.MaxAge >= 0 || attempt.Path != "/auth/oidc" {
		t.Errorf("attempt cookie = %+v, want cleared (Max-Age < 0) on /auth/oidc", attempt)
	}
}

func TestTheProviderSendingAnErrorAnswersTheSignInFailedPageAndClearsTheAttemptCookie(t *testing.T) {
	signIn := &signInRecorder{token: "token-ada"}
	handler := httpapi.NewHandler(httpapi.Options{SignIn: signIn, SignInProvider: fakeSignInProvider{state: "s1", attempt: "opaque-attempt"}})
	req := httptest.NewRequest(http.MethodGet, "/auth/oidc/callback?error=access_denied&state=s1", nil)
	req.AddCookie(&http.Cookie{Name: httpapi.AttemptCookie, Value: "s1.opaque-attempt"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest || rec.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Fatalf("status = %d, Content-Type = %q; want 400 and text/html", rec.Code, rec.Header().Get("Content-Type"))
	}
	body := rec.Body.String()
	for _, want := range []string{
		"<title>Sign-in failed</title>",
		"<h1>Sign-in failed</h1>",
		"Something went wrong while attempting to sign in. This can happen if the sign-in attempt was cancelled at the provider, or due to bad credentials, a timeout, or other problem. Please try again.",
		`<a href="/auth/sign-in">Try signing in again</a>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body lacks %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "access_denied") {
		t.Errorf("body shows the cause: %s", body)
	}
	if len(signIn.identities) != 0 || len(rec.Result().Cookies()) != 1 {
		t.Errorf("signed in %v, cookies %v; want no sign-in and only the cleared attempt cookie", signIn.identities, rec.Result().Cookies())
	}
	if c := rec.Result().Cookies()[0]; c.Name != httpapi.AttemptCookie || c.MaxAge >= 0 {
		t.Errorf("cookie = %+v, want the attempt cookie cleared", c)
	}
}
