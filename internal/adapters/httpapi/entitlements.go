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
	features := make(map[string]openapi.Feature, len(e.Features))
	for key, f := range e.Features {
		features[key] = toFeatureDTO(f)
	}
	return openapi.Entitlements{Plan: e.Plan, Limits: limits, Features: features}
}

func toFeatureDTO(f domain.Feature) openapi.Feature {
	return openapi.Feature{Name: f.Name, On: f.On, WhenOff: openapi.FeatureWhenOff(f.WhenOff), AvailableOn: availableOnDTO(f.AvailableOn)}
}

// availableOnDTO keeps the upgrade hint's plans a list, never null, as the
// contract requires of both the entitlements and the feature problem: with
// no hint it's empty.
func availableOnDTO(plans []string) []string {
	if plans == nil {
		return []string{}
	}
	return plans
}
