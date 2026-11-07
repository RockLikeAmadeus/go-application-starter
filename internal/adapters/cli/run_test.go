package cli_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/cli"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
)

// A command that fails says why on stderr, so its output, which scripts
// read, stays as it is, and it ends with a failing exit status.
func TestRunPrintsAFailedCommandsErrorOnStderr(t *testing.T) {
	var out, errOut bytes.Buffer
	cmd := cli.NewRootCmd(bootstrap.New, bootstrap.NewBilling, computer(t.TempDir(), nil), &out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"--storage", "disk", "info"})

	status := cli.Run(context.Background(), cmd)

	if status != 1 {
		t.Errorf("exit status = %d, want 1", status)
	}
	want := `Error: The storage setting "disk" is not recognized. Please use "memory", "jsonfile", or "sqlite".` + "\n"
	if got := errOut.String(); got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
	if out.Len() != 0 {
		t.Errorf("stdout holds %q, want nothing", out.String())
	}
}
