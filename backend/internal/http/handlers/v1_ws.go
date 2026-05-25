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
}

func NewV1WSHandler(
	manager *whatsapp.Manager,
	jwtCfg *config.JWTConfig,
	telefonoStore *storage.TelefonoStore,
	sessionStore *storage.SessionStore,
) *V1WSHandler {
	return &V1WSHandler{
		manager:       manager,
		jwtCfg:        jwtCfg,
		telefonoStore: telefonoStore,
		sessionStore:  sessionStore,
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

	defer func() {
		fmt.Printf("[INFO] V1 WS QR-link closed phone=%d account=%s reason=%v\n", phone.ID, accountID, ctx.Err())
		if h.sessionStore != nil && h.manager != nil {
			// Auditar el cierre del WebSocket en el historial de eventos en memoria
			reasonStr := "normal"
			if ctx.Err() != nil {
				reasonStr = ctx.Err().Error()
			}
			h.sessionStore.AppendEvent(phone.NumeroCompleto, "ws_closed", "WS cliente V1 cerrado: "+reasonStr)

			// Evitar carrera: si el cliente ya se conectó activamente en segundo plano, no borrarlo
			if client, ok := h.manager.Get(accountID); ok && client != nil && client.IsConnected() {
				return
			}
			if state, ok := h.sessionStore.Get(phone.NumeroCompleto); ok {
				if state.Status == "initializing" || state.Status == "qr_pending" {
					h.manager.Delete(accountID)
				}
			}
		}
	}()

	events, err := whatsapp.StartSession(h.manager, phone.NumeroCompleto)
	if err != nil {
		_ = writeWSEvent(c, "error", map[string]string{"message": "error al iniciar sesión: " + err.Error()})
		return
	}

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
