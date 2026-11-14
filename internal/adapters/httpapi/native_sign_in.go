package httpapi

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
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
// or straight there when it already has a session. With "sign-up=1" a
// browser with no session is sent to sign up instead of sign in.
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
			writeNativeAttemptGone(w)
			return
		}
		// The cookie can be planted by another site, so it is checked again.
		attempt, _ := url.ParseQuery(cookie.Value)
		port, failures := checkNativeAttempt(attempt)
		if len(failures) > 0 {
			writeNativeAttemptGone(w)
			return
		}
		if _, ok := accountFrom(r.Context()); !ok {
			writeProblem(w, problem(http.StatusUnauthorized, ProblemNoWorkspace, nativeNotSignedInDetail, nil))
			return
		}
		token, _ := sessionToken(r)
		code, err := signIn.IssueSignInCode(r.Context(), token, attempt.Get("challenge"))
		if err != nil {
			writeProblem(w, problemFor(err))
			return
		}
		callback := url.URL{
			Scheme:   "http",
			Host:     net.JoinHostPort("127.0.0.1", strconv.Itoa(port)),
			Path:     "/callback",
			RawQuery: url.Values{"code": {code}, "state": {attempt.Get("state")}}.Encode(),
		}
		http.Redirect(w, r, callback.String(), http.StatusFound)
	}
}

// nativeToken answers POST /auth/native/token, where the command line swaps
// its one-time code and PKCE verifier (a JSON body) for a new session token
// (a JSON answer). It refuses a bad body or a code that doesn't redeem with
// a bare 400, so a guesser learns nothing about why.
func nativeToken(signIn ports.SignInService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Code     string `json:"code"`
			Verifier string `json:"verifier"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		token, err := signIn.RedeemSignInCode(r.Context(), body.Code, body.Verifier)
		if errors.Is(err, domain.ErrNotFound) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err != nil {
			writeProblem(w, problemFor(err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}

func writeNativeAttemptGone(w http.ResponseWriter) {
	writeProblem(w, problem(http.StatusBadRequest, ProblemBadRequest, nativeAttemptGoneDetail, nil))
}

const nativeAttemptGoneDetail = "The sign-in request from the application that initiated it could not be found or is no longer valid. " +
	"The browser did not send the saved sign-in attempt (the `native_sign_in` cookie), or it has expired after 10 minutes, " +
	"or it does not hold a valid `port`, `state`, and `challenge`. Please start signing in again from that application. " +
	"If this keeps happening, check that your browser accepts cookies from this server, and that the application and the server are the same version."

const nativeNotSignedInDetail = "You are not signed in, so no sign-in code was issued to the application that initiated the sign-in. " +
	"Please start signing in again from that application, and finish signing in in the browser. " +
	"If this keeps happening, check that your browser accepts cookies from this server."

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
		detail := "The sign-in request from the application that initiated it is not valid. " + strings.Join(failures, " ") +
			" Please start signing in again from that application. If this keeps happening, check that the application and the server are the same version."
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
	page := "/auth/sign-in"
	if r.FormValue("sign-up") == "1" {
		page = "/auth/sign-up"
	}
	http.Redirect(w, r, page+"?return="+url.QueryEscape("/auth/native/finish"), http.StatusFound)
}
