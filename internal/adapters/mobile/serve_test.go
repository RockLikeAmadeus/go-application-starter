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
