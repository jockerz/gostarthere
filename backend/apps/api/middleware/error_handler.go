package middleware

import (
	"errors"
	"fmt"
	"runtime/debug"
	"vnti/extensions/logger"

	"github.com/gofiber/fiber/v3"
)

func ErrorHandler(ctx fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal server error"

	var e *fiber.Error

	if errors.As(err, &e) && e != nil {
		code = e.Code
		message = e.Message
	}

	logCtx := logger.APILogMessage(ctx, &logger.Logger, "", "")
	log := logCtx.Logger()
	log.Error().Msg(fmt.Sprintf("%d:%s", code, message))

	err = ctx.Status(code).JSON(map[string]any{
		"success": false,
		"error":   message,
	})
	return nil
}

func PanicHander(ctx fiber.Ctx, e any) error {
	logCtx := logger.APILogMessage(ctx, &logger.Logger, "", "")

	log := logCtx.Logger()
	log.Error().Msg(string(debug.Stack()))

	// Error to be handled by ErrorHandler
	if err, ok := e.(error); ok {
		return err
	}
	return errors.New("Internal server error")
}
