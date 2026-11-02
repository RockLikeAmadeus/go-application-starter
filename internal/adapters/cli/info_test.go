package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/cli"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
)

// runCLI executes the CLI with args, as cmd/starter wires it.
func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := cli.NewRootCmd(bootstrap.New, &out)
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(context.Background())
	return out.String(), err
}

func TestInfoNamesTheDataFolder(t *testing.T) {
	dir := t.TempDir()

	out, err := runCLI(t, "--data-dir", dir, "--storage", "jsonfile", "info")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if !strings.Contains(out, dir) {
		t.Errorf("info output doesn't name the data folder %s:\n%s", dir, out)
	}
}
