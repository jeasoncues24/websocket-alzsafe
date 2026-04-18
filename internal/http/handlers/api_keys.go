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
)

type ApiKeysHandler struct {
	apiKeyStore   *storage.ApiKeyStore
	telefonoStore *storage.TelefonoStore
	empresaStore  domain.EmpresaStoreInterface
}

func NewApiKeysHandler(apiKeyStore *storage.ApiKeyStore, telefonoStore *storage.TelefonoStore, empresaStore domain.EmpresaStoreInterface) *ApiKeysHandler {
	return &ApiKeysHandler{apiKeyStore: apiKeyStore, telefonoStore: telefonoStore, empresaStore: empresaStore}
}

type createApiKeyRequest struct {
	Nombre    string   `json:"nombre"`
	Scopes    []string `json:"scopes"`
	ExpiresAt string   `json:"expires_at,omitempty"`
}

func (h *ApiKeysHandler) ListByTelefono(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	telefonoID := extractTelefonoKeyID(r.URL.Path, "/api/admin/telefonos/", "/api-keys")
	if telefonoID <= 0 {
		http.Error(w, `{"ok": false, "error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	phone, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil || phone == nil {
		http.Error(w, `{"ok": false, "error": "Teléfono no encontrado"}`, http.StatusNotFound)
		return
	}
	if !h.canAccessTelefono(r, phone.EmpresaID) {
		http.Error(w, `{"ok": false, "error": "Acceso denegado a este teléfono"}`, http.StatusForbidden)
		return
	}

	keys, err := h.apiKeyStore.GetByTelefonoID(telefonoID)
	if err != nil {
		http.Error(w, `{"ok": false, "error": "Error al obtener API keys"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.ApiKeyListResponse{OK: true, ApiKeys: keys})
}

func (h *ApiKeysHandler) CreateForTelefono(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	telefonoID := extractTelefonoKeyID(r.URL.Path, "/api/admin/telefonos/", "/api-keys")
	if telefonoID <= 0 {
		http.Error(w, `{"ok": false, "error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	phone, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil || phone == nil {
		http.Error(w, `{"ok": false, "error": "Teléfono no encontrado"}`, http.StatusNotFound)
		return
	}
	if !h.canAccessTelefono(r, phone.EmpresaID) {
		http.Error(w, `{"ok": false, "error": "Acceso denegado a este teléfono"}`, http.StatusForbidden)
		return
	}

	var req createApiKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"ok": false, "error": "JSON inválido"}`, http.StatusBadRequest)
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
			http.Error(w, `{"ok": false, "error": "expires_at inválido"}`, http.StatusBadRequest)
			return
		}
		expiresAt = &parsed
	}

	prefix, rawKey, err := auth.GenerateAPIKeyMaterial()
	if err != nil {
		http.Error(w, `{"ok": false, "error": "Error al generar la API key"}`, http.StatusInternalServerError)
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
		http.Error(w, `{"ok": false, "error": "Error al crear API key"}`, http.StatusInternalServerError)
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
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	apiKeyID := extractAPIKeyID(r.URL.Path)
	if apiKeyID <= 0 {
		http.Error(w, `{"ok": false, "error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	apiKey, err := h.apiKeyStore.GetByID(apiKeyID)
	if err != nil || apiKey == nil {
		http.Error(w, `{"ok": false, "error": "API key no encontrada"}`, http.StatusNotFound)
		return
	}
	if !h.canAccessTelefono(r, apiKey.EmpresaID) {
		http.Error(w, `{"ok": false, "error": "Acceso denegado a esta API key"}`, http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.ApiKeyResponse{OK: true, ApiKey: apiKey})
}

func (h *ApiKeysHandler) Rotate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	apiKeyID := extractAPIKeyID(strings.TrimSuffix(r.URL.Path, "/rotate"))
	if apiKeyID <= 0 {
		http.Error(w, `{"ok": false, "error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	oldKey, err := h.apiKeyStore.GetByID(apiKeyID)
	if err != nil || oldKey == nil {
		http.Error(w, `{"ok": false, "error": "API key no encontrada"}`, http.StatusNotFound)
		return
	}
	if !h.canAccessTelefono(r, oldKey.EmpresaID) {
		http.Error(w, `{"ok": false, "error": "Acceso denegado a esta API key"}`, http.StatusForbidden)
		return
	}

	claims, _ := domain.GetAdminJWTClaims(r.Context())
	prefix, rawKey, err := auth.GenerateAPIKeyMaterial()
	if err != nil {
		http.Error(w, `{"ok": false, "error": "Error al generar la API key"}`, http.StatusInternalServerError)
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
		http.Error(w, `{"ok": false, "error": "Error al rotar API key"}`, http.StatusInternalServerError)
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
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	apiKeyID := extractAPIKeyID(strings.TrimSuffix(r.URL.Path, "/revoke"))
	if apiKeyID <= 0 {
		http.Error(w, `{"ok": false, "error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	apiKey, err := h.apiKeyStore.GetByID(apiKeyID)
	if err != nil || apiKey == nil {
		http.Error(w, `{"ok": false, "error": "API key no encontrada"}`, http.StatusNotFound)
		return
	}
	if !h.canAccessTelefono(r, apiKey.EmpresaID) {
		http.Error(w, `{"ok": false, "error": "Acceso denegado a esta API key"}`, http.StatusForbidden)
		return
	}

	if err := h.apiKeyStore.Revoke(apiKeyID); err != nil {
		http.Error(w, `{"ok": false, "error": "Error al revocar API key"}`, http.StatusInternalServerError)
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
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	apiKeyID := extractAPIKeyID(strings.TrimSuffix(r.URL.Path, "/usage"))
	if apiKeyID <= 0 {
		http.Error(w, `{"ok": false, "error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	apiKey, err := h.apiKeyStore.GetByID(apiKeyID)
	if err != nil || apiKey == nil {
		http.Error(w, `{"ok": false, "error": "API key no encontrada"}`, http.StatusNotFound)
		return
	}
	if !h.canAccessTelefono(r, apiKey.EmpresaID) {
		http.Error(w, `{"ok": false, "error": "Acceso denegado a esta API key"}`, http.StatusForbidden)
		return
	}

	usage, err := h.apiKeyStore.GetUsageDailyByKey(apiKeyID)
	if err != nil {
		http.Error(w, `{"ok": false, "error": "Error al obtener uso"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "usage": usage})
}

func (h *ApiKeysHandler) Audit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	apiKeyID := extractAPIKeyID(strings.TrimSuffix(r.URL.Path, "/audit"))
	if apiKeyID <= 0 {
		http.Error(w, `{"ok": false, "error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	apiKey, err := h.apiKeyStore.GetByID(apiKeyID)
	if err != nil || apiKey == nil {
		http.Error(w, `{"ok": false, "error": "API key no encontrada"}`, http.StatusNotFound)
		return
	}
	if !h.canAccessTelefono(r, apiKey.EmpresaID) {
		http.Error(w, `{"ok": false, "error": "Acceso denegado a esta API key"}`, http.StatusForbidden)
		return
	}

	audit, err := h.apiKeyStore.GetAuditEventsByKey(apiKeyID)
	if err != nil {
		http.Error(w, `{"ok": false, "error": "Error al obtener auditoría"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "audit": audit})
}

func (h *ApiKeysHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	claims, ok := domain.GetApiKeyClaims(r.Context())
	if !ok {
		http.Error(w, `{"ok": false, "error": "API key requerida"}`, http.StatusUnauthorized)
		return
	}

	apiKey, err := h.apiKeyStore.GetByID(claims.ApiKeyID)
	if err != nil || apiKey == nil {
		http.Error(w, `{"ok": false, "error": "API key no encontrada"}`, http.StatusNotFound)
		return
	}

	phone, err := h.telefonoStore.GetByID(claims.TelefonoID)
	if err != nil || phone == nil {
		http.Error(w, `{"ok": false, "error": "Teléfono no encontrado"}`, http.StatusNotFound)
		return
	}

	company, err := h.empresaStore.GetByID(claims.EmpresaID)
	if err != nil || company == nil {
		http.Error(w, `{"ok": false, "error": "Empresa no encontrada"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":       true,
		"api_key":  apiKey,
		"empresa":  company,
		"telefono": phone,
	})
}

func (h *ApiKeysHandler) canAccessTelefono(r *http.Request, empresaID int64) bool {
	claims, _ := domain.GetAdminJWTClaims(r.Context())
	if claims == nil {
		return false
	}
	if claims.IsRoot || claims.Rol == domain.RoleSuperAdmin {
		return true
	}
	return claims.EmpresaID != nil && *claims.EmpresaID == empresaID
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
