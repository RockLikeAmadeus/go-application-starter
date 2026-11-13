package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
)

const (
	nativeIntro     = "The sign-in request from the command line is not valid."
	nativePortText  = "`port` is missing or is not a port number from 1 to 65535."
	nativeStateText = "`state` is missing or malformed."
	nativeChalText  = "`challenge` is missing or is not a PKCE S256 challenge (43 base64url characters)."
	nativeOutro     = "Please run the login command again. If this keeps happening, check that the command-line application and the server are the same version."

	// The RFC 7636 example challenge: 43 base64url characters.
	validChallenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	validState     = "abcdefghijklmnopqrstuvwxyz012345"
)

func TestABadNativeSignInRequestIsRefusedListingEveryFailingValue(t *testing.T) {
	handler := httpapi.NewHandler(httpapi.Options{SignIn: &signInRecorder{}, IsDevSignInOn: true})
	for name, tc := range map[string]struct {
		query url.Values
		want  []string
	}{
		"a port of zero": {
			url.Values{"port": {"0"}, "state": {validState}, "challenge": {validChallenge}},
			[]string{nativePortText},
		},
		"a port above 65535": {
			url.Values{"port": {"65536"}, "state": {validState}, "challenge": {validChallenge}},
			[]string{nativePortText},
		},
		"a port that is not a number": {
			url.Values{"port": {"http"}, "state": {validState}, "challenge": {validChallenge}},
			[]string{nativePortText},
		},
		"a malformed state": {
			url.Values{"port": {"4000"}, "state": {"not valid!"}, "challenge": {validChallenge}},
			[]string{nativeStateText},
		},
		"a challenge of the wrong length": {
			url.Values{"port": {"4000"}, "state": {validState}, "challenge": {validChallenge + "A"}},
			[]string{nativeChalText},
		},
		"nothing at all": {
			url.Values{},
			[]string{nativePortText, nativeStateText, nativeChalText},
		},
	} {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/auth/native?"+tc.query.Encode(), nil))

			var body struct {
				Detail string `json:"detail"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode %q: %v", rec.Body.String(), err)
			}
			want := strings.Join(append(append([]string{nativeIntro}, tc.want...), nativeOutro), " ")
			if rec.Code != http.StatusBadRequest || body.Detail != want {
				t.Errorf("status = %d, detail = %q; want %d, %q", rec.Code, body.Detail, http.StatusBadRequest, want)
			}
		})
	}
}
