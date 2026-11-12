// Package oidctest is a fake sign-in provider for tests: an httptest server
// that speaks the parts of OpenID Connect the oidc adapter uses. Only tests
// import it.
package oidctest

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/go-jose/go-jose/v4"
)

// Provider is a running fake provider with its own signing key.
type Provider struct {
	server    *httptest.Server
	key       *rsa.PrivateKey
	discovery atomic.Int64
}

// New starts a provider and stops it when the test ends.
func New(t testing.TB) *Provider {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating the fake provider's key: %v", err)
	}
	p := &Provider{key: key}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /.well-known/openid-configuration", p.serveDiscovery)
	mux.HandleFunc("GET /keys", p.serveKeys)
	p.server = httptest.NewServer(mux)
	t.Cleanup(p.server.Close)
	return p
}

// Issuer is the provider's issuer URL.
func (p *Provider) Issuer() string { return p.server.URL }

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

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
