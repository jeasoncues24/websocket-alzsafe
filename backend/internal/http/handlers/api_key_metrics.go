package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
)

type ApiKeyMetricsHandler struct {
	telemetryStore *storage.TelemetryStore
}

func NewApiKeyMetricsHandler(telemetryStore *storage.TelemetryStore) *ApiKeyMetricsHandler {
	return &ApiKeyMetricsHandler{telemetryStore: telemetryStore}
}

func (h *ApiKeyMetricsHandler) UsageStats(w http.ResponseWriter, r *http.Request) {
	apiKeyID, ok := extractIDFromPath(r.URL.Path, "/api/admin/api-keys/", "/usage/stats")
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	desde, hasta := parseDateRange(r)
	stats, err := h.telemetryStore.GetUsageStats(apiKeyID, desde, hasta)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al obtener estadísticas de uso")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.TelemetryMetricsResponse{
		OK:    true,
		Stats: stats,
	})
}

func (h *ApiKeyMetricsHandler) UsageTimeSeries(w http.ResponseWriter, r *http.Request) {
	apiKeyID, ok := extractIDFromPath(r.URL.Path, "/api/admin/api-keys/", "/usage/timeseries")
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	desde, hasta := parseDateRange(r)
	granularidad := r.URL.Query().Get("granularidad")
	if granularidad == "" {
		granularidad = "daily"
	}

	series, err := h.telemetryStore.GetUsageTimeSeries(apiKeyID, desde, hasta, granularidad)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al obtener serie temporal")
		return
	}
	if series == nil {
		series = []domain.TelemetryTimeSeriesPoint{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.TelemetryMetricsResponse{
		OK:     true,
		Series: series,
	})
}

func (h *ApiKeyMetricsHandler) AuditStats(w http.ResponseWriter, r *http.Request) {
	apiKeyID, ok := extractIDFromPath(r.URL.Path, "/api/admin/api-keys/", "/audit/stats")
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	stats, err := h.telemetryStore.GetAuditStats(apiKeyID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al obtener estadísticas de auditoría")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.TelemetryAuditResponse{
		OK:    true,
		Stats: stats,
	})
}

func extractIDFromPath(path, prefix, suffix string) (int64, bool) {
	path = strings.TrimSuffix(path, "/")
	path = strings.TrimPrefix(path, prefix)
	path = strings.TrimSuffix(path, suffix)
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func parseDateRange(r *http.Request) (time.Time, time.Time) {
	now := time.Now()
	hasta := now
	desde := now.AddDate(0, 0, -30)

	if v := strings.TrimSpace(r.URL.Query().Get("desde")); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			desde = t
		}
	}
	if v := strings.TrimSpace(r.URL.Query().Get("hasta")); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			hasta = t
		}
	}

	return desde, hasta
}
