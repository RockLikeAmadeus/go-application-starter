package httpapi_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
)

func TestSignInGoesToTheDevSignInWhenItIsOn(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{SignIn: &signInRecorder{}, IsDevSignInOn: true})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/auth/sign-in", nil))

	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/auth/dev" {
		t.Errorf("status = %d, Location = %q; want %d to %q", rec.Code, rec.Header().Get("Location"), http.StatusFound, "/auth/dev")
	}
}

func TestSignInPassesALocalReturnPathOnToTheDevSignInAndDropsAnyOther(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{SignIn: &signInRecorder{}, IsDevSignInOn: true})
	for returnTo, want := range map[string]string{
		"/auth/native/finish?x=1": "/auth/dev?return=" + url.QueryEscape("/auth/native/finish?x=1"),
		"https://evil.example/":   "/auth/dev",
		"//evil.example/":         "/auth/dev",
		"":                        "/auth/dev",
	} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/auth/sign-in?return="+url.QueryEscape(returnTo), nil))
		if got := rec.Header().Get("Location"); rec.Code != http.StatusFound || got != want {
			t.Errorf("return=%q: status = %d, Location = %q; want %d to %q", returnTo, rec.Code, got, http.StatusFound, want)
		}
	}
}

// postSignOut posts a sign-out to the handler with the given session token in
// the session cookie.
func postSignOut(handler http.Handler, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/auth/sign-out", nil)
	req.AddCookie(&http.Cookie{Name: httpapi.SessionCookie, Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestSigningOutEndsTheSessionClearsTheCookieAndGoesToTheWebApp(t *testing.T) {
	signIn := &signInRecorder{}
	handler := httpapi.NewHandler(httpapi.Options{SignIn: signIn})
	rec := postSignOut(handler, "token-ada")

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/" {
		t.Fatalf("status = %d, Location = %q; want %d to %q", rec.Code, rec.Header().Get("Location"), http.StatusSeeOther, "/")
	}
	if len(signIn.signedOut) != 1 || signIn.signedOut[0] != "token-ada" {
		t.Errorf("signed out %q, want [token-ada]", signIn.signedOut)
	}
	cookie := sessionCookieOf(t, rec)
	if cookie.Value != "" || cookie.MaxAge >= 0 || cookie.Path != "/" {
		t.Errorf("cookie = %+v, want an empty value on path / that has expired (Max-Age < 0)", cookie)
	}
}

func TestASignOutThatFailsIsAProblemAndKeepsTheCookie(t *testing.T) {
	signIn := &signInRecorder{signOutErr: errors.New("storage can't be reached")}
	handler := httpapi.NewHandler(httpapi.Options{SignIn: signIn})
	rec := postSignOut(handler, "token-ada")

	if rec.Code != http.StatusInternalServerError || rec.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("status = %d, Content-Type = %q; want %d, %q", rec.Code, rec.Header().Get("Content-Type"), http.StatusInternalServerError, "application/problem+json")
	}
	if cookies := rec.Header().Values("Set-Cookie"); len(cookies) != 0 {
		t.Errorf("Set-Cookie = %q, want none, so the browser keeps the session cookie", cookies)
	}
}

func TestSigningOutWithABearerTokenEndsThatSessionWithNoCookieOrOrigin(t *testing.T) {
	signIn := &signInRecorder{}
	handler := httpapi.NewHandler(httpapi.Options{SignIn: signIn})
	req := httptest.NewRequest(http.MethodPost, "http://starter.example/auth/sign-out", nil)
	req.Header.Set("Authorization", "Bearer token-ada")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body %q", rec.Code, http.StatusSeeOther, rec.Body.String())
	}
	if len(signIn.signedOut) != 1 || signIn.signedOut[0] != "token-ada" {
		t.Errorf("signed out %q, want [token-ada]", signIn.signedOut)
	}
}
