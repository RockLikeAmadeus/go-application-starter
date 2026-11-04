package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/cli"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// A command prints the app's warnings on stderr, so its output, which
// scripts read, stays as it is.
func TestCommandsPrintTheAppsWarningsOnStderr(t *testing.T) {
	cases := []struct {
		name     string
		warnings []string
		want     string
	}{
		{"none", nil, ""},
		{"two", []string{"first warning", "second warning"}, "warning: first warning\nwarning: second warning\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			newApp := func(cfg config.Config) (*bootstrap.App, error) {
				app, err := bootstrap.New(cfg)
				if err != nil {
					return nil, err
				}
				app.Warnings = tc.warnings
				return app, nil
			}
			var out, errOut bytes.Buffer
			cmd := cli.NewRootCmd(newApp, computer(t.TempDir(), nil), &out)
			cmd.SetErr(&errOut)
			cmd.SetArgs([]string{"--storage", "memory", "info"})

			if err := cmd.ExecuteContext(context.Background()); err != nil {
				t.Fatalf("info: %v", err)
			}
			if got := errOut.String(); got != tc.want {
				t.Errorf("stderr = %q, want %q", got, tc.want)
			}
			if strings.Contains(out.String(), "warning") {
				t.Errorf("stdout holds a warning:\n%s", out.String())
			}
		})
	}
}
