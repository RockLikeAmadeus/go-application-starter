package oidc_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
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

func TestNewFailsForAnUnreachableIssuer(t *testing.T) {
	fake := oidctest.New(t)
	issuer := fake.Issuer()
	fake.Close()

	p, err := oidc.New(context.Background(), oidc.Settings{Issuer: issuer})
	if err == nil {
		t.Fatal("New succeeded for an issuer that is down")
	}
	if p != nil {
		t.Errorf("New returned a provider along with its error: %v", p)
	}
}

func startedProvider(t *testing.T) (*oidc.Provider, *oidctest.Provider) {
	t.Helper()
	fake := oidctest.New(t)
	p, err := oidc.New(context.Background(), oidc.Settings{
		Issuer:       fake.Issuer(),
		ClientID:     "the-client",
		ClientSecret: "the-secret",
		RedirectURL:  "https://app.example.com/auth/oidc/callback",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return p, fake
}

func TestStartGivesTheAuthorizationURLAndAnAttemptHoldingItsSecrets(t *testing.T) {
	p, fake := startedProvider(t)

	rawURL, attempt := p.Start()

	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parsing %q: %v", rawURL, err)
	}
	if got, want := u.Scheme+"://"+u.Host+u.Path, fake.Issuer()+"/authorize"; got != want {
		t.Errorf("authorization endpoint = %q, want %q", got, want)
	}
	q := u.Query()
	for key, want := range map[string]string{
		"response_type":         "code",
		"client_id":             "the-client",
		"redirect_uri":          "https://app.example.com/auth/oidc/callback",
		"scope":                 "openid profile email",
		"state":                 attempt.State,
		"nonce":                 attempt.Nonce,
		"code_challenge_method": "S256",
	} {
		if got := q.Get(key); got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
	for name, v := range map[string]string{"state": attempt.State, "nonce": attempt.Nonce, "verifier": attempt.Verifier} {
		if v == "" {
			t.Errorf("attempt has no %s", name)
		}
	}
	sum := sha256.Sum256([]byte(attempt.Verifier))
	if got, want := q.Get("code_challenge"), base64.RawURLEncoding.EncodeToString(sum[:]); got != want {
		t.Errorf("code_challenge = %q, want the S256 of the verifier, %q", got, want)
	}
}

func TestTwoStartsShareNothing(t *testing.T) {
	p, _ := startedProvider(t)

	_, a := p.Start()
	_, b := p.Start()

	if a.State == b.State || a.Nonce == b.Nonce || a.Verifier == b.Verifier {
		t.Errorf("two attempts share a secret: %+v and %+v", a, b)
	}
}
