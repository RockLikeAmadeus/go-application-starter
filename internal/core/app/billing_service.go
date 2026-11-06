package app

import (
	"context"

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
	known, _ := s.catalog.HasPlan(ctx, plan)
	if !known {
		return &domain.ValidationError{Field: "plan", Message: "does not exist"}
	}
	return s.workspacePlans.SavePlan(ctx, workspace, plan)
}
