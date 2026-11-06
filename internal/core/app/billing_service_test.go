package app_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/app"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/testsupport"
)

// assertStoredPlans fails the test unless stored holds exactly the plans in want.
func assertStoredPlans(t *testing.T, stored, want testsupport.StoredPlans) {
	t.Helper()
	if !reflect.DeepEqual(stored, want) {
		t.Errorf("stored plans = %v, want %v", stored, want)
	}
}

func TestChangePlan_SavesThePlanForTheGivenWorkspace(t *testing.T) {
	stored := testsupport.StoredPlans{"ws-b": "free"}
	svc := app.NewBillingService(stored, testsupport.KnownPlans{"free", "pro"})

	if err := svc.ChangePlan(context.Background(), "ws-a", "pro"); err != nil {
		t.Fatalf("ChangePlan: %v", err)
	}

	assertStoredPlans(t, stored, testsupport.StoredPlans{"ws-a": "pro", "ws-b": "free"})
}

func TestChangePlan_RefusesAPlanTheCatalogDoesNotHave(t *testing.T) {
	stored := testsupport.StoredPlans{"ws-a": "free"}
	svc := app.NewBillingService(stored, testsupport.KnownPlans{"free", "pro"})

	err := svc.ChangePlan(context.Background(), "ws-a", "platinum")

	var verr *domain.ValidationError
	if !errors.As(err, &verr) || verr.Field != "plan" {
		t.Fatalf("ChangePlan(platinum) = %v, want a *ValidationError on plan", err)
	}
	assertStoredPlans(t, stored, testsupport.StoredPlans{"ws-a": "free"})
}

// errCatalog is the error unreadableCatalog fails with.
var errCatalog = errors.New("catalog unavailable")

// unreadableCatalog is a ports.PlanCatalog whose lookup always fails.
type unreadableCatalog struct{}

func (unreadableCatalog) HasPlan(context.Context, string) (bool, error) {
	return false, errCatalog
}

func TestChangePlan_RefusesWhenTheCatalogCannotBeRead(t *testing.T) {
	stored := testsupport.StoredPlans{"ws-a": "free"}
	svc := app.NewBillingService(stored, unreadableCatalog{})

	err := svc.ChangePlan(context.Background(), "ws-a", "pro")

	if !errors.Is(err, errCatalog) {
		t.Fatalf("ChangePlan = %v, want the catalog's error", err)
	}
	assertStoredPlans(t, stored, testsupport.StoredPlans{"ws-a": "free"})
}

func TestChangePlan_RefusesAnEmptyWorkspaceBeforeTouchingTheCatalogOrStorage(t *testing.T) {
	stored := testsupport.StoredPlans{"ws-a": "free"}
	svc := app.NewBillingService(stored, unreadableCatalog{})

	err := svc.ChangePlan(context.Background(), "", "pro")

	if !errors.Is(err, domain.ErrNoWorkspace) {
		t.Fatalf("ChangePlan = %v, want ErrNoWorkspace", err)
	}
	assertStoredPlans(t, stored, testsupport.StoredPlans{"ws-a": "free"})
}
