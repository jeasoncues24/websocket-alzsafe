package http

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
)

type V1WebhooksHandler struct {
	webhookStore *storage.WebhookStore
	maxWebhooks  int
}

func NewV1WebhooksHandler(webhookStore *storage.WebhookStore, maxWebhooks int) *V1WebhooksHandler {
	return &V1WebhooksHandler{
		webhookStore: webhookStore,
		maxWebhooks:  maxWebhooks,
	}
}

type createWebhookRequest struct {
	URL     string                `json:"url"`
	Eventos []domain.WebhookEvent `json:"eventos"`
}

func init() {
	validWebhookEvents = map[domain.WebhookEvent]bool{
		domain.WebhookEventMessageReceived:     true,
		domain.WebhookEventMessageStatus:       true,
		domain.WebhookEventSessionConnected:    true,
		domain.WebhookEventSessionDisconnected: true,
	}
}

var validWebhookEvents map[domain.WebhookEvent]bool

func isValidWebhookURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return false
	}
	return parsed.Scheme == "https" && parsed.Host != ""
}

func (h *V1WebhooksHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	claims, ok := domain.GetApiKeyClaims(r.Context())
	if !ok {
		writeV1Error(w, http.StatusUnauthorized, "API_KEY_REQUIRED", "API key requerida")
		return
	}

	var req createWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeV1Error(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	req.URL = strings.TrimSpace(req.URL)
	if !isValidWebhookURL(req.URL) {
		writeV1Error(w, http.StatusBadRequest, "INVALID_URL", "URL debe ser HTTPS")
		return
	}

	if len(req.Eventos) == 0 {
		writeV1Error(w, http.StatusBadRequest, "INVALID_EVENTOS", "Debe especificar al menos un evento")
		return
	}

	for _, ev := range req.Eventos {
		if !validWebhookEvents[ev] {
			writeV1Error(w, http.StatusBadRequest, "INVALID_EVENTOS", "Evento no válido: "+string(ev))
			return
		}
	}

	existing, err := h.webhookStore.ListByApiKey(claims.ApiKeyID)
	if err != nil {
		writeV1Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al verificar webhooks existentes")
		return
	}

	activeCount := 0
	for _, wh := range existing {
		if wh.Activo {
			activeCount++
		}
	}
	if activeCount >= h.maxWebhooks {
		writeV1Error(w, http.StatusBadRequest, "MAX_WEBHOOKS_REACHED", "Límite de webhooks alcanzado")
		return
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		writeV1Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al generar secret")
		return
	}
	secret := hex.EncodeToString(b)

	wh := &domain.Webhook{
		EmpresaID:  claims.EmpresaID,
		TelefonoID: claims.TelefonoID,
		ApiKeyID:   claims.ApiKeyID,
		URL:        req.URL,
		Secret:     secret,
		Eventos:    req.Eventos,
		Activo:     true,
	}

	if err := h.webhookStore.Create(wh); err != nil {
		writeV1Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al crear webhook")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": true,
		"data": map[string]interface{}{
			"id":     wh.ID,
			"secret": secret,
		},
	})
}

func (h *V1WebhooksHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	claims, ok := domain.GetApiKeyClaims(r.Context())
	if !ok {
		writeV1Error(w, http.StatusUnauthorized, "API_KEY_REQUIRED", "API key requerida")
		return
	}

	webhooks, err := h.webhookStore.ListByApiKey(claims.ApiKeyID)
	if err != nil {
		writeV1Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al obtener webhooks")
		return
	}

	result := make([]map[string]interface{}, 0, len(webhooks))
	for _, wh := range webhooks {
		result = append(result, map[string]interface{}{
			"id":              wh.ID,
			"telefono_id":     wh.TelefonoID,
			"api_key_id":      wh.ApiKeyID,
			"url":             wh.URL,
			"eventos":         wh.Eventos,
			"activo":          wh.Activo,
			"failure_count":   wh.FailureCount,
			"last_error":      wh.LastError,
			"last_success_at": wh.LastSuccessAt,
			"created_at":      wh.CreatedAt,
			"updated_at":      wh.UpdatedAt,
		})
	}

	writeV1Success(w, map[string]interface{}{
		"webhooks": result,
		"total":    len(result),
	}, claims.EmpresaID)
}


func (h *V1WebhooksHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	claims, ok := domain.GetApiKeyClaims(r.Context())
	if !ok {
		writeV1Error(w, http.StatusUnauthorized, "API_KEY_REQUIRED", "API key requerida")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeV1Error(w, http.StatusBadRequest, "INVALID_ID", "ID inválido")
		return
	}

	wh, err := h.webhookStore.GetByID(id)
	if err != nil {
		writeV1Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al obtener webhook")
		return
	}
	if wh == nil || wh.ApiKeyID != claims.ApiKeyID {
		writeV1Error(w, http.StatusNotFound, "NOT_FOUND", "Webhook no encontrado")
		return
	}

	if err := h.webhookStore.Delete(id); err != nil {
		writeV1Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al eliminar webhook")
		return
	}

	writeV1Success(w, map[string]interface{}{
		"deleted": true,
	}, claims.EmpresaID)
}
