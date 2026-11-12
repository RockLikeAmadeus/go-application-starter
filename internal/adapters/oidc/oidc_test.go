package oidc_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/url"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/oidc"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/oidc/oidctest"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
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
		ClientID:     oidctest.ClientID,
		ClientSecret: oidctest.ClientSecret,
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
		"client_id":             oidctest.ClientID,
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

// authorize plays the browser: it opens the authorization URL at the fake
// provider and returns the code (and state) it sends the person back with.
func authorize(t *testing.T, authURL string) url.Values {
	t.Helper()
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.Get(authURL)
	if err != nil {
		t.Fatalf("opening the authorization URL: %v", err)
	}
	defer resp.Body.Close()
	back, err := url.Parse(resp.Header.Get("Location"))
	if err != nil || back.Query().Get("code") == "" {
		t.Fatalf("the provider sent no code back: status %d, Location %q", resp.StatusCode, resp.Header.Get("Location"))
	}
	return back.Query()
}

func TestFinishSwapsTheCodeAndReturnsTheIdentity(t *testing.T) {
	p, fake := startedProvider(t)
	fake.SetPerson(oidctest.Person{Subject: "ada-123", Name: "Ada Lovelace"})
	authURL, attempt := p.Start()
	back := authorize(t, authURL)

	identity, err := p.Finish(context.Background(), attempt, back.Get("code"))

	if err != nil {
		t.Fatalf("Finish: %v", err)
	}
	want := domain.Identity{Issuer: fake.Issuer(), Subject: "ada-123", Name: "Ada Lovelace"}
	if identity != want {
		t.Errorf("identity = %+v, want %+v", identity, want)
	}
}

func TestFinishNamesThePersonByTheBestClaimTheProviderSent(t *testing.T) {
	for _, tc := range []struct {
		name   string
		person oidctest.Person
		want   string
	}{
		{"the name claim", oidctest.Person{Subject: "s1", Name: "Ada Lovelace", PreferredUsername: "ada", Email: "ada@example.com"}, "Ada Lovelace"},
		{"the preferred username", oidctest.Person{Subject: "s1", PreferredUsername: "ada", Email: "ada@example.com"}, "ada"},
		{"the email", oidctest.Person{Subject: "s1", Email: "ada@example.com"}, "ada@example.com"},
		{"the subject", oidctest.Person{Subject: "s1"}, "s1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, fake := startedProvider(t)
			fake.SetPerson(tc.person)
			authURL, attempt := p.Start()
			back := authorize(t, authURL)

			identity, err := p.Finish(context.Background(), attempt, back.Get("code"))

			if err != nil {
				t.Fatalf("Finish: %v", err)
			}
			if identity.Name != tc.want {
				t.Errorf("name = %q, want %q", identity.Name, tc.want)
			}
		})
	}
}

func TestFinishRefusesAnIDTokenThatCannotBeTrusted(t *testing.T) {
	for _, tc := range []struct {
		name  string
		fault oidctest.Fault
	}{
		{"signed with another key", oidctest.SignedWithAnotherKey},
		{"meant for another client", oidctest.MeantForAnotherClient},
		{"expired", oidctest.Expired},
		{"from another issuer", oidctest.FromAnotherIssuer},
		{"with another nonce", oidctest.WithAnotherNonce},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, fake := startedProvider(t)
			fake.SetPerson(oidctest.Person{Subject: "ada-123", Name: "Ada Lovelace"})
			fake.SetFault(tc.fault)
			authURL, attempt := p.Start()
			back := authorize(t, authURL)

			identity, err := p.Finish(context.Background(), attempt, back.Get("code"))

			if err == nil {
				t.Fatalf("Finish succeeded with %+v", identity)
			}
			if identity != (domain.Identity{}) {
				t.Errorf("Finish returned %+v along with its error", identity)
			}
		})
	}
}
