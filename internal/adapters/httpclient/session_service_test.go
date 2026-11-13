package httpclient_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient/apiclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/app"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/testsupport"
)

// The client reads the server's session endpoint back: with the session
// cookie, sign-in is on and the session is the signed-in account's.
func TestCurrentSession_NamesTheSignedInAccount(t *testing.T) {
	ctx := context.Background()
	signIn := app.NewSignInService(testsupport.StoredAccounts{}, testsupport.StoredSessions{}, nil, testsupport.FixedToken("token-a"),
		testsupport.NewSequentialIDs("id"), testsupport.NewFixedClock(time.Date(2026, 11, 11, 9, 0, 0, 0, time.UTC)))
	token, err := signIn.SignIn(ctx, domain.Identity{Issuer: "dev", Subject: "Ada", Name: "Ada"})
	if err != nil {
		t.Fatalf("SignIn: %v", err)
	}
	srv := httptest.NewServer(httpapi.NewHandler(httpapi.Options{SignIn: signIn}))
	t.Cleanup(srv.Close)
	withCookie := apiclient.WithRequestEditorFn(func(_ context.Context, r *http.Request) error {
		r.AddCookie(&http.Cookie{Name: httpapi.SessionCookie, Value: token})
		return nil
	})
	client, err := httpclient.NewSessionService(srv.URL, withCookie)
	if err != nil {
		t.Fatal(err)
	}

	got, err := client.CurrentSession(ctx)
	if err != nil {
		t.Fatalf("CurrentSession: %v", err)
	}
	want := httpclient.Session{IsSignInOn: true, Name: "Ada"}
	if got != want {
		t.Errorf("CurrentSession = %+v, want %+v", got, want)
	}
}

// The client ends the session its token source names: afterwards the server no
// longer knows the token.
func TestEndSession_EndsTheSessionTheTokenSourceNames(t *testing.T) {
	ctx := context.Background()
	signIn := app.NewSignInService(testsupport.StoredAccounts{}, testsupport.StoredSessions{}, nil, testsupport.FixedToken("token-a"),
		testsupport.NewSequentialIDs("id"), testsupport.NewFixedClock(time.Date(2026, 11, 11, 9, 0, 0, 0, time.UTC)))
	token, err := signIn.SignIn(ctx, domain.Identity{Issuer: "dev", Subject: "Ada", Name: "Ada"})
	if err != nil {
		t.Fatalf("SignIn: %v", err)
	}
	srv := httptest.NewServer(httpapi.NewHandler(httpapi.Options{SignIn: signIn}))
	t.Cleanup(srv.Close)
	client, err := httpclient.NewSessionService(srv.URL, httpclient.WithTokenSource(func(context.Context) (string, error) { return token, nil }))
	if err != nil {
		t.Fatal(err)
	}

	if err := client.EndSession(ctx); err != nil {
		t.Fatalf("EndSession: %v", err)
	}

	got, err := client.CurrentSession(ctx)
	if err != nil {
		t.Fatalf("CurrentSession: %v", err)
	}
	if got.Name != "" {
		t.Errorf("CurrentSession name = %q after EndSession, want no one signed in", got.Name)
	}
}

// The server answers a sign-out with a redirect to the web app. The client
// takes that redirect as the answer, and doesn't follow it.
func TestEndSession_TakesTheServersRedirectAsSuccessWithoutFollowingIt(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/sign-out", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
	srv := httptest.NewServer(mux) // "/" is unknown here, so following the redirect would give a 404
	t.Cleanup(srv.Close)
	client, err := httpclient.NewSessionService(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	if err := client.EndSession(context.Background()); err != nil {
		t.Errorf("EndSession = %v, want nil", err)
	}
}

// A refused sign-out comes back as the domain error the server's problem
// document names.
func TestEndSession_ReturnsTheErrorOfARefusedSignOut(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/sign-out", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"type":"no_workspace","title":"No workspace","status":401}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client, err := httpclient.NewSessionService(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	if err := client.EndSession(context.Background()); !errors.Is(err, domain.ErrNoWorkspace) {
		t.Errorf("EndSession = %v, want domain.ErrNoWorkspace", err)
	}
}
