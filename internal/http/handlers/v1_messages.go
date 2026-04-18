package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
)

type V1MessagesHandler struct {
	msgRepo       storage.MessagesRepository
	telefonoStore *storage.TelefonoStore
}

func NewV1MessagesHandler(msgRepo storage.MessagesRepository, telefonoStore *storage.TelefonoStore) *V1MessagesHandler {
	return &V1MessagesHandler{
		msgRepo:       msgRepo,
		telefonoStore: telefonoStore,
	}
}

func (h *V1MessagesHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
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

	telefonoIDStr := r.URL.Query().Get("telefono_id")
	var telefonoID int64
	if telefonoIDStr != "" {
		var err error
		telefonoID, err = strconv.ParseInt(telefonoIDStr, 10, 64)
		if err != nil {
			writeV1Error(w, http.StatusBadRequest, "INVALID_TELEFONO_ID", "telefono_id inválido")
			return
		}
		if apiClaims != nil && telefonoID != apiClaims.TelefonoID {
			writeV1Error(w, http.StatusForbidden, "FORBIDDEN", "La API key solo puede usarse con su teléfono asignado")
			return
		}
		belongs, _ := h.telefonoStore.BelongsToEmpresa(telefonoID, claims.EmpresaID)
		if !belongs {
			writeV1Error(w, http.StatusForbidden, "FORBIDDEN", "El teléfono no pertenece a esta empresa")
			return
		}
	} else if apiClaims != nil {
		telefonoID = apiClaims.TelefonoID
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	messages, _, err := h.msgRepo.GetByEmpresa(claims.EmpresaID, "", "", limit, 0)
	if err != nil {
		writeV1Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al obtener mensajes")
		return
	}

	result := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		if telefonoID == 0 || msg.TelefonoID == telefonoID {
			result = append(result, map[string]interface{}{
				"id":          msg.ID,
				"telefono_id": msg.TelefonoID,
				"destino":     msg.Destino,
				"contenido":   msg.Contenido,
				"estado":      msg.Estado,
				"tiempo":      msg.TiempoEnvio,
			})
		}
	}

	writeV1Success(w, map[string]interface{}{
		"messages": result,
		"total":    len(result),
	}, claims.EmpresaID)
}

func (h *V1MessagesHandler) PostMessage(w http.ResponseWriter, r *http.Request) {
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
		TelefonoID int64  `json:"telefono_id"`
		Destino    string `json:"destino"`
		Contenido  string `json:"contenido"`
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

	if req.TelefonoID == 0 || req.Destino == "" || req.Contenido == "" {
		writeV1Error(w, http.StatusBadRequest, "MISSING_FIELDS", "telefono_id, destino y contenido requeridos")
		return
	}

	belongs, _ := h.telefonoStore.BelongsToEmpresa(req.TelefonoID, claims.EmpresaID)
	if !belongs {
		writeV1Error(w, http.StatusForbidden, "FORBIDDEN", "El teléfono no pertenece a esta empresa")
		return
	}

	phone, err := h.telefonoStore.GetByID(req.TelefonoID)
	if err != nil || phone == nil {
		writeV1Error(w, http.StatusNotFound, "TELEFONO_NOT_FOUND", "Teléfono no encontrado")
		return
	}

	if phone.Status != domain.TelefonoStatusActive {
		writeV1Error(w, http.StatusBadRequest, "SESSION_NOT_ACTIVE", "El teléfono no está activo")
		return
	}

	msg := &domain.Message{
		EmpresaID:   claims.EmpresaID,
		TelefonoID:  req.TelefonoID,
		Destino:     req.Destino,
		Contenido:   req.Contenido,
		Estado:      "pending",
		TiempoEnvio: time.Now(),
	}
	_ = h.msgRepo.Create(msg)

	writeV1Success(w, map[string]interface{}{
		"status": "sent",
	}, claims.EmpresaID)
}
