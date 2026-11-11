package httpapi

import (
	"net/http"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/workspace"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// SessionCookie is the name of the cookie that holds the session's token.
const SessionCookie = "session"

// sessionMiddleware runs a request with a valid session cookie in its
// account's workspace (workspace.WithID). Any other request runs with no
// workspace, so the use cases refuse it with domain.ErrNoWorkspace: signed out
// fails closed.
func sessionMiddleware(signIn ports.SignInService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie(SessionCookie); err == nil {
			if account, err := signIn.CurrentAccount(r.Context(), cookie.Value); err == nil {
				r = r.WithContext(workspace.WithID(r.Context(), account.WorkspaceID))
			}
		}
		next.ServeHTTP(w, r)
	})
}
