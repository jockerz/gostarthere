package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var AccessLogger zerolog.Logger

const (
	accessLogAppName = "BACKEND"
)

func InitLogger(appName, logType string, isDebug bool) {
	var logger zerolog.Logger

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if isDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.DateTime}
		output.FormatLevel = func(i any) string {
			return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
		}
		logger = zerolog.New(output)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		logger = zerolog.New(os.Stdout)
	}

	log.Logger = logger.With().
		Timestamp().
		Str("log_app", appName).
		Str("log_type", logType).
		Logger()
}

func InitAccessLogger(isDebug bool) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if isDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.DateTime}
		output.FormatLevel = func(i any) string {
			return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
		}
		AccessLogger = zerolog.New(output)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		AccessLogger = zerolog.New(os.Stdout)
	}
	AccessLogger = AccessLogger.With().
		Timestamp().Str("log_app", accessLogAppName).Str("log_type", "ACCESS").
		Logger()
}
