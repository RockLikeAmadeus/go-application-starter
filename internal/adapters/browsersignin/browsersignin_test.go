package browsersignin_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/browsersignin"
)

func TestStart_ListensOnALoopbackPortAndHandsTheNativeSignInAddressToTheOpener(t *testing.T) {
	var opened string
	open := func(address string) error {
		opened = address
		return nil
	}

	attempt, err := browsersignin.Start(context.Background(), "https://app.example.com", open)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer attempt.Close()

	address, err := url.Parse(opened)
	if err != nil {
		t.Fatalf("opened address %q: %v", opened, err)
	}
	if address.Scheme != "https" || address.Host != "app.example.com" || address.Path != "/auth/native" {
		t.Errorf("opened %q, want https://app.example.com/auth/native", opened)
	}
	query := address.Query()
	if !regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`).MatchString(query.Get("state")) {
		t.Errorf("state %q is not 16 to 128 base64url characters", query.Get("state"))
	}
	if !regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`).MatchString(query.Get("challenge")) {
		t.Errorf("challenge %q is not 43 base64url characters", query.Get("challenge"))
	}
	port, err := strconv.Atoi(query.Get("port"))
	if err != nil || port < 1 || port > 65535 {
		t.Fatalf("port %q is not a port number", query.Get("port"))
	}
	conn, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", query.Get("port")))
	if err != nil {
		t.Fatalf("nothing listens on 127.0.0.1:%d: %v", port, err)
	}
	conn.Close()
}

func TestWait_SwapsTheCodeOfACallbackWithTheRightStateForATokenAndAnswersTheSuccessPage(t *testing.T) {
	var challenge string
	var swapped struct{ Code, Verifier string }
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/auth/native/token" {
			http.NotFound(w, r)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&swapped); err != nil {
			t.Errorf("token request body: %v", err)
		}
		sum := sha256.Sum256([]byte(swapped.Verifier))
		if base64.RawURLEncoding.EncodeToString(sum[:]) != challenge {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "session-token"})
	}))
	defer server.Close()

	var opened string
	attempt, err := browsersignin.Start(context.Background(), server.URL, func(address string) error {
		opened = address
		return nil
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer attempt.Close()
	query := mustParse(t, opened).Query()
	challenge = query.Get("challenge")

	type result struct {
		token string
		err   error
	}
	done := make(chan result, 1)
	go func() {
		token, err := attempt.Wait(context.Background())
		done <- result{token, err}
	}()

	callback := "http://127.0.0.1:" + query.Get("port") + "/callback?" +
		url.Values{"code": {"the-code"}, "state": {query.Get("state")}}.Encode()
	response, err := (&http.Client{Timeout: 2 * time.Second}).Get(callback)
	if err != nil {
		t.Fatalf("callback: %v", err)
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)

	const want = "Sign-in complete. Please close this window and return to the application that initiated the sign-in."
	if response.StatusCode != http.StatusOK {
		t.Errorf("callback status %d, want 200", response.StatusCode)
	}
	if !strings.Contains(string(body), want) {
		t.Errorf("callback page %q does not contain %q", body, want)
	}
	got := <-done
	if got.err != nil || got.token != "session-token" {
		t.Errorf("Wait gave %q, %v, want session-token", got.token, got.err)
	}
	if swapped.Code != "the-code" {
		t.Errorf("server was sent code %q, want the-code", swapped.Code)
	}
}

func TestWait_RefusesACallbackWithAWrongOrMissingStateSwapsNothingAndKeepsWaiting(t *testing.T) {
	var swaps atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaps.Add(1)
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "session-token"})
	}))
	defer server.Close()

	var opened string
	attempt, err := browsersignin.Start(context.Background(), server.URL, func(address string) error {
		opened = address
		return nil
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer attempt.Close()
	query := mustParse(t, opened).Query()

	done := make(chan string, 1)
	go func() {
		token, _ := attempt.Wait(context.Background())
		done <- token
	}()

	const want = "This sign-in request was refused, because its state value did not match the sign-in that the application started. No session token was issued. If you started a sign-in from the application, please start signing in again from there. Otherwise, you may close this window."
	base := "http://127.0.0.1:" + query.Get("port") + "/callback?"
	client := &http.Client{Timeout: 2 * time.Second}
	for name, values := range map[string]url.Values{
		"wrong":   {"code": {"the-code"}, "state": {"not-the-state"}},
		"missing": {"code": {"the-code"}},
	} {
		response, err := client.Get(base + values.Encode())
		if err != nil {
			t.Fatalf("%s state callback: %v", name, err)
		}
		body, _ := io.ReadAll(response.Body)
		response.Body.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Errorf("%s state: status %d, want 400", name, response.StatusCode)
		}
		if !strings.Contains(string(body), want) {
			t.Errorf("%s state: page %q does not contain %q", name, body, want)
		}
	}
	if n := swaps.Load(); n != 0 {
		t.Errorf("the server was asked to swap %d times, want 0", n)
	}

	response, err := client.Get(base + url.Values{"code": {"the-code"}, "state": {query.Get("state")}}.Encode())
	if err != nil {
		t.Fatalf("right state callback: %v", err)
	}
	response.Body.Close()
	select {
	case token := <-done:
		if token != "session-token" {
			t.Errorf("Wait gave %q, want session-token", token)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Wait did not go on to take the callback with the right state")
	}
}

func TestWait_ReturnsTheContextsErrorWhenTheContextEndsBeforeAnyCallback(t *testing.T) {
	attempt, err := browsersignin.Start(context.Background(), "https://app.example.com", func(string) error { return nil })
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer attempt.Close()

	ctx, cancel := context.WithCancel(context.Background())
	errs := make(chan error, 1)
	go func() {
		_, err := attempt.Wait(ctx)
		errs <- err
	}()
	cancel()

	select {
	case err := <-errs:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Wait returned %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Wait did not return after its context ended")
	}
}

func TestWait_AnswersTheFailurePageAndEndsWithErrCodeRefusedWhenTheServerRefusesTheSwap(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusNotFound)
	}))
	defer server.Close()

	var opened string
	attempt, err := browsersignin.Start(context.Background(), server.URL, func(address string) error {
		opened = address
		return nil
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer attempt.Close()
	query := mustParse(t, opened).Query()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := attempt.Wait(ctx)
		done <- err
	}()

	callback := "http://127.0.0.1:" + query.Get("port") + "/callback?" +
		url.Values{"code": {"the-code"}, "state": {query.Get("state")}}.Encode()
	response, err := (&http.Client{Timeout: 2 * time.Second}).Get(callback)
	if err != nil {
		t.Fatalf("callback: %v", err)
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)

	const want = "Sign-in failed. The server refused to exchange the sign-in code for a session token. The code may have expired (codes last 5 minutes), may already have been used, or may not match the application that requested it. Please start signing in again from the application that initiated the sign-in."
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("callback status %d, want 400", response.StatusCode)
	}
	if !strings.Contains(string(body), want) {
		t.Errorf("callback page %q does not contain %q", body, want)
	}
	if err := <-done; !errors.Is(err, browsersignin.ErrCodeRefused) {
		t.Errorf("Wait ended with %v, want ErrCodeRefused", err)
	}
}

func mustParse(t *testing.T, address string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(address)
	if err != nil {
		t.Fatalf("address %q: %v", address, err)
	}
	return parsed
}
