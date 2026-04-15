package config

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
)

var GlobalLogger *zerolog.Logger

func InitLogger() *zerolog.Logger {
	level := strings.ToLower(os.Getenv("LOG_LEVEL"))
	var zerologLevel zerolog.Level

	switch level {
	case "debug":
		zerologLevel = zerolog.DebugLevel
	case "warn":
		zerologLevel = zerolog.WarnLevel
	case "error":
		zerologLevel = zerolog.ErrorLevel
	default:
		zerologLevel = zerolog.InfoLevel
	}

	logger := zerolog.New(os.Stdout).
		Level(zerologLevel).
		With().
		Timestamp().
		Caller().
		Logger()

	GlobalLogger = &logger
	return &logger
}

func GetLogger() *zerolog.Logger {
	if GlobalLogger == nil || GlobalLogger.GetLevel() == zerolog.NoLevel {
		return InitLogger()
	}
	return GlobalLogger
}

type LoggerContext struct {
	CorrelationID string
	RUCEmpresa    string
	ReferenceID   string
}

func (c LoggerContext) AddToLog(e *zerolog.Event) *zerolog.Event {
	if c.CorrelationID != "" {
		e.Str("correlation_id", c.CorrelationID)
	}
	if c.RUCEmpresa != "" {
		e.Str("ruc_empresa", c.RUCEmpresa)
	}
	if c.ReferenceID != "" {
		e.Str("reference_id", c.ReferenceID)
	}
	return e
}
