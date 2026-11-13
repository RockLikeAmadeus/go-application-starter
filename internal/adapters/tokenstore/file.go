// Package tokenstore keeps the session tokens a client application (the
// command line, the terminal app) was given by servers, one per server.
package tokenstore

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ErrNoToken is what Read returns when no token is kept for the server.
var ErrNoToken = errors.New("no token kept for this server")

const fileName = "tokens.json"

// File keeps tokens in a file in a folder, readable by its owner only.
type File struct {
	folder string
	mu     sync.Mutex
}

// NewFile returns a File keeping its tokens in the folder.
func NewFile(folder string) *File { return &File{folder: folder} }

// Save keeps the token for the server.
func (f *File) Save(serverURL, token string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	tokens, err := f.load()
	if err != nil {
		return err
	}
	tokens[normalise(serverURL)] = token
	data, err := json.Marshal(tokens)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(f.folder, 0o700); err != nil {
		return err
	}
	// Write a private temporary file and rename it over the real one, so a
	// file that already exists with wider permissions is replaced, and a
	// crash never leaves a half-written one.
	temporary, err := os.CreateTemp(f.folder, fileName+".*")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporary.Name(), filepath.Join(f.folder, fileName))
}

// Read returns the token kept for the server, or ErrNoToken.
func (f *File) Read(serverURL string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	tokens, err := f.load()
	if err != nil {
		return "", err
	}
	token, ok := tokens[normalise(serverURL)]
	if !ok {
		return "", ErrNoToken
	}
	return token, nil
}

func (f *File) load() (map[string]string, error) {
	tokens := map[string]string{}
	data, err := os.ReadFile(filepath.Join(f.folder, fileName))
	if errors.Is(err, os.ErrNotExist) {
		return tokens, nil
	}
	if err != nil {
		return nil, err
	}
	return tokens, json.Unmarshal(data, &tokens)
}

// normalise makes the addresses people type for one server the same key.
// The scheme and host are not case-sensitive, but the path is.
func normalise(serverURL string) string {
	address, err := url.Parse(serverURL)
	if err != nil {
		return strings.TrimRight(serverURL, "/")
	}
	address.Scheme = strings.ToLower(address.Scheme)
	address.Host = strings.ToLower(address.Host)
	address.Path = strings.TrimRight(address.Path, "/")
	address.RawPath = ""
	return address.String()
}
