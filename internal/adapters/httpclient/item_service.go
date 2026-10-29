// Package httpclient implements the ports.ItemService driving port over
// HTTP, so the CLI and TUI can run against a remote server with exactly the
// same code they use against the embedded core.
package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

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
	opts = append([]apiclient.ClientOption{
		apiclient.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
	}, opts...)
	api, err := apiclient.NewClientWithResponses(strings.TrimRight(baseURL, "/"), opts...)
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

// toError maps a problem document back to the domain error the server
// started from, so callers can use errors.Is regardless of transport.
func toError(status int, p *apiclient.Problem) error {
	if p == nil {
		return fmt.Errorf("unexpected HTTP status %d", status)
	}
	typ, detail, field := deref(p.Type), deref(p.Detail), deref(p.Field)
	switch typ {
	case "not_found":
		return domain.ErrNotFound
	case "validation":
		return &domain.ValidationError{Field: field, Message: strings.TrimSpace(strings.TrimPrefix(detail, field))}
	default:
		return fmt.Errorf("server error (%d): %s", status, detail)
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
