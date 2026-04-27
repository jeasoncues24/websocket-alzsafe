package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"wsapi/internal/auth"
	"wsapi/internal/config"
	"wsapi/internal/whatsapp"

	"github.com/coder/websocket"
)

type V1WSHandler struct {
	manager *whatsapp.Manager
	jwtCfg  *config.JWTConfig
}

func NewV1WSHandler(manager *whatsapp.Manager, jwtCfg *config.JWTConfig) *V1WSHandler {
	return &V1WSHandler{manager: manager, jwtCfg: jwtCfg}
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

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}
	defer c.CloseNow()

	ctx := r.Context()
	for {
		_, data, err := c.Read(ctx)
		if err != nil {
			return
		}
		h.processMessage(ctx, c, data, claims.EmpresaID)
	}
}

type wsInboundPayload struct {
	Event string          `json:"type"`
	Data  json.RawMessage `json:"data"`
}

func (h *V1WSHandler) processMessage(ctx context.Context, c *websocket.Conn, data []byte, empresaID int64) {
	var payload wsInboundPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		_ = writeWSEvent(c, "error", map[string]string{"message": "invalid payload"})
		return
	}

	switch payload.Event {
	case "ping":
		_ = writeWSEvent(c, "pong", nil)
	case "subscribe":
		var req struct {
			PhoneID int64 `json:"phone_id"`
		}
		if json.Unmarshal(payload.Data, &req) == nil {
			_ = writeWSEvent(c, "subscribed", map[string]int64{"phone_id": req.PhoneID})
		}
	default:
		_ = writeWSEvent(c, "error", map[string]string{"message": "unknown event"})
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

var _ = whatsapp.NormalizeAccountID
