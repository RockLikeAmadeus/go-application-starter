package httpapi

import (
	"bytes"
	"context"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi/openapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// The Items endpoints.

func (h *Handlers) ListItems(ctx context.Context, _ openapi.ListItemsRequestObject) (openapi.ListItemsResponseObject, error) {
	items, err := h.items.ListItems(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]openapi.Item, 0, len(items))
	for _, it := range items {
		out = append(out, toItemDTO(it))
	}
	return openapi.ListItems200JSONResponse{Items: out}, nil
}

func (h *Handlers) CreateItem(ctx context.Context, req openapi.CreateItemRequestObject) (openapi.CreateItemResponseObject, error) {
	item, err := h.items.CreateItem(ctx, req.Body.Name)
	if err != nil {
		return nil, err
	}
	return openapi.CreateItem201JSONResponse(toItemDTO(item)), nil
}

func (h *Handlers) GetItem(ctx context.Context, req openapi.GetItemRequestObject) (openapi.GetItemResponseObject, error) {
	item, err := h.items.GetItem(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return openapi.GetItem200JSONResponse(toItemDTO(item)), nil
}

func (h *Handlers) RenameItem(ctx context.Context, req openapi.RenameItemRequestObject) (openapi.RenameItemResponseObject, error) {
	item, err := h.items.RenameItem(ctx, req.Id, req.Body.Name)
	if err != nil {
		return nil, err
	}
	return openapi.RenameItem200JSONResponse(toItemDTO(item)), nil
}

func (h *Handlers) DeleteItem(ctx context.Context, req openapi.DeleteItemRequestObject) (openapi.DeleteItemResponseObject, error) {
	if err := h.items.DeleteItem(ctx, req.Id); err != nil {
		return nil, err
	}
	return openapi.DeleteItem204Response{}, nil
}

func (h *Handlers) ExportItems(ctx context.Context, _ openapi.ExportItemsRequestObject) (openapi.ExportItemsResponseObject, error) {
	csv, err := h.items.ExportItems(ctx)
	if err != nil {
		return nil, err
	}
	return openapi.ExportItems200TextcsvResponse{Body: bytes.NewReader(csv), ContentLength: int64(len(csv))}, nil
}

func toItemDTO(i domain.Item) openapi.Item {
	return openapi.Item{Id: i.ID, Name: i.Name, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt}
}
