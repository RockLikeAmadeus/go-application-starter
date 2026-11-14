package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

const (
	nativeIntro     = "The sign-in request from the application that initiated it is not valid."
	nativePortText  = "`port` is missing or is not a port number from 1 to 65535."
	nativeStateText = "`state` is missing or malformed."
	nativeChalText  = "`challenge` is missing or is not a PKCE S256 challenge (43 base64url characters)."
	nativeOutro     = "Please start signing in again from that application. If this keeps happening, check that the application and the server are the same version."

	// The RFC 7636 example challenge: 43 base64url characters.
	validChallenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	validState     = "abcdefghijklmnopqrstuvwxyz012345"
)

func TestABadNativeSignInRequestIsRefusedListingEveryFailingValue(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{SignIn: &signInRecorder{}, IsDevSignInOn: true})
	for name, tc := range map[string]struct {
		query url.Values
		want  []string
	}{
		"a port of zero": {
			url.Values{"port": {"0"}, "state": {validState}, "challenge": {validChallenge}},
			[]string{nativePortText},
		},
		"a port above 65535": {
			url.Values{"port": {"65536"}, "state": {validState}, "challenge": {validChallenge}},
			[]string{nativePortText},
		},
		"a port that is not a number": {
			url.Values{"port": {"http"}, "state": {validState}, "challenge": {validChallenge}},
			[]string{nativePortText},
		},
		"a malformed state": {
			url.Values{"port": {"4000"}, "state": {"not valid!"}, "challenge": {validChallenge}},
			[]string{nativeStateText},
		},
		"a challenge of the wrong length": {
			url.Values{"port": {"4000"}, "state": {validState}, "challenge": {validChallenge + "A"}},
			[]string{nativeChalText},
		},
		"nothing at all": {
			url.Values{},
			[]string{nativePortText, nativeStateText, nativeChalText},
		},
	} {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/auth/native?"+tc.query.Encode(), nil))

			var body struct {
				Detail string `json:"detail"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode %q: %v", rec.Body.String(), err)
			}
			want := strings.Join(append(append([]string{nativeIntro}, tc.want...), nativeOutro), " ")
			if rec.Code != http.StatusBadRequest || body.Detail != want {
				t.Errorf("status = %d, detail = %q; want %d, %q", rec.Code, body.Detail, http.StatusBadRequest, want)
			}
		})
	}
}

func TestAGoodNativeSignInRequestWithNoSessionIsKeptInACookieAndSentToSignIn(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{SignIn: &signInRecorder{}, IsDevSignInOn: true})
	query := url.Values{"port": {"4000"}, "state": {validState}, "challenge": {validChallenge}}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/auth/native?"+query.Encode(), nil))

	location, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse location: %v", err)
	}
	if rec.Code != http.StatusFound || location.Path != "/auth/sign-in" || location.Query().Get("return") != "/auth/native/finish" {
		t.Errorf("status = %d, location = %q; want %d, /auth/sign-in?return=/auth/native/finish",
			rec.Code, rec.Header().Get("Location"), http.StatusFound)
	}

	var cookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == "native_sign_in" {
			cookie = c
		}
	}
	if cookie == nil {
		t.Fatal("no native_sign_in cookie")
	}
	kept, err := url.ParseQuery(cookie.Value)
	if err != nil || kept.Get("port") != "4000" || kept.Get("state") != validState || kept.Get("challenge") != validChallenge {
		t.Errorf("cookie value = %q, want the port, state, and challenge", cookie.Value)
	}
	if cookie.Path != "/auth/native" || !cookie.HttpOnly || cookie.MaxAge <= 0 {
		t.Errorf("cookie path = %q, HttpOnly = %v, MaxAge = %d; want /auth/native, HttpOnly, a lifetime", cookie.Path, cookie.HttpOnly, cookie.MaxAge)
	}
}

func TestANativeSignInRequestToSignUpWithNoSessionIsSentToSignUp(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{SignIn: &signInRecorder{}, IsDevSignInOn: true})
	query := url.Values{"port": {"4000"}, "state": {validState}, "challenge": {validChallenge}, "sign-up": {"1"}}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/auth/native?"+query.Encode(), nil))

	location, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse location: %v", err)
	}
	if rec.Code != http.StatusFound || location.Path != "/auth/sign-up" || location.Query().Get("return") != "/auth/native/finish" {
		t.Errorf("status = %d, location = %q; want %d, /auth/sign-up?return=/auth/native/finish",
			rec.Code, rec.Header().Get("Location"), http.StatusFound)
	}
}

func TestAGoodNativeSignInRequestWithASessionGoesStraightToFinish(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{
		SignIn:        sessions{"token-a": {ID: "account-a", WorkspaceID: "ws-a"}},
		IsDevSignInOn: true,
	})
	query := url.Values{"port": {"4000"}, "state": {validState}, "challenge": {validChallenge}}
	req := httptest.NewRequest(http.MethodGet, "/auth/native?"+query.Encode(), nil)
	req.AddCookie(&http.Cookie{Name: httpapi.SessionCookie, Value: "token-a"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/auth/native/finish" {
		t.Errorf("status = %d, location = %q; want %d, /auth/native/finish",
			rec.Code, rec.Header().Get("Location"), http.StatusFound)
	}
}

func TestNativeFinishIssuesACodeForTheBrowsersSessionAndGoesToTheCommandLinesLoopbackPort(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{
		SignIn:        sessions{"token-a": {ID: "account-a", WorkspaceID: "ws-a"}},
		IsDevSignInOn: true,
	})
	attempt := url.Values{"port": {"4000"}, "state": {validState}, "challenge": {validChallenge}}
	req := httptest.NewRequest(http.MethodGet, "/auth/native/finish", nil)
	req.AddCookie(&http.Cookie{Name: httpapi.SessionCookie, Value: "token-a"})
	req.AddCookie(&http.Cookie{Name: httpapi.NativeSignInCookie, Value: attempt.Encode()})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	location, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse location: %v", err)
	}
	wantCode := "code-for-token-a-" + validChallenge
	if rec.Code != http.StatusFound || location.Scheme != "http" || location.Host != "127.0.0.1:4000" ||
		location.Path != "/callback" || location.Query().Get("code") != wantCode || location.Query().Get("state") != validState {
		t.Errorf("status = %d, location = %q; want %d, http://127.0.0.1:4000/callback?code=%s&state=%s",
			rec.Code, rec.Header().Get("Location"), http.StatusFound, wantCode, validState)
	}
}

func TestNativeFinishNeverRedirectsToAnyHostButLoopbackWhateverTheAttemptCookieHolds(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{
		SignIn:        sessions{"token-a": {ID: "account-a", WorkspaceID: "ws-a"}},
		IsDevSignInOn: true,
	})
	attempt := url.Values{"port": {"1@evil.example"}, "state": {validState}, "challenge": {validChallenge}}
	req := httptest.NewRequest(http.MethodGet, "/auth/native/finish", nil)
	req.AddCookie(&http.Cookie{Name: httpapi.SessionCookie, Value: "token-a"})
	req.AddCookie(&http.Cookie{Name: httpapi.NativeSignInCookie, Value: attempt.Encode()})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if location := rec.Header().Get("Location"); location != "" || rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, location = %q; want %d and no redirect", rec.Code, location, http.StatusBadRequest)
	}
	if strings.Contains(rec.Body.String(), "code-for-") {
		t.Errorf("body = %q, want no code issued", rec.Body.String())
	}
}

const nativeAttemptGoneText = "The sign-in request from the application that initiated it could not be found or is no longer valid. " +
	"The browser did not send the saved sign-in attempt (the `native_sign_in` cookie), or it has expired after 10 minutes, " +
	"or it does not hold a valid `port`, `state`, and `challenge`. Please start signing in again from that application. " +
	"If this keeps happening, check that your browser accepts cookies from this server, and that the application and the server are the same version."

func TestNativeFinishRefusesAMissingOrInvalidAttemptCookieWithOneMessage(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{
		SignIn:        sessions{"token-a": {ID: "account-a", WorkspaceID: "ws-a"}},
		IsDevSignInOn: true,
	})
	invalid := url.Values{"port": {"0"}, "state": {"bad"}, "challenge": {"bad"}}
	for name, cookie := range map[string]*http.Cookie{
		"no attempt cookie":         nil,
		"an invalid attempt cookie": {Name: httpapi.NativeSignInCookie, Value: invalid.Encode()},
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/auth/native/finish", nil)
			req.AddCookie(&http.Cookie{Name: httpapi.SessionCookie, Value: "token-a"})
			if cookie != nil {
				req.AddCookie(cookie)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			var body struct {
				Detail string `json:"detail"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode %q: %v", rec.Body.String(), err)
			}
			if rec.Code != http.StatusBadRequest || body.Detail != nativeAttemptGoneText {
				t.Errorf("status = %d, detail = %q; want %d, %q", rec.Code, body.Detail, http.StatusBadRequest, nativeAttemptGoneText)
			}
		})
	}
}

const nativeNotSignedInText = "You are not signed in, so no sign-in code was issued to the application that initiated the sign-in. " +
	"Please start signing in again from that application, and finish signing in in the browser. " +
	"If this keeps happening, check that your browser accepts cookies from this server."

func TestNativeFinishWithNoSessionIsRefusedIssuingNoCode(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{
		SignIn:        sessions{"token-a": {ID: "account-a", WorkspaceID: "ws-a"}},
		IsDevSignInOn: true,
	})
	attempt := url.Values{"port": {"4000"}, "state": {validState}, "challenge": {validChallenge}}
	for name, session := range map[string]*http.Cookie{
		"no session cookie":  nil,
		"an unknown session": {Name: httpapi.SessionCookie, Value: "token-gone"},
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/auth/native/finish", nil)
			req.AddCookie(&http.Cookie{Name: httpapi.NativeSignInCookie, Value: attempt.Encode()})
			if session != nil {
				req.AddCookie(session)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			var body struct {
				Detail string `json:"detail"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode %q: %v", rec.Body.String(), err)
			}
			if rec.Code != http.StatusUnauthorized || body.Detail != nativeNotSignedInText || rec.Header().Get("Location") != "" {
				t.Errorf("status = %d, location = %q, detail = %q; want %d, no redirect, %q",
					rec.Code, rec.Header().Get("Location"), body.Detail, http.StatusUnauthorized, nativeNotSignedInText)
			}
		})
	}
}

// failingToIssue is a sign-in service whose sessions are all valid but whose
// codes can't be kept.
type failingToIssue struct{ sessions }

func (failingToIssue) IssueSignInCode(context.Context, string, string) (string, error) {
	return "", errors.New("storage is down")
}

func TestNativeFinishWhenTheCodeCannotBeIssuedFailsWithoutRedirecting(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{
		SignIn:        failingToIssue{sessions{"token-a": {ID: "account-a", WorkspaceID: "ws-a"}}},
		IsDevSignInOn: true,
	})
	attempt := url.Values{"port": {"4000"}, "state": {validState}, "challenge": {validChallenge}}
	req := httptest.NewRequest(http.MethodGet, "/auth/native/finish", nil)
	req.AddCookie(&http.Cookie{Name: httpapi.SessionCookie, Value: "token-a"})
	req.AddCookie(&http.Cookie{Name: httpapi.NativeSignInCookie, Value: attempt.Encode()})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError || rec.Header().Get("Location") != "" {
		t.Errorf("status = %d, location = %q; want %d and no redirect", rec.Code, rec.Header().Get("Location"), http.StatusInternalServerError)
	}
	if strings.Contains(rec.Body.String(), "storage is down") {
		t.Errorf("body = %q, want the error's text kept from the browser", rec.Body.String())
	}
}

// redeemingCodes is a sign-in service that swaps one code and verifier for a
// token, and refuses every other pair as a missing code.
type redeemingCodes struct{ sessions }

func (redeemingCodes) RedeemSignInCode(_ context.Context, code, verifier string) (string, error) {
	if code == "good-code" && verifier == "good-verifier" {
		return "new-session-token", nil
	}
	return "", domain.ErrNotFound
}

func TestNativeTokenSwapsACodeAndVerifierForASessionTokenAndRefusesAnyOtherWithoutSayingWhy(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{SignIn: redeemingCodes{}, IsDevSignInOn: true})
	post := func(body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/auth/native/token", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	rec := post(`{"code":"good-code","verifier":"good-verifier"}`)
	var got struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil || rec.Code != http.StatusOK || got.Token != "new-session-token" {
		t.Errorf("status = %d, token = %q, err = %v; want %d and new-session-token", rec.Code, got.Token, err, http.StatusOK)
	}
	if cookies := rec.Result().Cookies(); len(cookies) != 0 {
		t.Errorf("cookies = %v, want none: the token is for the command line, not the browser", cookies)
	}

	for name, body := range map[string]string{
		"a wrong verifier": `{"code":"good-code","verifier":"other"}`,
		"an unknown code":  `{"code":"other","verifier":"good-verifier"}`,
		"no body fields":   `{}`,
		"malformed JSON":   `not json`,
	} {
		t.Run(name, func(t *testing.T) {
			rec := post(body)
			if rec.Code != http.StatusBadRequest || rec.Body.Len() != 0 {
				t.Errorf("status = %d, body = %q; want %d and an empty body, so the refusal says nothing about why",
					rec.Code, rec.Body.String(), http.StatusBadRequest)
			}
		})
	}
}
