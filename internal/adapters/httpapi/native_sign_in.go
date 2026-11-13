package httpapi

import (
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
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
// and the browser goes on to sign in and then back to /auth/native/finish.
func nativeSignIn(publicURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startNativeSignIn(w, r, publicURL)
	}
}

func startNativeSignIn(w http.ResponseWriter, r *http.Request, publicURL string) {
	var failures []string
	if port, err := strconv.Atoi(r.FormValue("port")); err != nil || port < 1 || port > 65535 {
		failures = append(failures, "`port` is missing or is not a port number from 1 to 65535.")
	}
	if !statePattern.MatchString(r.FormValue("state")) {
		failures = append(failures, "`state` is missing or malformed.")
	}
	if !challengePattern.MatchString(r.FormValue("challenge")) {
		failures = append(failures, "`challenge` is missing or is not a PKCE S256 challenge (43 base64url characters).")
	}
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
	http.Redirect(w, r, "/auth/sign-in?return="+url.QueryEscape("/auth/native/finish"), http.StatusFound)
}
