package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
)

type V1BroadcastsHandler struct {
	broadcastStore *storage.BroadcastStore
	telefonoStore  *storage.TelefonoStore
}

func NewV1BroadcastsHandler(broadcastStore *storage.BroadcastStore, telefonoStore *storage.TelefonoStore) *V1BroadcastsHandler {
	return &V1BroadcastsHandler{
		broadcastStore: broadcastStore,
		telefonoStore:  telefonoStore,
	}
}

func (h *V1BroadcastsHandler) GetBroadcasts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	claims, ok := domain.GetEmpresaJWTClaims(r.Context())
	apiClaims := (*domain.ApiKeyClaims)(nil)
	if !ok {
		if keyClaims, ok2 := domain.GetApiKeyClaims(r.Context()); ok2 {
			apiClaims = keyClaims
			claims = &domain.EmpresaJWTClaims{EmpresaID: keyClaims.EmpresaID, Permissions: keyClaims.Scopes}
		} else {
			writeV1Error(w, http.StatusUnauthorized, "TOKEN_REQUIRED", "JWT de empresa requerido")
			return
		}
	}

	jobs := h.broadcastStore.ListByEmpresa(claims.EmpresaID)

	result := make([]map[string]interface{}, 0, len(jobs))
	for _, job := range jobs {
		if apiClaims != nil && job.TelefonoID != apiClaims.TelefonoID {
			continue
		}
		result = append(result, map[string]interface{}{
			"reference_id": job.ReferenceID,
			"telefono_id":  job.TelefonoID,
			"total":        job.Total,
			"status":       job.Status,
			"created_at":   job.CreatedAt,
		})
	}

	writeV1Success(w, map[string]interface{}{
		"broadcasts": result,
		"total":      len(result),
	}, claims.EmpresaID)
}

func (h *V1BroadcastsHandler) GetBroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	claims, ok := domain.GetEmpresaJWTClaims(r.Context())
	if !ok {
		if keyClaims, ok2 := domain.GetApiKeyClaims(r.Context()); ok2 {
			claims = &domain.EmpresaJWTClaims{EmpresaID: keyClaims.EmpresaID, Permissions: keyClaims.Scopes}
		} else {
			writeV1Error(w, http.StatusUnauthorized, "TOKEN_REQUIRED", "JWT de empresa requerido")
			return
		}
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

	writeV1Success(w, map[string]interface{}{
		"reference_id": job.ReferenceID,
		"empresa_id":   job.EmpresaID,
		"telefono_id":  job.TelefonoID,
		"total":        job.Total,
		"status":       job.Status,
		"results":      job.Results,
		"created_at":   job.CreatedAt,
	}, claims.EmpresaID)
}

func (h *V1BroadcastsHandler) PostBroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	claims, ok := domain.GetEmpresaJWTClaims(r.Context())
	apiClaims := (*domain.ApiKeyClaims)(nil)
	if !ok {
		if keyClaims, ok2 := domain.GetApiKeyClaims(r.Context()); ok2 {
			apiClaims = keyClaims
			claims = &domain.EmpresaJWTClaims{EmpresaID: keyClaims.EmpresaID, Permissions: keyClaims.Scopes}
		} else {
			writeV1Error(w, http.StatusUnauthorized, "TOKEN_REQUIRED", "JWT de empresa requerido")
			return
		}
	}

	var req struct {
		TelefonoID int64    `json:"telefono_id"`
		Destinos   []string `json:"destinos"`
		Mensaje    string   `json:"mensaje"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeV1Error(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	if apiClaims != nil {
		if req.TelefonoID == 0 {
			req.TelefonoID = apiClaims.TelefonoID
		} else if req.TelefonoID != apiClaims.TelefonoID {
			writeV1Error(w, http.StatusForbidden, "FORBIDDEN", "La API key solo puede usarse con su teléfono asignado")
			return
		}
	}

	if req.TelefonoID == 0 || len(req.Destinos) == 0 || req.Mensaje == "" {
		writeV1Error(w, http.StatusBadRequest, "MISSING_FIELDS", "telefono_id, destinos y mensaje requeridos")
		return
	}

	belongs, _ := h.telefonoStore.BelongsToEmpresa(req.TelefonoID, claims.EmpresaID)
	if !belongs {
		writeV1Error(w, http.StatusForbidden, "FORBIDDEN", "El teléfono no pertenece a esta empresa")
		return
	}

	refID := "BC_" + strconv.FormatInt(claims.EmpresaID, 10) + "_" + strconv.FormatInt(int64(len(req.Destinos)), 10)

	job := &domain.BroadcastJob{
		ReferenceID: refID,
		EmpresaID:   claims.EmpresaID,
		TelefonoID:  req.TelefonoID,
		Total:       len(req.Destinos),
		Status:      domain.BroadcastStatusPending,
	}
	h.broadcastStore.Create(job)

	writeV1Success(w, map[string]interface{}{
		"reference_id": refID,
		"total":        len(req.Destinos),
		"status":       "pending",
	}, claims.EmpresaID)
}

func extractBroadcastID(r *http.Request) string {
	path := r.URL.Path
	segments := splitPathSegments(path)
	for i := len(segments) - 1; i >= 0; i-- {
		if segments[i] != "" && segments[i] != "broadcast" {
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
