package tokenstore_test

import (
	"errors"
	"testing"

	"github.com/zalando/go-keyring"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/tokenstore"
)

func TestSystem_SavesInTheKeyringAndFallsBackToTheFileWhenTheKeyringIsUnavailable(t *testing.T) {
	const server = "https://app.example.com"

	t.Run("keyring available", func(t *testing.T) {
		keyring.MockInit()
		folder := t.TempDir()
		kept, err := tokenstore.NewSystem(folder).Save(server, "token-a")
		if err != nil {
			t.Fatalf("Save: %v", err)
		}
		if kept != tokenstore.InKeyring {
			t.Errorf("Save kept the token in %v; want the keyring", kept)
		}
		if got, err := tokenstore.NewKeyring().Read(server); err != nil || got != "token-a" {
			t.Errorf("keyring Read = %q, %v; want token-a", got, err)
		}
		if _, err := tokenstore.NewFile(folder).Read(server); !errors.Is(err, tokenstore.ErrNoToken) {
			t.Errorf("file Read = %v; want ErrNoToken", err)
		}
	})

	t.Run("keyring unavailable", func(t *testing.T) {
		keyring.MockInitWithError(errors.New("no secret service"))
		folder := t.TempDir()
		kept, err := tokenstore.NewSystem(folder).Save(server, "token-b")
		if err != nil {
			t.Fatalf("Save: %v", err)
		}
		if kept != tokenstore.InFile {
			t.Errorf("Save kept the token in %v; want the file", kept)
		}
		if got, err := tokenstore.NewFile(folder).Read(server); err != nil || got != "token-b" {
			t.Errorf("file Read = %q, %v; want token-b", got, err)
		}
	})
}

func TestSystem_ReadsATokenFromTheKeyringElseFromTheFileAndReportsNoTokenWhenNeitherHasOne(t *testing.T) {
	const server = "https://app.example.com"

	t.Run("in the keyring", func(t *testing.T) {
		keyring.MockInit()
		folder := t.TempDir()
		if err := tokenstore.NewKeyring().Save(server, "token-a"); err != nil {
			t.Fatalf("keyring Save: %v", err)
		}
		if got, err := tokenstore.NewSystem(folder).Read(server); err != nil || got != "token-a" {
			t.Errorf("Read = %q, %v; want token-a", got, err)
		}
	})

	t.Run("only in the file", func(t *testing.T) {
		keyring.MockInit()
		folder := t.TempDir()
		if err := tokenstore.NewFile(folder).Save(server, "token-b"); err != nil {
			t.Fatalf("file Save: %v", err)
		}
		if got, err := tokenstore.NewSystem(folder).Read(server); err != nil || got != "token-b" {
			t.Errorf("Read = %q, %v; want token-b", got, err)
		}
	})

	t.Run("keyring unavailable", func(t *testing.T) {
		keyring.MockInitWithError(errors.New("no secret service"))
		folder := t.TempDir()
		system := tokenstore.NewSystem(folder)
		if _, err := system.Save(server, "token-c"); err != nil {
			t.Fatalf("Save: %v", err)
		}
		if got, err := system.Read(server); err != nil || got != "token-c" {
			t.Errorf("Read = %q, %v; want token-c", got, err)
		}
	})

	t.Run("in neither", func(t *testing.T) {
		keyring.MockInit()
		if _, err := tokenstore.NewSystem(t.TempDir()).Read(server); !errors.Is(err, tokenstore.ErrNoToken) {
			t.Errorf("Read = %v; want ErrNoToken", err)
		}
	})
}
