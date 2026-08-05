package logger

import (
	"context"

	"github.com/gofiber/fiber/v3/middleware/requestid"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func LogContext(ctx context.Context, ns, event string) zerolog.Logger {
	reqId := requestid.FromContext(ctx)
	return log.Logger.
		With().
		Ctx(ctx).
		Str("r", reqId).
		Str("event", event).
		Str("ns", ns).
		Logger()
}
