package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// signInRecorder is a ports.SignInService that records the identities it
// signs in and the tokens it signs out, and answers the same token for every
// sign-in.
type signInRecorder struct {
	token      string
	identities []domain.Identity
	signedOut  []string
}

func (s *signInRecorder) SignIn(_ context.Context, identity domain.Identity) (string, error) {
	s.identities = append(s.identities, identity)
	return s.token, nil
}

func (s *signInRecorder) CurrentAccount(context.Context, string) (domain.Account, error) {
	return domain.Account{}, domain.ErrNoWorkspace
}

func (s *signInRecorder) SignOut(_ context.Context, token string) error {
	s.signedOut = append(s.signedOut, token)
	return nil
}

// sessionCookieOf returns the response's session cookie, failing the test
// without one.
func sessionCookieOf(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == httpapi.SessionCookie {
			return cookie
		}
	}
	t.Fatalf("no %q cookie in %v", httpapi.SessionCookie, rec.Result().Header.Values("Set-Cookie"))
	return nil
}

// postDevSignIn posts the dev sign-in form with the given name to the handler.
func postDevSignIn(handler http.Handler, name string) *httptest.ResponseRecorder {
	return postDevSignInTo(handler, "/auth/dev", name)
}

// postDevSignInTo posts the dev sign-in form with the given name to target, a
// path or a full URL (an https one makes the request arrive over TLS).
func postDevSignInTo(handler http.Handler, target, name string) *httptest.ResponseRecorder {
	form := url.Values{"name": {name}}
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestTheDevSignInSignsInAsTheGivenNameAndSetsTheSessionCookie(t *testing.T) {
	signIn := &signInRecorder{token: "token-ada"}
	handler := httpapi.NewHandler(httpapi.Options{SignIn: signIn, IsDevSignInOn: true})
	rec := postDevSignIn(handler, "Ada")

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/" {
		t.Fatalf("status = %d, Location = %q; want %d to %q", rec.Code, rec.Header().Get("Location"), http.StatusSeeOther, "/")
	}
	want := domain.Identity{Issuer: httpapi.DevIssuer, Subject: "Ada", Name: "Ada"}
	if len(signIn.identities) != 1 || signIn.identities[0] != want {
		t.Errorf("signed in as %+v, want %+v", signIn.identities, want)
	}
	cookie := sessionCookieOf(t, rec)
	if cookie.Value != "token-ada" || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.Path != "/" {
		t.Errorf("cookie = %+v, want token-ada, HttpOnly, SameSite=Lax, and path /", cookie)
	}
}

func TestTheSessionCookieLastsAsLongAsTheSession(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{SignIn: &signInRecorder{token: "token-ada"}, IsDevSignInOn: true})
	rec := postDevSignIn(handler, "Ada")

	cookie := sessionCookieOf(t, rec)
	if want := int(domain.SessionLifetime.Seconds()); cookie.MaxAge != want {
		t.Errorf("cookie Max-Age = %d, want %d (the session's lifetime)", cookie.MaxAge, want)
	}
}

func TestTheSessionCookieIsSecureOnlyOverHTTPS(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{SignIn: &signInRecorder{token: "token-ada"}, IsDevSignInOn: true})
	for target, want := range map[string]bool{"https://example.com/auth/dev": true, "http://example.com/auth/dev": false} {
		rec := postDevSignInTo(handler, target, "Ada")
		if cookie := sessionCookieOf(t, rec); cookie.Secure != want {
			t.Errorf("posting to %s: cookie Secure = %v, want %v", target, cookie.Secure, want)
		}
	}
}

func TestTheDevSignInFormAsksForANameAndPostsToTheDevSignIn(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{SignIn: &signInRecorder{}, IsDevSignInOn: true})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/auth/dev", nil))

	if rec.Code != http.StatusOK || !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("status = %d, Content-Type = %q; want %d and an HTML page", rec.Code, rec.Header().Get("Content-Type"), http.StatusOK)
	}
	body := rec.Body.String()
	for _, want := range []string{`<form method="post" action="/auth/dev">`, `<input name="name"`, `<button type="submit">`} {
		if !strings.Contains(body, want) {
			t.Errorf("the page has no %s:\n%s", want, body)
		}
	}
}
