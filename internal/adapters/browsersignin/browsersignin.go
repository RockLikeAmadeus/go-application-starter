// Package browsersignin signs a client application in through the browser,
// following RFC 8252: it listens on a loopback port, sends the browser to the
// server's /auth/native with a state and a PKCE challenge, and takes the
// session token the browser's return brings. It has no Cobra, Bubble Tea, or
// Wails import, so any client can use it.
package browsersignin

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ErrCodeRefused is what Wait returns when the server refuses to swap the
// browser's code for a session token.
var ErrCodeRefused = errors.New("the server refused the sign-in code")

// Attempt is a sign-in in progress. Close it when done, to free the port.
type Attempt struct {
	listener  net.Listener
	state     string
	verifier  string
	serverURL string
}

// Start listens on a free loopback port and hands the address of the
// server's native sign-in page to open.
func Start(ctx context.Context, serverURL string, open func(address string) error) (*Attempt, error) {
	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	attempt := &Attempt{listener: listener, serverURL: strings.TrimRight(serverURL, "/")}
	if attempt.state, err = randomToken(); err != nil {
		listener.Close()
		return nil, err
	}
	if attempt.verifier, err = randomToken(); err != nil {
		listener.Close()
		return nil, err
	}
	challenge := sha256.Sum256([]byte(attempt.verifier))
	address := attempt.serverURL + "/auth/native?" + url.Values{
		"port":      {strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)},
		"state":     {attempt.state},
		"challenge": {base64.RawURLEncoding.EncodeToString(challenge[:])},
	}.Encode()
	if err := open(address); err != nil {
		listener.Close()
		return nil, err
	}
	return attempt, nil
}

// Close stops listening.
func (a *Attempt) Close() {
	a.listener.Close()
}

// randomToken is 32 random bytes in unpadded base64url: 43 characters, which
// meets the server's rules for a state and for a PKCE verifier (RFC 7636).
func randomToken() (string, error) {
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(random), nil
}

// Wait blocks until the browser returns to the loopback address, and gives
// the session token the server swapped for its code.
func (a *Attempt) Wait(ctx context.Context) (string, error) {
	tokens := make(chan string, 1)
	failures := make(chan error, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /callback", func(w http.ResponseWriter, r *http.Request) {
		// Compared in constant time, so the wait leaks nothing about the state.
		if subtle.ConstantTimeCompare([]byte(r.URL.Query().Get("state")), []byte(a.state)) != 1 {
			writePage(w, http.StatusBadRequest, "Sign-in refused", wrongStatePage)
			return
		}
		token, err := a.swap(r.Context(), r.URL.Query().Get("code"))
		if err != nil {
			writePage(w, http.StatusBadRequest, "Sign-in failed", swapRefusedPage)
			select {
			case failures <- err:
			default:
			}
			return
		}
		writePage(w, http.StatusOK, "Sign-in complete", signedInPage)
		tokens <- token
	})
	// Serve ends when Close closes the listener; its error says only that.
	go func() { _ = http.Serve(a.listener, mux) }()
	select {
	case token := <-tokens:
		return token, nil
	case err := <-failures:
		return "", err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func writePage(w http.ResponseWriter, status int, title, text string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprint(w, "<!doctype html><html><head><meta charset=\"utf-8\"><title>"+title+"</title></head>"+
		"<body><p>"+text+"</p></body></html>")
}

const wrongStatePage = "This sign-in request was refused, because its state value did not match the sign-in that the application started. No session token was issued. If you started a sign-in from the application, please start signing in again from there. Otherwise, you may close this window."

const swapRefusedPage = "Sign-in failed. The server refused to exchange the sign-in code for a session token. The code may have expired (codes last 5 minutes), may already have been used, or may not match the application that requested it. Please start signing in again from the application that initiated the sign-in."

const signedInPage = "Sign-in complete. Please close this window and return to the application that initiated the sign-in."

// swap exchanges the one-time code and the verifier for a session token.
func (a *Attempt) swap(ctx context.Context, code string) (string, error) {
	body, err := json.Marshal(map[string]string{"code": code, "verifier": a.verifier})
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, a.serverURL+"/auth/native/token", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", ErrCodeRefused
	}
	var token struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(response.Body).Decode(&token); err != nil {
		return "", err
	}
	return token.Token, nil
}
