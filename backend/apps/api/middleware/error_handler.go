package middleware

import (
	"errors"
	"fmt"
	"runtime/debug"
	"vnti/internal/logger"

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

	l := logger.LogContext(ctx, "", "error")
	l.Err(e).Msg(fmt.Sprintf("%d:%s", code, message))

	return ctx.Status(code).JSON(map[string]any{
		"success": false,
		"error":   message,
	})
}

func PanicHander(ctx fiber.Ctx, e any) error {
	l := logger.LogContext(ctx, "", "panic")
	l.Error().Msg(string(debug.Stack()))

	// Error to be handled by ErrorHandler
	if err, ok := e.(error); ok {
		l.Err(err).Msg(string(debug.Stack()))
		return err
	} else {
		l.Error().Msg(string(debug.Stack()))
	}
	return fiber.ErrInternalServerError
}
