package httpapi

import (
	"context"
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// AttemptCookie is the cookie that keeps a sign-in attempt between leaving for
// the provider and coming back.
const AttemptCookie = "oidc_sign_in"

// attemptLifetime is how long a person has to finish at the provider.
const attemptLifetime = 10 * time.Minute

// SignInProvider is the sign-in provider the OIDC routes talk to; the oidc
// adapter's Provider implements it. Start gives the authorization URL, the
// state the provider will send back, and the attempt, a string only the
// provider understands that it gets back in Finish.
type SignInProvider interface {
	Start() (authURL, state, attempt string)
	// StartSignUp is Start for someone who is new: it opens at the provider's
	// sign-up page when the provider has one.
	StartSignUp() (authURL, state, attempt string)
	Finish(ctx context.Context, attempt, code string) (domain.Identity, error)
}

// oidcStart answers GET /auth/oidc and GET /auth/oidc/sign-up: it begins an
// attempt with start (the provider's Start or StartSignUp), keeps it in a
// cookie, and sends the browser to the provider.
func oidcStart(start func() (authURL, state, attempt string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authURL, state, attempt := start()
		// Lax still sends the cookie on the provider's top-level redirect back.
		http.SetCookie(w, &http.Cookie{
			Name:     AttemptCookie,
			Value:    attemptCookieValue(state, attempt),
			Path:     "/auth/oidc",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   int(attemptLifetime.Seconds()),
		})
		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

// oidcCallback answers GET /auth/oidc/callback: it finishes the attempt kept in
// the cookie with the provider's code, signs the person in, and goes to the web
// app.
func oidcCallback(log *slog.Logger, provider SignInProvider, signIn ports.SignInService, publicURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Whatever the outcome, the attempt is over.
		clearAttemptCookie(w)
		fail := func(cause string) {
			log.Warn("oidc sign-in failed", "cause", cause)
			signInFailed(w)
		}
		if providerError := r.URL.Query().Get("error"); providerError != "" {
			fail("the provider answered an error: " + providerError)
			return
		}
		cookie, err := r.Cookie(AttemptCookie)
		if err != nil {
			fail("no attempt cookie")
			return
		}
		state, attempt, _ := strings.Cut(cookie.Value, ".")
		if subtle.ConstantTimeCompare([]byte(state), []byte(r.URL.Query().Get("state"))) != 1 {
			fail("state does not match")
			return
		}
		identity, err := provider.Finish(r.Context(), attempt, r.URL.Query().Get("code"))
		if err != nil {
			fail("finishing the attempt failed: " + err.Error())
			return
		}
		token, err := signIn.SignIn(r.Context(), identity)
		if err != nil {
			writeProblem(w, problemFor(err))
			return
		}
		setSessionCookie(w, r, publicURL, token)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// clearAttemptCookie tells the browser to forget the attempt cookie at once.
func clearAttemptCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: AttemptCookie, Path: "/auth/oidc", MaxAge: -1})
}

// attemptCookieValue keeps the state and the attempt together as
// "<state>.<attempt>". The state is base32 and the attempt base64url, so
// neither holds a dot, and the callback can split them again.
func attemptCookieValue(state, attempt string) string {
	return state + "." + attempt
}

// signInFailedPage is what a person sees when signing in through the provider
// did not work. It never says why: the log holds the cause.
const signInFailedPage = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Sign-in failed</title>
</head>
<body>
<h1>Sign-in failed</h1>
<p>Something went wrong while attempting to sign in. This can happen if the sign-in attempt was cancelled at the provider, or due to bad credentials, a timeout, or other problem. Please try again.</p>
<p><a href="/auth/sign-in">Try signing in again</a></p>
</body>
</html>
`

// signInFailed answers the sign-in failed page with status 400.
func signInFailed(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(signInFailedPage))
}
