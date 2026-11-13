package oidc_test

import (
	"context"
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

	rawURL, state, attempt := p.Start()

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
		"state":                 state,
		"code_challenge_method": "S256",
	} {
		if got := q.Get(key); got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
	// What the attempt holds is checked by Finish's tests: the nonce and the
	// PKCE verifier have to come back right for it to sign anyone in.
	for _, key := range []string{"state", "nonce", "code_challenge"} {
		if q.Get(key) == "" {
			t.Errorf("the URL has no %s", key)
		}
	}
	if attempt == "" {
		t.Error("Start gave no attempt")
	}
}

func queryOf(t *testing.T, rawURL string) url.Values {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parsing %q: %v", rawURL, err)
	}
	return u.Query()
}

func TestTwoStartsShareNothing(t *testing.T) {
	p, _ := startedProvider(t)

	urlA, stateA, a := p.Start()
	urlB, stateB, b := p.Start()

	qA, qB := queryOf(t, urlA), queryOf(t, urlB)
	if stateA == stateB || a == b || qA.Get("nonce") == qB.Get("nonce") || qA.Get("code_challenge") == qB.Get("code_challenge") {
		t.Errorf("two starts share a secret: %q and %q", urlA, urlB)
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
	authURL, _, attempt := p.Start()
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
			authURL, _, attempt := p.Start()
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
			authURL, _, attempt := p.Start()
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

func TestFinishRefusesWhenTheProviderRejectsTheCode(t *testing.T) {
	t.Run("a wrong PKCE verifier", func(t *testing.T) {
		p, fake := startedProvider(t)
		fake.SetPerson(oidctest.Person{Subject: "ada-123", Name: "Ada Lovelace"})
		authURL, _, _ := p.Start()
		back := authorize(t, authURL)
		_, _, attempt := p.Start() // another attempt, so another verifier

		identity, err := p.Finish(context.Background(), attempt, back.Get("code"))

		if err == nil {
			t.Fatalf("Finish succeeded with %+v", identity)
		}
		if identity != (domain.Identity{}) {
			t.Errorf("Finish returned %+v along with its error", identity)
		}
	})
	t.Run("a code used twice", func(t *testing.T) {
		p, fake := startedProvider(t)
		fake.SetPerson(oidctest.Person{Subject: "ada-123", Name: "Ada Lovelace"})
		authURL, _, attempt := p.Start()
		code := authorize(t, authURL).Get("code")
		if _, err := p.Finish(context.Background(), attempt, code); err != nil {
			t.Fatalf("first Finish: %v", err)
		}

		identity, err := p.Finish(context.Background(), attempt, code)

		if err == nil {
			t.Fatalf("second Finish succeeded with %+v", identity)
		}
		if identity != (domain.Identity{}) {
			t.Errorf("second Finish returned %+v along with its error", identity)
		}
	})
}

func TestStartSignUpAsksForTheProvidersSignUpPageOnlyWhenItListsPromptCreate(t *testing.T) {
	for name, supported := range map[string]bool{"listed": true, "not listed": false} {
		t.Run(name, func(t *testing.T) {
			fake := oidctest.New(t)
			fake.SetPromptCreateSupported(supported)
			p, err := oidc.New(context.Background(), oidc.Settings{
				Issuer:       fake.Issuer(),
				ClientID:     oidctest.ClientID,
				ClientSecret: oidctest.ClientSecret,
				RedirectURL:  "https://app.example.com/auth/oidc/callback",
			})
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			rawURL, _, _ := p.StartSignUp()

			client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
			resp, err := client.Get(rawURL)
			if err != nil {
				t.Fatalf("following %q: %v", rawURL, err)
			}
			_ = resp.Body.Close()
			want := ""
			if supported {
				want = "create"
			}
			if got := fake.LastPrompt(); got != want {
				t.Errorf("prompt sent = %q, want %q", got, want)
			}
		})
	}
}
