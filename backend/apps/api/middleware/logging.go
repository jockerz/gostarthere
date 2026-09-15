package middleware

import (
	// "vnti/extensions/logger"

	"vnti/internal"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"

	"github.com/rs/zerolog/log"
)

type internalLogCtx struct{}

// Un-used
func LoggerToCtxMiddleware(config *internal.Config) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		reqId := requestid.FromContext(ctx)
		log.Debug().Msg("Before")
		nCtx := log.With().Str("r", reqId).Logger().WithContext(ctx)

		err := ctx.Next()

		// TODO: remove
		log.Ctx(nCtx).Warn().Msg("Logger yo-")
		log.Ctx(ctx).Warn().Msg("Logger yo---")
		log.Debug().Msg("After")

		return err
	}
}
