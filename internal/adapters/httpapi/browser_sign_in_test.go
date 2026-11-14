package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
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

// clientSession is an httpapi.ClientSignIn that answers a fixed session, and
// records how it was asked to sign in.
type clientSession struct {
	session    httpapi.ClientSession
	signInErr  error
	isSignUp   *bool
	openedWith *string
}

func (c clientSession) CurrentSession(context.Context) (httpapi.ClientSession, error) {
	return c.session, nil
}

func (c clientSession) SignInThroughBrowser(_ context.Context, isSignUp bool, open func(address string) error) error {
	*c.isSignUp = isSignUp
	if err := open("http://browser.test/sign-in"); err != nil {
		return err
	}
	*c.openedWith = "http://browser.test/sign-in"
	return c.signInErr
}

func (c clientSession) CancelBrowserSignIn() {}

func (c clientSession) SignOut(context.Context) error { return nil }

// signOutRecorder is an httpapi.ClientSignIn that counts sign-outs and
// answers them with signOutErr.
type signOutRecorder struct {
	clientSession
	signOutErr error
	signedOut  *int
}

func (s signOutRecorder) SignOut(context.Context) error {
	*s.signedOut++
	return s.signOutErr
}

func TestSigningOutOfAHandlerWithAClientSignInSignsOutThroughItAndGoesToTheWebApp(t *testing.T) {
	cases := map[string]error{
		"the server ends the session":      nil,
		"the server can't end the session": httpapi.ErrSessionNotEnded,
	}
	for name, signOutErr := range cases {
		t.Run(name, func(t *testing.T) {
			var signedOut int
			handler := httpapi.NewHandler(httpapi.Options{
				ClientSignIn: signOutRecorder{signOutErr: signOutErr, signedOut: &signedOut},
			})
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/auth/sign-out", nil))

			if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/" {
				t.Errorf("status = %d, Location = %q; want %d to %q", rec.Code, rec.Header().Get("Location"), http.StatusSeeOther, "/")
			}
			if signedOut != 1 {
				t.Errorf("signed out %d times, want 1", signedOut)
			}
		})
	}
}

func TestAHandlerWithAClientSignInAnswersTheHostedServersSessionAndSaysToSignInThroughTheBrowser(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{
		ClientSignIn: clientSession{session: httpapi.ClientSession{IsSignInOn: true, Name: "Ada"}},
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

func TestStartingABrowserSignInSignsInThroughTheClientSignInAndAnswersTheSession(t *testing.T) {
	var isSignUp bool
	var opened string
	var handedAddress string
	handler := httpapi.NewHandler(httpapi.Options{
		ClientSignIn: clientSession{
			session:    httpapi.ClientSession{IsSignInOn: true, Name: "Ada"},
			isSignUp:   &isSignUp,
			openedWith: &opened,
		},
		OpenBrowser: func(address string) error { handedAddress = address; return nil },
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/browser-sign-in", strings.NewReader(`{"isSignUp":true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", rec.Code, rec.Body)
	}
	var got session
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Name == nil || *got.Name != "Ada" || got.SignInMethod != "browser" {
		t.Errorf("session = %+v, want name Ada, method browser", got)
	}
	if !isSignUp {
		t.Error("isSignUp = false, want true")
	}
	if handedAddress != "http://browser.test/sign-in" {
		t.Errorf("opener got %q, want the address the client sign-in opened", handedAddress)
	}
}

func TestAFailedBrowserSignInIsAProblemResponse(t *testing.T) {
	var isSignUp bool
	var opened string
	handler := httpapi.NewHandler(httpapi.Options{
		ClientSignIn: clientSession{signInErr: errors.New("the browser never came back"), isSignUp: &isSignUp, openedWith: &opened},
		OpenBrowser:  func(string) error { return nil },
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/browser-sign-in", strings.NewReader(`{"isSignUp":false}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	var body struct{ Type string }
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil || body.Type != httpapi.ProblemInternal {
		t.Errorf("body type = %q (err %v), want %q", body.Type, err, httpapi.ProblemInternal)
	}
}

// waitingSignIn is an httpapi.ClientSignIn whose browser sign-in waits until
// it is cancelled.
type waitingSignIn struct {
	clientSession
	waiting   chan struct{}
	cancelled chan struct{}
}

func (w waitingSignIn) SignInThroughBrowser(ctx context.Context, _ bool, _ func(string) error) error {
	close(w.waiting)
	select {
	case <-w.cancelled:
		return httpapi.ErrSignInCancelled
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w waitingSignIn) CancelBrowserSignIn() { close(w.cancelled) }

func TestCancellingABrowserSignInAnswersNoContentAndTheWaitingStartAProblem(t *testing.T) {
	signIn := waitingSignIn{waiting: make(chan struct{}), cancelled: make(chan struct{})}
	handler := httpapi.NewHandler(httpapi.Options{ClientSignIn: signIn})

	started := make(chan *httptest.ResponseRecorder)
	go func() {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/browser-sign-in", strings.NewReader(`{"isSignUp":false}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		started <- rec
	}()
	<-signIn.waiting

	cancel := httptest.NewRecorder()
	handler.ServeHTTP(cancel, httptest.NewRequest(http.MethodDelete, "/api/v1/browser-sign-in", nil))
	if cancel.Code != http.StatusNoContent {
		t.Fatalf("cancel status = %d, want 204", cancel.Code)
	}

	waiting := <-started
	if waiting.Code != http.StatusConflict {
		t.Fatalf("waiting status = %d, want 409", waiting.Code)
	}
	var body struct{ Type string }
	if err := json.NewDecoder(waiting.Body).Decode(&body); err != nil || body.Type != httpapi.ProblemCancelled {
		t.Errorf("body type = %q (err %v), want %q", body.Type, err, httpapi.ProblemCancelled)
	}
}
