package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
)

func TestSignInGoesToTheDevSignInWhenItIsOn(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{SignIn: &signInRecorder{}, IsDevSignInOn: true})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/auth/sign-in", nil))

	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/auth/dev" {
		t.Errorf("status = %d, Location = %q; want %d to %q", rec.Code, rec.Header().Get("Location"), http.StatusFound, "/auth/dev")
	}
}
