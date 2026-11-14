package httpclient

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient/apiclient"
)

// Session is what the server says about the request's session: whether it
// signs people in, and the signed-in account's name, empty when no one is
// signed in.
type Session struct {
	IsSignInOn bool
	Name       string
}

// SessionService reads the server's session endpoint, GET /api/v1/session.
// It isn't a driving port: the core keeps sessions behind
// ports.SignInService, and only a client of the server needs to ask it.
type SessionService struct {
	api *apiclient.ClientWithResponses
	raw *apiclient.Client
}

// NewSessionService returns a client for the server at baseURL, e.g.
// "http://localhost:8080".
func NewSessionService(baseURL string, opts ...apiclient.ClientOption) (*SessionService, error) {
	raw, err := newClient(baseURL, opts...)
	if err != nil {
		return nil, err
	}
	return &SessionService{api: &apiclient.ClientWithResponses{ClientInterface: raw}, raw: raw}, nil
}

// CurrentSession returns the session the server puts the request in.
func (s *SessionService) CurrentSession(ctx context.Context) (Session, error) {
	res, err := s.api.GetSessionWithResponse(ctx)
	if err != nil {
		return Session{}, err
	}
	if res.JSON200 == nil {
		return Session{}, toError(res.StatusCode(), res.ApplicationproblemJSONDefault)
	}
	return Session{IsSignInOn: res.JSON200.IsSignInOn, Name: deref(res.JSON200.Name)}, nil
}

// EndSession ends the session the request's token names, with
// POST /auth/sign-out. The server answers success with a redirect to the web
// app, and a refusal with a problem document.
func (s *SessionService) EndSession(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.raw.Server+"auth/sign-out", nil) // NewClient ends Server with a slash
	if err != nil {
		return err
	}
	for _, edit := range s.raw.RequestEditors {
		if err := edit(ctx, req); err != nil {
			return err
		}
	}
	res, err := s.signOutClient().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 400 {
		return nil
	}
	var problem apiclient.Problem
	if json.NewDecoder(res.Body).Decode(&problem) != nil {
		return toError(res.StatusCode, nil)
	}
	return toError(res.StatusCode, &problem)
}

// signOutClient is the API client's HTTP client, except that it follows no
// redirect: the server answers a sign-out with the redirect it sends
// browsers, and that answer is the result.
func (s *SessionService) signOutClient() apiclient.HttpRequestDoer {
	httpClient, ok := s.raw.Client.(*http.Client)
	if !ok {
		return s.raw.Client
	}
	noRedirect := *httpClient
	noRedirect.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &noRedirect
}
