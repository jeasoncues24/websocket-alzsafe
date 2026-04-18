package http

import (
	"net/http"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
)

type V1MetricsHandler struct {
	msgRepo       storage.MessagesRepository
	telefonoStore *storage.TelefonoStore
}

func NewV1MetricsHandler(msgRepo storage.MessagesRepository, telefonoStore *storage.TelefonoStore) *V1MetricsHandler {
	return &V1MetricsHandler{
		msgRepo:       msgRepo,
		telefonoStore: telefonoStore,
	}
}

func (h *V1MetricsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	claims, ok := domain.GetEmpresaJWTClaims(r.Context())
	if !ok {
		writeV1Error(w, http.StatusUnauthorized, "TOKEN_REQUIRED", "JWT de empresa requerido")
		return
	}

	metrics, _ := h.msgRepo.GetMessageMetricsByEmpresa(claims.EmpresaID)

	phones, _ := h.telefonoStore.GetByEmpresa(claims.EmpresaID)
	activePhones := 0
	for _, p := range phones {
		if p.Status == domain.TelefonoStatusActive {
			activePhones++
		}
	}

	writeV1Success(w, map[string]interface{}{
		"messages_sent":   metrics.MensajesExitosos,
		"messages_failed": metrics.MensajesFallidos,
		"messages_total":  metrics.TotalMensajes,
		"messages_today":  metrics.MensajesHoy,
		"success_rate":    calculateSuccessRate(metrics.MensajesExitosos, metrics.TotalMensajes),
		"active_phones":   activePhones,
		"total_phones":    len(phones),
	}, claims.EmpresaID)
}

func calculateSuccessRate(success, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(success) / float64(total) * 100
}
