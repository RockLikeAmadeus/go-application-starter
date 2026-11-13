// Package browsersignin signs a client application in through the browser,
// following RFC 8252: it listens on a loopback port, sends the browser to the
// server's /auth/native with a state and a PKCE challenge, and takes the
// session token the browser's return brings. It has no Cobra, Bubble Tea, or
// Wails import, so any client can use it.
package browsersignin

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// Attempt is a sign-in in progress. Close it when done, to free the port.
type Attempt struct {
	listener net.Listener
	state    string
	verifier string
}

// Start listens on a free loopback port and hands the address of the
// server's native sign-in page to open.
func Start(ctx context.Context, serverURL string, open func(address string) error) (*Attempt, error) {
	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	attempt := &Attempt{listener: listener}
	if attempt.state, err = randomToken(); err != nil {
		listener.Close()
		return nil, err
	}
	if attempt.verifier, err = randomToken(); err != nil {
		listener.Close()
		return nil, err
	}
	challenge := sha256.Sum256([]byte(attempt.verifier))
	address := strings.TrimRight(serverURL, "/") + "/auth/native?" + url.Values{
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
