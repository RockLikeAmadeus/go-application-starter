package app

import (
	"context"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// EntitlementsService implements ports.EntitlementsService.
type EntitlementsService struct {
	entitlements ports.Entitlements
	workspaces   ports.WorkspaceResolver
}

var _ ports.EntitlementsService = (*EntitlementsService)(nil)

// NewEntitlementsService wires the use case to its driven ports.
func NewEntitlementsService(entitlements ports.Entitlements, workspaces ports.WorkspaceResolver) *EntitlementsService {
	return &EntitlementsService{entitlements: entitlements, workspaces: workspaces}
}

// CurrentEntitlements returns what the current workspace's plan allows it.
func (s *EntitlementsService) CurrentEntitlements(ctx context.Context) (domain.Entitlements, error) {
	workspace, _ := s.workspaces.CurrentWorkspace(ctx)
	return s.entitlements.For(ctx, workspace)
}
