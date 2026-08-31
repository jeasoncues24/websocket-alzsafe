package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestLoggingMiddlewareAllowsStreamingFlush protege contra la regresión que rompía
// SSE: el wrapper statusRecorder debe exponer Flush/Unwrap para que
// http.ResponseController pueda vaciar la respuesta de forma incremental a través
// del logging middleware. Sin esto, un handler SSE que trata el fallo de Flush
// como fatal cierra la conexión tras el primer evento.
func TestLoggingMiddlewareAllowsStreamingFlush(t *testing.T) {
	var flushErr error

	h := LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc := http.NewResponseController(w)
		if _, err := w.Write([]byte("event: a\n\n")); err != nil {
			t.Fatalf("write: %v", err)
		}
		flushErr = rc.Flush()
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/admin/sesiones/stream", nil))

	if flushErr != nil {
		t.Fatalf("ResponseController.Flush a través de LoggingMiddleware devolvió error: %v", flushErr)
	}
	if !rec.Flushed {
		t.Fatal("se esperaba que la respuesta se hubiera vaciado (Flushed=true)")
	}
}

// TestStatusRecorderImplementsStreamingInterfaces documenta el contrato que
// necesitan los handlers de streaming.
func TestStatusRecorderImplementsStreamingInterfaces(t *testing.T) {
	var sr any = &statusRecorder{ResponseWriter: httptest.NewRecorder()}

	if _, ok := sr.(http.Flusher); !ok {
		t.Error("statusRecorder debe implementar http.Flusher")
	}
	if _, ok := sr.(interface{ Unwrap() http.ResponseWriter }); !ok {
		t.Error("statusRecorder debe implementar Unwrap() http.ResponseWriter")
	}
}
