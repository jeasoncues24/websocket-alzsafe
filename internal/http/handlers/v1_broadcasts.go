package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
	"wsapi/internal/whatsapp"
)

type V1BroadcastsHandler struct {
	broadcastStore  *storage.BroadcastStore
	telefonoStore   *storage.TelefonoStore
	broadcastWorker *whatsapp.BroadcastWorker
}

func NewV1BroadcastsHandler(broadcastStore *storage.BroadcastStore, telefonoStore *storage.TelefonoStore, broadcastWorker *whatsapp.BroadcastWorker) *V1BroadcastsHandler {
	return &V1BroadcastsHandler{
		broadcastStore:  broadcastStore,
		telefonoStore:   telefonoStore,
		broadcastWorker: broadcastWorker,
	}
}

func (h *V1BroadcastsHandler) GetBroadcasts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	apiClaims, ok := domain.GetApiKeyClaims(r.Context())
	if !ok {
		writeV1Error(w, http.StatusUnauthorized, "API_KEY_REQUIRED", "API key requerida")
		return
	}

	jobs := h.broadcastStore.ListByEmpresa(apiClaims.EmpresaID)

	result := make([]map[string]interface{}, 0, len(jobs))
	for _, job := range jobs {
		if job.TelefonoID != apiClaims.TelefonoID {
			continue
		}
		result = append(result, map[string]interface{}{
			"reference_id": job.ReferenceID,
			"telefono_id":  job.TelefonoID,
			"total":        job.Total,
			"adjuntos":     job.Adjuntos,
			"status":       job.Status,
			"created_at":   job.CreatedAt,
		})
	}

	writeV1Success(w, map[string]interface{}{
		"broadcasts": result,
		"total":      len(result),
	}, apiClaims.EmpresaID)
}

func (h *V1BroadcastsHandler) GetBroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	apiClaims, ok := domain.GetApiKeyClaims(r.Context())
	if !ok {
		writeV1Error(w, http.StatusUnauthorized, "API_KEY_REQUIRED", "API key requerida")
		return
	}

	refID := extractBroadcastID(r)
	if refID == "" {
		writeV1Error(w, http.StatusBadRequest, "MISSING_BROADCAST_ID", "broadcast_id requerido")
		return
	}

	job, ok := h.broadcastStore.Get(refID)
	if !ok || job == nil {
		writeV1Error(w, http.StatusNotFound, "BROADCAST_NOT_FOUND", "Difusión no encontrada")
		return
	}

	if job.TelefonoID != apiClaims.TelefonoID {
		writeV1Error(w, http.StatusForbidden, "FORBIDDEN", "La difusión no pertenece a esta API key")
		return
	}

	writeV1Success(w, map[string]interface{}{
		"reference_id": job.ReferenceID,
		"empresa_id":   job.EmpresaID,
		"telefono_id":  job.TelefonoID,
		"total":        job.Total,
		"adjuntos":     job.Adjuntos,
		"status":       job.Status,
		"results":      job.Results,
		"created_at":   job.CreatedAt,
	}, apiClaims.EmpresaID)
}

func (h *V1BroadcastsHandler) PostBroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	apiClaims, ok := domain.GetApiKeyClaims(r.Context())
	if !ok {
		writeV1Error(w, http.StatusUnauthorized, "API_KEY_REQUIRED", "API key requerida")
		return
	}

	var req struct {
		Destinos []string                   `json:"destinos"`
		Mensaje  string                     `json:"mensaje"`
		Adjuntos []domain.AttachmentPayload `json:"adjuntos,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeV1Error(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	if len(req.Destinos) == 0 {
		writeV1Error(w, http.StatusBadRequest, "MISSING_FIELDS", "destinos son requeridos")
		return
	}

	items := make([]domain.BroadcastItem, len(req.Destinos))
	for i, d := range req.Destinos {
		items[i] = domain.BroadcastItem{Destino: d, Mensaje: req.Mensaje}
	}

	broadcastReq := domain.BroadcastRequest{
		TelefonoID:    apiClaims.TelefonoID,
		Adjuntos:      req.Adjuntos,
		ListaDifusion: items,
	}
	if err := domain.ValidateBroadcastRequest(&broadcastReq); err != nil {
		writeV1Error(w, http.StatusBadRequest, err.Code, err.Message)
		return
	}

	phone, err := h.telefonoStore.GetByID(apiClaims.TelefonoID)
	if err != nil || phone == nil {
		writeV1Error(w, http.StatusNotFound, "TELEFONO_NOT_FOUND", "Teléfono no encontrado")
		return
	}

	if phone.Status != domain.TelefonoStatusActive {
		writeV1Error(w, http.StatusBadRequest, "SESSION_NOT_ACTIVE", "El teléfono no está activo")
		return
	}

	refID := uuid.New().String()

	job := &domain.BroadcastJob{
		ReferenceID: refID,
		EmpresaID:   apiClaims.EmpresaID,
		TelefonoID:  apiClaims.TelefonoID,
		Adjuntos:    nil,
		Total:       len(items),
		Status:      domain.BroadcastStatusPending,
		CreatedAt:   time.Now(),
	}
	infos, infoErr := buildAttachmentInfos(req.Adjuntos)
	if infoErr != nil {
		writeV1Error(w, http.StatusBadRequest, "INVALID_ATTACHMENT", "Adjunto inválido")
		return
	}
	job.Adjuntos = infos
	h.broadcastStore.Create(job)

	// Build items and submit to worker for async real sending
	workerJob := whatsapp.BroadcastJob{
		ReferenceID: refID,
		RUCEmpresa:  phone.NumeroCompleto,
		AccountID:   phone.NumeroCompleto,
		Attachments: req.Adjuntos,
		Items:       items,
		ResultChan:  make(chan whatsapp.BroadcastResult, len(items)+1),
	}
	h.broadcastWorker.SubmitAsync(workerJob)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": true,
		"data": map[string]interface{}{
			"reference_id": refID,
			"total":        len(req.Destinos),
			"estado":       string(domain.BroadcastStatusPending),
		},
		"meta": map[string]interface{}{
			"empresa_id": apiClaims.EmpresaID,
		},
	})
}

func extractBroadcastID(r *http.Request) string {
	path := r.URL.Path
	segments := splitPathSegments(path)
	for i := len(segments) - 1; i >= 0; i-- {
		if segments[i] != "" && segments[i] != "broadcast" && segments[i] != "difusiones" {
			return segments[i]
		}
	}
	return ""
}

func splitPathSegments(path string) []string {
	if path == "" {
		return nil
	}
	path = stripPrefix(path, "/v1/broadcast/")
	segs := splitSimple(path, "/")
	parts := make([]string, 0, len(segs))
	for _, p := range segs {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func stripPrefix(s, prefix string) string {
	if len(prefix) > 0 && len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}

func splitSimple(s, sep string) []string {
	if s == "" {
		return nil
	}
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			if i > start {
				result = append(result, s[start:i])
			}
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}
