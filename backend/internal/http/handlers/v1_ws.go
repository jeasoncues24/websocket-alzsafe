package http

import (
	"context"
	"encoding/json"
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
		writeV1Error(w, http.StatusUnauthorized, "TOKEN_REQUIRED", "Token requerido")
		return
	}
	claims, err := auth.ParseEmpresaJWT(token, h.jwtCfg.Secret)
	if err != nil {
		writeV1Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "Token inválido o expirado")
		return
	}

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	defer c.CloseNow()

	ctx := r.Context()

	// — Resolver phone según tipo de token —
	var phoneID int64

	if claims.Scope == "qr_link" {
		// Token provisional: auto-suscribir al teléfono del token
		if claims.PhoneID <= 0 {
			_ = writeWSEvent(c, "error", map[string]string{"message": "token QR inválido: phone_id ausente"})
			return
		}
		phoneID = claims.PhoneID
	} else {
		// Empresa JWT regular: esperar mensaje subscribe con phone_id
		_, data, err := c.Read(ctx)
		if err != nil {
			return
		}
		var payload struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(data, &payload); err != nil || payload.Type != "subscribe" {
			_ = writeWSEvent(c, "error", map[string]string{"message": "primer mensaje debe ser subscribe"})
			return
		}
		var sub struct {
			PhoneID int64 `json:"phone_id"`
		}
		if err := json.Unmarshal(payload.Data, &sub); err != nil || sub.PhoneID <= 0 {
			_ = writeWSEvent(c, "error", map[string]string{"message": "phone_id requerido"})
			return
		}
		belongs, _ := h.telefonoStore.BelongsToEmpresa(sub.PhoneID, claims.EmpresaID)
		if !belongs {
			_ = writeWSEvent(c, "error", map[string]string{"message": "forbidden"})
			return
		}
		phoneID = sub.PhoneID
	}

	// — Cargar teléfono (path común) —
	phone, err := h.telefonoStore.GetByID(phoneID)
	if err != nil || phone == nil {
		_ = writeWSEvent(c, "error", map[string]string{"message": "teléfono no encontrado"})
		return
	}
	accountID := whatsapp.NormalizeAccountID(phone.NumeroCompleto)

	if claims.Scope == "qr_link" {
		fmt.Printf("[INFO] V1 WS QR-link opened phone=%d account=%s\n", phone.ID, accountID)
	} else {
		fmt.Printf("[INFO] V1 WS opened empresa=%d phone=%d account=%s\n", claims.EmpresaID, phone.ID, accountID)
	}

	defer func() {
		if claims.Scope == "qr_link" {
			fmt.Printf("[INFO] V1 WS QR-link closed phone=%d account=%s reason=%v\n", phone.ID, accountID, ctx.Err())
		} else {
			fmt.Printf("[INFO] V1 WS closed empresa=%d account=%s reason=%v\n", claims.EmpresaID, accountID, ctx.Err())
		}
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
	msgBytes, _ := json.Marshal(msg)
	return c.Write(context.Background(), websocket.MessageText, msgBytes)
}
