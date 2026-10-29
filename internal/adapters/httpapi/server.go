package httpapi

import (
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
	Items  ports.ItemService
	Logger *slog.Logger
	// WebAssets is the built SPA (web/dist). When nil, "/" serves a hint page.
	WebAssets fs.FS
}

// NewHandler builds the full HTTP surface: the REST API, health probes and
// the single-page app.
func NewHandler(opts Options) http.Handler {
	log := opts.Logger
	if log == nil {
		log = slog.Default()
	}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if _, err := opts.Items.ListItems(r.Context()); err != nil {
			http.Error(w, "storage unavailable", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("ready"))
	})

	strict := openapi.NewStrictHandlerWithOptions(NewHandlers(opts.Items), nil, openapi.StrictHTTPServerOptions{
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

	mux.Handle("/", spaHandler(opts.WebAssets))

	return recoverer(log, requestLogger(log, mux))
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
				writeProblem(w, problem(http.StatusInternalServerError, ProblemInternal, "internal server error", nil))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
