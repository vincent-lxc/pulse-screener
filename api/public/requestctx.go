package public

import (
	"context"

	servertypes "github.com/digitalwayhk/core/pkg/server/types"
	"github.com/digitalwayhk/pulse/models"
)

func reqContext(req servertypes.IRequest) context.Context {
	if req != nil {
		if httpReq, ok := req.(servertypes.IRequestHttp); ok {
			if raw := httpReq.GetHttpRequest(); raw != nil {
				return raw.Context()
			}
		}
	}
	return context.Background()
}

func dtoQuoteIDRequired() error {
	return models.NewValidationError("请提供 id/cmcID 或 symbol")
}
