package http

import (
	"net/http"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
)

type V1PhonesHandler struct {
	telefonoStore *storage.TelefonoStore
	sessionStore  *storage.SessionStore
}

func NewV1PhonesHandler(telefonoStore *storage.TelefonoStore, sessionStore *storage.SessionStore) *V1PhonesHandler {
	return &V1PhonesHandler{
		telefonoStore: telefonoStore,
		sessionStore:  sessionStore,
	}
}

func (h *V1PhonesHandler) GetPhones(w http.ResponseWriter, r *http.Request) {
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

	result := make([]map[string]interface{}, 0, len(phones))
	for _, phone := range phones {
		result = append(result, map[string]interface{}{
			"telefono_id":    phone.ID,
			"numero":         phone.Numero,
			"codigo_pais":    phone.CodigoPais,
			"numeroCompleto": phone.NumeroCompleto,
			"status":         phone.Status,
			"lastConnected":  phone.LastConnected,
		})
	}

	writeV1Success(w, map[string]interface{}{
		"phones": result,
		"total":  len(result),
	}, claims.EmpresaID)
}

func (h *V1PhonesHandler) PostPhoneQr(w http.ResponseWriter, r *http.Request) {
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

	belongs, _ := h.telefonoStore.BelongsToEmpresa(telefonoID, claims.EmpresaID)
	if !belongs {
		writeV1Error(w, http.StatusForbidden, "FORBIDDEN", "El teléfono no pertenece a esta empresa")
		return
	}

	phone, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil || phone == nil {
		writeV1Error(w, http.StatusNotFound, "TELEFONO_NOT_FOUND", "Teléfono no encontrado")
		return
	}

	qrString := ""
	if state, ok := h.sessionStore.Get(phone.NumeroCompleto); ok && state.QRString != "" {
		qrString = state.QRString
		h.telefonoStore.UpdateQRString(telefonoID, qrString)
	}

	writeV1Success(w, map[string]interface{}{
		"telefono_id": telefonoID,
		"qr_string":   qrString,
		"expires_in":  60,
	}, claims.EmpresaID)
}

func (h *V1PhonesHandler) GetPhoneStatus(w http.ResponseWriter, r *http.Request) {
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

	belongs, _ := h.telefonoStore.BelongsToEmpresa(telefonoID, claims.EmpresaID)
	if !belongs {
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
