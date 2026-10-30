package mobile_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/mobile"
)

func TestServeSendsTheRequestAndReturnsTheReply(t *testing.T) {
	var got *http.Request
	var gotBody []byte
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"name":"Widget"}`)
	})

	reply := mobile.Serve(h, http.MethodPost, "/api/v1/items", `{"name":"Widget"}`)

	if got == nil {
		t.Fatal("the handler was not called")
	}
	if got.Method != http.MethodPost || got.URL.Path != "/api/v1/items" {
		t.Errorf("handler got %s %s, want POST /api/v1/items", got.Method, got.URL.Path)
	}
	if ct := got.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("handler got Content-Type %q, want application/json", ct)
	}
	if string(gotBody) != `{"name":"Widget"}` {
		t.Errorf("handler got body %q, want the request's JSON", gotBody)
	}
	want := mobile.Reply{Status: http.StatusCreated, ContentType: "application/json", Body: `{"name":"Widget"}`}
	if reply != want {
		t.Errorf("reply = %+v, want %+v", reply, want)
	}
}

func TestServePassesTheQueryAndReturnsAnErrorReplyUnchanged(t *testing.T) {
	const problem = `{"title":"Validation failed","status":422}`
	var gotQuery string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = io.WriteString(w, problem)
	})

	reply := mobile.Serve(h, http.MethodGet, "/api/v1/items?limit=2&after=x", "")

	if gotQuery != "limit=2&after=x" {
		t.Errorf("handler got query %q, want limit=2&after=x", gotQuery)
	}
	want := mobile.Reply{Status: http.StatusUnprocessableEntity, ContentType: "application/problem+json", Body: problem}
	if reply != want {
		t.Errorf("reply = %+v, want %+v", reply, want)
	}
}

// A request Serve can't build must come back as a reply: a panic would crash
// the bridge call.
func TestServeAnswersAMalformedPathWithBadRequest(t *testing.T) {
	h := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("the handler was called for a malformed path")
	})

	reply := mobile.Serve(h, http.MethodGet, "/api/v1/items/%zz", "")

	if reply.Status != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", reply.Status, http.StatusBadRequest)
	}
}
