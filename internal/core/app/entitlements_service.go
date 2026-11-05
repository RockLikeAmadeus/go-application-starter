package app

import (
	"context"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// EntitlementsService implements ports.EntitlementsService.
type EntitlementsService struct {
	entitlementsProvider ports.EntitlementsProvider
	workspaces           ports.WorkspaceResolver
}

var _ ports.EntitlementsService = (*EntitlementsService)(nil)

// NewEntitlementsService wires the use case to its driven ports.
func NewEntitlementsService(entitlementsProvider ports.EntitlementsProvider, workspaces ports.WorkspaceResolver) *EntitlementsService {
	return &EntitlementsService{entitlementsProvider: entitlementsProvider, workspaces: workspaces}
}

// CurrentEntitlements returns what the current workspace's plan allows it.
func (s *EntitlementsService) CurrentEntitlements(ctx context.Context) (domain.Entitlements, error) {
	workspace, err := s.workspaces.CurrentWorkspace(ctx)
	if err != nil {
		return domain.Entitlements{}, err
	}
	return s.entitlementsProvider.For(ctx, workspace)
}
