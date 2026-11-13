package httpapi

import (
	"net/http"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// signInGoesTo answers GET /auth/sign-in by sending the browser to the
// configured sign-in adapter's own page, such as the dev sign-in's form.
func signInGoesTo(adapterPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, adapterPath, http.StatusFound)
	}
}

// signOut answers POST /auth/sign-out: it ends the session (the bearer
// token's, else the cookie's), clears the
// session cookie, and goes to the web app.
func signOut(signIn ports.SignInService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token, ok := sessionToken(r); ok {
			if err := signIn.SignOut(r.Context(), token); err != nil {
				writeProblem(w, problemFor(err))
				return
			}
		}
		clearSessionCookie(w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}
