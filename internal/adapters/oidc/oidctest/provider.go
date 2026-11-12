// Package oidctest is a fake sign-in provider for tests: an httptest server
// that speaks the parts of OpenID Connect the oidc adapter uses. Only tests
// import it.
package oidctest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

// The client the fake provider knows.
const (
	ClientID     = "the-client"
	ClientSecret = "the-secret"
)

// Person is who signs in at the fake provider.
type Person struct {
	Subject           string
	Name              string
	PreferredUsername string
	Email             string
}

// Fault is a way the fake provider can misbehave when it issues an ID token.
type Fault int

// The faults the fake provider can be told to commit.
const (
	NoFault Fault = iota
	SignedWithAnotherKey
	MeantForAnotherClient
	Expired
	FromAnotherIssuer
	WithAnotherNonce
)

// Provider is a running fake provider with its own signing key.
type Provider struct {
	server    *httptest.Server
	key       *rsa.PrivateKey
	discovery atomic.Int64

	mu     sync.Mutex
	person Person
	fault  Fault
	grants map[string]grant
}

// grant is what the provider remembers about an issued code.
type grant struct {
	nonce     string
	challenge string
}

// New starts a provider and stops it when the test ends.
func New(t testing.TB) *Provider {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating the fake provider's key: %v", err)
	}
	p := &Provider{key: key, grants: map[string]grant{}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /.well-known/openid-configuration", p.serveDiscovery)
	mux.HandleFunc("GET /keys", p.serveKeys)
	mux.HandleFunc("GET /authorize", p.serveAuthorize)
	mux.HandleFunc("POST /token", p.serveToken)
	p.server = httptest.NewServer(mux)
	t.Cleanup(p.server.Close)
	return p
}

// Issuer is the provider's issuer URL.
func (p *Provider) Issuer() string { return p.server.URL }

// Close stops the provider before the test ends, to play a provider that is down.
func (p *Provider) Close() { p.server.Close() }

// SetPerson chooses who signs in next.
func (p *Provider) SetPerson(person Person) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.person = person
}

// SetFault makes the provider misbehave in every ID token it issues from now on.
func (p *Provider) SetFault(fault Fault) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.fault = fault
}

// DiscoveryRequests counts the discovery documents the provider has served.
func (p *Provider) DiscoveryRequests() int { return int(p.discovery.Load()) }

func (p *Provider) serveDiscovery(w http.ResponseWriter, _ *http.Request) {
	p.discovery.Add(1)
	writeJSON(w, map[string]any{
		"issuer":                                p.Issuer(),
		"authorization_endpoint":                p.Issuer() + "/authorize",
		"token_endpoint":                        p.Issuer() + "/token",
		"jwks_uri":                              p.Issuer() + "/keys",
		"response_types_supported":              []string{"code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
	})
}

func (p *Provider) serveKeys(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{
		Key: &p.key.PublicKey, KeyID: "fake-key", Algorithm: "RS256", Use: "sig",
	}}})
}

// serveAuthorize signs the person in at once and sends them back with a code.
func (p *Provider) serveAuthorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	code := rand.Text()
	p.mu.Lock()
	p.grants[code] = grant{nonce: q.Get("nonce"), challenge: q.Get("code_challenge")}
	p.mu.Unlock()
	back, err := url.Parse(q.Get("redirect_uri"))
	if err != nil {
		http.Error(w, "bad redirect_uri", http.StatusBadRequest)
		return
	}
	v := back.Query()
	v.Set("code", code)
	v.Set("state", q.Get("state"))
	back.RawQuery = v.Encode()
	http.Redirect(w, r, back.String(), http.StatusFound)
}

// serveToken swaps a code for an ID token, if the client and PKCE verifier are
// right and the code has not been used before.
func (p *Provider) serveToken(w http.ResponseWriter, r *http.Request) {
	id, secret, ok := r.BasicAuth()
	if !ok {
		id, secret = r.PostFormValue("client_id"), r.PostFormValue("client_secret")
	}
	if id != ClientID || secret != ClientSecret {
		writeError(w, "invalid_client")
		return
	}
	p.mu.Lock()
	code := r.PostFormValue("code")
	g, found := p.grants[code]
	delete(p.grants, code) // a code works once, even when the verifier is wrong
	person, fault := p.person, p.fault
	p.mu.Unlock()
	sum := sha256.Sum256([]byte(r.PostFormValue("code_verifier")))
	if !found || base64.RawURLEncoding.EncodeToString(sum[:]) != g.challenge {
		writeError(w, "invalid_grant")
		return
	}
	claims := map[string]any{
		"iss":                p.Issuer(),
		"sub":                person.Subject,
		"aud":                ClientID,
		"exp":                time.Now().Add(time.Hour).Unix(),
		"iat":                time.Now().Unix(),
		"nonce":              g.nonce,
		"name":               person.Name,
		"preferred_username": person.PreferredUsername,
		"email":              person.Email,
	}
	signingKey := p.key
	switch fault {
	case SignedWithAnotherKey:
		signingKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	case MeantForAnotherClient:
		claims["aud"] = "another-client"
	case Expired:
		claims["exp"] = time.Now().Add(-time.Hour).Unix()
	case FromAnotherIssuer:
		claims["iss"] = "https://another-issuer.example.com"
	case WithAnotherNonce:
		claims["nonce"] = "another-nonce"
	}
	idToken, err := signIDToken(signingKey, claims)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{
		"access_token": rand.Text(),
		"token_type":   "Bearer",
		"id_token":     idToken,
	})
}

func signIDToken(key *rsa.PrivateKey, claims map[string]any) (string, error) {
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: key},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", "fake-key"),
	)
	if err != nil {
		return "", err
	}
	return jwt.Signed(signer).Claims(claims).Serialize()
}

func writeError(w http.ResponseWriter, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
