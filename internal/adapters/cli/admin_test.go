package cli_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/cli"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

func TestAdminSetPlanReportsThatBillingRefused(t *testing.T) {
	refused := errors.New("refused")
	newBilling := func(config.Config) (ports.BillingService, error) {
		return nil, refused
	}
	cmd := cli.NewRootCmd(bootstrap.New, newBilling, computer(t.TempDir(), nil), &bytes.Buffer{})
	cmd.SetArgs([]string{"admin", "set-plan", "ws-a", "pro"})

	if err := cmd.ExecuteContext(context.Background()); !errors.Is(err, refused) {
		t.Errorf("admin set-plan = %v, want the refusal", err)
	}
}
