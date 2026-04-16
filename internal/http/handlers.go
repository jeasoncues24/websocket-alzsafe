package http

import (
	"context"
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
	"wsapi/internal/whatsapp"
)

type Handler struct {
	manager         *whatsapp.Manager
	sessionStore    *storage.SessionStore
	msgRepo         storage.MessagesRepository
	empresaStore    domain.EmpresaStoreInterface
	broadcastWorker *whatsapp.BroadcastWorker
	broadcastStore  *storage.BroadcastStore
}

type inboundPayload struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type initSessionData struct {
	RUCEmpresa string `json:"ruc_empresa"`
}

type sessionReadyData struct {
	RUCEmpresa string `json:"ruc_empresa"`
}

type sessionDisconnectedData struct {
	RUCEmpresa string `json:"ruc_empresa"`
	Reason     string `json:"reason"`
}

type outboundPayload struct {
	Event string         `json:"event"`
	Data  map[string]any `json:"data"`
}

func NewHandler(manager *whatsapp.Manager, sessionStore *storage.SessionStore, msgRepo storage.MessagesRepository, empresaStore domain.EmpresaStoreInterface) *Handler {
	return &Handler{manager: manager, sessionStore: sessionStore, msgRepo: msgRepo, empresaStore: empresaStore}
}

func NewHandlerWithBroadcast(manager *whatsapp.Manager, sessionStore *storage.SessionStore, msgRepo storage.MessagesRepository, empresaStore domain.EmpresaStoreInterface, bw *whatsapp.BroadcastWorker, bs *storage.BroadcastStore) *Handler {
	return &Handler{manager: manager, sessionStore: sessionStore, msgRepo: msgRepo, empresaStore: empresaStore, broadcastWorker: bw, broadcastStore: bs}
}

func (h *Handler) HandleWS(w stdhttp.ResponseWriter, r *stdhttp.Request) {
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

		if err := h.processMessage(ctx, c, data); err != nil {
			_ = writeEvent(ctx, c, outboundPayload{
				Event: "error-event",
				Data:  map[string]any{"message": err.Error()},
			})
		}
	}
}

func (h *Handler) processMessage(ctx context.Context, c *websocket.Conn, data []byte) error {
	var payload inboundPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return errors.New("mensaje invalido")
	}

	switch payload.Event {
	case "init-session":
		return h.processInitSession(ctx, c, payload.Data)
	case "session-ready":
		return h.processSessionReady(ctx, c, payload.Data)
	case "session-disconnected":
		return h.processSessionDisconnected(ctx, c, payload.Data)
	case "session-logout":
		return h.processSessionLogout(ctx, c, payload.Data)
	default:
		return errors.New("evento no soportado")
	}
}

func (h *Handler) processInitSession(ctx context.Context, c *websocket.Conn, data json.RawMessage) error {
	var req initSessionData
	if err := json.Unmarshal(data, &req); err != nil {
		return errors.New("payload de init-session invalido")
	}

	ruc := strings.TrimSpace(req.RUCEmpresa)
	if len(ruc) < 8 {
		return errors.New("ruc_empresa invalido")
	}

	if err := whatsapp.StartSession(h.manager, ruc); err != nil {
		return err
	}

	h.sessionStore.SetInitializing(ruc)

	qr := whatsapp.GenerateQRCode(ruc)
	h.sessionStore.SetQRPending(ruc, qr)
	if err := writeEvent(ctx, c, outboundPayload{
		Event: "qr-" + ruc,
		Data: map[string]any{
			"message":  "Escanee el codigo QR para iniciar sesion.",
			"qrString": qr,
		},
	}); err != nil {
		return err
	}

	return writeEvent(ctx, c, outboundPayload{
		Event: "active-" + ruc,
		Data: map[string]any{
			"message":  "Sesion en proceso de inicializacion",
			"isActive": false,
		},
	})
}

func (h *Handler) processSessionReady(ctx context.Context, c *websocket.Conn, data json.RawMessage) error {
	var req sessionReadyData
	if err := json.Unmarshal(data, &req); err != nil {
		return errors.New("payload de session-ready invalido")
	}

	ruc := strings.TrimSpace(req.RUCEmpresa)
	if len(ruc) < 8 {
		return errors.New("ruc_empresa invalido")
	}

	h.sessionStore.SetActive(ruc)
	return writeEvent(ctx, c, outboundPayload{
		Event: "active-" + ruc,
		Data: map[string]any{
			"message":  "Sesion activa",
			"isActive": true,
		},
	})
}

func (h *Handler) processSessionDisconnected(ctx context.Context, c *websocket.Conn, data json.RawMessage) error {
	var req sessionDisconnectedData
	if err := json.Unmarshal(data, &req); err != nil {
		return errors.New("payload de session-disconnected invalido")
	}

	ruc := strings.TrimSpace(req.RUCEmpresa)
	if len(ruc) < 8 {
		return errors.New("ruc_empresa invalido")
	}

	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "disconnected"
	}

	h.manager.Delete(ruc)
	h.sessionStore.SetDisconnected(ruc, reason)

	return writeEvent(ctx, c, outboundPayload{
		Event: "active-" + ruc,
		Data: map[string]any{
			"message":       "Sesion desconectada",
			"isActive":      false,
			"reason":        reason,
			"requiresNewQR": false,
		},
	})
}

func (h *Handler) processSessionLogout(ctx context.Context, c *websocket.Conn, data json.RawMessage) error {
	var req sessionDisconnectedData
	if err := json.Unmarshal(data, &req); err != nil {
		return errors.New("payload de session-logout invalido")
	}

	ruc := strings.TrimSpace(req.RUCEmpresa)
	if len(ruc) < 8 {
		return errors.New("ruc_empresa invalido")
	}

	h.manager.Delete(ruc)
	h.sessionStore.SetDisconnected(ruc, "logout")

	return writeEvent(ctx, c, outboundPayload{
		Event: "active-" + ruc,
		Data: map[string]any{
			"message":       "Sesion cerrada por logout",
			"isActive":      false,
			"reason":        "logout",
			"requiresNewQR": true,
		},
	})
}

func writeEvent(ctx context.Context, c *websocket.Conn, payload outboundPayload) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.Write(ctx, websocket.MessageText, b)
}

// HandlePostMessage handles HTTP POST /message requests for direct message sending
func (h *Handler) HandlePostMessage(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get empresa filter from token
	filter, ok := domain.GetEmpresaFilter(r.Context(), r.Header.Get("X-Empresa-ID"))
	if !ok {
		w.WriteHeader(stdhttp.StatusUnauthorized)
		json.NewEncoder(w).Encode(domain.MessageResponse{
			OK:    false,
			Error: "UNAUTHORIZED",
		})
		return
	}

	// Parse request body
	var req domain.MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(stdhttp.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.MessageResponse{
			OK:      false,
			Error:   domain.ErrorCodeInvalidJSON,
			Details: "Invalid JSON in request body",
		})
		return
	}

	// Validate message request
	if validationErr := ValidateMessageRequest(&req); validationErr != nil {
		w.WriteHeader(stdhttp.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.MessageResponse{
			OK:      false,
			Error:   validationErr.Code,
			Details: validationErr.Message,
		})
		return
	}

	// Get RUC from filter
	ruc := ""
	if filter.IsRoot && filter.EmpresaID == nil {
		// Root sin header puede usar el ruc del body
		ruc = whatsapp.NormalizeAccountID(req.RUCEmpresa)
	} else {
		// Usuario normal debe usar su empresa del token
		empresa, err := domain.GetRUCFromContext(r.Context(), filter, h.empresaStore)
		if err != nil || empresa == "" {
			w.WriteHeader(stdhttp.StatusForbidden)
			json.NewEncoder(w).Encode(domain.MessageResponse{
				OK:      false,
				Error:   "NO_EMPRESA_ACCESS",
				Details: "No empresa access",
			})
			return
		}
		ruc = whatsapp.NormalizeAccountID(empresa)
	}

	// Check if session is active
	sessionState, ok := h.sessionStore.Get(ruc)
	if !ok || sessionState.Status != "active" || !sessionState.IsActive {
		w.WriteHeader(stdhttp.StatusForbidden)
		json.NewEncoder(w).Encode(domain.MessageResponse{
			OK:      false,
			Error:   domain.ErrorCodeSessionNotActive,
			Details: "Session is not active for this empresa",
		})
		return
	}

	// Create message
	message := domain.NewMessage(ruc, strings.TrimSpace(req.Destino), strings.TrimSpace(req.Mensaje))

	// [QUÉ] Persistir el mensaje en DB antes de responder al cliente.
	// [POR QUÉ] Guardamos primero (estado 'pending') para garantizar trazabilidad incluso si
	// el envío posterior falla o la conexión WhatsApp cae. Fail-fast: si no podemos persistir,
	// no prometemos un 202 que no podremos rastrear.
	// msgRepo puede ser nil en entornos de test/desarrollo sin DB configurada.
	if h.msgRepo != nil {
		if err := h.msgRepo.Create(message); err != nil {
			w.WriteHeader(stdhttp.StatusInternalServerError)
			json.NewEncoder(w).Encode(domain.MessageResponse{
				OK:      false,
				Error:   "PERSISTENCE_ERROR",
				Details: "Failed to persist message before sending",
			})
			return
		}
	}

	w.WriteHeader(stdhttp.StatusAccepted)

	var usuarioID *int64
	var empresaNombre string

	if claims, ok := domain.GetTokenClaims(r.Context()); ok {
		usuarioID = &claims.UserID
		if claims.EmpresaNombre != nil {
			empresaNombre = *claims.EmpresaNombre
		}
	}

	json.NewEncoder(w).Encode(domain.MessageResponse{
		OK:            true,
		Message:       "Message accepted for processing",
		ReferenceID:   message.ReferenceID,
		UsuarioID:     usuarioID,
		RUCEmpresa:    ruc,
		EmpresaNombre: empresaNombre,
		SessionID:     sessionState.SessionID,
	})
}

// HandleGetMessages handles HTTP GET /messages for message listing and audit.
// [QUÉ] Endpoint de consulta/auditoría: lista mensajes de una empresa con paginación y filtros.
// [POR QUÉ] Permite al frontend mostrar el historial y al operador auditar envíos.
//
// Query params:
//   - page (opcional, default 1)
//   - limit (opcional, default 20, max 100)
//   - desde (opcional, formato YYYY-MM-DD)
//   - hasta (opcional, formato YYYY-MM-DD)
//   - telefono (opcional, búsqueda parcial)
//   - empresa_id (opcional, solo para super_admin)
//   - estado (opcional: pending, sent, delivered, failed, rejected)
func (h *Handler) HandleGetMessages(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.msgRepo == nil {
		w.WriteHeader(stdhttp.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(domain.MessagesListResponse{
			OK:    false,
			Error: "SERVICE_UNAVAILABLE",
		})
		return
	}

	filter, ok := domain.GetEmpresaFilter(r.Context(), r.Header.Get("X-Empresa-ID"))
	if !ok {
		w.WriteHeader(stdhttp.StatusUnauthorized)
		json.NewEncoder(w).Encode(domain.MessagesListResponse{
			OK:    false,
			Error: "UNAUTHORIZED",
		})
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
	empresaIDStr := strings.TrimSpace(r.URL.Query().Get("empresa_id"))
	estado := strings.TrimSpace(r.URL.Query().Get("estado"))

	if empresaIDStr != "" && !filter.IsRoot {
		empresaIDStr = ""
	}

	if estado != "" {
		switch domain.MessageState(estado) {
		case domain.MessageStatePending, domain.MessageStateSent, domain.MessageStateDelivered, domain.MessageStateFailed, domain.MessageStateRejected:
		default:
			w.WriteHeader(stdhttp.StatusBadRequest)
			json.NewEncoder(w).Encode(domain.MessagesListResponse{
				OK:      false,
				Error:   "INVALID_STATE_FILTER",
				Details: "estado must be one of: pending, sent, delivered, failed, rejected",
			})
			return
		}
	}

	var targetEmpresaID *int64
	if empresaIDStr != "" && filter.IsRoot {
		if id, err := strconv.ParseInt(empresaIDStr, 10, 64); err == nil && id > 0 {
			targetEmpresaID = &id
		}
	}

	var messages []domain.Message
	var total int
	var err error

	if filter.IsRoot && targetEmpresaID == nil && filter.EmpresaID == nil {
		rucs, err := domain.GetAllEmpresaRUCs(h.empresaStore)
		if err != nil {
			w.WriteHeader(stdhttp.StatusInternalServerError)
			json.NewEncoder(w).Encode(domain.MessagesListResponse{
				OK:    false,
				Error: "QUERY_ERROR",
			})
			return
		}

		allMessages := []domain.Message{}
		totalCount := 0
		for _, ruc := range rucs {
			if desdeStr != "" || hastaStr != "" {
				var startDate, endDate time.Time
				if desdeStr != "" {
					startDate, err = time.Parse("2006-01-02", desdeStr)
					if err != nil {
						w.WriteHeader(stdhttp.StatusBadRequest)
						json.NewEncoder(w).Encode(domain.MessagesListResponse{
							OK:      false,
							Error:   "INVALID_DATE_FORMAT",
							Details: "desde must be in YYYY-MM-DD format",
						})
						return
					}
				}
				if hastaStr != "" {
					endDate, err = time.Parse("2006-01-02", hastaStr)
					if err != nil {
						w.WriteHeader(stdhttp.StatusBadRequest)
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
				msgs, cnt, err := h.msgRepo.GetByEmpresaAndDateRange(ruc, startDate, endDate, estado, telefono, limit, offset)
				if err != nil {
					continue
				}
				totalCount += cnt
				allMessages = append(allMessages, msgs...)
			} else {
				msgs, cnt, err := h.msgRepo.GetByEmpresa(ruc, estado, telefono, limit, offset)
				if err != nil {
					continue
				}
				totalCount += cnt
				allMessages = append(allMessages, msgs...)
			}
		}
		messages = allMessages
		total = totalCount
	} else {
		empresa, err := domain.GetRUCFromContext(r.Context(), filter, h.empresaStore)
		if err != nil || empresa == "" {
			w.WriteHeader(stdhttp.StatusForbidden)
			json.NewEncoder(w).Encode(domain.MessagesListResponse{
				OK:      false,
				Error:   "NO_EMPRESA_ACCESS",
				Details: "No empresa access",
			})
			return
		}

		sessionState, ok := h.sessionStore.Get(empresa)
		if !ok || sessionState.Status != "active" || !sessionState.IsActive {
			w.WriteHeader(stdhttp.StatusForbidden)
			json.NewEncoder(w).Encode(domain.MessagesListResponse{
				OK:      false,
				Error:   domain.ErrorCodeSessionNotActive,
				Details: "Session is not active for this empresa",
			})
			return
		}

		if desdeStr != "" || hastaStr != "" {
			var startDate, endDate time.Time

			if desdeStr != "" {
				startDate, err = time.Parse("2006-01-02", desdeStr)
				if err != nil {
					w.WriteHeader(stdhttp.StatusBadRequest)
					json.NewEncoder(w).Encode(domain.MessagesListResponse{
						OK:      false,
						Error:   "INVALID_DATE_FORMAT",
						Details: "desde must be in YYYY-MM-DD format",
					})
					return
				}
			}

			if hastaStr != "" {
				endDate, err = time.Parse("2006-01-02", hastaStr)
				if err != nil {
					w.WriteHeader(stdhttp.StatusBadRequest)
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
			messages, total, err = h.msgRepo.GetByEmpresaAndDateRange(empresa, startDate, endDate, estado, telefono, limit, offset)
		} else {
			messages, total, err = h.msgRepo.GetByEmpresa(empresa, estado, telefono, limit, offset)
		}
	}

	if err != nil {
		w.WriteHeader(stdhttp.StatusInternalServerError)
		json.NewEncoder(w).Encode(domain.MessagesListResponse{
			OK:    false,
			Error: "QUERY_ERROR",
		})
		return
	}

	var usuarioID *int64
	var empresaNombre string

	if claims, ok := domain.GetTokenClaims(r.Context()); ok {
		usuarioID = &claims.UserID
		if claims.EmpresaNombre != nil {
			empresaNombre = *claims.EmpresaNombre
		}
	}

	totalPages := 0
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}

	w.WriteHeader(stdhttp.StatusOK)
	json.NewEncoder(w).Encode(domain.MessagesListResponse{
		OK:            true,
		Messages:      messages,
		Total:         total,
		Page:          page,
		Limit:         limit,
		TotalPages:    totalPages,
		UsuarioID:     usuarioID,
		RUCEmpresa:    "",
		EmpresaNombre: empresaNombre,
	})
}

// parseIntParam convierte un string de query param a int con un valor por defecto.
// [APRENDE] Los query params siempre llegan como string; hay que convertirlos manualmente.
func parseIntParam(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return v
}

// HandlePostBroadcast handles HTTP POST /broadcast requests for mass broadcasting.
// Validates the payload and accepts the request for asynchronous processing (Story 3.2).
// In this story, only validation and 202 acceptance are implemented; actual processing
// (worker pool, per-recipient results, persistence) belong to Stories 3.2 and 3.3.
func (h *Handler) HandlePostBroadcast(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get empresa filter from token
	filter, ok := domain.GetEmpresaFilter(r.Context(), r.Header.Get("X-Empresa-ID"))
	if !ok {
		w.WriteHeader(stdhttp.StatusUnauthorized)
		json.NewEncoder(w).Encode(domain.BroadcastResponse{
			OK:    false,
			Error: "UNAUTHORIZED",
		})
		return
	}

	var rawReq struct {
		RUCEmpresa    string          `json:"ruc_empresa"`
		ListaDifusion json.RawMessage `json:"lista_difusion"`
	}

	if err := json.NewDecoder(r.Body).Decode(&rawReq); err != nil {
		w.WriteHeader(stdhttp.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.BroadcastResponse{
			OK:      false,
			Error:   domain.ErrorCodeInvalidJSON,
			Details: "Invalid JSON in request body",
		})
		return
	}

	if len(rawReq.ListaDifusion) == 0 {
		w.WriteHeader(stdhttp.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.BroadcastResponse{
			OK:      false,
			Error:   domain.ErrorCodeValidation,
			Details: "lista_difusion is required and must be a non-empty array",
		})
		return
	}

	var listaDifusion []domain.BroadcastItem
	if err := json.Unmarshal(rawReq.ListaDifusion, &listaDifusion); err != nil {
		w.WriteHeader(stdhttp.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.BroadcastResponse{
			OK:      false,
			Error:   domain.ErrorCodeValidation,
			Details: "lista_difusion must be a non-empty array",
		})
		return
	}

	// Get RUC from filter
	ruc := ""
	if filter.IsRoot && filter.EmpresaID == nil {
		ruc = whatsapp.NormalizeAccountID(rawReq.RUCEmpresa)
	} else {
		empresa, err := domain.GetRUCFromContext(r.Context(), filter, h.empresaStore)
		if err != nil || empresa == "" {
			w.WriteHeader(stdhttp.StatusForbidden)
			json.NewEncoder(w).Encode(domain.BroadcastResponse{
				OK:      false,
				Error:   "NO_EMPRESA_ACCESS",
				Details: "No empresa access",
			})
			return
		}
		ruc = whatsapp.NormalizeAccountID(empresa)
	}

	req := domain.BroadcastRequest{
		RUCEmpresa:    ruc,
		ListaDifusion: listaDifusion,
	}

	if validationErr := ValidateBroadcastRequest(&req); validationErr != nil {
		w.WriteHeader(stdhttp.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.BroadcastResponse{
			OK:      false,
			Error:   validationErr.Code,
			Details: validationErr.Message,
		})
		return
	}

	sessionState, ok := h.sessionStore.Get(ruc)
	if !ok || sessionState.Status != "active" || !sessionState.IsActive {
		w.WriteHeader(stdhttp.StatusForbidden)
		json.NewEncoder(w).Encode(domain.BroadcastResponse{
			OK:      false,
			Error:   domain.ErrorCodeSessionNotActive,
			Details: "Session is not active for this empresa",
		})
		return
	}

	referenceID := uuid.New().String()

	if h.broadcastWorker != nil && h.broadcastStore != nil {
		job := &domain.BroadcastJob{
			ReferenceID: referenceID,
			RUCEmpresa:  ruc,
			Total:       len(req.ListaDifusion),
		}
		h.broadcastStore.Create(job)

		resultChan := make(chan whatsapp.BroadcastResult, len(req.ListaDifusion))
		wJob := whatsapp.BroadcastJob{
			ReferenceID: referenceID,
			RUCEmpresa:  ruc,
			Items:       req.ListaDifusion,
			ResultChan:  resultChan,
		}
		h.broadcastWorker.SubmitAsync(wJob)

		go func() {
			for result := range resultChan {
				domainResult := domain.BroadcastResult{
					Index:      result.Index,
					Destino:    result.Destino,
					RUCEmpresa: ruc,
					State:      result.State,
					Error:      result.Error,
					Timestamp:  result.Timestamp,
				}
				h.broadcastStore.AppendResult(referenceID, domainResult)
			}

			h.broadcastStore.UpdateStatus(referenceID, domain.BroadcastStatusCompleted)
		}()
	}

	var empresaNombre string
	if claims, ok := domain.GetTokenClaims(r.Context()); ok {
		if claims.EmpresaNombre != nil {
			empresaNombre = *claims.EmpresaNombre
		}
	}

	w.WriteHeader(stdhttp.StatusAccepted)
	json.NewEncoder(w).Encode(domain.BroadcastResponse{
		OK:            true,
		ReferenceID:   referenceID,
		Total:         len(req.ListaDifusion),
		RUCEmpresa:    ruc,
		EmpresaNombre: empresaNombre,
	})
}

// HandleGetBroadcast handles HTTP GET /broadcast/{reference_id} requests
func (h *Handler) HandleGetBroadcast(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	w.Header().Set("Content-Type", "application/json")

	referenceID := strings.TrimPrefix(r.URL.Path, "/broadcast/")
	if referenceID == "" || referenceID == r.URL.Path {
		w.WriteHeader(stdhttp.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.BroadcastDetailResponse{
			OK:    false,
			Error: "MISSING_REFERENCE_ID",
		})
		return
	}

	if h.broadcastStore == nil {
		w.WriteHeader(stdhttp.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(domain.BroadcastDetailResponse{
			OK:    false,
			Error: "SERVICE_UNAVAILABLE",
		})
		return
	}

	job, ok := h.broadcastStore.Get(referenceID)
	if !ok {
		w.WriteHeader(stdhttp.StatusNotFound)
		json.NewEncoder(w).Encode(domain.BroadcastDetailResponse{
			OK:    false,
			Error: "BROADCAST_NOT_FOUND",
		})
		return
	}

	sessionState, ok := h.sessionStore.Get(job.RUCEmpresa)
	if !ok || sessionState.Status != "active" || !sessionState.IsActive {
		w.WriteHeader(stdhttp.StatusForbidden)
		json.NewEncoder(w).Encode(domain.BroadcastDetailResponse{
			OK:    false,
			Error: "SESSION_NOT_ACTIVE",
		})
		return
	}

	w.WriteHeader(stdhttp.StatusOK)
	json.NewEncoder(w).Encode(domain.BroadcastDetailResponse{
		OK:          true,
		ReferenceID: job.ReferenceID,
		RUCEmpresa:  job.RUCEmpresa,
		Total:       job.Total,
		Status:      string(job.Status),
		Results:     job.Results,
	})
}
