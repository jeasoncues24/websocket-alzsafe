package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"wsapi/internal/auth"
	"wsapi/internal/config"
	"wsapi/internal/storage"
	"wsapi/internal/whatsapp"

	"github.com/coder/websocket"
)

type V1WSHandler struct {
	manager       *whatsapp.Manager
	jwtCfg        *config.JWTConfig
	telefonoStore *storage.TelefonoStore
	sessionStore  *storage.SessionStore
	apiKeyStore   *storage.ApiKeyStore
}

func NewV1WSHandler(
	manager *whatsapp.Manager,
	jwtCfg *config.JWTConfig,
	telefonoStore *storage.TelefonoStore,
	sessionStore *storage.SessionStore,
	apiKeyStore *storage.ApiKeyStore,
) *V1WSHandler {
	return &V1WSHandler{
		manager:       manager,
		jwtCfg:        jwtCfg,
		telefonoStore: telefonoStore,
		sessionStore:  sessionStore,
		apiKeyStore:   apiKeyStore,
	}
}

func (h *V1WSHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	if token == "" {
		secProtocols := r.Header.Get("Sec-WebSocket-Protocol")
		if secProtocols != "" {
			parts := strings.Split(secProtocols, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					token = p
					break
				}
			}
		}
	}
	if token == "" {
		writeV1Error(w, http.StatusUnauthorized, "TOKEN_REQUIRED", "Token requerido")
		return
	}
	claims, err := auth.ParseQRLinkToken(token, h.jwtCfg.Secret)
	if err != nil {
		writeV1Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "Token QR inválido o expirado")
		return
	}

	acceptOpts := &websocket.AcceptOptions{InsecureSkipVerify: true}
	// Si el token provino del header Sec-WebSocket-Protocol, debemos negociarlo como subprotocolo
	if r.Header.Get("Sec-WebSocket-Protocol") != "" {
		acceptOpts.Subprotocols = []string{token}
	}

	c, err := websocket.Accept(w, r, acceptOpts)
	if err != nil {
		return
	}
	defer c.CloseNow()

	ctx := r.Context()

	// — Resolver phone según token —
	phoneID := claims.PhoneID

	// — Cargar teléfono (path común) —
	phone, err := h.telefonoStore.GetByID(phoneID)
	if err != nil || phone == nil {
		c.Close(websocket.StatusPolicyViolation, "teléfono no encontrado")
		return
	}
	accountID := whatsapp.NormalizeAccountID(phone.NumeroCompleto)

	fmt.Printf("[INFO] V1 WS QR-link opened phone=%d account=%s\n", phone.ID, accountID)

	events, unsubscribe, err := whatsapp.StartSession(h.manager, phone.NumeroCompleto)
	if err != nil {
		_ = writeWSEvent(c, "error", map[string]string{"message": "error al iniciar sesión: " + err.Error()})
		return
	}

	defer func() {
		// Darse de baja como observador. El runtime WhatsApp sigue vivo para otros
		// observadores y para que la sesión persista; si era una sesión en QR sin
		// completar, whatsmeow la termina sola al expirar el QR.
		unsubscribe()
		fmt.Printf("[INFO] V1 WS QR-link closed phone=%d account=%s reason=%v\n", phone.ID, accountID, ctx.Err())
		if h.sessionStore != nil {
			reasonStr := "normal"
			if ctx.Err() != nil {
				reasonStr = ctx.Err().Error()
			}
			h.sessionStore.AppendEvent(phone.NumeroCompleto, "ws_closed", "WS cliente V1 cerrado: "+reasonStr)
		}
	}()

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}
			if err := writeWSEvent(c, mapV1EventType(event.Event, event.Data), event.Data); err != nil {
				fmt.Printf("[WARN] V1 WS write event failed account=%s: %v\n", accountID, err)
				return
			}
		case <-ticker.C:
			if err := writeWSEvent(c, "ping", nil); err != nil {
				fmt.Printf("[WARN] V1 WS ping failed account=%s: %v\n", accountID, err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (h *V1WSHandler) HandleConnectWS(w http.ResponseWriter, r *http.Request) {
	if h.apiKeyStore == nil {
		writeV1Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Key store no inicializado")
		return
	}
	if h.telefonoStore == nil {
		writeV1Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Phone store no inicializado")
		return
	}

	rawKey, fromProtocol := extractConnectAPIKey(r)
	if rawKey == "" {
		writeV1Error(w, http.StatusUnauthorized, "API_KEY_REQUIRED", "API key requerida")
		return
	}

	key, err := h.apiKeyStore.Validate(rawKey)
	if err != nil || key == nil {
		writeV1Error(w, http.StatusUnauthorized, "INVALID_API_KEY", "API key inválida o expirada")
		return
	}

	phone, err := h.telefonoStore.GetByID(key.TelefonoID)
	if err != nil || phone == nil {
		writeV1Error(w, http.StatusUnauthorized, "TELEFONO_NOT_FOUND", "Teléfono no encontrado")
		return
	}
	if phone.NumeroCompleto == "" {
		writeV1Error(w, http.StatusBadRequest, "INVALID_PHONE", "Número de teléfono vacío")
		return
	}
	accountID := whatsapp.NormalizeAccountID(phone.NumeroCompleto)

	acceptOpts := &websocket.AcceptOptions{InsecureSkipVerify: true}
	if fromProtocol {
		acceptOpts.Subprotocols = []string{rawKey}
	}

	c, err := websocket.Accept(w, r, acceptOpts)
	if err != nil {
		return
	}
	defer c.CloseNow()

	ctx := r.Context()

	fmt.Printf("[INFO] V1 WS connect opened phone=%d account=%s empresa=%d\n", phone.ID, accountID, key.EmpresaID)

	events, unsubscribe, err := whatsapp.StartSession(h.manager, phone.NumeroCompleto)
	if err != nil {
		_ = writeWSEvent(c, "error", map[string]string{"message": "error al iniciar sesión: " + err.Error()})
		return
	}

	var writeErr error
	defer func() {
		unsubscribe()
		reasonStr := "normal"
		if ctx.Err() != nil {
			reasonStr = ctx.Err().Error()
		} else if writeErr != nil {
			reasonStr = writeErr.Error()
		}
		fmt.Printf("[INFO] V1 WS connect closed phone=%d account=%s reason=%s\n", phone.ID, accountID, reasonStr)
		if h.sessionStore != nil {
			h.sessionStore.AppendEvent(phone.NumeroCompleto, "ws_closed", "WS cliente V1 connect cerrado: "+reasonStr)
		}
	}()

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}
			mappedType := mapV1EventType(event.Event, event.Data)
			payload := event.Data
			if mappedType == "qr" && payload != nil {
				// Clonamos para no mutar el event.Data original
				newPayload := make(map[string]any)
				for k, v := range payload {
					newPayload[k] = v
				}
				if val, ok := newPayload["qrString"]; ok {
					newPayload["qr_string"] = val
					delete(newPayload, "qrString")
				}
				newPayload["message"] = "Escanea el código QR"
				payload = newPayload
			}
			if err := writeWSEvent(c, mappedType, payload); err != nil {
				writeErr = err
				fmt.Printf("[WARN] V1 WS connect write event failed account=%s: %v\n", accountID, err)
				return
			}
		case <-ticker.C:
			if err := writeWSEvent(c, "ping", nil); err != nil {
				writeErr = err
				fmt.Printf("[WARN] V1 WS connect ping failed account=%s: %v\n", accountID, err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func extractConnectAPIKey(r *http.Request) (rawKey string, fromProtocol bool) {
	if v := strings.TrimSpace(r.Header.Get("X-API-Key")); v != "" {
		return v, false
	}
	authHdr := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHdr != "" {
		parts := strings.SplitN(authHdr, " ", 2)
		if len(parts) == 2 && (strings.EqualFold(parts[0], "ApiKey") || strings.EqualFold(parts[0], "Bearer")) {
			return strings.TrimSpace(parts[1]), false
		}
	}
	if v := r.URL.Query().Get("api_key"); v != "" {
		return v, false
	}
	if sec := r.Header.Get("Sec-WebSocket-Protocol"); sec != "" {
		for _, p := range strings.Split(sec, ",") {
			if p = strings.TrimSpace(p); p != "" {
				return p, true
			}
		}
	}
	return "", false
}

func mapV1EventType(event string, data map[string]any) string {
	switch {
	case strings.HasPrefix(event, "qr-"):
		return "qr"
	case strings.HasPrefix(event, "active-"):
		if data != nil {
			if isActive, ok := data["isActive"].(bool); ok && isActive {
				return "connected"
			}
		}
		return "disconnected"
	default:
		return event
	}
}

func writeWSEvent(c *websocket.Conn, eventType string, data interface{}) error {
	msg := map[string]interface{}{"type": eventType}
	if data != nil {
		msg["data"] = data
	}
	return WriteWSJSON(context.Background(), c, msg)
}
