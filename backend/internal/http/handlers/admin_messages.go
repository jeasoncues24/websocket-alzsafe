package http

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
	"wsapi/internal/whatsapp"
)

type adminMessageDTO struct {
	ID          int                     `json:"id"`
	ReferenceID string                  `json:"reference_id,omitempty"`
	AccountID   string                  `json:"account_id"`
	To          string                  `json:"to"`
	Content     string                  `json:"content"`
	Status      string                  `json:"status"`
	ErrorReason *string                 `json:"error_reason,omitempty"`
	RetryCount  *int                    `json:"retry_count,omitempty"`
	Adjuntos    []domain.AttachmentInfo `json:"adjuntos,omitempty"`
	CreatedAt   time.Time               `json:"created_at"`
}

// AdminMessagesHandler maneja los endpoints admin de mensajes con dependencias inyectadas.
type AdminMessagesHandler struct {
	msgRepo       storage.MessagesRepository
	empresaStore  domain.EmpresaStoreInterface
	telefonoStore *storage.TelefonoStore
	manager       *whatsapp.Manager
}

func NewAdminMessagesHandler(
	msgRepo storage.MessagesRepository,
	empresaStore domain.EmpresaStoreInterface,
	telefonoStore *storage.TelefonoStore,
	manager *whatsapp.Manager,
) *AdminMessagesHandler {
	return &AdminMessagesHandler{
		msgRepo:       msgRepo,
		empresaStore:  empresaStore,
		telefonoStore: telefonoStore,
		manager:       manager,
	}
}

func (h *AdminMessagesHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limit := 50
	if l := query.Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	messages := []adminMessageDTO{}

	if h.msgRepo != nil && h.empresaStore != nil {
		accountID := query.Get("account_id")
		status := query.Get("status")

		appendMessages := func(empresaID int64, ruc string) {
			items, _, err := h.msgRepo.GetByEmpresa(empresaID, status, "", limit, 0)
			if err != nil {
				return
			}
			for _, m := range items {
				msg := adminMessageDTO{
					ID:          int(m.ID),
					ReferenceID: m.ReferenceID,
					AccountID:   ruc,
					To:          m.Destino,
					Content:     m.Contenido,
					Status:      string(m.Estado),
					Adjuntos:    m.Adjuntos,
					CreatedAt:   m.TiempoEnvio,
				}
				if m.ErrorReason != "" {
					er := m.ErrorReason
					msg.ErrorReason = &er
				}
				if m.RetryCount > 0 {
					rc := m.RetryCount
					msg.RetryCount = &rc
				}
				messages = append(messages, msg)
			}
		}

		if accountID != "" {
			if empresa, err := h.empresaStore.GetByRUC(accountID); err == nil && empresa != nil {
				appendMessages(empresa.ID, empresa.RUC)
			}
		} else {
			empresas, _, err := h.empresaStore.GetAll(1, 1000, "", nil)
			if err == nil {
				for i := range empresas {
					appendMessages(empresas[i].ID, empresas[i].RUC)
				}
				sort.Slice(messages, func(i, j int) bool {
					return messages[i].CreatedAt.After(messages[j].CreatedAt)
				})
				if len(messages) > limit {
					messages = messages[:limit]
				}
			}
		}
	}

	writeHandlerJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"messages": messages,
		"total":    len(messages),
	})
}

func (h *AdminMessagesHandler) RetryMessage(w http.ResponseWriter, r *http.Request) {
	access, ok := domain.GetPanelAccess(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	refID := r.PathValue("id")
	if refID == "" {
		writeAPIError(w, http.StatusBadRequest, "missing reference_id")
		return
	}

	if h.msgRepo == nil || h.telefonoStore == nil {
		writeAPIError(w, http.StatusInternalServerError, "service not available")
		return
	}

	msg, err := h.msgRepo.GetByReferenceID(refID)
	if err != nil || msg == nil {
		writeAPIError(w, http.StatusNotFound, "message not found")
		return
	}

	if !access.CanAccessEmpresa(msg.EmpresaID) {
		writeAPIError(w, http.StatusForbidden, "forbidden")
		return
	}

	if msg.Estado == domain.MessageStateSent || msg.Estado == domain.MessageStateDelivered {
		writeAPIError(w, http.StatusBadRequest, "message already sent")
		return
	}
	if len(msg.Adjuntos) > 0 {
		writeAPIError(w, http.StatusBadRequest, "media retry unsupported")
		return
	}

	telefono, err := h.telefonoStore.GetByID(msg.TelefonoID)
	if err != nil || telefono == nil {
		writeAPIError(w, http.StatusNotFound, "telefono not found")
		return
	}

	if telefono.Status != domain.TelefonoStatusActive {
		writeAPIError(w, http.StatusBadRequest, "session not active")
		return
	}

	if err := h.msgRepo.IncrementRetryCount(refID); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "error preparing retry")
		return
	}

	sendErr := whatsapp.SendRichMessageWithReference(r.Context(), h.manager, telefono.NumeroCompleto, msg.Destino, msg.Contenido, nil, refID)

	if sendErr != nil {
		_ = h.msgRepo.UpdateEstado(refID, domain.MessageStateFailed, sendErr.Error())
		writeHandlerJSON(w, http.StatusOK, map[string]interface{}{
			"ok":           false,
			"reference_id": refID,
			"estado":       string(domain.MessageStateFailed),
			"error":        sendErr.Error(),
		})
		return
	}

	_ = h.msgRepo.UpdateEstado(refID, domain.MessageStateSent, "")
	writeHandlerJSON(w, http.StatusOK, map[string]interface{}{
		"ok":           true,
		"reference_id": refID,
		"estado":       string(domain.MessageStateSent),
	})
}
