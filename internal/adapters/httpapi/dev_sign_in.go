package httpapi

import (
	"net/http"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// DevIssuer is the issuer of the identities the dev sign-in signs in: anyone
// can sign in as any name, so it's only for development and tests.
const DevIssuer = "dev"

// devSignInPage is the dev sign-in's form: a plain page asking for a name,
// posting to POST /auth/dev. It's for developers only, so it has no styling.
const devSignInPage = `<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>Dev sign-in</title></head>
<body>
<h1>Dev sign-in</h1>
<form method="post" action="/auth/dev">
<label>Name <input name="name" required autofocus></label>
<button type="submit">Sign in</button>
</form>
</body>
</html>
`

// devSignInForm answers GET /auth/dev with devSignInPage.
func devSignInForm(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(devSignInPage))
}

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
