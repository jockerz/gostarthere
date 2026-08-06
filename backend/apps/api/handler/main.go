package handler

import (
	"context"
	"errors"
	"vnti/apps/api/presenter"
	"vnti/internal/logger"
)

func APICheck(ctx context.Context, _ *struct{}) (*presenter.SuccessResponse, error) {
	l := logger.LogContext(ctx, "api:handler:main:apiCheck", "text")
	l.Debug().Msg("Test - main")

	// return &presenter.SuccessResponse{Body: presenter.SuccessBody{
	// 	Success: true,
	// 	Message: "Success",
	// }}, nil
	return nil, errors.New("Yo")
}
