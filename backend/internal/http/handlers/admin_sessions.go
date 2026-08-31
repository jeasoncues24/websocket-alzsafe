package http

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"wsapi/internal/auth"
	"wsapi/internal/config"
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
	Total          int `json:"total"`
	Active         int `json:"active"`
	Disconnected   int `json:"disconnected"`
	Mismatch       int `json:"mismatch"`
	QRPending      int `json:"qr_pending"`
	Initializing   int `json:"initializing"`
	ClientOutdated int `json:"client_outdated"`
}

// sessionsSnapshotDTO es la foto completa del reporte de sesiones. La comparten el
// endpoint REST (GET /api/admin/sesiones) y el stream SSE, que consumen exactamente
// la misma ruta de construcción vía buildSessionsSnapshot.
type sessionsSnapshotDTO struct {
	OK       bool              `json:"ok"`
	Summary  sessionSummaryDTO `json:"summary"`
	Sessions []sessionInfoDTO  `json:"sessions"`
}

type AdminSessionsHandler struct {
	empresaStore  domain.EmpresaStoreInterface
	telefonoStore *storage.TelefonoStore
	manager       *whatsapp.Manager
	sessionStore  *storage.SessionStore
	jwtCfg        *config.JWTConfig

	hub *fleetHub
}

func NewAdminSessionsHandler(
	empresaStore domain.EmpresaStoreInterface,
	telefonoStore *storage.TelefonoStore,
	manager *whatsapp.Manager,
	sessionStore *storage.SessionStore,
	jwtCfg *config.JWTConfig,
) *AdminSessionsHandler {
	return &AdminSessionsHandler{
		empresaStore:  empresaStore,
		telefonoStore: telefonoStore,
		manager:       manager,
		sessionStore:  sessionStore,
		jwtCfg:        jwtCfg,
	}
}

func (h *AdminSessionsHandler) GenerateQRLink(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	telefonoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || telefonoID <= 0 {
		writeAPIError(w, http.StatusBadRequest, "telefono_id inválido")
		return
	}
	if h.telefonoStore == nil || h.jwtCfg == nil {
		writeAPIError(w, http.StatusInternalServerError, "servicio no disponible")
		return
	}
	phone, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil || phone == nil {
		writeAPIError(w, http.StatusNotFound, "Teléfono no encontrado")
		return
	}
	token, err := auth.GenerateQRLinkToken(phone.EmpresaID, phone.ID, h.jwtCfg.Secret)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error generando token")
		return
	}
	writeHandlerJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"token":      token,
		"phone_id":   phone.ID,
		"expires_in": 600,
	})
}

func (h *AdminSessionsHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	writeHandlerJSON(w, http.StatusOK, h.buildSessionsSnapshot())
}

// buildSessionsSnapshot construye la foto completa del reporte de sesiones cruzando
// el estado persistido (telefonos + empresa), el runtime en memoria (Manager) y el
// SessionStore. Es la única ruta de construcción: la comparten el endpoint REST y el
// stream SSE. Una sola query resuelve todos los teléfonos con su empresa (sin N+1).
func (h *AdminSessionsHandler) buildSessionsSnapshot() sessionsSnapshotDTO {
	sessions := []sessionInfoDTO{}
	if h.telefonoStore != nil {
		telefonos, err := h.telefonoStore.ListAllConEmpresa()
		if err == nil {
			for _, row := range telefonos {
				t := row.Telefono
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
					EmpresaID:        t.EmpresaID,
					EmpresaNombre:    row.EmpresaNombre,
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
	sort.SliceStable(sessions, func(i, j int) bool {
		// 1. Runtime conectado (online) primero
		if sessions[i].RuntimeConnected != sessions[j].RuntimeConnected {
			return sessions[i].RuntimeConnected
		}
		// 2. Status 'active' en BD primero
		iActive := sessions[i].Status == string(domain.TelefonoStatusActive)
		jActive := sessions[j].Status == string(domain.TelefonoStatusActive)
		if iActive != jActive {
			return iActive
		}
		// 3. Desempate por nombre de empresa / ID
		if sessions[i].EmpresaNombre != sessions[j].EmpresaNombre {
			return strings.ToLower(sessions[i].EmpresaNombre) < strings.ToLower(sessions[j].EmpresaNombre)
		}
		return sessions[i].TelefonoID < sessions[j].TelefonoID
	})

	return sessionsSnapshotDTO{
		OK:       true,
		Summary:  computeSessionSummary(sessions),
		Sessions: sessions,
	}
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
		case "client_outdated":
			s.ClientOutdated++
		}
		if sess.Mismatch && !sess.Reconnecting {
			s.Mismatch++
		}
	}
	return s
}
