package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/coder/websocket"

	"wsapi/internal/domain"
)

func writeV1Error(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":      false,
		"error":   code,
		"message": message,
	})
}

func writeV1Success(w http.ResponseWriter, data map[string]interface{}, empresaID int64) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"ok":   true,
		"data": data,
		"meta": map[string]interface{}{
			"empresa_id": empresaID,
			"timestamp":  time.Now().Format(time.RFC3339),
		},
	}
	json.NewEncoder(w).Encode(response)
}


func extractTelefonoID(r *http.Request) (int64, error) {
	if v := r.URL.Query().Get("telefono_id"); v != "" {
		return strconv.ParseInt(v, 10, 64)
	}
	if claims, ok := domain.GetApiKeyClaims(r.Context()); ok && claims.TelefonoID > 0 {
		return claims.TelefonoID, nil
	}
	path := strings.TrimSuffix(r.URL.Path, "/")
	segments := strings.Split(path, "/")
	for i := len(segments) - 1; i >= 0; i-- {
		if id, err := strconv.ParseInt(segments[i], 10, 64); err == nil && id > 0 {
			return id, nil
		}
	}
	return 0, http.ErrNoCookie
}

// WriteWSJSON serializa y escribe un payload JSON a través de un WebSocket.
func WriteWSJSON(ctx context.Context, c *websocket.Conn, payload interface{}) error {
	msgBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.Write(ctx, websocket.MessageText, msgBytes)
}
