package oidc

import (
	"cmp"
	"context"
	"crypto/rand"
	"errors"
	"fmt"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
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
	verifier *gooidc.IDTokenVerifier
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
		verifier: p.Verifier(&gooidc.Config{ClientID: s.ClientID}),
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

// Finish completes a sign-in: it swaps the code for tokens, sending the
// attempt's PKCE verifier, checks the ID token, and returns who the provider
// says signed in.
func (p *Provider) Finish(ctx context.Context, attempt Attempt, code string) (domain.Identity, error) {
	token, err := p.config.Exchange(ctx, code, oauth2.VerifierOption(attempt.Verifier))
	if err != nil {
		return domain.Identity{}, fmt.Errorf("swapping the code: %w", err)
	}
	raw, ok := token.Extra("id_token").(string)
	if !ok {
		return domain.Identity{}, errors.New("the provider sent no ID token")
	}
	idToken, err := p.verifier.Verify(ctx, raw)
	if err != nil {
		return domain.Identity{}, fmt.Errorf("checking the ID token: %w", err)
	}
	var claims struct {
		Name              string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
		Email             string `json:"email"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return domain.Identity{}, fmt.Errorf("reading the ID token's claims: %w", err)
	}
	// A provider that sends no name claim still gives the person something to be called.
	name := cmp.Or(claims.Name, claims.PreferredUsername, claims.Email, idToken.Subject)
	return domain.Identity{
		Issuer:  idToken.Issuer,
		Subject: idToken.Subject,
		Name:    name,
	}, nil
}
