package httpapi

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// signInGoesTo answers GET /auth/sign-in by sending the browser to the
// configured sign-in adapter's own page, such as the dev sign-in's form. A
// local "return" path goes along, so the adapter can return to it afterwards.
func signInGoesTo(adapterPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		target := adapterPath
		if r.FormValue("return") != "" {
			if path := returnPath(r); path != "/" {
				target += "?return=" + url.QueryEscape(path)
			}
		}
		http.Redirect(w, r, target, http.StatusFound)
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

// clientSignOut answers POST /auth/sign-out in an app's in-process handler: it
// signs out through the client sign-in and goes to the web app. A server that
// couldn't end the session doesn't stop it, as the token is forgotten anyway.
func clientSignOut(signIn ClientSignIn) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := signIn.SignOut(r.Context()); err != nil && !errors.Is(err, ErrSessionNotEnded) {
			writeProblem(w, problemFor(err))
			return
		}
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
	return localPath(r.FormValue("return"))
}

// localPath is path when it is a path on this server, else "/". See returnPath
// for what it refuses.
func localPath(path string) string {
	parsed, err := url.Parse(path)
	if err != nil || parsed.Scheme != "" || parsed.Host != "" ||
		!strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.HasPrefix(path, `/\`) {
		return "/"
	}
	return path
}
