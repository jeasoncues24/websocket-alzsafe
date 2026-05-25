package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

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
