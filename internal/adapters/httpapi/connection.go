package httpapi

import (
	"context"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi/openapi"
)

// The Connection endpoint.

func (h *Handlers) GetConnection(_ context.Context, _ openapi.GetConnectionRequestObject) (openapi.GetConnectionResponseObject, error) {
	return openapi.GetConnection200JSONResponse{Required: !h.localData}, nil
}
