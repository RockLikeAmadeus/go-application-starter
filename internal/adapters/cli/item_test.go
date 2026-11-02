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
	cmd := cli.NewRootCmd(bootstrap.New, noLocalDataDir(t), &out)
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
