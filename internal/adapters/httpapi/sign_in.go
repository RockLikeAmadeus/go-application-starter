package httpapi

import (
	"net/http"
	"net/url"
	"strings"

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

// returnPath is where to send the browser after it signs in: the request's
// "return" value when it is a path on this server, else the web app. Anything
// else (another site's address, a protocol-relative "//host", a backslash
// that browsers read as a slash, or a control character such as a tab, which
// browsers drop from addresses and url.Parse refuses) would let a link make the sign-in an open
// redirect.
func returnPath(r *http.Request) string {
	path := r.FormValue("return")
	parsed, err := url.Parse(path)
	if err != nil || parsed.Scheme != "" || parsed.Host != "" ||
		!strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.HasPrefix(path, `/\`) {
		return "/"
	}
	return path
}
