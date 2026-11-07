package app

import (
	"context"
	"fmt"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// BillingService implements ports.BillingService.
type BillingService struct {
	workspacePlans ports.WorkspacePlanRepository
	catalog        ports.PlanCatalog
}

var _ ports.BillingService = (*BillingService)(nil)

// NewBillingService wires the use case to its driven ports.
func NewBillingService(workspacePlans ports.WorkspacePlanRepository, catalog ports.PlanCatalog) *BillingService {
	return &BillingService{workspacePlans: workspacePlans, catalog: catalog}
}

// ChangePlan moves the workspace to the plan.
func (s *BillingService) ChangePlan(ctx context.Context, workspace domain.WorkspaceID, plan string) error {
	if workspace == "" {
		return domain.ErrNoWorkspace
	}
	known, err := s.catalog.HasPlan(ctx, plan)
	if err != nil {
		return fmt.Errorf("look up plan: %w", err)
	}
	if !known {
		return &domain.ValidationError{Field: "plan", Message: "plan does not exist"}
	}
	return s.workspacePlans.SavePlan(ctx, workspace, plan)
}
