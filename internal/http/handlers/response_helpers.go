package http

import (
	"encoding/json"
	"net/http"
)

type handlerAPIErrorResponse struct {
	OK      bool   `json:"ok"`
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func writeHandlerJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeAPIError(w http.ResponseWriter, status int, message string) {
	writeHandlerJSON(w, status, handlerAPIErrorResponse{OK: false, Error: message, Message: message})
}
