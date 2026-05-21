package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestV1Health_GetReturns200WithExpectedShape(t *testing.T) {
	h := NewV1HealthHandler("dev", 1000)
	req := httptest.NewRequest(http.MethodGet, "/api/service/v1/health", nil)
	rr := httptest.NewRecorder()

	h.GetHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("esperado 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type esperado application/json, got %s", ct)
	}

	var body map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("error decodificando respuesta: %v", err)
	}

	if ok, _ := body["ok"].(bool); !ok {
		t.Errorf("campo 'ok' debe ser true")
	}
	if svc, _ := body["service"].(string); svc != "wsapi" {
		t.Errorf("campo 'service' esperado 'wsapi', got %q", svc)
	}
	if ver, _ := body["version"].(string); ver != "dev" {
		t.Errorf("campo 'version' esperado 'dev', got %q", ver)
	}
	ts, _ := body["timestamp"].(string)
	if ts == "" {
		t.Fatal("campo 'timestamp' está vacío")
	}
	if _, err := time.Parse(time.RFC3339, ts); err != nil {
		t.Errorf("timestamp no es RFC3339 válido: %q — %v", ts, err)
	}
}

func TestV1Health_NonGetReturns405(t *testing.T) {
	h := NewV1HealthHandler("dev", 1000)
	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/service/v1/health", nil)
			rr := httptest.NewRecorder()
			h.GetHealth(rr, req)

			if rr.Code != http.StatusMethodNotAllowed {
				t.Fatalf("%s: esperado 405, got %d", method, rr.Code)
			}
			var body map[string]any
			json.NewDecoder(rr.Body).Decode(&body)
			if errCode, _ := body["error"].(string); errCode != "METHOD_NOT_ALLOWED" {
				t.Errorf("%s: error code esperado METHOD_NOT_ALLOWED, got %q", method, errCode)
			}
		})
	}
}

func TestV1Health_VersionFromConfig(t *testing.T) {
	h := NewV1HealthHandler("v9.9.9", 1000)
	req := httptest.NewRequest(http.MethodGet, "/api/service/v1/health", nil)
	rr := httptest.NewRecorder()

	h.GetHealth(rr, req)

	var body map[string]any
	json.NewDecoder(rr.Body).Decode(&body)
	if ver, _ := body["version"].(string); ver != "v9.9.9" {
		t.Errorf("versión esperada v9.9.9, got %q", ver)
	}
}

func TestV1Health_RateLimit429(t *testing.T) {
	h := NewV1HealthHandler("dev", 2)
	// Inyectar reloj fijo para que la ventana no avance
	fixed := time.Now()
	h.nowFunc = func() time.Time { return fixed }

	makeReq := func() int {
		req := httptest.NewRequest(http.MethodGet, "/api/service/v1/health", nil)
		req.RemoteAddr = "1.2.3.4:5678"
		rr := httptest.NewRecorder()
		h.GetHealth(rr, req)
		return rr.Code
	}

	if code := makeReq(); code != http.StatusOK {
		t.Fatalf("request 1: esperado 200, got %d", code)
	}
	if code := makeReq(); code != http.StatusOK {
		t.Fatalf("request 2: esperado 200, got %d", code)
	}
	if code := makeReq(); code != http.StatusTooManyRequests {
		t.Fatalf("request 3: esperado 429, got %d", code)
	}
}

func TestV1Health_RateLimitRetryAfterHeader(t *testing.T) {
	h := NewV1HealthHandler("dev", 1)
	fixed := time.Now()
	h.nowFunc = func() time.Time { return fixed }

	makeReq := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/service/v1/health", nil)
		req.RemoteAddr = "9.9.9.9:1234"
		rr := httptest.NewRecorder()
		h.GetHealth(rr, req)
		return rr
	}

	makeReq() // consume el único slot
	rr := makeReq()

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("esperado 429, got %d", rr.Code)
	}
	if ra := rr.Header().Get("Retry-After"); ra != "60" {
		t.Errorf("Retry-After esperado '60', got %q", ra)
	}
}

func TestV1Health_RateLimitWindowReset(t *testing.T) {
	h := NewV1HealthHandler("dev", 1)
	start := time.Now()
	h.nowFunc = func() time.Time { return start }

	req := func() int {
		r := httptest.NewRequest(http.MethodGet, "/api/service/v1/health", nil)
		r.RemoteAddr = "5.5.5.5:80"
		rr := httptest.NewRecorder()
		h.GetHealth(rr, r)
		return rr.Code
	}

	if c := req(); c != 200 {
		t.Fatalf("primero: esperado 200, got %d", c)
	}
	if c := req(); c != 429 {
		t.Fatalf("segundo (mismo window): esperado 429, got %d", c)
	}

	// Avanzar el reloj más de 60s → ventana nueva
	advanced := start.Add(61 * time.Second)
	h.nowFunc = func() time.Time { return advanced }

	if c := req(); c != 200 {
		t.Fatalf("tras reset de ventana: esperado 200, got %d", c)
	}
}

func TestV1Health_NoSensitiveInfoInResponse(t *testing.T) {
	h := NewV1HealthHandler("dev", 1000)
	req := httptest.NewRequest(http.MethodGet, "/api/service/v1/health", nil)
	rr := httptest.NewRecorder()
	h.GetHealth(rr, req)

	var body map[string]any
	json.NewDecoder(rr.Body).Decode(&body)

	forbidden := []string{"db_ok", "uptime", "whatsapp_status", "host", "port", "empresa"}
	for _, key := range forbidden {
		if _, exists := body[key]; exists {
			t.Errorf("campo sensible %q no debe aparecer en la respuesta", key)
		}
	}
	// Solo deben existir los 4 campos esperados
	expected := map[string]bool{"ok": true, "service": true, "version": true, "timestamp": true}
	for key := range body {
		if !expected[key] {
			t.Errorf("campo inesperado %q en la respuesta", key)
		}
	}
}
