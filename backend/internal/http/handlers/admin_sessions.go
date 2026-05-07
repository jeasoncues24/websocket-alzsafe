package http

import (
	"encoding/json"
	"net/http"
	"time"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
	"wsapi/internal/whatsapp"
)

type sessionEventDTO struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Details   string    `json:"details,omitempty"`
}

type sessionInfoDTO struct {
	TelefonoID       int64             `json:"telefono_id"`
	EmpresaID        int64             `json:"empresa_id"`
	EmpresaNombre    string            `json:"empresa_nombre"`
	AccountID        string            `json:"account_id"`
	Status           string            `json:"status"`
	RuntimeConnected bool              `json:"runtime_connected"`
	Mismatch         bool              `json:"mismatch"`
	Reconnecting     bool              `json:"reconnecting,omitempty"`
	QRString         string            `json:"qr_string,omitempty"`
	LastConnected    *time.Time        `json:"last_connected,omitempty"`
	UpdatedAt        time.Time         `json:"updated_at"`
	Events           []sessionEventDTO `json:"events,omitempty"`
}

type sessionSummaryDTO struct {
	Total        int `json:"total"`
	Active       int `json:"active"`
	Disconnected int `json:"disconnected"`
	Mismatch     int `json:"mismatch"`
	QRPending    int `json:"qr_pending"`
	Initializing int `json:"initializing"`
}

type AdminSessionsHandler struct {
	empresaStore  domain.EmpresaStoreInterface
	telefonoStore *storage.TelefonoStore
	manager       *whatsapp.Manager
	sessionStore  *storage.SessionStore
}

func NewAdminSessionsHandler(
	empresaStore domain.EmpresaStoreInterface,
	telefonoStore *storage.TelefonoStore,
	manager *whatsapp.Manager,
	sessionStore *storage.SessionStore,
) *AdminSessionsHandler {
	return &AdminSessionsHandler{
		empresaStore:  empresaStore,
		telefonoStore: telefonoStore,
		manager:       manager,
		sessionStore:  sessionStore,
	}
}

func (h *AdminSessionsHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	sessions := []sessionInfoDTO{}
	if h.empresaStore != nil && h.telefonoStore != nil {
		empresas, _, err := h.empresaStore.GetAll(1, 1000, "", nil)
		if err == nil {
			for i := range empresas {
				telefonos, err := h.telefonoStore.GetByEmpresa(empresas[i].ID)
				if err != nil {
					continue
				}
				nombre := empresas[i].NombreComercial
				if nombre == "" {
					nombre = empresas[i].Nombre
				}
				for _, t := range telefonos {
					accountID := whatsapp.NormalizeAccountID(t.NumeroCompleto)
					runtimeConnected := false
					if h.manager != nil {
						if client, ok := h.manager.Get(accountID); ok && client != nil && client.IsConnected() {
							runtimeConnected = true
						}
					}
					var events []sessionEventDTO
					var storeStatus string
					if h.sessionStore != nil {
						if state, ok := h.sessionStore.Get(t.NumeroCompleto); ok {
							storeStatus = state.Status
							last := state.Events
							if len(last) > 10 {
								last = last[len(last)-10:]
							}
							for _, e := range last {
								events = append(events, sessionEventDTO{
									Timestamp: e.Timestamp,
									Type:      e.Type,
									Details:   e.Details,
								})
							}
						}
					}
					reconnecting := (t.Status == domain.TelefonoStatusActive) &&
						!runtimeConnected &&
						(storeStatus == "initializing" || storeStatus == "qr_pending")
					mismatch := (t.Status == domain.TelefonoStatusActive) != runtimeConnected && !reconnecting
					qr := ""
					if t.Status == domain.TelefonoStatusQRPending {
						qr = t.QRString
					}
					sessions = append(sessions, sessionInfoDTO{
						TelefonoID:       t.ID,
						EmpresaID:        empresas[i].ID,
						EmpresaNombre:    nombre,
						AccountID:        t.NumeroCompleto,
						Status:           string(t.Status),
						RuntimeConnected: runtimeConnected,
						Mismatch:         mismatch,
						Reconnecting:     reconnecting,
						QRString:         qr,
						LastConnected:    t.LastConnected,
						UpdatedAt:        t.UpdatedAt,
						Events:           events,
					})
				}
			}
		}
	}
	summary := computeSessionSummary(sessions)
	writeHandlerJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"summary":  summary,
		"sessions": sessions,
	})
}

func (h *AdminSessionsHandler) PostSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AccountID string `json:"account_id"`
		Action    string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Action == "disconnect" && h.telefonoStore != nil {
		if telefono, err := h.telefonoStore.GetByNumeroCompleto(req.AccountID); err == nil && telefono != nil {
			_ = h.telefonoStore.SetDisconnected(telefono.ID)
			if h.sessionStore != nil {
				h.sessionStore.AppendEvent(req.AccountID, "disconnected", "manual_admin")
			}
		}
	}

	writeHandlerJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"status": "ok",
	})
}

func computeSessionSummary(sessions []sessionInfoDTO) sessionSummaryDTO {
	s := sessionSummaryDTO{Total: len(sessions)}
	for _, sess := range sessions {
		switch sess.Status {
		case "active":
			s.Active++
		case "disconnected":
			s.Disconnected++
		case "qr_pending":
			s.QRPending++
		case "initializing":
			s.Initializing++
		}
		if sess.Mismatch && !sess.Reconnecting {
			s.Mismatch++
		}
	}
	return s
}
