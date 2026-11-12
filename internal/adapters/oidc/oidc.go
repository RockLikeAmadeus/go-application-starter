package oidc

import (
	"context"
	"crypto/rand"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// Settings says which provider to sign in through, and as which client.
type Settings struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// Provider is a discovered sign-in provider.
type Provider struct {
	provider *gooidc.Provider
	config   oauth2.Config
}

// New discovers the provider from its issuer.
func New(ctx context.Context, s Settings) (*Provider, error) {
	p, err := gooidc.NewProvider(ctx, s.Issuer)
	if err != nil {
		return nil, err
	}
	return &Provider{
		provider: p,
		config: oauth2.Config{
			ClientID:     s.ClientID,
			ClientSecret: s.ClientSecret,
			RedirectURL:  s.RedirectURL,
			Endpoint:     p.Endpoint(),
			Scopes:       []string{gooidc.ScopeOpenID, "profile", "email"},
		},
	}, nil
}

// Attempt is the secrets of one sign-in in progress, kept until the provider
// sends the person back.
type Attempt struct {
	State    string
	Nonce    string
	Verifier string
}

// Start begins a sign-in: the URL to send the person to, and the attempt to
// keep until they return.
func (p *Provider) Start() (string, Attempt) {
	attempt := Attempt{
		State:    rand.Text(),
		Nonce:    rand.Text(),
		Verifier: oauth2.GenerateVerifier(),
	}
	authURL := p.config.AuthCodeURL(attempt.State,
		gooidc.Nonce(attempt.Nonce),
		oauth2.S256ChallengeOption(attempt.Verifier),
	)
	return authURL, attempt
}
