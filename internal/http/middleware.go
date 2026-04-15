package http

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"wsapi/internal/config"
)

type contextKey string

const (
	CorrelationIDKey contextKey = "correlation_id"
	LoggerKey        contextKey = "logger"
)

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func CorrelationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corrID := r.Header.Get("X-Correlation-ID")
		if corrID == "" {
			corrID = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), CorrelationIDKey, corrID)

		logger := config.GetLogger().With().
			Str("correlation_id", corrID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Logger()

		ctx = context.WithValue(ctx, LoggerKey, logger)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		logger := GetLoggerFromContext(r.Context())
		logger.Info().Msg("request started")

		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)

		duration := time.Since(start)

		logger.Info().
			Int("status", rec.statusCode).
			Str("duration", duration.String()).
			Msg("request completed")
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusRecorder) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func GetCorrelationID(ctx context.Context) string {
	if corrID, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return corrID
	}
	return ""
}

func GetLoggerFromContext(ctx context.Context) *zerolog.Logger {
	if logger, ok := ctx.Value(LoggerKey).(*zerolog.Logger); ok {
		return logger
	}
	return config.GetLogger()
}

func LogWithContext(ctx context.Context) *zerolog.Logger {
	logger := GetLoggerFromContext(ctx)

	corrID := GetCorrelationID(ctx)
	if corrID != "" {
		l := logger.With().Str("correlation_id", corrID).Logger()
		return &l
	}

	return logger
}
