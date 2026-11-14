package tokenstore_test

import (
	"errors"
	"testing"

	"github.com/zalando/go-keyring"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/tokenstore"
)

// store is what every token store offers, whatever it keeps tokens in.
type store interface {
	Save(serverURL, token string) error
	Read(serverURL string) (string, error)
	Delete(serverURL string) error
}

// runContract checks a store against what every token store promises. The
// factory returns a fresh, empty store each time it is called, and one that
// shares its storage with the earlier ones from the same call of runContract.
func runContract(t *testing.T, newStore func(t *testing.T) func() store) {
	t.Run("keeps each server's token apart under its normalised address", func(t *testing.T) {
		open := newStore(t)
		s := open()
		if err := s.Save("https://App.Example.com/", "token-a"); err != nil {
			t.Fatalf("Save a: %v", err)
		}
		if err := s.Save("http://localhost:8080", "token-b"); err != nil {
			t.Fatalf("Save b: %v", err)
		}

		fresh := open()
		if got, err := fresh.Read("https://app.example.com"); err != nil || got != "token-a" {
			t.Errorf("Read a = %q, %v; want token-a", got, err)
		}
		if got, err := fresh.Read("http://localhost:8080/"); err != nil || got != "token-b" {
			t.Errorf("Read b = %q, %v; want token-b", got, err)
		}
		if _, err := fresh.Read("https://other.example.com"); !errors.Is(err, tokenstore.ErrNoToken) {
			t.Errorf("Read of an unknown server = %v; want ErrNoToken", err)
		}
	})

	t.Run("deletes one server's token, leaving the others", func(t *testing.T) {
		open := newStore(t)
		s := open()
		if err := s.Save("https://a.example.com", "token-a"); err != nil {
			t.Fatal(err)
		}
		if err := s.Save("https://b.example.com", "token-b"); err != nil {
			t.Fatal(err)
		}

		if err := s.Delete("https://A.example.com/"); err != nil {
			t.Fatalf("Delete a: %v", err)
		}
		if err := s.Delete("https://unknown.example.com"); err != nil {
			t.Errorf("Delete of an unknown server = %v; want nil", err)
		}

		fresh := open()
		if _, err := fresh.Read("https://a.example.com"); !errors.Is(err, tokenstore.ErrNoToken) {
			t.Errorf("Read a after Delete = %v; want ErrNoToken", err)
		}
		if got, err := fresh.Read("https://b.example.com"); err != nil || got != "token-b" {
			t.Errorf("Read b = %q, %v; want token-b", got, err)
		}
	})
}

func TestFileStoreKeepsTheContract(t *testing.T) {
	runContract(t, func(t *testing.T) func() store {
		folder := t.TempDir()
		return func() store { return tokenstore.NewFile(folder) }
	})
}

func TestKeyringStoreKeepsTheContract(t *testing.T) {
	runContract(t, func(t *testing.T) func() store {
		keyring.MockInit()
		return func() store { return tokenstore.NewKeyring() }
	})
}
