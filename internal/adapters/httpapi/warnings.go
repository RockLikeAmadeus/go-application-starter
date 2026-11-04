package httpapi

import (
	"context"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi/openapi"
)

// The Warnings endpoint. Warnings aren't a business capability, so they come
// from Options as given (bootstrap.App.Warnings), not from a driving port.

func (h *Handlers) GetWarnings(_ context.Context, _ openapi.GetWarningsRequestObject) (openapi.GetWarningsResponseObject, error) {
	// Never nil: an empty list, not null, when there are none.
	warnings := make([]openapi.Warning, len(h.warnings))
	for i, message := range h.warnings {
		warnings[i] = openapi.Warning{Message: message}
	}
	return openapi.GetWarnings200JSONResponse{Warnings: warnings}, nil
}
