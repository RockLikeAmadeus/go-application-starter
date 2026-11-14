package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi/openapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/workspace"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// SessionCookie is the name of the cookie that holds the session's token.
const SessionCookie = "session"

// sessionToken returns the request's session token: the bearer token the
// command line sends, else the browser's session cookie. A bearer token that
// turns out invalid is not replaced by the cookie.
func sessionToken(r *http.Request) (string, bool) {
	if token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok && token != "" {
		return token, true
	}
	if cookie, err := r.Cookie(SessionCookie); err == nil {
		return cookie.Value, true
	}
	return "", false
}

// sessionMiddleware runs a request with a valid session token in its
// account's workspace (workspace.WithID). Any other request runs with no
// workspace, so the use cases refuse it with domain.ErrNoWorkspace: signed out
// fails closed.
func sessionMiddleware(signIn ports.SignInService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token, ok := sessionToken(r); ok {
			if account, err := signIn.CurrentAccount(r.Context(), token); err == nil {
				r = r.WithContext(withAccount(r.Context(), account))
			}
		}
		next.ServeHTTP(w, r)
	})
}

// setSessionCookie gives the browser the session's token in a cookie that
// page scripts can't read (HttpOnly) and other sites' links and forms don't
// send, except top-level navigation (SameSite=Lax).
func setSessionCookie(w http.ResponseWriter, r *http.Request, publicURL, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(domain.SessionLifetime.Seconds()),
		Secure:   isSecure(r, publicURL),
	})
}

// isSecure says whether cookies should be marked Secure: the request came over
// TLS, or the public URL is https:// (a proxy may end TLS before the server).
func isSecure(r *http.Request, publicURL string) bool {
	return r.TLS != nil || strings.HasPrefix(publicURL, "https://")
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
	if h.clientSignIn != nil {
		hosted, err := h.hostedSession(ctx)
		return openapi.GetSession200JSONResponse(hosted), err
	}
	session := openapi.GetSession200JSONResponse{IsSignInOn: h.isSignInOn, SignInMethod: openapi.Page}
	if account, ok := accountFrom(ctx); ok {
		session.Name = &account.Identity.Name
	}
	return session, nil
}

// hostedSession answers the session of the hosted server this handler is the
// app's window to. The app signs in through the browser.
func (h *Handlers) hostedSession(ctx context.Context) (openapi.Session, error) {
	hosted, err := h.clientSignIn.CurrentSession(ctx)
	if err != nil {
		return openapi.Session{}, err
	}
	session := openapi.Session{IsSignInOn: hosted.IsSignInOn, SignInMethod: openapi.Browser}
	if hosted.Name != "" {
		session.Name = &hosted.Name
	}
	return session, nil
}

// The browser sign-in endpoints belong to the apps' in-process handler. A
// server, which has no client sign-in, answers both as if they didn't exist.

func (h *Handlers) StartBrowserSignIn(ctx context.Context, request openapi.StartBrowserSignInRequestObject) (openapi.StartBrowserSignInResponseObject, error) {
	if h.clientSignIn == nil {
		return nil, domain.ErrNotFound
	}
	if err := h.clientSignIn.SignInThroughBrowser(ctx, request.Body.IsSignUp, h.openBrowser); err != nil {
		return nil, err
	}
	session, err := h.hostedSession(ctx)
	if err != nil {
		return nil, err
	}
	return openapi.StartBrowserSignIn200JSONResponse(session), nil
}

func (h *Handlers) CancelBrowserSignIn(context.Context, openapi.CancelBrowserSignInRequestObject) (openapi.CancelBrowserSignInResponseObject, error) {
	return nil, domain.ErrNotFound
}
