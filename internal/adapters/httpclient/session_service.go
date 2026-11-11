package httpclient

import (
	"context"

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
}

// NewSessionService returns a client for the server at baseURL, e.g.
// "http://localhost:8080".
func NewSessionService(baseURL string, opts ...apiclient.ClientOption) (*SessionService, error) {
	api, err := newAPI(baseURL, opts...)
	if err != nil {
		return nil, err
	}
	return &SessionService{api: api}, nil
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
