package httpclient_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient"
)

// The token source is asked on each request, so a token that appears,
// changes, or is forgotten takes effect without building a new client.
func TestWithTokenSource_SendsTheTokenItReturnsOnEachRequestAndNoHeaderForNone(t *testing.T) {
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"isSignInOn":true}`))
	}))
	defer srv.Close()

	tokens := []string{"", "first", "second", ""}
	next := 0
	client, err := httpclient.NewSessionService(srv.URL, httpclient.WithTokenSource(func(context.Context) (string, error) {
		token := tokens[next]
		next++
		return token, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	for range tokens {
		if _, err := client.CurrentSession(context.Background()); err != nil {
			t.Fatalf("CurrentSession: %v", err)
		}
	}

	want := []string{"", "Bearer first", "Bearer second", ""}
	for i := range want {
		if seen[i] != want[i] {
			t.Errorf("request %d Authorization = %q, want %q", i, seen[i], want[i])
		}
	}
}
