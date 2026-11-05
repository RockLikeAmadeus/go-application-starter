package httpclient

import (
	"context"
	"net/http"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient/apiclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

type ItemService struct {
	api *apiclient.ClientWithResponses
}

var _ ports.ItemService = (*ItemService)(nil)

// New returns a client for the server at baseURL, e.g. "http://localhost:8080".
func New(baseURL string, opts ...apiclient.ClientOption) (*ItemService, error) {
	api, err := newAPI(baseURL, opts...)
	if err != nil {
		return nil, err
	}
	return &ItemService{api: api}, nil
}

func (s *ItemService) CreateItem(ctx context.Context, name string) (domain.Item, error) {
	res, err := s.api.CreateItemWithResponse(ctx, apiclient.ItemInput{Name: name})
	if err != nil {
		return domain.Item{}, err
	}
	if res.JSON201 == nil {
		return domain.Item{}, toError(res.StatusCode(), res.ApplicationproblemJSONDefault)
	}
	return toDomain(*res.JSON201), nil
}

func (s *ItemService) GetItem(ctx context.Context, id string) (domain.Item, error) {
	res, err := s.api.GetItemWithResponse(ctx, id)
	if err != nil {
		return domain.Item{}, err
	}
	if res.JSON200 == nil {
		return domain.Item{}, toError(res.StatusCode(), res.ApplicationproblemJSONDefault)
	}
	return toDomain(*res.JSON200), nil
}

func (s *ItemService) ListItems(ctx context.Context) ([]domain.Item, error) {
	res, err := s.api.ListItemsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if res.JSON200 == nil {
		return nil, toError(res.StatusCode(), res.ApplicationproblemJSONDefault)
	}
	items := make([]domain.Item, 0, len(res.JSON200.Items))
	for _, it := range res.JSON200.Items {
		items = append(items, toDomain(it))
	}
	return items, nil
}

func (s *ItemService) RenameItem(ctx context.Context, id, name string) (domain.Item, error) {
	res, err := s.api.RenameItemWithResponse(ctx, id, apiclient.ItemInput{Name: name})
	if err != nil {
		return domain.Item{}, err
	}
	if res.JSON200 == nil {
		return domain.Item{}, toError(res.StatusCode(), res.ApplicationproblemJSONDefault)
	}
	return toDomain(*res.JSON200), nil
}

func (s *ItemService) DeleteItem(ctx context.Context, id string) error {
	res, err := s.api.DeleteItemWithResponse(ctx, id)
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusNoContent {
		return toError(res.StatusCode(), res.ApplicationproblemJSONDefault)
	}
	return nil
}

func toDomain(i apiclient.Item) domain.Item {
	return domain.Item{ID: i.Id, Name: i.Name, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt}
}

// ExportItems fetches the workspace's items as CSV, as the server encodes them.
func (s *ItemService) ExportItems(ctx context.Context) ([]byte, error) {
	res, err := s.api.ExportItemsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, toError(res.StatusCode(), res.ApplicationproblemJSONDefault)
	}
	return res.Body, nil
}
