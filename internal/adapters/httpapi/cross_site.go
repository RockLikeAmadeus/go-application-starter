package httpapi

import "net/http"

// refuseCrossSite refuses a request that changes something (not GET, HEAD,
// or OPTIONS) unless it comes from the server's own origin, as the browser
// reports in Sec-Fetch-Site or, without it, Origin. The session cookie's
// SameSite=Lax already keeps it off most requests from other sites; this
// closes the rest, such as another origin on the same site. Clients that
// aren't browsers send neither header, so they pass.
func refuseCrossSite(next http.Handler) http.Handler {
	protection := http.NewCrossOriginProtection()
	protection.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeProblem(w, problem(http.StatusForbidden, ProblemCrossSite,
			"The server refused this request because it did not come from the application's own pages. Please try again from the application.", nil))
	}))
	return protection.Handler(next)
}
