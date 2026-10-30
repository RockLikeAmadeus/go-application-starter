package mobile

import "net/http"

// AssetHandler adapts next for Wails' Android asset path. Android asks for
// /index.html, which Go's file server answers with a redirect to "./", and
// Wails treats any status but 200 as "not found". So /index.html is served
// as /.
func AssetHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/index.html" {
			r = r.Clone(r.Context())
			r.URL.Path = "/"
		}
		next.ServeHTTP(w, r)
	})
}
