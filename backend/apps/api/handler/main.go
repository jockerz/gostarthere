package handler

import (
	"context"
	"vnti/apps/api/presenter"
)

func APICheck(ctx context.Context, _ *struct{}) (*presenter.SuccessResponse, error) {
	return &presenter.SuccessResponse{Body: presenter.SuccessBody{
		Success: true,
		Message: "Success",
	}}, nil
}
