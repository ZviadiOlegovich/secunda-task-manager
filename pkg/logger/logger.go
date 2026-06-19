package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func ParseLevel(value string) zerolog.Level {
	switch strings.ToLower(value) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "trace":
		return zerolog.TraceLevel
	default:
		log.Warn().Msgf("unknown log level %q, using info", value)
		return zerolog.InfoLevel
	}
}

func New(forcePlain bool, level zerolog.Level, serviceName string) zerolog.Logger {
	if !forcePlain {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	}

	zerolog.SetGlobalLevel(level)

	log.Logger = log.With().
		Str("service", serviceName).
		Caller().
		Logger()

	zerolog.DefaultContextLogger = &log.Logger

	return log.Logger
}
