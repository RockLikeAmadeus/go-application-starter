package httpapi

import (
	"context"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi/openapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// The Entitlements endpoint.

func (h *Handlers) GetEntitlements(ctx context.Context, _ openapi.GetEntitlementsRequestObject) (openapi.GetEntitlementsResponseObject, error) {
	entitlements, err := h.entitlements.CurrentEntitlements(ctx)
	if err != nil {
		return nil, err
	}
	return openapi.GetEntitlements200JSONResponse(toEntitlementsDTO(entitlements)), nil
}

func toEntitlementsDTO(e domain.Entitlements) openapi.Entitlements {
	limits := make(map[string]openapi.Limit, len(e.Limits))
	for name, limit := range e.Limits {
		limits[name] = openapi.Limit{Max: limit.Max, WhenOver: openapi.LimitWhenOver(limit.WhenOver)}
	}
	return openapi.Entitlements{Plan: e.Plan, Limits: limits}
}
