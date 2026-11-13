package tokenstore_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/tokenstore"
)

func TestFile_KeepsEachServersTokenApartInAPrivateFileUnderItsNormalisedAddress(t *testing.T) {
	folder := t.TempDir()
	store := tokenstore.NewFile(folder)

	if err := store.Save("https://App.Example.com/", "token-a"); err != nil {
		t.Fatalf("Save a: %v", err)
	}
	if err := store.Save("http://localhost:8080", "token-b"); err != nil {
		t.Fatalf("Save b: %v", err)
	}

	// Read through a fresh store, so the token came from the file.
	fresh := tokenstore.NewFile(folder)
	got, err := fresh.Read("https://app.example.com")
	if err != nil || got != "token-a" {
		t.Errorf("Read a = %q, %v; want token-a", got, err)
	}
	got, err = fresh.Read("http://localhost:8080/")
	if err != nil || got != "token-b" {
		t.Errorf("Read b = %q, %v; want token-b", got, err)
	}
	if _, err := fresh.Read("https://other.example.com"); !errors.Is(err, tokenstore.ErrNoToken) {
		t.Errorf("Read of an unknown server = %v; want ErrNoToken", err)
	}

	entries, err := os.ReadDir(folder)
	if err != nil || len(entries) != 1 {
		t.Fatalf("settings folder holds %d entries (%v); want one file", len(entries), err)
	}
	info, err := os.Stat(filepath.Join(folder, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("file mode = %o; want 600", mode)
	}
}

func TestFile_KeepsServersWhoseAddressesDifferOnlyInTheCaseOfTheirPathsApart(t *testing.T) {
	store := tokenstore.NewFile(t.TempDir())
	if err := store.Save("https://Example.com/App", "token-upper"); err != nil {
		t.Fatal(err)
	}
	if err := store.Save("https://example.com/app/", "token-lower"); err != nil {
		t.Fatal(err)
	}

	for address, want := range map[string]string{
		"https://example.com/App":  "token-upper",
		"https://EXAMPLE.com/app/": "token-lower",
	} {
		if got, err := store.Read(address); err != nil || got != want {
			t.Errorf("Read(%q) = %q, %v; want %q", address, got, err, want)
		}
	}
}

func TestFile_MakesAnExistingWiderFilePrivateWhenItSavesATokenInIt(t *testing.T) {
	folder := t.TempDir()
	path := filepath.Join(folder, "tokens.json")
	if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := tokenstore.NewFile(folder).Save("https://app.example.com", "token"); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("file mode = %o; want 600", mode)
	}
}
