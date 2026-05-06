package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"wsapi/internal/auth"
	"wsapi/internal/domain"
	"wsapi/internal/storage"
	"wsapi/internal/whatsapp"
)

type ApiKeysHandler struct {
	apiKeyStore   *storage.ApiKeyStore
	telefonoStore *storage.TelefonoStore
	empresaStore  domain.EmpresaStoreInterface
	manager       *whatsapp.Manager
}

func NewApiKeysHandler(apiKeyStore *storage.ApiKeyStore, telefonoStore *storage.TelefonoStore, empresaStore domain.EmpresaStoreInterface, manager *whatsapp.Manager) *ApiKeysHandler {
	return &ApiKeysHandler{apiKeyStore: apiKeyStore, telefonoStore: telefonoStore, empresaStore: empresaStore, manager: manager}
}

type createApiKeyRequest struct {
	Nombre    string   `json:"nombre"`
	Scopes    []string `json:"scopes"`
	ExpiresAt string   `json:"expires_at,omitempty"`
}

func (h *ApiKeysHandler) ListByTelefono(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	telefonoID := extractTelefonoKeyID(r.URL.Path, "/api/admin/telefonos/", "/api-keys")
	if telefonoID <= 0 {
		writeAPIError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	phone, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil || phone == nil {
		writeAPIError(w, http.StatusNotFound, "Teléfono no encontrado")
		return
	}
	if !h.canAccessTelefono(r, phone.EmpresaID) {
		writeAPIError(w, http.StatusForbidden, "Acceso denegado a este teléfono")
		return
	}

	keys, err := h.apiKeyStore.GetByTelefonoID(telefonoID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al obtener API keys")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.ApiKeyListResponse{OK: true, ApiKeys: keys})
}

func (h *ApiKeysHandler) CreateForTelefono(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	telefonoID := extractTelefonoKeyID(r.URL.Path, "/api/admin/telefonos/", "/api-keys")
	if telefonoID <= 0 {
		writeAPIError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	phone, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil || phone == nil {
		writeAPIError(w, http.StatusNotFound, "Teléfono no encontrado")
		return
	}
	if !h.canAccessTelefono(r, phone.EmpresaID) {
		writeAPIError(w, http.StatusForbidden, "Acceso denegado a este teléfono")
		return
	}

	var req createApiKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	if req.Nombre == "" {
		req.Nombre = fmt.Sprintf("Key %s", phone.NumeroCompleto)
	}

	claims, _ := domain.GetAdminJWTClaims(r.Context())
	var expiresAt *time.Time
	if strings.TrimSpace(req.ExpiresAt) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(req.ExpiresAt))
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "expires_at inválido")
			return
		}
		expiresAt = &parsed
	}

	prefix, rawKey, err := auth.GenerateAPIKeyMaterial()
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al generar la API key")
		return
	}

	apiKey := &domain.ApiKey{
		EmpresaID:       phone.EmpresaID,
		TelefonoID:      phone.ID,
		Nombre:          req.Nombre,
		KeyPrefix:       prefix,
		Scopes:          req.Scopes,
		Activo:          true,
		CreatedByUserID: nil,
		ExpiresAt:       expiresAt,
	}
	if claims != nil {
		apiKey.CreatedByUserID = &claims.UserID
	}

	secret, err := h.apiKeyStore.Create(apiKey, rawKey)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al crear API key")
		return
	}

	metadata, _ := json.Marshal(map[string]any{"key_prefix": prefix, "nombre": apiKey.Nombre})
	_ = h.apiKeyStore.RecordAuditEvent(&domain.ApiKeyAuditEvent{
		ApiKeyID:    apiKey.ID,
		EmpresaID:   apiKey.EmpresaID,
		TelefonoID:  apiKey.TelefonoID,
		Action:      "created",
		ActorUserID: apiKey.CreatedByUserID,
		Metadata:    metadata,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(domain.ApiKeyCreateResponse{
		OK:      true,
		ApiKey:  apiKey,
		Secret:  secret,
		Message: "API key creada exitosamente. Guárdala en un lugar seguro.",
	})
}

func (h *ApiKeysHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	apiKeyID := extractAPIKeyID(r.URL.Path)
	if apiKeyID <= 0 {
		writeAPIError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	apiKey, err := h.apiKeyStore.GetByID(apiKeyID)
	if err != nil || apiKey == nil {
		writeAPIError(w, http.StatusNotFound, "API key no encontrada")
		return
	}
	if !h.canAccessTelefono(r, apiKey.EmpresaID) {
		writeAPIError(w, http.StatusForbidden, "Acceso denegado a esta API key")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.ApiKeyResponse{OK: true, ApiKey: apiKey})
}

func (h *ApiKeysHandler) Rotate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	apiKeyID := extractAPIKeyID(strings.TrimSuffix(r.URL.Path, "/rotate"))
	if apiKeyID <= 0 {
		writeAPIError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	oldKey, err := h.apiKeyStore.GetByID(apiKeyID)
	if err != nil || oldKey == nil {
		writeAPIError(w, http.StatusNotFound, "API key no encontrada")
		return
	}
	if !h.canAccessTelefono(r, oldKey.EmpresaID) {
		writeAPIError(w, http.StatusForbidden, "Acceso denegado a esta API key")
		return
	}

	claims, _ := domain.GetAdminJWTClaims(r.Context())
	prefix, rawKey, err := auth.GenerateAPIKeyMaterial()
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al generar la API key")
		return
	}

	newKey := &domain.ApiKey{
		EmpresaID:       oldKey.EmpresaID,
		TelefonoID:      oldKey.TelefonoID,
		Nombre:          oldKey.Nombre,
		KeyPrefix:       prefix,
		Scopes:          oldKey.Scopes,
		Activo:          true,
		CreatedByUserID: nil,
		ExpiresAt:       oldKey.ExpiresAt,
	}
	if claims != nil {
		newKey.CreatedByUserID = &claims.UserID
	}
	newKey.RotatedFromID = &oldKey.ID

	secret, err := h.apiKeyStore.Rotate(oldKey, newKey, rawKey)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al rotar API key")
		return
	}

	metadata, _ := json.Marshal(map[string]any{"rotated_from": oldKey.ID, "new_key_prefix": prefix})
	_ = h.apiKeyStore.RecordAuditEvent(&domain.ApiKeyAuditEvent{
		ApiKeyID:    newKey.ID,
		EmpresaID:   newKey.EmpresaID,
		TelefonoID:  newKey.TelefonoID,
		Action:      "rotated",
		ActorUserID: newKey.CreatedByUserID,
		Metadata:    metadata,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.ApiKeyCreateResponse{
		OK:      true,
		ApiKey:  newKey,
		Secret:  secret,
		Message: "API key rotada exitosamente. La anterior quedó revocada.",
	})
}

func (h *ApiKeysHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	apiKeyID := extractAPIKeyID(strings.TrimSuffix(r.URL.Path, "/revoke"))
	if apiKeyID <= 0 {
		writeAPIError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	apiKey, err := h.apiKeyStore.GetByID(apiKeyID)
	if err != nil || apiKey == nil {
		writeAPIError(w, http.StatusNotFound, "API key no encontrada")
		return
	}
	if !h.canAccessTelefono(r, apiKey.EmpresaID) {
		writeAPIError(w, http.StatusForbidden, "Acceso denegado a esta API key")
		return
	}

	if err := h.apiKeyStore.Revoke(apiKeyID); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al revocar API key")
		return
	}

	claims, _ := domain.GetAdminJWTClaims(r.Context())
	metadata, _ := json.Marshal(map[string]any{"api_key_id": apiKeyID})
	_ = h.apiKeyStore.RecordAuditEvent(&domain.ApiKeyAuditEvent{
		ApiKeyID:   apiKeyID,
		EmpresaID:  apiKey.EmpresaID,
		TelefonoID: apiKey.TelefonoID,
		Action:     "revoked",
		ActorUserID: func() *int64 {
			if claims != nil {
				return &claims.UserID
			}
			return nil
		}(),
		Metadata: metadata,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.ApiKeyResponse{OK: true, ApiKey: apiKey})
}

func (h *ApiKeysHandler) Usage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	apiKeyID := extractAPIKeyID(strings.TrimSuffix(r.URL.Path, "/usage"))
	if apiKeyID <= 0 {
		writeAPIError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	apiKey, err := h.apiKeyStore.GetByID(apiKeyID)
	if err != nil || apiKey == nil {
		writeAPIError(w, http.StatusNotFound, "API key no encontrada")
		return
	}
	if !h.canAccessTelefono(r, apiKey.EmpresaID) {
		writeAPIError(w, http.StatusForbidden, "Acceso denegado a esta API key")
		return
	}

	usage, err := h.apiKeyStore.GetUsageDailyByKey(apiKeyID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al obtener uso")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "usage": usage})
}

func (h *ApiKeysHandler) Audit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	apiKeyID := extractAPIKeyID(strings.TrimSuffix(r.URL.Path, "/audit"))
	if apiKeyID <= 0 {
		writeAPIError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	apiKey, err := h.apiKeyStore.GetByID(apiKeyID)
	if err != nil || apiKey == nil {
		writeAPIError(w, http.StatusNotFound, "API key no encontrada")
		return
	}
	if !h.canAccessTelefono(r, apiKey.EmpresaID) {
		writeAPIError(w, http.StatusForbidden, "Acceso denegado a esta API key")
		return
	}

	audit, err := h.apiKeyStore.GetAuditEventsByKey(apiKeyID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al obtener auditoría")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "audit": audit})
}

func (h *ApiKeysHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	claims, ok := domain.GetApiKeyClaims(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "API key requerida")
		return
	}

	apiKey, err := h.apiKeyStore.GetByID(claims.ApiKeyID)
	if err != nil || apiKey == nil {
		writeAPIError(w, http.StatusNotFound, "API key no encontrada")
		return
	}

	phone, err := h.telefonoStore.GetByID(claims.TelefonoID)
	if err != nil || phone == nil {
		writeAPIError(w, http.StatusNotFound, "Teléfono no encontrado")
		return
	}

	company, err := h.empresaStore.GetByID(claims.EmpresaID)
	if err != nil || company == nil {
		writeAPIError(w, http.StatusNotFound, "Empresa no encontrada")
		return
	}

	runtime := h.runtimeSessionInfo(phone)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":              true,
		"api_key":         apiKey,
		"empresa":         company,
		"telefono":        phone,
		"session_runtime": runtime,
	})
}

func (h *ApiKeysHandler) Session(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	claims, ok := domain.GetApiKeyClaims(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "API key requerida")
		return
	}

	phone, err := h.telefonoStore.GetByID(claims.TelefonoID)
	if err != nil || phone == nil {
		writeAPIError(w, http.StatusNotFound, "Teléfono no encontrado")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":   true,
		"data": h.runtimeSessionInfo(phone),
	})
}

func (h *ApiKeysHandler) runtimeSessionInfo(phone *domain.Telefono) map[string]any {
	accountID := whatsapp.NormalizeAccountID(phone.NumeroCompleto)
	runtimeConnected := false
	if h.manager != nil {
		if client, ok := h.manager.Get(accountID); ok && client != nil && client.IsConnected() {
			runtimeConnected = true
		}
	}

	return buildSessionRuntimeInfo(phone, accountID, runtimeConnected)
}

func buildSessionRuntimeInfo(phone *domain.Telefono, accountID string, runtimeConnected bool) map[string]any {
	dbStatus := string(phone.Status)
	runtimeStatus := "disconnected"
	if runtimeConnected {
		runtimeStatus = "connected"
	}

	expectedActive := phone.Status == domain.TelefonoStatusActive
	mismatch := expectedActive != runtimeConnected
	reason := ""
	if mismatch {
		if expectedActive {
			reason = "db_active_runtime_disconnected"
		} else {
			reason = "db_not_active_runtime_connected"
		}
	}

	return map[string]any{
		"telefono_id":        phone.ID,
		"account_id":         accountID,
		"status_db":          dbStatus,
		"status_runtime":     runtimeStatus,
		"runtime_connected":  runtimeConnected,
		"mismatch":           mismatch,
		"mismatch_reason":    reason,
		"recommended_action": recommendedSessionAction(dbStatus, runtimeConnected),
	}
}

func recommendedSessionAction(dbStatus string, runtimeConnected bool) string {
	if runtimeConnected {
		return "none"
	}
	if dbStatus == string(domain.TelefonoStatusActive) {
		return "reanudar_conexion"
	}
	return "iniciar_conexion"
}

func (h *ApiKeysHandler) canAccessTelefono(r *http.Request, empresaID int64) bool {
	access, ok := domain.GetPanelAccess(r.Context())
	if !ok {
		return false
	}
	return access.CanAccessEmpresa(empresaID)
}

func extractAPIKeyID(path string) int64 {
	path = strings.TrimSuffix(path, "/")
	segments := strings.Split(path, "/")
	for i := len(segments) - 1; i >= 0; i-- {
		if id, err := strconv.ParseInt(segments[i], 10, 64); err == nil && id > 0 {
			return id
		}
	}
	return 0
}

func extractTelefonoKeyID(path, prefix, suffix string) int64 {
	path = strings.TrimSuffix(path, "/")
	path = strings.TrimPrefix(path, prefix)
	path = strings.TrimSuffix(path, suffix)
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return 0
	}
	id, _ := strconv.ParseInt(parts[0], 10, 64)
	return id
}
