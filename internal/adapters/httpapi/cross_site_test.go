package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
)

func TestASignOutFromAnotherSiteIsRefused(t *testing.T) {
	signIn := &signInRecorder{}
	handler := httpapi.NewHandler(httpapi.Options{SignIn: signIn})
	req := httptest.NewRequest(http.MethodPost, "/auth/sign-out", nil)
	req.AddCookie(&http.Cookie{Name: httpapi.SessionCookie, Value: "token-ada"})
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden || rec.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("status = %d, Content-Type = %q; want %d, %q", rec.Code, rec.Header().Get("Content-Type"), http.StatusForbidden, "application/problem+json")
	}
	var got struct {
		Type   string `json:"type"`
		Detail string `json:"detail"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	wantDetail := "The server refused this request because it did not come from the application's own pages. Please try again from the application."
	if got.Type != httpapi.ProblemCrossSite || got.Detail != wantDetail {
		t.Errorf("type = %q, detail = %q; want %q, %q", got.Type, got.Detail, httpapi.ProblemCrossSite, wantDetail)
	}
	if len(signIn.signedOut) != 0 {
		t.Errorf("signed out %q, want none", signIn.signedOut)
	}
}
