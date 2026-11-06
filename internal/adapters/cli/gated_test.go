package cli_test

import (
	"bytes"
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/cli"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// The help lists a gated command by its feature's state on the plan, and
// reads only the entitlements: it never opens the data folder.
func TestHelpNotesGatedCommandsByThePlan(t *testing.T) {
	const listed = "report Make a report"
	tests := []struct {
		name string
		// plans is the local plans file ("" for none, so it can't be read);
		// server, when set, answers the entitlements instead.
		plans  string
		server *domain.Entitlements
		want   string // the command's line in the help, "" when not listed
	}{
		{"on", plansWith(`"show-locked"`, `"reports"`), nil, listed},
		{"locked, in the local edition", plansWith(`"show-locked"`, ``), nil, listed + " (disabled on your plan)"},
		{"hidden", plansWith(`"hide"`, ``), nil, ""},
		{"locked, with a server's hint", "", &domain.Entitlements{Plan: "free", Features: map[string]domain.Feature{
			"reports": {Key: "reports", Name: "Reports", WhenOff: domain.ShowLocked, AvailableOn: []string{"pro"}},
		}}, listed + " (not included in your current plan; available on the pro plan)"},
		{"locked, with several plans in a server's hint", "", &domain.Entitlements{Plan: "free", Features: map[string]domain.Feature{
			"reports": {Key: "reports", Name: "Reports", WhenOff: domain.ShowLocked, AvailableOn: []string{"pro", "team"}},
		}}, listed + " (not included in your current plan; available on the following plans: pro or team)"},
		{"entitlements unreadable", "", nil, listed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := map[string]string{"STARTER_PLANS_FILE": filepath.Join(t.TempDir(), "missing.jsonc")}
			if tt.plans != "" {
				vars["STARTER_PLANS_FILE"] = writeFile(t, "plans.jsonc", tt.plans)
			}
			data := filepath.Join(t.TempDir(), "data")
			args := []string{"--storage", "jsonfile", "--data-dir", data, "--help"}
			if tt.server != nil {
				args = append([]string{"--server", entitlementsServer(t, *tt.server)}, args...)
			}
			var out bytes.Buffer
			root := cli.NewRootCmd(bootstrap.New, bootstrap.NewBilling, computer(t.TempDir(), vars), &out)
			root.AddCommand(cli.Gated("reports", &cobra.Command{
				Use: "report", Short: "Make a report",
				RunE: func(*cobra.Command, []string) error { return nil },
			}))
			root.SetArgs(args)

			if err := root.ExecuteContext(context.Background()); err != nil {
				t.Fatalf("help: %v", err)
			}
			if got := commandLine(out.String(), "report"); got != tt.want {
				t.Errorf("help lists %q, want %q:\n%s", got, tt.want, out.String())
			}
			if _, err := os.Stat(data); !os.IsNotExist(err) {
				t.Errorf("help touched the data folder %s (stat: %v)", data, err)
			}
		})
	}
}

// plansWith is a plans file whose one feature, reports, is whenOff when
// off, and whose local plan lists features (e.g. `"reports"`, or nothing).
func plansWith(whenOff, features string) string {
	return `{
	"features": {"reports": {"name": "Reports", "whenOff": ` + whenOff + `}},
	"plans": {"local": {"limits": {}, "features": [` + features + `]}}
}`
}

// commandLine returns the help's line for the command name, with its spaces
// collapsed, or "" when the help doesn't list it.
func commandLine(help, name string) string {
	for _, line := range strings.Split(help, "\n") {
		if fields := strings.Fields(line); len(fields) > 0 && fields[0] == name {
			return strings.Join(fields, " ")
		}
	}
	return ""
}

// writeFile writes content to a file called name in a new folder, and
// returns its path.
func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// entitlementsServer starts a server answering only the entitlements, with
// e, and returns its URL.
func entitlementsServer(t *testing.T, e domain.Entitlements) string {
	t.Helper()
	srv := httptest.NewServer(httpapi.NewHandler(httpapi.Options{Entitlements: fixedEntitlements(e)}))
	t.Cleanup(srv.Close)
	return srv.URL
}

// fixedEntitlements is a ports.EntitlementsService that always answers
// itself.
type fixedEntitlements domain.Entitlements

func (f fixedEntitlements) CurrentEntitlements(context.Context) (domain.Entitlements, error) {
	return domain.Entitlements(f), nil
}
