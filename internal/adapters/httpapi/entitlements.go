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

func toEntitlementsDTO(entitlements domain.Entitlements) openapi.Entitlements {
	limits := make(map[string]openapi.Limit, len(entitlements.Limits))
	for name, limit := range entitlements.Limits {
		limits[name] = openapi.Limit{Max: limit.Max, WhenOver: openapi.OverLimitBehavior(limit.WhenOver)}
	}
	features := make(map[string]openapi.Feature, len(entitlements.Features))
	for key, feature := range entitlements.Features {
		features[key] = toFeatureDTO(feature)
	}
	return openapi.Entitlements{Plan: entitlements.Plan, Limits: limits, Features: features}
}

func toFeatureDTO(feature domain.Feature) openapi.Feature {
	return openapi.Feature{Name: feature.Name, IsOn: feature.IsOn, WhenOff: openapi.FeatureOffBehavior(feature.WhenOff), AvailableOn: availableOnDTO(feature.AvailableOn)}
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
