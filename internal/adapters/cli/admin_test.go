package cli_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/cli"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// recordingBilling remembers the plan change it was asked for, and answers
// with refusal.
type recordingBilling struct {
	workspace domain.WorkspaceID
	plan      string
	refusal   error
}

func (b *recordingBilling) ChangePlan(_ context.Context, workspace domain.WorkspaceID, plan string) error {
	b.workspace, b.plan = workspace, plan
	return b.refusal
}

// usingBilling is a billing factory that hands out billing.
func usingBilling(billing ports.BillingService) cli.BillingFactory {
	return func(config.Config) (ports.BillingService, error) {
		return billing, nil
	}
}

// runAdmin runs `starter admin ARGS...` with the given billing factory.
func runAdmin(t *testing.T, newBilling cli.BillingFactory, out io.Writer, args ...string) error {
	t.Helper()
	cmd := cli.NewRootCmd(bootstrap.New, newBilling, computer(t.TempDir(), nil), out)
	cmd.SetArgs(append([]string{"admin"}, args...))
	return cmd.ExecuteContext(context.Background())
}

func TestAdminHelpSaysWhoItsCommandsAreFor(t *testing.T) {
	out, err := runCLI(t, "admin", "--help")
	if err != nil {
		t.Fatalf("admin --help: %v", err)
	}
	if want := "Commands for product administrators"; !strings.Contains(out, want) {
		t.Errorf("admin --help doesn't say %q:\n%s", want, out)
	}
}

func TestAdminSetPlanReportsThatBillingRefused(t *testing.T) {
	refused := errors.New("refused")
	newBilling := func(config.Config) (ports.BillingService, error) {
		return nil, refused
	}

	if err := runAdmin(t, newBilling, &bytes.Buffer{}, "set-plan", "ws-a", "pro"); !errors.Is(err, refused) {
		t.Errorf("admin set-plan = %v, want the refusal", err)
	}
}

func TestAdminSetPlanMovesTheWorkspaceAndSaysWhichPlanItIsOn(t *testing.T) {
	billing := &recordingBilling{}
	var out bytes.Buffer

	if err := runAdmin(t, usingBilling(billing), &out, "set-plan", "ws-a", "pro"); err != nil {
		t.Fatalf("admin set-plan: %v", err)
	}
	if billing.workspace != "ws-a" || billing.plan != "pro" {
		t.Errorf("ChangePlan(%q, %q), want (\"ws-a\", \"pro\")", billing.workspace, billing.plan)
	}
	if want := "ws-a is now on the pro plan"; !strings.Contains(out.String(), want) {
		t.Errorf("output = %q, want it to contain %q", out.String(), want)
	}
}

func TestAdminSetPlanReportsChangePlansRefusalAndDoesNotSayTheWorkspaceMoved(t *testing.T) {
	refusal := errors.New("no such plan")
	billing := &recordingBilling{refusal: refusal}
	var out bytes.Buffer

	if err := runAdmin(t, usingBilling(billing), &out, "set-plan", "ws-a", "platinum"); !errors.Is(err, refusal) {
		t.Errorf("admin set-plan = %v, want the refusal", err)
	}
	if strings.Contains(out.String(), "is now on") {
		t.Errorf("output = %q, want no claim that the workspace moved", out.String())
	}
}

func TestAdminSetPlanReportsSettingsItCannotLoadAndChangesNothing(t *testing.T) {
	billing := &recordingBilling{}

	err := runAdmin(t, usingBilling(billing), &bytes.Buffer{}, "set-plan", "ws-a", "pro", "--storage", "bogus")

	if err == nil || !strings.Contains(err.Error(), "bogus") {
		t.Errorf("admin set-plan = %v, want the settings error naming %q", err, "bogus")
	}
	if billing.workspace != "" || billing.plan != "" {
		t.Errorf("ChangePlan(%q, %q), want no call", billing.workspace, billing.plan)
	}
}

func TestAdminSetPlanRefusesAnythingButAWorkspaceAndAPlan(t *testing.T) {
	cases := map[string][]string{
		"three arguments":  {"ws-a", "pro", "extra"},
		"only a workspace": {"ws-a"},
		"no arguments":     {},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			billing := &recordingBilling{}

			err := runAdmin(t, usingBilling(billing), &bytes.Buffer{}, append([]string{"set-plan"}, args...)...)

			if err == nil {
				t.Errorf("admin set-plan %v = <nil>, want a usage error", args)
			}
			if billing.workspace != "" || billing.plan != "" {
				t.Errorf("ChangePlan(%q, %q), want no call", billing.workspace, billing.plan)
			}
		})
	}
}

func TestAdminSetPlanOverAServerIsRefused(t *testing.T) {
	err := runAdmin(t, bootstrap.NewBilling, &bytes.Buffer{}, "set-plan", "ws-a", "pro", "--server", "http://localhost:1")

	if !errors.Is(err, bootstrap.ErrNoPlanChangesOverAServer) {
		t.Errorf("admin set-plan --server = %v, want ErrNoPlanChangesOverAServer", err)
	}
}
