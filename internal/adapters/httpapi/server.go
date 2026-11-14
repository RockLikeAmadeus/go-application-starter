package httpapi

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi/openapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// Options configures the HTTP handler.
type Options struct {
	// sample:start
	Items ports.ItemService
	// sample:end
	// Entitlements answers GET /api/v1/entitlements: the current
	// workspace's plan and its limits.
	Entitlements ports.EntitlementsService
	Logger       *slog.Logger
	// Ready reports whether storage can be reached, for GET /readyz. When
	// nil, the server is always ready (as with memory storage).
	Ready func(context.Context) error
	// WebAssets is the built SPA (web/dist). When nil, "/" serves a hint page.
	WebAssets fs.FS
	// IsDataLocal is set when the data is on the same device as the web app
	// (the desktop app's local-only mode), so it needs no connection.
	IsDataLocal bool
	// Warnings answers GET /api/v1/warnings: messages every front end shows
	// as given, such as bootstrap.App.Warnings. None when nil.
	Warnings []string
	// SignIn knows each session's account. When set, a request with a
	// session cookie runs in that account's workspace.
	SignIn ports.SignInService
	// IsDevSignInOn serves the dev sign-in under /auth/dev. It needs SignIn.
	IsDevSignInOn bool
	// SignInProvider serves the OIDC sign-in under /auth/oidc. It needs SignIn.
	SignInProvider SignInProvider
	// PublicURL is the address people reach the server at. When it starts with
	// https://, cookies are Secure even over plain HTTP (a proxy ended TLS).
	PublicURL string
	// ClientSignIn is set by the apps' in-process handler in remote mode,
	// where the session is the hosted server's, not this handler's.
	ClientSignIn ClientSignIn
	// OpenBrowser opens an address in the system browser, for ClientSignIn.
	OpenBrowser func(address string) error
}

// ClientSession is what a hosted server says about the current session.
type ClientSession struct {
	IsSignInOn bool
	Name       string
}

// ErrSignInCancelled is what a ClientSignIn returns from SignInThroughBrowser
// when CancelBrowserSignIn ended it.
var ErrSignInCancelled = errors.New("the sign-in was cancelled")

// ClientSignIn signs an app in to a hosted server.
type ClientSignIn interface {
	CurrentSession(ctx context.Context) (ClientSession, error)
	// SignInThroughBrowser signs in, or with isSignUp creates an account,
	// handing the address to open, and returns when it finishes.
	SignInThroughBrowser(ctx context.Context, isSignUp bool, open func(address string) error) error
	// CancelBrowserSignIn ends the sign-in in progress, if any, so that
	// SignInThroughBrowser returns ErrSignInCancelled.
	CancelBrowserSignIn()
}

// NewHandler builds the full HTTP surface: the REST API, health probes and
// the single-page app.
func NewHandler(opts Options) http.Handler {
	log := opts.Logger
	if log == nil {
		log = slog.Default()
	}
	ready := opts.Ready
	if ready == nil {
		ready = func(context.Context) error { return nil }
	}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := ready(r.Context()); err != nil {
			log.ErrorContext(r.Context(), "not ready", "err", err)
			http.Error(w, "storage unavailable", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("ready"))
	})

	strict := openapi.NewStrictHandlerWithOptions(NewHandlers(opts), nil, openapi.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  requestErrorHandler,
		ResponseErrorHandlerFunc: responseErrorHandler(log),
	})
	openapi.HandlerWithOptions(strict, openapi.StdHTTPServerOptions{
		BaseRouter:       mux,
		ErrorHandlerFunc: requestErrorHandler,
	})
	// Unknown API routes get a problem document rather than the SPA.
	mux.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) {
		writeProblem(w, problem(http.StatusNotFound, ProblemNotFound, "no such endpoint", nil))
	})

	if opts.IsDevSignInOn {
		mux.HandleFunc("GET /auth/sign-in", signInGoesTo("/auth/dev"))
		mux.HandleFunc("GET /auth/sign-up", signInGoesTo("/auth/dev"))
		mux.HandleFunc("GET /auth/dev", devSignInForm)
		mux.Handle("POST /auth/dev", devSignIn(opts.SignIn, opts.PublicURL))
	}
	if opts.SignInProvider != nil {
		mux.HandleFunc("GET /auth/sign-in", signInGoesTo("/auth/oidc"))
		mux.HandleFunc("GET /auth/sign-up", signInGoesTo("/auth/oidc/sign-up"))
		mux.HandleFunc("GET /auth/oidc", oidcStart(opts.SignInProvider.Start, opts.PublicURL))
		mux.HandleFunc("GET /auth/oidc/sign-up", oidcStart(opts.SignInProvider.StartSignUp, opts.PublicURL))
		mux.HandleFunc("GET /auth/oidc/callback", oidcCallback(log, opts.SignInProvider, opts.SignIn, opts.PublicURL))
	}
	if opts.SignIn != nil {
		mux.Handle("POST /auth/sign-out", signOut(opts.SignIn))
		mux.HandleFunc("GET /auth/native", nativeSignIn(opts.PublicURL))
		mux.HandleFunc("GET /auth/native/finish", nativeFinish(opts.SignIn))
		mux.HandleFunc("POST /auth/native/token", nativeToken(opts.SignIn))
	}

	mux.Handle("/", spaHandler(opts.WebAssets))

	var handler http.Handler = mux
	if opts.SignIn != nil {
		handler = sessionMiddleware(opts.SignIn, handler)
		handler = refuseCrossSite(handler)
	}
	return recoverer(log, requestLogger(log, handler))
}

// spaHandler serves static files and falls back to index.html so client-side
// routes work on reload.
func spaHandler(assets fs.FS) http.Handler {
	if assets == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`<p>API is running. The web UI was not embedded in this build; ` +
				`run <code>task web:dev</code> or build with <code>-tags embedweb</code>.</p>`))
		})
	}
	fileServer := http.FileServerFS(assets)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" {
			name = "index.html"
		}
		if _, err := fs.Stat(assets, name); errors.Is(err, fs.ErrNotExist) {
			r.URL.Path = "/"
		}
		// No MIME table Go reads knows the web app manifest (nor does the
		// distroless image), so the file server would sniff it as text/plain.
		if path.Ext(name) == ".webmanifest" {
			w.Header().Set("Content-Type", "application/manifest+json")
		}
		fileServer.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func requestLogger(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.InfoContext(r.Context(), "http request",
			"method", r.Method, "path", r.URL.Path, "status", rec.status, "duration", time.Since(start))
	})
}

func recoverer(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				log.ErrorContext(r.Context(), "panic", "value", v, "path", r.URL.Path)
				writeProblem(w, internalProblem())
			}
		}()
		next.ServeHTTP(w, r)
	})
}
