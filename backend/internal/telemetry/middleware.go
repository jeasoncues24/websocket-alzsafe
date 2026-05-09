package telemetry

import (
	"log"
	"net/http"
	"strings"
	"time"

	"wsapi/internal/domain"
)

type TelemetryMiddleware struct {
	store Store
	cfg   Config
}

func NewTelemetryMiddleware(store Store, cfg Config) *TelemetryMiddleware {
	return &TelemetryMiddleware{store: store, cfg: cfg}
}

type telemetryResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	wroteHeader  bool
}

func (w *telemetryResponseWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.statusCode = code
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *telemetryResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func (m *TelemetryMiddleware) Capture(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, ServicePrefix) {
			next.ServeHTTP(w, r)
			return
		}

		started := time.Now()
		rw := &telemetryResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		event := &domain.TelemetryEvent{
			ContractName: ExtractContractName(r.URL.Path),
			Endpoint:     r.URL.Path,
			Method:       r.Method,
			StatusCode:   rw.statusCode,
			LatencyMS:    int(time.Since(started).Milliseconds()),
			CreatedAt:    started,
		}

		if claims, ok := domain.GetApiKeyClaims(r.Context()); ok {
			event.ApiKeyID = claims.ApiKeyID
			event.EmpresaID = claims.EmpresaID
			event.TelefonoID = claims.TelefonoID
		}

		if rw.statusCode >= 400 {
			event.ErrorCode = http.StatusText(rw.statusCode)
			event.ErrorMessage = strings.TrimSpace(http.StatusText(rw.statusCode))
		}

		if err := m.store.Record(event); err != nil {
			log.Printf("[telemetry] error recording event: %v", err)
		}
	})
}

func (m *TelemetryMiddleware) WrapClientAuth(clientAuth func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return clientAuth(m.Capture(next))
	}
}
