package cli_test

import (
	"bytes"
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/cli"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// A command that fails says why on stderr, so its output, which scripts
// read, stays as it is, and it ends with a failing exit status.
func TestRunPrintsAFailedCommandsErrorOnStderr(t *testing.T) {
	var out, errOut bytes.Buffer
	cmd := cli.NewRootCmd(bootstrap.New, bootstrap.NewBilling, bootstrap.NewClientSignIn, computer(t.TempDir(), nil), &out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"--storage", "disk", "info"})

	status := cli.Run(context.Background(), cmd)

	if status != 1 {
		t.Errorf("exit status = %d, want 1", status)
	}
	want := `Error: The storage setting "disk" is not recognized. Please use "memory", "jsonfile", "sqlite", or "postgres".` + "\n"
	if got := errOut.String(); got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
	if out.Len() != 0 {
		t.Errorf("stdout holds %q, want nothing", out.String())
	}
}

// An error from starting the application reads as it is given, straight after
// the prefix. A bad plans file is one that reaches the CLI past its settings.
func TestAnErrorStartingTheApplicationFollowsThePrefixDirectly(t *testing.T) {
	plansFile := filepath.Join(t.TempDir(), "plans.jsonc")
	content := `{"plans": {"local": {"limits": {"items": {"max": -1, "whenOver": "block-new"}}}}}`
	if err := os.WriteFile(plansFile, []byte(content), 0o600); err != nil {
		t.Fatalf("write the plans file: %v", err)
	}
	env := computer(t.TempDir(), map[string]string{"STARTER_PLANS_FILE": plansFile})
	var out, errOut bytes.Buffer
	cmd := cli.NewRootCmd(bootstrap.New, bootstrap.NewBilling, bootstrap.NewClientSignIn, env, &out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"--storage", "memory", "info"})

	cli.Run(context.Background(), cmd)

	want := `Error: plans file: plan "local", limit "items": The max value cannot be negative.` + "\n"
	if got := errOut.String(); got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

// refusingEntitlements is a ports.EntitlementsService that refuses every
// request for want of a session, as a server does for a signed-out caller.
type refusingEntitlements struct{}

func (refusingEntitlements) CurrentEntitlements(context.Context) (domain.Entitlements, error) {
	return domain.Entitlements{}, domain.ErrNoWorkspace
}

// A request that a server refuses for want of a session tells the user to run
// starter login, naming the server, for any command.
func TestARequestRefusedForWantOfASessionTellsTheUserToRunLogin(t *testing.T) {
	srv := httptest.NewServer(httpapi.NewHandler(httpapi.Options{Entitlements: refusingEntitlements{}}))
	t.Cleanup(srv.Close)
	var out, errOut bytes.Buffer
	cmd := cli.NewRootCmd(bootstrap.New, bootstrap.NewBilling, bootstrap.NewClientSignIn, computer(t.TempDir(), nil), &out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"--server", srv.URL, "info"})

	status := cli.Run(context.Background(), cmd)

	if status != 1 {
		t.Errorf("exit status = %d, want 1", status)
	}
	want := "Error: The server at " + srv.URL + " refused this request, because there is no valid session for it: " +
		"you are not signed in, or the session token has expired or been ended. " +
		"Please run starter login to sign in, then run this command again.\n"
	if got := errOut.String(); got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}
