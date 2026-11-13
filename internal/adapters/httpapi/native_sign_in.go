package httpapi

import (
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// NativeSignInCookie is the cookie that keeps a command-line sign-in's port,
// state, and challenge while the browser signs in.
const NativeSignInCookie = "native_sign_in"

// nativeSignInLifetime is how long a person has to sign in.
const nativeSignInLifetime = 10 * time.Minute

// A state is 16 to 128 base64url characters, enough to be unguessable; a
// challenge is the 43 characters of a SHA-256 digest in unpadded base64url.
var (
	statePattern     = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)
	challengePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)
)

// nativeSignIn answers GET /auth/native, where the command line starts a
// sign-in: the browser arrives with the command line's loopback "port", a
// "state", and a PKCE S256 "challenge". A request with a bad value is refused
// with a problem naming each one. A good one is kept in a short-lived cookie,
// and the browser goes on to sign in and then back to /auth/native/finish,
// or straight there when it already has a session.
func nativeSignIn(publicURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startNativeSignIn(w, r, publicURL)
	}
}

// nativeFinish answers GET /auth/native/finish, where the signed-in browser
// is sent on: it issues a one-time code for the browser's session, and sends
// the browser to the command line's loopback address with the code and the
// state.
func nativeFinish(signIn ports.SignInService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(NativeSignInCookie)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		// The cookie can be planted by another site, so it is checked again.
		attempt, _ := url.ParseQuery(cookie.Value)
		port, failures := checkNativeAttempt(attempt)
		if len(failures) > 0 {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		token, _ := sessionToken(r)
		code, _ := signIn.IssueSignInCode(r.Context(), token, attempt.Get("challenge"))
		callback := url.URL{
			Scheme:   "http",
			Host:     net.JoinHostPort("127.0.0.1", strconv.Itoa(port)),
			Path:     "/callback",
			RawQuery: url.Values{"code": {code}, "state": {attempt.Get("state")}}.Encode(),
		}
		http.Redirect(w, r, callback.String(), http.StatusFound)
	}
}

// checkNativeAttempt checks a command-line sign-in's "port", "state", and
// "challenge", and returns the port as a number with a sentence for each
// failing value.
func checkNativeAttempt(attempt url.Values) (port int, failures []string) {
	port, err := strconv.Atoi(attempt.Get("port"))
	if err != nil || port < 1 || port > 65535 {
		failures = append(failures, "`port` is missing or is not a port number from 1 to 65535.")
	}
	if !statePattern.MatchString(attempt.Get("state")) {
		failures = append(failures, "`state` is missing or malformed.")
	}
	if !challengePattern.MatchString(attempt.Get("challenge")) {
		failures = append(failures, "`challenge` is missing or is not a PKCE S256 challenge (43 base64url characters).")
	}
	return port, failures
}

func startNativeSignIn(w http.ResponseWriter, r *http.Request, publicURL string) {
	_ = r.ParseForm()
	_, failures := checkNativeAttempt(r.Form)
	if len(failures) > 0 {
		detail := "The sign-in request from the command line is not valid. " + strings.Join(failures, " ") +
			" Please run the login command again. If this keeps happening, check that the command-line application and the server are the same version."
		writeProblem(w, problem(http.StatusBadRequest, ProblemBadRequest, detail, nil))
		return
	}
	attempt := url.Values{
		"port":      {r.FormValue("port")},
		"state":     {r.FormValue("state")},
		"challenge": {r.FormValue("challenge")},
	}
	http.SetCookie(w, &http.Cookie{
		Name:     NativeSignInCookie,
		Value:    attempt.Encode(),
		Path:     "/auth/native",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(nativeSignInLifetime.Seconds()),
		Secure:   isSecure(r, publicURL),
	})
	if _, ok := accountFrom(r.Context()); ok {
		http.Redirect(w, r, "/auth/native/finish", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/auth/sign-in?return="+url.QueryEscape("/auth/native/finish"), http.StatusFound)
}
