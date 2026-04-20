package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
	"wsapi/internal/whatsapp"
)

type V1MessagesHandler struct {
	msgRepo       storage.MessagesRepository
	telefonoStore *storage.TelefonoStore
	manager       *whatsapp.Manager
}

func NewV1MessagesHandler(msgRepo storage.MessagesRepository, telefonoStore *storage.TelefonoStore, manager *whatsapp.Manager) *V1MessagesHandler {
	return &V1MessagesHandler{
		msgRepo:       msgRepo,
		telefonoStore: telefonoStore,
		manager:       manager,
	}
}

func (h *V1MessagesHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	apiClaims, ok := domain.GetApiKeyClaims(r.Context())
	if !ok {
		writeV1Error(w, http.StatusUnauthorized, "API_KEY_REQUIRED", "API key requerida")
		return
	}

	telefonoIDStr := r.URL.Query().Get("telefono_id")
	if telefonoIDStr != "" {
		tid, err := strconv.ParseInt(telefonoIDStr, 10, 64)
		if err != nil {
			writeV1Error(w, http.StatusBadRequest, "INVALID_TELEFONO_ID", "telefono_id inválido")
			return
		}
		if tid != apiClaims.TelefonoID {
			writeV1Error(w, http.StatusForbidden, "FORBIDDEN", "La API key solo puede usarse con su teléfono asignado")
			return
		}
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	messages, _, err := h.msgRepo.GetByEmpresa(apiClaims.EmpresaID, "", "", limit, 0)
	if err != nil {
		writeV1Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al obtener mensajes")
		return
	}

	result := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		if msg.TelefonoID == apiClaims.TelefonoID {
			result = append(result, map[string]interface{}{
				"reference_id": msg.ReferenceID,
				"telefono_id":  msg.TelefonoID,
				"destino":      msg.Destino,
				"contenido":    msg.Contenido,
				"estado":       msg.Estado,
				"tiempo":       msg.TiempoEnvio,
			})
		}
	}

	writeV1Success(w, map[string]interface{}{
		"messages": result,
		"total":    len(result),
	}, apiClaims.EmpresaID)
}

func (h *V1MessagesHandler) PostMessage(w http.ResponseWriter, r *http.Request) {
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
		Destino   string `json:"destino"`
		Contenido string `json:"contenido"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeV1Error(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	if req.Destino == "" || req.Contenido == "" {
		writeV1Error(w, http.StatusBadRequest, "MISSING_FIELDS", "destino y contenido son requeridos")
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

	msg := domain.NewMessage(apiClaims.EmpresaID, apiClaims.TelefonoID, req.Destino, req.Contenido)

	if err := h.msgRepo.Create(msg); err != nil {
		writeV1Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al registrar el mensaje")
		return
	}

	sendErr := whatsapp.SendTextMessage(r.Context(), h.manager, phone.NumeroCompleto, req.Destino, req.Contenido)

	if sendErr != nil {
		_ = h.msgRepo.UpdateEstado(msg.ReferenceID, domain.MessageStateFailed, sendErr.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok": false,
			"data": map[string]interface{}{
				"reference_id": msg.ReferenceID,
				"estado":       string(domain.MessageStateFailed),
				"error":        sendErr.Error(),
			},
			"meta": map[string]interface{}{
				"empresa_id": apiClaims.EmpresaID,
			},
		})
		return
	}

	_ = h.msgRepo.UpdateEstado(msg.ReferenceID, domain.MessageStateSent, "")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": true,
		"data": map[string]interface{}{
			"reference_id": msg.ReferenceID,
			"estado":       string(domain.MessageStateSent),
		},
		"meta": map[string]interface{}{
			"empresa_id": apiClaims.EmpresaID,
		},
	})
}

func (h *V1MessagesHandler) GetMessageByReference(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	apiClaims, ok := domain.GetApiKeyClaims(r.Context())
	if !ok {
		writeV1Error(w, http.StatusUnauthorized, "API_KEY_REQUIRED", "API key requerida")
		return
	}

	referenceID := extractReferenceID(r.URL.Path)
	if referenceID == "" {
		writeV1Error(w, http.StatusBadRequest, "MISSING_REFERENCE_ID", "Reference ID requerido")
		return
	}

	msg, err := h.msgRepo.GetByReferenceID(referenceID)
	if err != nil {
		writeV1Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al obtener mensaje")
		return
	}
	if msg == nil {
		writeV1Error(w, http.StatusNotFound, "MESSAGE_NOT_FOUND", "Mensaje no encontrado")
		return
	}

	if msg.TelefonoID != apiClaims.TelefonoID || msg.EmpresaID != apiClaims.EmpresaID {
		writeV1Error(w, http.StatusForbidden, "FORBIDDEN", "No tienes acceso a este mensaje")
		return
	}

	writeV1Success(w, map[string]interface{}{
		"message": map[string]interface{}{
			"reference_id":   msg.ReferenceID,
			"telefono_id":   msg.TelefonoID,
			"destino":       msg.Destino,
			"contenido":     msg.Contenido,
			"estado":        string(msg.Estado),
			"error_reason":  msg.ErrorReason,
			"retry_count":   msg.RetryCount,
			"created_at":    msg.TiempoEnvio,
			"timestamp_sent": msg.TimestampSent,
			"last_attempt":  msg.LastAttemptAt,
		},
	}, apiClaims.EmpresaID)
}

func (h *V1MessagesHandler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	apiClaims, ok := domain.GetApiKeyClaims(r.Context())
	if !ok {
		writeV1Error(w, http.StatusUnauthorized, "API_KEY_REQUIRED", "API key requerida")
		return
	}

	referenceID := extractReferenceID(r.URL.Path)
	if referenceID == "" {
		writeV1Error(w, http.StatusBadRequest, "MISSING_REFERENCE_ID", "Reference ID requerido")
		return
	}

	msg, err := h.msgRepo.GetByReferenceID(referenceID)
	if err != nil || msg == nil {
		writeV1Error(w, http.StatusNotFound, "MESSAGE_NOT_FOUND", "Mensaje no encontrado")
		return
	}

	if msg.TelefonoID != apiClaims.TelefonoID || msg.EmpresaID != apiClaims.EmpresaID {
		writeV1Error(w, http.StatusForbidden, "FORBIDDEN", "No tienes acceso a este mensaje")
		return
	}

	if msg.Estado == domain.MessageStateSent || msg.Estado == domain.MessageStateDelivered {
		writeV1Error(w, http.StatusBadRequest, "INVALID_STATE", "No se puede editar un mensaje ya enviado")
		return
	}

	var req struct {
		Contenido string `json:"contenido"`
		Destino   string `json:"destino"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeV1Error(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	if req.Contenido == "" && req.Destino == "" {
		writeV1Error(w, http.StatusBadRequest, "MISSING_FIELDS", "Se requiere contenido o destino")
		return
	}

	if req.Contenido != "" {
		if err := h.msgRepo.UpdateContenido(referenceID, req.Contenido); err != nil {
			writeV1Error(w, http.StatusInternalServerError, "UPDATE_ERROR", "Error al actualizar mensaje")
			return
		}
	}

	writeV1Success(w, map[string]interface{}{
		"ok":          true,
		"reference_id": referenceID,
		"message":      "Mensaje actualizado",
	}, apiClaims.EmpresaID)
}

func (h *V1MessagesHandler) RetryMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	apiClaims, ok := domain.GetApiKeyClaims(r.Context())
	if !ok {
		writeV1Error(w, http.StatusUnauthorized, "API_KEY_REQUIRED", "API key requerida")
		return
	}

	referenceID := extractReferenceID(r.URL.Path)
	if referenceID == "" {
		writeV1Error(w, http.StatusBadRequest, "MISSING_REFERENCE_ID", "Reference ID requerido")
		return
	}

	msg, err := h.msgRepo.GetByReferenceID(referenceID)
	if err != nil || msg == nil {
		writeV1Error(w, http.StatusNotFound, "MESSAGE_NOT_FOUND", "Mensaje no encontrado")
		return
	}

	if msg.TelefonoID != apiClaims.TelefonoID || msg.EmpresaID != apiClaims.EmpresaID {
		writeV1Error(w, http.StatusForbidden, "FORBIDDEN", "No tienes acceso a este mensaje")
		return
	}

	if msg.Estado == domain.MessageStateSent || msg.Estado == domain.MessageStateDelivered {
		writeV1Error(w, http.StatusBadRequest, "INVALID_STATE", "El mensaje ya fue enviado")
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

	if err := h.msgRepo.IncrementRetryCount(referenceID); err != nil {
		writeV1Error(w, http.StatusInternalServerError, "RETRY_ERROR", "Error al preparar reintento")
		return
	}

	sendErr := whatsapp.SendTextMessage(r.Context(), h.manager, phone.NumeroCompleto, msg.Destino, msg.Contenido)

	if sendErr != nil {
		_ = h.msgRepo.UpdateEstado(referenceID, domain.MessageStateFailed, sendErr.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok": false,
			"data": map[string]interface{}{
				"reference_id": referenceID,
				"estado":       string(domain.MessageStateFailed),
				"error":         sendErr.Error(),
			},
			"meta": map[string]interface{}{
				"empresa_id": apiClaims.EmpresaID,
			},
		})
		return
	}

	_ = h.msgRepo.UpdateEstado(referenceID, domain.MessageStateSent, "")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": true,
		"data": map[string]interface{}{
			"reference_id": referenceID,
			"estado":       string(domain.MessageStateSent),
		},
		"meta": map[string]interface{}{
			"empresa_id": apiClaims.EmpresaID,
		},
	})
}

func extractReferenceID(path string) string {
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	for i, p := range parts {
		if p == "mensajes" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
