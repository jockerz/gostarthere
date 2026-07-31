package middleware

import (
	// "vnti/extensions/logger"
	"vnti/internal"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/rs/zerolog/log"
)

func NewLoggingMiddleware(config *internal.Config) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		reqId := ctx.Get(requestid.ConfigDefault.Header)
		log.With().Str("r", reqId)
		// logger.Logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		// 	return c.Str("request_id", reqId)
		// })
		// logger.Logger.
		return ctx.Next()
	}
}
