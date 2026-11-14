package tokenstore

import (
	"errors"

	"github.com/zalando/go-keyring"
)

// keyringService is the name the credential store files the tokens under.
const keyringService = "go-application-starter"

// Keyring keeps tokens in the operating system's credential store: the
// Keychain on macOS, the Credential Manager on Windows, and the Secret
// Service (over D-Bus) elsewhere.
type Keyring struct{}

// NewKeyring returns a Keyring.
func NewKeyring() *Keyring { return &Keyring{} }

// Save keeps the token for the server.
func (k *Keyring) Save(serverURL, token string) error {
	return keyring.Set(keyringService, normalise(serverURL), token)
}

// Read returns the token kept for the server, or ErrNoToken.
func (k *Keyring) Read(serverURL string) (string, error) {
	token, err := keyring.Get(keyringService, normalise(serverURL))
	if errors.Is(err, keyring.ErrNotFound) {
		return "", ErrNoToken
	}
	return token, err
}

// Delete forgets the token kept for the server. Forgetting an unknown one is
// not an error.
func (k *Keyring) Delete(serverURL string) error {
	err := keyring.Delete(keyringService, normalise(serverURL))
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}
