package mobile

import (
	"net/http"
	"strings"
)

// AssetHandler adapts next for Wails' Android asset path, which sends every
// request to Go as a GET with no body and treats any status but 200 as "not
// found".
//   - /api/ paths answer 404. API calls must go over the bridge (see Serve),
//     so one that reaches the asset path fails loudly instead of being
//     answered as a GET.
//   - /index.html, which Android asks for, is served as /, because Go's file
//     server answers it with a redirect to "./".
func AssetHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/index.html" {
			r = r.Clone(r.Context())
			r.URL.Path = "/"
		}
		next.ServeHTTP(w, r)
	})
}
