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

func TestChangePlan_SavesThePlanForTheGivenWorkspace(t *testing.T) {
	stored := testsupport.StoredPlans{"ws-b": "free"}
	svc := app.NewBillingService(stored, testsupport.KnownPlans{"free", "pro"})

	if err := svc.ChangePlan(context.Background(), "ws-a", "pro"); err != nil {
		t.Fatalf("ChangePlan: %v", err)
	}

	want := testsupport.StoredPlans{"ws-a": "pro", "ws-b": "free"}
	if !reflect.DeepEqual(stored, want) {
		t.Errorf("stored plans = %v, want %v", stored, want)
	}
}

func TestChangePlan_RefusesAPlanTheCatalogDoesNotHave(t *testing.T) {
	stored := testsupport.StoredPlans{"ws-a": "free"}
	svc := app.NewBillingService(stored, testsupport.KnownPlans{"free", "pro"})

	err := svc.ChangePlan(context.Background(), "ws-a", "platinum")

	var verr *domain.ValidationError
	if !errors.As(err, &verr) || verr.Field != "plan" {
		t.Fatalf("ChangePlan(platinum) = %v, want a *ValidationError on plan", err)
	}
	want := testsupport.StoredPlans{"ws-a": "free"}
	if !reflect.DeepEqual(stored, want) {
		t.Errorf("stored plans = %v, want %v", stored, want)
	}
}
