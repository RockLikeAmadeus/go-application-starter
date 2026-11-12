package httpapi

import (
	"context"
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
	Finish(ctx context.Context, attempt, code string) (domain.Identity, error)
}

// oidcStart answers GET /auth/oidc: it keeps the attempt in a cookie and sends
// the browser to the provider.
func oidcStart(provider SignInProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authURL, state, attempt := provider.Start()
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
func oidcCallback(provider SignInProvider, signIn ports.SignInService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(AttemptCookie)
		if err != nil {
			writeProblem(w, problemFor(err))
			return
		}
		_, attempt, _ := strings.Cut(cookie.Value, ".")
		identity, err := provider.Finish(r.Context(), attempt, r.URL.Query().Get("code"))
		if err != nil {
			writeProblem(w, problemFor(err))
			return
		}
		token, err := signIn.SignIn(r.Context(), identity)
		if err != nil {
			writeProblem(w, problemFor(err))
			return
		}
		clearAttemptCookie(w)
		setSessionCookie(w, r, token)
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
