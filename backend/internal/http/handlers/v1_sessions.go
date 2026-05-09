package http

import (
	"encoding/json"
	"net/http"
	"time"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
	"wsapi/internal/whatsapp"
)

type V1SessionsHandler struct {
	telefonoStore *storage.TelefonoStore
	sessionStore  *storage.SessionStore
	manager       *whatsapp.Manager
}

func NewV1SessionsHandler(telefonoStore *storage.TelefonoStore, sessionStore *storage.SessionStore, manager *whatsapp.Manager) *V1SessionsHandler {
	return &V1SessionsHandler{
		telefonoStore: telefonoStore,
		sessionStore:  sessionStore,
		manager:       manager,
	}
}

func (h *V1SessionsHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	claims, ok := domain.GetEmpresaJWTClaims(r.Context())
	if !ok {
		writeV1Error(w, http.StatusUnauthorized, "TOKEN_REQUIRED", "JWT de empresa requerido")
		return
	}

	phones, err := h.telefonoStore.GetByEmpresa(claims.EmpresaID)
	if err != nil {
		writeV1Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al obtener teléfonos")
		return
	}

	sessions := make([]map[string]interface{}, 0, len(phones))
	for _, phone := range phones {
		sessions = append(sessions, map[string]interface{}{
			"telefono_id":    phone.ID,
			"numero":         phone.Numero,
			"codigo_pais":    phone.CodigoPais,
			"numeroCompleto": phone.NumeroCompleto,
			"status":         phone.Status,
			"lastConnected":  phone.LastConnected,
		})
	}

	writeV1Success(w, map[string]interface{}{
		"sessions": sessions,
		"total":    len(sessions),
	}, claims.EmpresaID)
}

func (h *V1SessionsHandler) PostSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	claims, ok := domain.GetEmpresaJWTClaims(r.Context())
	if !ok {
		writeV1Error(w, http.StatusUnauthorized, "TOKEN_REQUIRED", "JWT de empresa requerido")
		return
	}

	var req struct {
		CodigoPais string `json:"codigo_pais"`
		Numero     string `json:"numero"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeV1Error(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	if req.CodigoPais == "" || req.Numero == "" {
		writeV1Error(w, http.StatusBadRequest, "MISSING_FIELDS", "codigo_pais y numero requeridos")
		return
	}

	numeroCompleto := req.CodigoPais + req.Numero

	phone := &domain.Telefono{
		EmpresaID:      claims.EmpresaID,
		CodigoPais:     req.CodigoPais,
		Numero:         req.Numero,
		NumeroCompleto: numeroCompleto,
		Status:         domain.TelefonoStatusQRPending,
	}

	id, err := h.telefonoStore.Create(phone)
	if err != nil {
		writeV1Error(w, http.StatusInternalServerError, "CREATE_ERROR", "Error al crear teléfono")
		return
	}

	qrString := ""
	if state, ok := h.sessionStore.Get(numeroCompleto); ok {
		if state.QRString != "" {
			h.telefonoStore.UpdateQRString(id, state.QRString)
			qrString = state.QRString
		}
	}

	writeV1Success(w, map[string]interface{}{
		"telefono_id":    id,
		"numeroCompleto": numeroCompleto,
		"status":         "qr_pending",
		"expires_in":     60,
		"qr_string":      qrString,
	}, claims.EmpresaID)
}

func (h *V1SessionsHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	telefonoID, err := extractTelefonoID(r)
	if err != nil {
		writeV1Error(w, http.StatusBadRequest, "MISSING_TELEFONO_ID", "telefono_id requerido")
		return
	}

	claims, ok := domain.GetEmpresaJWTClaims(r.Context())
	if !ok {
		writeV1Error(w, http.StatusUnauthorized, "TOKEN_REQUIRED", "JWT de empresa requerido")
		return
	}

	belongs, err := h.telefonoStore.BelongsToEmpresa(telefonoID, claims.EmpresaID)
	if err != nil || !belongs {
		writeV1Error(w, http.StatusForbidden, "FORBIDDEN", "El teléfono no pertenece a esta empresa")
		return
	}

	phone, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil || phone == nil {
		writeV1Error(w, http.StatusNotFound, "TELEFONO_NOT_FOUND", "Teléfono no encontrado")
		return
	}

	writeV1Success(w, map[string]interface{}{
		"telefono_id":    phone.ID,
		"numeroCompleto": phone.NumeroCompleto,
		"status":         phone.Status,
		"lastConnected":  phone.LastConnected,
		"qr_string":      phone.QRString,
	}, claims.EmpresaID)
}

func (h *V1SessionsHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	telefonoID, err := extractTelefonoID(r)
	if err != nil {
		writeV1Error(w, http.StatusBadRequest, "MISSING_TELEFONO_ID", "telefono_id requerido")
		return
	}

	claims, ok := domain.GetEmpresaJWTClaims(r.Context())
	if !ok {
		writeV1Error(w, http.StatusUnauthorized, "TOKEN_REQUIRED", "JWT de empresa requerido")
		return
	}

	belongs, err := h.telefonoStore.BelongsToEmpresa(telefonoID, claims.EmpresaID)
	if err != nil || !belongs {
		writeV1Error(w, http.StatusForbidden, "FORBIDDEN", "El teléfono no pertenece a esta empresa")
		return
	}

	phone, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil || phone == nil {
		writeV1Error(w, http.StatusNotFound, "TELEFONO_NOT_FOUND", "Teléfono no encontrado")
		return
	}

	h.manager.Delete(phone.NumeroCompleto)
	h.sessionStore.SetDisconnected(phone.NumeroCompleto, "v1_disconnect")
	h.telefonoStore.SetDisconnected(telefonoID)

	writeV1Success(w, map[string]interface{}{
		"telefono_id": telefonoID,
		"status":      "disconnected",
	}, claims.EmpresaID)
}

func (h *V1SessionsHandler) StartPhoneConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	telefonoID, err := extractTelefonoID(r)
	if err != nil {
		writeV1Error(w, http.StatusBadRequest, "MISSING_TELEFONO_ID", "telefono_id requerido")
		return
	}

	claims, ok := domain.GetEmpresaJWTClaims(r.Context())
	if !ok {
		writeV1Error(w, http.StatusUnauthorized, "TOKEN_REQUIRED", "JWT de empresa requerido")
		return
	}

	belongs, err := h.telefonoStore.BelongsToEmpresa(telefonoID, claims.EmpresaID)
	if err != nil || !belongs {
		writeV1Error(w, http.StatusForbidden, "FORBIDDEN", "El teléfono no pertenece a esta empresa")
		return
	}

	phone, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil || phone == nil {
		writeV1Error(w, http.StatusNotFound, "TELEFONO_NOT_FOUND", "Teléfono no encontrado")
		return
	}

	if state, ok := h.sessionStore.Get(phone.NumeroCompleto); ok && state.Status == "active" {
		writeV1Success(w, map[string]interface{}{
			"telefono_id":    phone.ID,
			"numeroCompleto": phone.NumeroCompleto,
			"status":         "active",
			"lastConnected":  phone.LastConnected,
			"qr_string":      phone.QRString,
		}, claims.EmpresaID)
		return
	}

	events, err := whatsapp.StartSession(h.manager, phone.NumeroCompleto)
	if err != nil {
		writeV1Error(w, http.StatusInternalServerError, "CONNECT_FAILED", err.Error())
		return
	}

	go func() {
		for range events {
		}
	}()

	writeV1Success(w, map[string]interface{}{
		"telefono_id":    phone.ID,
		"numeroCompleto": phone.NumeroCompleto,
		"status":         "initializing",
		"qr_string":      phone.QRString,
		"expires_in":     60,
	}, claims.EmpresaID)
}

var _ = time.Now
