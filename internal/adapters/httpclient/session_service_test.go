package httpclient_test

import (
	"context"
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
