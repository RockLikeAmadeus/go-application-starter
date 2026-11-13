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
