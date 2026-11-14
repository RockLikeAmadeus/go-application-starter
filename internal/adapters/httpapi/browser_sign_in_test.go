package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
)

func TestAServerWithNoClientSignInAnswersBothBrowserSignInOperationsNotFound(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{SignIn: &signInRecorder{}, IsDevSignInOn: true})
	for _, method := range []string{http.MethodPost, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/browser-sign-in", strings.NewReader(`{"isSignUp":false}`))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", rec.Code)
			}
			var body struct{ Type string }
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil || body.Type != httpapi.ProblemNotFound {
				t.Errorf("body type = %q (err %v), want %q", body.Type, err, httpapi.ProblemNotFound)
			}
		})
	}
}
