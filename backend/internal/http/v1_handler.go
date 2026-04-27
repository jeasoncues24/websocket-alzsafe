package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"wsapi/internal/domain"
	"wsapi/internal/storage"
	"wsapi/internal/whatsapp"
)

// V1Handler sirve los endpoints /v1/* autenticados con JWT de empresa.
// Cada handler extrae EmpresaJWTClaims del contexto (inyectadas por EmpresaAuthMiddleware).
type V1Handler struct {
	manager         *whatsapp.Manager
	sessionStore    *storage.SessionStore
	msgRepo         storage.MessagesRepository
	telefonoStore   *storage.TelefonoStore
	broadcastWorker *whatsapp.BroadcastWorker
	broadcastStore  *storage.BroadcastStore
}

func NewV1Handler(
	manager *whatsapp.Manager,
	sessionStore *storage.SessionStore,
	msgRepo storage.MessagesRepository,
	telefonoStore *storage.TelefonoStore,
	broadcastWorker *whatsapp.BroadcastWorker,
	broadcastStore *storage.BroadcastStore,
) *V1Handler {
	return &V1Handler{
		manager:         manager,
		sessionStore:    sessionStore,
		msgRepo:         msgRepo,
		telefonoStore:   telefonoStore,
		broadcastWorker: broadcastWorker,
		broadcastStore:  broadcastStore,
	}
}

// V1PostMessage maneja POST /v1/messages
func (h *V1Handler) V1PostMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := domain.GetEmpresaJWTClaims(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(domain.MessageResponse{OK: false, Error: "UNAUTHORIZED"})
		return
	}

	var req domain.MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.MessageResponse{
			OK:      false,
			Error:   domain.ErrorCodeInvalidJSON,
			Details: "Invalid JSON in request body",
		})
		return
	}

	if validationErr := ValidateMessageRequest(&req); validationErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.MessageResponse{
			OK:      false,
			Error:   validationErr.Code,
			Details: validationErr.Message,
		})
		return
	}

	// Verificar que el telefono pertenece a la empresa del JWT
	belongs, err := h.telefonoStore.BelongsToEmpresa(req.TelefonoID, claims.EmpresaID)
	if err != nil || !belongs {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(domain.MessageResponse{
			OK:      false,
			Error:   "TELEFONO_NOT_OWNED",
			Details: "Telefono does not belong to this empresa",
		})
		return
	}

	// Obtener el numero_completo del teléfono para lookup de sesión
	telefono, err := h.telefonoStore.GetByID(req.TelefonoID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(domain.MessageResponse{
			OK:      false,
			Error:   "TELEFONO_NOT_FOUND",
			Details: "Telefono not found",
		})
		return
	}

	sessionKey := whatsapp.NormalizeAccountID(telefono.NumeroCompleto)
	sessionState, ok := h.sessionStore.Get(sessionKey)
	if !ok || sessionState.Status != "active" || !sessionState.IsActive {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(domain.MessageResponse{
			OK:      false,
			Error:   domain.ErrorCodeSessionNotActive,
			Details: "Session is not active for this telefono",
		})
		return
	}

	message := domain.NewMessage(claims.EmpresaID, req.TelefonoID, strings.TrimSpace(req.Destino), strings.TrimSpace(req.Mensaje))
	infos, infoErr := buildAttachmentInfos(req.Adjuntos)
	if infoErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.MessageResponse{OK: false, Error: "INVALID_ATTACHMENT", Details: "Adjunto inválido"})
		return
	}
	message.Adjuntos = infos

	if h.msgRepo != nil {
		if err := h.msgRepo.Create(message); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(domain.MessageResponse{
				OK:      false,
				Error:   "PERSISTENCE_ERROR",
				Details: "Failed to persist message before sending",
			})
			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(domain.MessageResponse{
		OK:            true,
		Message:       "Message accepted for processing",
		ReferenceID:   message.ReferenceID,
		EmpresaID:     message.EmpresaID,
		EmpresaNombre: claims.EmpresaNombre,
		SessionID:     sessionState.SessionID,
	})
}

// V1GetMessages maneja GET /v1/messages
func (h *V1Handler) V1GetMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.msgRepo == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(domain.MessagesListResponse{OK: false, Error: "SERVICE_UNAVAILABLE"})
		return
	}

	claims, ok := domain.GetEmpresaJWTClaims(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(domain.MessagesListResponse{OK: false, Error: "UNAUTHORIZED"})
		return
	}

	page := parseIntParam(r.URL.Query().Get("page"), 1)
	limit := parseIntParam(r.URL.Query().Get("limit"), 20)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	desdeStr := strings.TrimSpace(r.URL.Query().Get("desde"))
	hastaStr := strings.TrimSpace(r.URL.Query().Get("hasta"))
	telefono := strings.TrimSpace(r.URL.Query().Get("telefono"))
	estado := strings.TrimSpace(r.URL.Query().Get("estado"))

	if estado != "" {
		switch domain.MessageState(estado) {
		case domain.MessageStatePending, domain.MessageStateSent, domain.MessageStateDelivered, domain.MessageStateFailed, domain.MessageStateRejected:
		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(domain.MessagesListResponse{
				OK:      false,
				Error:   "INVALID_STATE_FILTER",
				Details: "estado must be one of: pending, sent, delivered, failed, rejected",
			})
			return
		}
	}

	var messages []domain.Message
	var total int

	if desdeStr != "" || hastaStr != "" {
		var startDate, endDate time.Time
		var parseErr error

		if desdeStr != "" {
			startDate, parseErr = time.Parse("2006-01-02", desdeStr)
			if parseErr != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(domain.MessagesListResponse{
					OK:      false,
					Error:   "INVALID_DATE_FORMAT",
					Details: "desde must be in YYYY-MM-DD format",
				})
				return
			}
		}
		if hastaStr != "" {
			endDate, parseErr = time.Parse("2006-01-02", hastaStr)
			if parseErr != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(domain.MessagesListResponse{
					OK:      false,
					Error:   "INVALID_DATE_FORMAT",
					Details: "hasta must be in YYYY-MM-DD format",
				})
				return
			}
			endDate = endDate.Add(24*time.Hour - time.Millisecond)
		}
		if desdeStr != "" && hastaStr == "" {
			endDate = startDate.Add(24*time.Hour - time.Millisecond)
		}

		var queryErr error
		messages, total, queryErr = h.msgRepo.GetByEmpresaAndDateRange(claims.EmpresaID, startDate, endDate, estado, telefono, limit, offset)
		if queryErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(domain.MessagesListResponse{OK: false, Error: "QUERY_ERROR"})
			return
		}
	} else {
		var queryErr error
		messages, total, queryErr = h.msgRepo.GetByEmpresa(claims.EmpresaID, estado, telefono, limit, offset)
		if queryErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(domain.MessagesListResponse{OK: false, Error: "QUERY_ERROR"})
			return
		}
	}

	json.NewEncoder(w).Encode(domain.MessagesListResponse{
		OK:       true,
		Messages: messages,
		Total:    total,
		Page:     page,
		Limit:    limit,
	})
}

// V1PostBroadcast maneja POST /v1/broadcasts
func (h *V1Handler) V1PostBroadcast(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := domain.GetEmpresaJWTClaims(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(domain.BroadcastResponse{OK: false, Error: "UNAUTHORIZED"})
		return
	}

	var req domain.BroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.BroadcastResponse{
			OK:      false,
			Error:   domain.ErrorCodeInvalidJSON,
			Details: "Invalid JSON in request body",
		})
		return
	}

	if validationErr := ValidateBroadcastRequest(&req); validationErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.BroadcastResponse{
			OK:      false,
			Error:   validationErr.Code,
			Details: validationErr.Message,
		})
		return
	}

	belongs, err := h.telefonoStore.BelongsToEmpresa(req.TelefonoID, claims.EmpresaID)
	if err != nil || !belongs {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(domain.BroadcastResponse{
			OK:      false,
			Error:   "TELEFONO_NOT_OWNED",
			Details: "Telefono does not belong to this empresa",
		})
		return
	}

	telefono, err := h.telefonoStore.GetByID(req.TelefonoID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(domain.BroadcastResponse{
			OK:      false,
			Error:   "TELEFONO_NOT_FOUND",
			Details: "Telefono not found",
		})
		return
	}

	sessionKey := whatsapp.NormalizeAccountID(telefono.NumeroCompleto)
	sessionState, ok := h.sessionStore.Get(sessionKey)
	if !ok || sessionState.Status != "active" || !sessionState.IsActive {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(domain.BroadcastResponse{
			OK:      false,
			Error:   domain.ErrorCodeSessionNotActive,
			Details: "Session is not active for this telefono",
		})
		return
	}

	referenceID := uuid.New().String()

	job := &domain.BroadcastJob{
		ReferenceID: referenceID,
		EmpresaID:   claims.EmpresaID,
		TelefonoID:  req.TelefonoID,
		Adjuntos:    nil,
		Total:       len(req.ListaDifusion),
	}
	infos, infoErr := buildAttachmentInfos(req.Adjuntos)
	if infoErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.BroadcastResponse{OK: false, Error: "INVALID_ATTACHMENT", Details: "Adjunto inválido"})
		return
	}
	job.Adjuntos = infos
	h.broadcastStore.Create(job)

	resultChan := make(chan whatsapp.BroadcastResult, len(req.ListaDifusion))
	wJob := whatsapp.BroadcastJob{
		ReferenceID: referenceID,
		RUCEmpresa:  sessionKey,
		Attachments: req.Adjuntos,
		Items:       req.ListaDifusion,
		ResultChan:  resultChan,
	}

	go h.v1CollectBroadcastResults(referenceID, claims.EmpresaID, resultChan)
	h.broadcastWorker.SubmitAsync(wJob)

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(domain.BroadcastResponse{
		OK:            true,
		ReferenceID:   referenceID,
		Total:         len(req.ListaDifusion),
		EmpresaID:     claims.EmpresaID,
		EmpresaNombre: claims.EmpresaNombre,
	})
}

// V1GetBroadcast maneja GET /v1/broadcasts/{reference_id}
func (h *V1Handler) V1GetBroadcast(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := domain.GetEmpresaJWTClaims(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(domain.BroadcastDetailResponse{OK: false, Error: "UNAUTHORIZED"})
		return
	}

	// Extraer reference_id del path: /v1/broadcasts/{reference_id}
	path := r.URL.Path
	referenceID := ""
	if idx := strings.LastIndex(path, "/broadcasts/"); idx >= 0 {
		referenceID = strings.TrimSpace(path[idx+len("/broadcasts/"):])
	}

	if referenceID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.BroadcastDetailResponse{OK: false, Error: "MISSING_REFERENCE_ID"})
		return
	}

	job, ok := h.broadcastStore.Get(referenceID)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(domain.BroadcastDetailResponse{OK: false, Error: "BROADCAST_NOT_FOUND"})
		return
	}

	// Verificar que el job pertenece a la empresa del JWT
	if job.EmpresaID != claims.EmpresaID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(domain.BroadcastDetailResponse{OK: false, Error: "FORBIDDEN"})
		return
	}

	json.NewEncoder(w).Encode(domain.BroadcastDetailResponse{
		OK:          true,
		ReferenceID: job.ReferenceID,
		EmpresaID:   job.EmpresaID,
		TelefonoID:  job.TelefonoID,
		Adjuntos:    job.Adjuntos,
		Total:       job.Total,
		Status:      string(job.Status),
		Results:     job.Results,
	})
}

// V1ListTelefonos maneja GET /v1/telefonos
func (h *V1Handler) V1ListTelefonos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := domain.GetEmpresaJWTClaims(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(domain.TelefonosListResponse{OK: false, Error: "UNAUTHORIZED"})
		return
	}

	telefonos, err := h.telefonoStore.GetByEmpresa(claims.EmpresaID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(domain.TelefonosListResponse{OK: false, Error: "QUERY_ERROR"})
		return
	}

	json.NewEncoder(w).Encode(domain.TelefonosListResponse{
		OK:        true,
		Total:     len(telefonos),
		Telefonos: telefonos,
	})
}

// V1GetTelefono maneja GET /v1/telefonos/{id}
func (h *V1Handler) V1GetTelefono(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := domain.GetEmpresaJWTClaims(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(domain.TelefonoResponse{OK: false, Error: "UNAUTHORIZED"})
		return
	}

	path := r.URL.Path
	idStr := ""
	if idx := strings.LastIndex(path, "/telefonos/"); idx >= 0 {
		idStr = strings.TrimSpace(path[idx+len("/telefonos/"):])
	}

	telefonoID, convErr := strconv.ParseInt(idStr, 10, 64)
	if convErr != nil || telefonoID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.TelefonoResponse{OK: false, Error: "INVALID_ID"})
		return
	}

	belongs, err := h.telefonoStore.BelongsToEmpresa(telefonoID, claims.EmpresaID)
	if err != nil || !belongs {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(domain.TelefonoResponse{OK: false, Error: "FORBIDDEN"})
		return
	}

	telefono, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(domain.TelefonoResponse{OK: false, Error: "NOT_FOUND"})
		return
	}

	json.NewEncoder(w).Encode(domain.TelefonoResponse{
		OK:       true,
		Telefono: telefono,
	})
}

// v1CollectBroadcastResults recoge los resultados del worker y actualiza el broadcastStore.
func (h *V1Handler) v1CollectBroadcastResults(referenceID string, empresaID int64, resultChan <-chan whatsapp.BroadcastResult) {
	for result := range resultChan {
		domainResult := domain.BroadcastResult{
			Index:     result.Index,
			Destino:   result.Destino,
			EmpresaID: empresaID,
			State:     result.State,
			Error:     result.Error,
			Timestamp: result.Timestamp,
		}
		h.broadcastStore.AppendResult(referenceID, domainResult)
	}
	h.broadcastStore.UpdateStatus(referenceID, domain.BroadcastStatusCompleted)
}
