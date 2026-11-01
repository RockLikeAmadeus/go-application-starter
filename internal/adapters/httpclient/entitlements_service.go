package httpclient

import (
	"context"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient/apiclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// EntitlementsService is ports.EntitlementsService over HTTP: the plan and
// limits of the workspace the server puts the request in.
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
	return domain.Entitlements{Plan: e.Plan, Limits: limits}
}
