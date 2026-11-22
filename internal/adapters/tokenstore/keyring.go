package tokenstore

import (
	"errors"

	"github.com/zalando/go-keyring"
)

// keyringService is the name the credential store files the tokens under.
const (
	keyringService    = "go-application-starter"
	appKeyringService = "go-application-starter-app"
)

// Keyring keeps tokens in the operating system's credential store: the
// Keychain on macOS, the Credential Manager on Windows, and the Secret
// Service (over D-Bus) elsewhere.
type Keyring struct{ service string }

// NewKeyring returns a Keyring.
func NewKeyring() *Keyring { return &Keyring{service: keyringService} }

// Save keeps the token for the server.
func (k *Keyring) Save(serverURL, token string) error {
	return keyring.Set(k.service, normalise(serverURL), token)
}

// Read returns the token kept for the server, or ErrNoToken.
func (k *Keyring) Read(serverURL string) (string, error) {
	token, err := keyring.Get(k.service, normalise(serverURL))
	if errors.Is(err, keyring.ErrNotFound) {
		return "", ErrNoToken
	}
	return token, err
}

// Delete forgets the token kept for the server. Forgetting an unknown one is
// not an error.
func (k *Keyring) Delete(serverURL string) error {
	err := keyring.Delete(k.service, normalise(serverURL))
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

// NewAppKeyring returns a Keyring for the desktop and Android apps, under a
// service name of their own.
func NewAppKeyring() *Keyring { return &Keyring{service: appKeyringService} }
