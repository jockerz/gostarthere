package logger

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/rs/zerolog"
)

var AccessLogger, Logger zerolog.Logger

func InitLogger(appName, logType string, isDebug bool, extra ...string) {
	var output zerolog.ConsoleWriter

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if isDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		output = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.DateTime}
		output.FormatLevel = func(i any) string {
			return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
		}
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		output = zerolog.ConsoleWriter{Out: os.Stdout}
	}
	Logger = zerolog.New(output).With().
		Timestamp().
		Str("log_app", appName).
		Str("log_type", logType).
		Logger()
	AccessLogger = zerolog.New(output).With().
		Timestamp().
		Str("log_app", appName).
		Str("log_type", "ACCESS").
		Logger()
}

func APILogMessage(ctx context.Context, logger *zerolog.Logger, event, namespace string, data ...any) zerolog.Context {
	requestId := requestid.FromContext(ctx)

	logCtx := logger.With().Str("r", requestId)
	if event != "" {
		logCtx = logCtx.Str("event", event)
	}
	if namespace != "" {
		logCtx = logCtx.Str("ns", namespace)
	}
	if len(data) > 0 {
		logger.With().Any("data", data[0])
	}

	return logCtx
}
