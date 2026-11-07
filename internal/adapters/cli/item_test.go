package cli_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/cli"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// run executes the CLI against a fresh data directory shared across calls.
func run(t *testing.T, dataDir string, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := cli.NewRootCmd(bootstrap.New, bootstrap.NewBilling, computer(t.TempDir(), nil), &out)
	cmd.SetArgs(append([]string{"--data-dir", dataDir, "--storage", string(config.StorageJSONFile)}, args...))
	err := cmd.ExecuteContext(context.Background())
	return out.String(), err
}

func TestAddThenList(t *testing.T) {
	dir := t.TempDir()

	if _, err := run(t, dir, "item", "add", "Widget"); err != nil {
		t.Fatalf("add: %v", err)
	}
	out, err := run(t, dir, "item", "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(out, "Widget") {
		t.Errorf("list output missing item:\n%s", out)
	}
}

func TestAddRejectsBlankName(t *testing.T) {
	_, err := run(t, t.TempDir(), "item", "add", " ")
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
}

func TestExportPrintsTheItemsAsCSV(t *testing.T) {
	dir := t.TempDir()
	if _, err := run(t, dir, "item", "add", "Widget, large"); err != nil {
		t.Fatalf("add: %v", err)
	}

	out, err := run(t, dir, "item", "export")
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if want := "name,createdAt\n\"Widget, large\","; !strings.HasPrefix(out, want) {
		t.Errorf("export output = %q, want it to start with %q", out, want)
	}
}

func TestDeleteConfirmsInAFullSentence(t *testing.T) {
	dir := t.TempDir()
	added, err := run(t, dir, "item", "add", "Widget")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	_, row, _ := strings.Cut(added, "\n")
	id, _, _ := strings.Cut(row, " ")

	out, err := run(t, dir, "item", "delete", id)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if want := "Item " + id + " has been deleted.\n"; out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}
