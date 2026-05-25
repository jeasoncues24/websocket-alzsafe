package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
	"wsapi/internal/whatsapp"
)

type V1BroadcastsHandler struct {
	jobRepo         storage.JobQueueRepository
	telefonoStore   *storage.TelefonoStore
	broadcastWorker *whatsapp.BroadcastWorker
}

func NewV1BroadcastsHandler(jobRepo storage.JobQueueRepository, telefonoStore *storage.TelefonoStore, broadcastWorker *whatsapp.BroadcastWorker) *V1BroadcastsHandler {
	return &V1BroadcastsHandler{
		jobRepo:         jobRepo,
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

	if h.jobRepo == nil {
		writeV1Success(w, map[string]interface{}{"broadcasts": []interface{}{}, "total": 0}, apiClaims.EmpresaID)
		return
	}

	jobs, err := h.jobRepo.ListByEmpresa(r.Context(), apiClaims.EmpresaID)
	if err != nil {
		writeV1Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al obtener difusiones")
		return
	}

	result := make([]map[string]interface{}, 0, len(jobs))
	for _, job := range jobs {
		if job.Type != domain.JobTypeBroadcast {
			continue
		}
		result = append(result, map[string]interface{}{
			"reference_id": job.EntityID,
			"total":        0, // se actualiza al consultar items si se necesita
			"status":       string(job.Status),
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

	if h.jobRepo == nil {
		writeV1Error(w, http.StatusNotFound, "BROADCAST_NOT_FOUND", "Difusión no encontrada")
		return
	}

	job, err := h.jobRepo.GetByEntityID(r.Context(), refID)
	if err != nil || job == nil {
		writeV1Error(w, http.StatusNotFound, "BROADCAST_NOT_FOUND", "Difusión no encontrada")
		return
	}

	// Verificar que pertenece al teléfono de esta API key (via empresa)
	if job.EmpresaID != apiClaims.EmpresaID {
		writeV1Error(w, http.StatusForbidden, "FORBIDDEN", "La difusión no pertenece a esta API key")
		return
	}

	items, err := h.jobRepo.GetAllItems(r.Context(), job.ID)
	if err != nil {
		writeV1Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al obtener resultados")
		return
	}

	results := make([]map[string]interface{}, 0, len(items))
	resultItems := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		destino, _, _ := storage.DecodeBroadcastPayload(item.Payload)
		errVal := interface{}(nil)
		if item.ErrorText != "" {
			errVal = item.ErrorText
		}
		results = append(results, map[string]interface{}{
			"destino": destino,
			"ok":      item.Status == domain.JobItemSent,
			"error":   errVal,
		})

		var processedAtStr *string
		if item.ProcessedAt != nil {
			s := item.ProcessedAt.Format(time.RFC3339)
			processedAtStr = &s
		}
		resultItems = append(resultItems, map[string]interface{}{
			"id":             item.ID,
			"sequence_order": item.SequenceOrder,
			"destino":        destino,
			"status":         string(item.Status),
			"error_text":     errVal,
			"processed_at":   processedAtStr,
		})
	}

	writeV1Success(w, map[string]interface{}{
		"reference_id": job.EntityID,
		"empresa_id":   job.EmpresaID,
		"total":        len(items),
		"status":       string(job.Status),
		"results":      results,
		"items":        resultItems,
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
	if len(req.Destinos) > domain.MaxBroadcastItems {
		writeV1Error(w, http.StatusBadRequest, "MAX_BROADCAST_EXCEEDED",
			"destinos supera el límite de 30")
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

	// Calcular tiempo estimado antes de encolar
	estimatedSeconds := domain.EstimateBroadcastSeconds(len(items), h.broadcastWorker.Config().TimingConfig())
	_ = estimatedSeconds

	// Persistir en DB si hay repositorio
	var jobID int64
	var itemIDs []int64

	if h.jobRepo != nil {
		jobDomain := &domain.Job{
			Type:        domain.JobTypeBroadcast,
			EntityID:    refID,
			Priority:    5,
			EmpresaID:   apiClaims.EmpresaID,
			MaxAttempts: 1,
		}
		jobItems := make([]domain.JobItem, len(items))
		for i, item := range items {
			payload, _ := storage.EncodeBroadcastPayload(item.Destino, item.Mensaje)
			jobItems[i] = domain.JobItem{
				SequenceOrder: i,
				Payload:       payload,
			}
		}
		if err := h.jobRepo.CreateJobWithItems(context.Background(), jobDomain, jobItems); err != nil {
			writeV1Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Error al registrar la difusión en la cola de tareas")
			return
		}
		jobID = jobDomain.ID
		// Recuperar IDs de items para que el worker pueda actualizar su estado
		dbItems, fetchErr := h.jobRepo.GetAllItems(context.Background(), jobID)
		if fetchErr != nil {
			writeV1Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Error al recuperar detalles de la difusión")
			return
		}
		itemIDs = make([]int64, len(dbItems))
		for i, it := range dbItems {
			itemIDs[i] = it.ID
		}
	}

	infos, infoErr := buildAttachmentInfos(req.Adjuntos)
	if infoErr != nil {
		writeV1Error(w, http.StatusBadRequest, "INVALID_ATTACHMENT", "Adjunto inválido")
		return
	}
	_ = infos // usados en la respuesta GET si se amplía

	workerJob := whatsapp.BroadcastJob{
		ReferenceID: refID,
		RUCEmpresa:  phone.NumeroCompleto,
		AccountID:   phone.NumeroCompleto,
		JobID:       jobID,
		Attachments: req.Adjuntos,
		Items:       items,
		ItemIDs:     itemIDs,
	}
	h.broadcastWorker.SubmitAsync(workerJob)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": true,
		"data": map[string]interface{}{
			"reference_id":      refID,
			"total":             len(req.Destinos),
			"estado":            string(domain.JobStatusPending),
			"estimated_seconds": estimatedSeconds,
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

