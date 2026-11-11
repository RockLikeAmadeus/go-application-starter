package httpapi

import (
	"net/http"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// DevIssuer is the issuer of the identities the dev sign-in signs in: anyone
// can sign in as any name, so it's only for development and tests.
const DevIssuer = "dev"

// devSignIn answers POST /auth/dev: it signs in as the form's name (the name
// is also the subject, so one name is one account), sets the session cookie,
// and goes to the web app.
func devSignIn(signIn ports.SignInService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		token, err := signIn.SignIn(r.Context(), domain.Identity{Issuer: DevIssuer, Subject: name, Name: name})
		if err != nil {
			writeProblem(w, problemFor(err))
			return
		}
		setSessionCookie(w, r, token)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}
