package httpapi

import (
	"context"
	"net/http"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi/openapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/workspace"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
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
				r = r.WithContext(withAccount(r.Context(), account))
			}
		}
		next.ServeHTTP(w, r)
	})
}

// setSessionCookie gives the browser the session's token in a cookie that
// page scripts can't read (HttpOnly) and other sites' links and forms don't
// send, except top-level navigation (SameSite=Lax).
func setSessionCookie(w http.ResponseWriter, r *http.Request, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(domain.SessionLifetime.Seconds()),
		Secure:   r.TLS != nil,
	})
}

// clearSessionCookie tells the browser to forget the session cookie at once
// (Max-Age < 0).
func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: SessionCookie, Path: "/", MaxAge: -1})
}

// accountKey is the context key for the signed-in account.
type accountKey struct{}

// withAccount runs ctx as the signed-in account: in its workspace, with the
// account itself for accountFrom.
func withAccount(ctx context.Context, account domain.Account) context.Context {
	ctx = workspace.WithID(ctx, account.WorkspaceID)
	return context.WithValue(ctx, accountKey{}, account)
}

// accountFrom returns the signed-in account withAccount put in ctx, and
// whether there is one.
func accountFrom(ctx context.Context) (domain.Account, bool) {
	account, ok := ctx.Value(accountKey{}).(domain.Account)
	return account, ok
}

// The Session endpoint.

func (h *Handlers) GetSession(ctx context.Context, _ openapi.GetSessionRequestObject) (openapi.GetSessionResponseObject, error) {
	account, ok := accountFrom(ctx)
	if !ok {
		return openapi.GetSession200JSONResponse{IsSignInOn: true}, nil
	}
	return openapi.GetSession200JSONResponse{IsSignInOn: true, Name: &account.Identity.Name}, nil
}
