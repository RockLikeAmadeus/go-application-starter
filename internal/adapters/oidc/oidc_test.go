package oidc_test

import (
	"context"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/oidc"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/oidc/oidctest"
)

func TestNewDiscoversTheProviderFromItsIssuer(t *testing.T) {
	fake := oidctest.New(t)

	p, err := oidc.New(context.Background(), oidc.Settings{Issuer: fake.Issuer()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if p == nil {
		t.Fatal("New returned no provider")
	}
	if fake.DiscoveryRequests() == 0 {
		t.Error("the provider's discovery document was never requested")
	}
}
