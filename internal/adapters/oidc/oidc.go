package oidc

import (
	"context"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
)

// Settings says which provider to sign in through, and as which client.
type Settings struct {
	Issuer string
}

// Provider is a discovered sign-in provider.
type Provider struct {
	provider *gooidc.Provider
}

// New discovers the provider from its issuer.
func New(ctx context.Context, s Settings) (*Provider, error) {
	p, err := gooidc.NewProvider(ctx, s.Issuer)
	if err != nil {
		return nil, err
	}
	return &Provider{provider: p}, nil
}
