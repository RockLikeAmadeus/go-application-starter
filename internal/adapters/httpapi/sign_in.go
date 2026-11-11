package httpapi

import "net/http"

// signInGoesTo answers GET /auth/sign-in by sending the browser to the
// configured sign-in adapter's own page, such as the dev sign-in's form.
func signInGoesTo(adapterPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, adapterPath, http.StatusFound)
	}
}
