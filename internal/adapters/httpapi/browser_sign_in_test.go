package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
)

func TestAServerWithNoClientSignInAnswersBothBrowserSignInOperationsNotFound(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{SignIn: &signInRecorder{}, IsDevSignInOn: true})
	for _, method := range []string{http.MethodPost, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/browser-sign-in", strings.NewReader(`{"isSignUp":false}`))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", rec.Code)
			}
			var body struct{ Type string }
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil || body.Type != httpapi.ProblemNotFound {
				t.Errorf("body type = %q (err %v), want %q", body.Type, err, httpapi.ProblemNotFound)
			}
		})
	}
}

// clientSession is an httpapi.ClientSignIn that answers a fixed session.
type clientSession struct{ session httpapi.ClientSession }

func (c clientSession) CurrentSession(context.Context) (httpapi.ClientSession, error) {
	return c.session, nil
}

func TestAHandlerWithAClientSignInAnswersTheHostedServersSessionAndSaysToSignInThroughTheBrowser(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{
		ClientSignIn: clientSession{httpapi.ClientSession{IsSignInOn: true, Name: "Ada"}},
	})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/session", nil))

	var got session
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if !got.IsSignInOn || got.Name == nil || *got.Name != "Ada" || got.SignInMethod != "browser" {
		t.Errorf("session = %+v, want sign-in on, name Ada, method browser", got)
	}
}
