package httpclient

import (
	"context"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient/apiclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// EntitlementsService is ports.EntitlementsService over HTTP: the plan,
// limits and features of the workspace the server puts the request in.
type EntitlementsService struct {
	api *apiclient.ClientWithResponses
}

var _ ports.EntitlementsService = (*EntitlementsService)(nil)

// NewEntitlementsService returns a client for the server at baseURL, e.g.
// "http://localhost:8080".
func NewEntitlementsService(baseURL string, opts ...apiclient.ClientOption) (*EntitlementsService, error) {
	api, err := newAPI(baseURL, opts...)
	if err != nil {
		return nil, err
	}
	return &EntitlementsService{api: api}, nil
}

func (s *EntitlementsService) CurrentEntitlements(ctx context.Context) (domain.Entitlements, error) {
	res, err := s.api.GetEntitlementsWithResponse(ctx)
	if err != nil {
		return domain.Entitlements{}, err
	}
	if res.JSON200 == nil {
		return domain.Entitlements{}, toError(res.StatusCode(), res.ApplicationproblemJSONDefault)
	}
	return toEntitlements(*res.JSON200), nil
}

func toEntitlements(e apiclient.Entitlements) domain.Entitlements {
	limits := make(map[string]domain.Limit, len(e.Limits))
	for name, limit := range e.Limits {
		limits[name] = domain.Limit{Name: name, Max: limit.Max, WhenOver: domain.OverLimit(limit.WhenOver)}
	}
	features := make(map[string]domain.Feature, len(e.Features))
	for key, f := range e.Features {
		features[key] = toFeature(key, f)
	}
	return domain.Entitlements{Plan: e.Plan, Limits: limits, Features: features}
}

// toFeature reads a feature back as the core built it.
func toFeature(key string, f apiclient.Feature) domain.Feature {
	return domain.Feature{Key: key, Name: f.Name, On: f.On, WhenOff: domain.FeatureOff(f.WhenOff), AvailableOn: fromAvailableOn(f.AvailableOn)}
}

// fromAvailableOn reads an upgrade hint back as the core built it: the API
// always sends availableOn as a list, and an empty one (no hint) becomes nil
// again.
func fromAvailableOn(plans []string) []string {
	if len(plans) == 0 {
		return nil
	}
	return plans
}
