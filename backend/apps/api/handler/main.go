package handler

import (
	"context"
	"vnti/apps/api/presenter"
	"vnti/extensions/logger"
	// "github.com/rs/zerolog/log"
)

func APICheck(ctx context.Context, _ *struct{}) (*presenter.SuccessResponse, error) {
	logCtx := logger.APILogMessage(ctx, &logger.Logger, "test", "api:handler:main:apiCheck")

	log := logCtx.Logger()
	log.Debug().Msg("Text log")
	panic("panic")
	// return nil, errors.Join(errors.New("tes"), errors.New("tes2"), errors.New("tes3"))

	return &presenter.SuccessResponse{Body: presenter.SuccessBody{
		Success: true,
		Message: "Success",
	}}, nil
}
