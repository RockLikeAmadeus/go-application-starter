package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
)

// fakeSignInProvider is an httpapi.SignInProvider that starts every sign-in
// with the same URL, state, and attempt.
type fakeSignInProvider struct {
	url, state, attempt string
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
