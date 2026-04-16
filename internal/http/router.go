package http

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"wsapi/internal/config"
	"wsapi/internal/domain"
	httpHandlers "wsapi/internal/http/handlers"
	"wsapi/internal/http/middleware"
	"wsapi/internal/metrics"
	"wsapi/internal/storage"
	"wsapi/internal/whatsapp"
)

func NewRouter() http.Handler {
	manager := whatsapp.NewManager()
	sessionStore := storage.NewSessionStore()

	broadcastWorker := whatsapp.NewBroadcastWorker(whatsapp.DefaultWorkerConfig)
	broadcastWorker.Start(whatsapp.DefaultWorkerConfig.MaxWorkersGlobal)

	broadcastStore := storage.NewBroadcastStore()

	cfg := config.Load()
	jwtCfg := config.LoadJWT()

	var msgRepo storage.MessagesRepository
	var empresaStore domain.EmpresaStoreInterface
	var db *sql.DB
	if cfg.DBHost != "" {
		var err error
		db, err = storage.NewDB(cfg)
		if err != nil {
			fmt.Printf("[WARN] DB no disponible: %v\n", err)
		} else {
			msgRepo = storage.NewMessagesRepository(db)
			empresaStore = storage.NewEmpresaStore(db)
			fmt.Printf("[INFO] DB conectada a %s:%s/%s\n", cfg.DBHost, cfg.DBPort, cfg.DBName)
		}
	}

	h := NewHandlerWithBroadcast(manager, sessionStore, msgRepo, empresaStore, broadcastWorker, broadcastStore)

	// Dashboard handler
	var dashboardHandler *DashboardHandler
	if msgRepo != nil {
		dashboardHandler = NewDashboardHandler(msgRepo, sessionStore, empresaStore)
	}

	// Auth handler and middleware
	var authHandler *httpHandlers.AuthHandler
	var companiesHandler *httpHandlers.CompaniesHandler
	var authMiddleware *middleware.AuthMiddleware
	if db != nil {
		userStore := storage.NewAdminUserStore(db)
		blacklistStore := storage.NewTokenBlacklistStore(db)
		authHandler = httpHandlers.NewAuthHandler(userStore, empresaStore, blacklistStore, jwtCfg)
		companiesHandler = httpHandlers.NewCompaniesHandler(empresaStore)
		authMiddleware = middleware.NewAuthMiddleware(jwtCfg, blacklistStore)
	}

	mux := http.NewServeMux()

	// Admin users/roles/modules handler
	var adminHandler *AdminHandler
	if db != nil {
		adminHandler = NewAdminHandler(db)
	}

	// Auth routes (no auth required)
	if authHandler != nil {
		mux.HandleFunc("POST /api/auth/login", authHandler.Login)
		mux.HandleFunc("POST /api/auth/logout", authHandler.Logout)
		mux.HandleFunc("POST /api/auth/refresh", authHandler.Refresh)
		if authMiddleware != nil {
			mux.Handle("GET /api/auth/me", authMiddleware.RequireAuth()(http.HandlerFunc(authHandler.Me)))
		}
	}

	// Admin routes (users, roles, modules) - protected
	if authMiddleware != nil && adminHandler != nil {
		protected := authMiddleware.RequireAuth()
		mux.Handle("GET /api/admin/users", protected(http.HandlerFunc(adminHandler.ListUsers)))
		mux.Handle("GET /api/admin/users/", protected(http.HandlerFunc(adminHandler.GetUser)))
		mux.Handle("POST /api/admin/users", protected(http.HandlerFunc(adminHandler.CreateUser)))
		mux.Handle("PUT /api/admin/users/", protected(http.HandlerFunc(adminHandler.UpdateUser)))
		mux.Handle("DELETE /api/admin/users/", protected(http.HandlerFunc(adminHandler.DeleteUser)))
		mux.Handle("GET /api/admin/roles", protected(http.HandlerFunc(adminHandler.ListRoles)))
		mux.Handle("GET /api/admin/modules", protected(http.HandlerFunc(adminHandler.ListModules)))
		mux.Handle("POST /api/admin/users/promote/", protected(http.HandlerFunc(adminHandler.PromoteUser)))
		mux.Handle("PUT /api/admin/users/modules/", protected(http.HandlerFunc(adminHandler.AssignUserModules)))
	}

	// Companies routes (protected with auth middleware)
	if authMiddleware != nil && companiesHandler != nil {
		protected := authMiddleware.RequireAuth()
		mux.Handle("/api/companies", protected(http.HandlerFunc(companiesHandler.List)))
		mux.Handle("/api/companies/", protected(http.HandlerFunc(companiesHandler.Get)))
		mux.Handle("POST /api/companies", protected(http.HandlerFunc(companiesHandler.Create)))
		mux.Handle("PUT /api/companies/", protected(http.HandlerFunc(companiesHandler.Update)))
		mux.Handle("DELETE /api/companies/", protected(http.HandlerFunc(companiesHandler.Delete)))

		// Protected message/session/broadcast endpoints
		mux.Handle("POST /api/message", protected(http.HandlerFunc(h.HandlePostMessage)))
		mux.Handle("GET /api/messages", protected(http.HandlerFunc(h.HandleGetMessages)))
		mux.Handle("POST /api/broadcast", protected(http.HandlerFunc(h.HandlePostBroadcast)))
		mux.Handle("GET /api/broadcast/", protected(http.HandlerFunc(h.HandleGetBroadcast)))
		mux.Handle("GET /api/sessions", protected(http.HandlerFunc(HandleGetAdminSessions)))
		mux.Handle("POST /api/sessions", protected(http.HandlerFunc(HandlePostAdminSessions)))
		mux.Handle("GET /api/admin/messages", protected(http.HandlerFunc(HandleGetAdminMessages)))
		mux.Handle("GET /api/admin/broadcasts", protected(http.HandlerFunc(HandleGetAdminBroadcasts)))

		// Dashboard metrics endpoint
		if dashboardHandler != nil {
			mux.Handle("GET /api/dashboard/metricas", protected(http.HandlerFunc(dashboardHandler.GetMetrics)))
		}
	}

	// Public endpoints (no auth required for now)
	mux.HandleFunc("/ws", h.HandleWS)
	mux.HandleFunc("POST /message", h.HandlePostMessage)
	mux.HandleFunc("GET /messages", h.HandleGetMessages)
	mux.HandleFunc("POST /broadcast", h.HandlePostBroadcast)
	mux.HandleFunc("GET /broadcast/", h.HandleGetBroadcast)
	mux.HandleFunc("GET /metrics", HandleGetMetrics)
	mux.HandleFunc("GET /companies", HandleGetCompanies)
	mux.HandleFunc("GET /admin/messages", HandleGetAdminMessages)
	mux.HandleFunc("GET /admin/sessions", HandleGetAdminSessions)
	mux.HandleFunc("POST /admin/sessions", HandlePostAdminSessions)
	mux.HandleFunc("GET /admin/broadcasts", HandleGetAdminBroadcasts)
	mux.HandleFunc("POST /admin/login", HandleAdminLogin)

	// Health check endpoint
	mux.HandleFunc("GET /health", HandleHealth)

	// Register catch-all routes for 404/405 handling

	// Wrap with catchAll for 404 handling
	wrappedMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if path is a registered route
		path := r.URL.Path

		// Check for GET / (root) - handle specially
		if r.Method == "GET" && path == "/" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"status":   "ok",
				"service":  "wsapi",
				"message":  "WhatsApp API running",
				"frontend": "Serve separately on port 3000 or configure nginx",
			})
			return
		}

		// For root path with non-GET method, return 405
		if path == "/" {
			w.Header().Set("Allow", "GET")
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// For other paths, use the mux
		mux.ServeHTTP(w, r)
	})

	return LoggingMiddleware(CorrelationIDMiddleware(CORSMiddleware(wrappedMux)))
}

type DashboardMetrics struct {
	ActiveCompanies   int     `json:"active_companies"`
	MessagesToday     int     `json:"messages_today"`
	BroadcastsToday   int     `json:"broadcasts_today"`
	SuccessRate       float64 `json:"success_rate"`
	LastUpdate        string  `json:"last_update"`
	SessionsActive    int     `json:"sessions_active"`
	MessagesSent      int     `json:"messages_sent"`
	MessagesFailed    int     `json:"messages_failed"`
	BroadcastsCreated int     `json:"broadcasts_created"`
	Alerts            []Alert `json:"alerts,omitempty"`
}

type Alert struct {
	Type    string `json:"type"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

func HandleGetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := metrics.GetCounters()

	m := DashboardMetrics{
		ActiveCompanies:   int(c.SessionsActive),
		MessagesToday:     int(c.MessagesSent),
		BroadcastsToday:   int(c.BroadcastsCreated),
		SessionsActive:    int(c.SessionsActive),
		MessagesSent:      int(c.MessagesSent),
		MessagesFailed:    int(c.MessagesFailed),
		BroadcastsCreated: int(c.BroadcastsCreated),
		LastUpdate:        time.Now().Format(time.RFC3339),
		Alerts:            []Alert{},
	}

	totalMsgs := c.MessagesSent + c.MessagesFailed
	if totalMsgs > 0 {
		m.SuccessRate = float64(c.MessagesSent) / float64(totalMsgs) * 100
	}

	if c.SessionsActive == 0 {
		m.Alerts = append(m.Alerts, Alert{
			Type:    "sessions",
			Level:   "warning",
			Message: "No hay sesiones activas",
		})
	} else if c.SessionsActive < 3 {
		m.Alerts = append(m.Alerts, Alert{
			Type:    "sessions",
			Level:   "info",
			Message: fmt.Sprintf("%d sesiones activas", c.SessionsActive),
		})
	}

	if c.MessagesFailed > 10 {
		m.Alerts = append(m.Alerts, Alert{
			Type:    "messages",
			Level:   "warning",
			Message: fmt.Sprintf("%d mensajes fallidos", c.MessagesFailed),
		})
	}

	json.NewEncoder(w).Encode(m)
}

type Company struct {
	AccountID   string    `json:"account_id"`
	Status      string    `json:"status"`
	LastMessage time.Time `json:"last_message,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func HandleGetCompanies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	manager := whatsapp.NewManager()
	companies := manager.ListKeys()

	result := make([]Company, 0, len(companies))
	sessionStore := storage.NewSessionStore()

	for _, accountID := range companies {
		state, ok := sessionStore.Get(accountID)
		status := "inactive"
		if ok {
			status = state.Status
		}
		result = append(result, Company{
			AccountID: accountID,
			Status:    status,
			UpdatedAt: state.UpdatedAt,
		})
	}

	json.NewEncoder(w).Encode(map[string][]Company{
		"companies": result,
	})
}

type AdminMessage struct {
	ID        int       `json:"id"`
	AccountID string    `json:"account_id"`
	To        string    `json:"to"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func HandleGetAdminMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()
	limit := 50
	if l := query.Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	messages := []AdminMessage{}

	cfg := config.Load()
	if cfg.DBHost != "" {
		db, err := storage.NewDB(cfg)
		if err == nil {
			msgRepo := storage.NewMessagesRepository(db)
			msgs, _, err := msgRepo.GetByEmpresa(query.Get("account_id"), query.Get("status"), "", limit, 0)
			if err == nil {
				for _, m := range msgs {
					messages = append(messages, AdminMessage{
						ID:        int(m.ID),
						AccountID: m.RUCEmpresa,
						To:        m.Destino,
						Content:   m.Contenido,
						Status:    string(m.Estado),
						CreatedAt: m.TiempoEnvio,
					})
				}
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
		"total":    len(messages),
	})
}

type SessionInfo struct {
	AccountID string    `json:"account_id"`
	Status    string    `json:"status"`
	QRString  string    `json:"qr_string,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

func HandleGetAdminSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sessionStore := storage.NewSessionStore()
	manager := whatsapp.NewManager()

	result := []SessionInfo{}

	companies := manager.ListKeys()
	for _, accountID := range companies {
		state, ok := sessionStore.Get(accountID)
		if !ok {
			state = storage.SessionState{AccountID: accountID, Status: "inactive", UpdatedAt: time.Now()}
		}
		qr := ""
		if state.Status == "qr_pending" {
			qr = state.QRString
		}
		result = append(result, SessionInfo{
			AccountID: accountID,
			Status:    state.Status,
			QRString:  qr,
			UpdatedAt: state.UpdatedAt,
		})
	}

	json.NewEncoder(w).Encode(map[string][]SessionInfo{
		"sessions": result,
	})
}

func HandlePostAdminSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		AccountID string `json:"account_id"`
		Action    string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	sessionStore := storage.NewSessionStore()

	if req.Action == "disconnect" {
		sessionStore.SetDisconnected(req.AccountID, "admin_disconnect")
		manager := whatsapp.NewManager()
		manager.Delete(req.AccountID)
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

type BroadcastInfo struct {
	ReferenceID string    `json:"reference_id"`
	RUCEmpresa  string    `json:"ruc_empresa"`
	Total       int       `json:"total"`
	Status      string    `json:"status"`
	Success     int       `json:"success"`
	Failed      int       `json:"failed"`
	CreatedAt   time.Time `json:"created_at"`
}

func HandleGetAdminBroadcasts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	broadcastStore := storage.NewBroadcastStore()

	query := r.URL.Query()
	ruc := query.Get("account_id")

	var jobs []*domain.BroadcastJob
	if ruc != "" {
		jobs = broadcastStore.ListByRUC(ruc)
	} else {
	}

	result := make([]BroadcastInfo, 0, len(jobs))
	for _, job := range jobs {
		success := 0
		failed := 0
		for _, r := range job.Results {
			if r.Error == "" {
				success++
			} else {
				failed++
			}
		}
		result = append(result, BroadcastInfo{
			ReferenceID: job.ReferenceID,
			RUCEmpresa:  job.RUCEmpresa,
			Total:       job.Total,
			Status:      string(job.Status),
			Success:     success,
			Failed:      failed,
			CreatedAt:   job.CreatedAt,
		})
	}

	json.NewEncoder(w).Encode(map[string][]BroadcastInfo{
		"broadcasts": result,
	})
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type DashboardMetricsResponse struct {
	OK                   bool  `json:"ok"`
	TotalMensajes        int64 `json:"total_mensajes"`
	MensajesHoy          int64 `json:"mensajes_hoy"`
	MensajesSemana       int64 `json:"mensajes_semana"`
	MensajesExitosos     int64 `json:"mensajes_exitosos"`
	MensajesFallidos     int64 `json:"mensajes_fallidos"`
	SesionesActivas      int   `json:"sesiones_activas"`
	BroadcastsEjecutados int64 `json:"broadcasts_ejecutados"`
}

type DashboardHandler struct {
	msgRepo      storage.MessagesRepository
	sessionStore *storage.SessionStore
	empresaStore domain.EmpresaStoreInterface
}

func NewDashboardHandler(msgRepo storage.MessagesRepository, sessionStore *storage.SessionStore, empresaStore domain.EmpresaStoreInterface) *DashboardHandler {
	return &DashboardHandler{
		msgRepo:      msgRepo,
		sessionStore: sessionStore,
		empresaStore: empresaStore,
	}
}

func (h *DashboardHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	filter, ok := domain.GetEmpresaFilter(r.Context(), r.Header.Get("X-Empresa-ID"))
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(DashboardMetricsResponse{OK: false})
		return
	}

	if h.msgRepo == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(DashboardMetricsResponse{OK: false})
		return
	}

	var metrics *storage.MessageMetrics
	var err error

	if filter.IsRoot && filter.EmpresaID == nil {
		empresaIDStr := strings.TrimSpace(r.URL.Query().Get("empresa_id"))
		if empresaIDStr != "" {
			if empresaID, err := strconv.ParseInt(empresaIDStr, 10, 64); err == nil && empresaID > 0 {
				empresa, err := h.empresaStore.GetByID(empresaID)
				if err != nil || empresa == nil {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(DashboardMetricsResponse{OK: false})
					return
				}
				metrics, err = h.msgRepo.GetMessageMetricsByEmpresa(empresa.RUC)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(DashboardMetricsResponse{OK: false})
					return
				}
			} else {
				metrics, err = h.msgRepo.GetAllMessageMetrics()
			}
		} else {
			metrics, err = h.msgRepo.GetAllMessageMetrics()
		}
	} else {
		empresa, err := domain.GetRUCFromContext(r.Context(), filter, h.empresaStore)
		if err != nil || empresa == "" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(DashboardMetricsResponse{OK: false})
			return
		}
		metrics, err = h.msgRepo.GetMessageMetricsByEmpresa(empresa)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(DashboardMetricsResponse{OK: false})
			return
		}
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(DashboardMetricsResponse{OK: false})
		return
	}

	sessionCount := 0
	if filter.IsRoot && filter.EmpresaID == nil {
		sessionCount = h.sessionStore.ActiveCount()
	} else {
		empresa, _ := domain.GetRUCFromContext(r.Context(), filter, h.empresaStore)
		if empresa != "" {
			if state, ok := h.sessionStore.Get(empresa); ok && state.Status == "active" && state.IsActive {
				sessionCount = 1
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(DashboardMetricsResponse{
		OK:                   true,
		TotalMensajes:        metrics.TotalMensajes,
		MensajesHoy:          metrics.MensajesHoy,
		MensajesSemana:       metrics.MensajesSemana,
		MensajesExitosos:     metrics.MensajesExitosos,
		MensajesFallidos:     metrics.MensajesFallidos,
		SesionesActivas:      sessionCount,
		BroadcastsEjecutados: metrics.BroadcastsEjecutados,
	})
}

func HandleAdminLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Demo: accept admin/admin123
	// In production, this should query the database with hashed passwords
	if req.Username == "admin" && req.Password == "admin123" {
		token := "demo-token-" + time.Now().Format("20060102150405")
		json.NewEncoder(w).Encode(domain.LoginResponse{
			OK:    true,
			Token: token,
		})
		return
	}

	http.Error(w, "invalid credentials", http.StatusUnauthorized)
}

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "API is running",
	})
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Path

	if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/admin/") {
		if r.Method == "GET" || r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" {
			if !routeExists(path, r.Method) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
		}
		if !routeExists(path, r.Method) {
			allowedMethods := getAllowedMethods(path)
			if len(allowedMethods) == 0 {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}

	http.Error(w, "not found", http.StatusNotFound)
}

func handleOtherMethods(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Path
	allowedMethods := getAllowedMethods(path)

	if len(allowedMethods) == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

var registeredRoutes = map[string][]string{
	"/api/auth/login":           {"POST"},
	"/api/auth/logout":          {"POST"},
	"/api/auth/refresh":         {"POST"},
	"/api/auth/me":              {"GET"},
	"/api/companies":            {"GET", "POST"},
	"/api/companies/":           {"GET", "PUT", "DELETE"},
	"/api/message":              {"POST"},
	"/api/messages":             {"GET"},
	"/api/broadcast":            {"POST"},
	"/api/broadcast/":           {"GET"},
	"/api/sessions":             {"GET", "POST"},
	"/api/admin/messages":       {"GET"},
	"/api/admin/broadcasts":     {"GET"},
	"/api/admin/users":          {"GET", "POST"},
	"/api/admin/users/":         {"GET", "PUT", "DELETE"},
	"/api/admin/users/promote/": {"POST"},
	"/api/admin/users/modules/": {"PUT"},
	"/api/admin/roles":          {"GET"},
	"/api/admin/modules":        {"GET"},
	"/api/dashboard/metricas":   {"GET"},
	"/message":                  {"POST"},
	"/messages":                 {"GET"},
	"/broadcast":                {"POST"},
	"/broadcast/":               {"GET"},
	"/metrics":                  {"GET"},
	"/companies":                {"GET"},
	"/admin/messages":           {"GET"},
	"/admin/sessions":           {"GET", "POST"},
	"/admin/broadcasts":         {"GET"},
	"/admin/login":              {"POST"},
	"/ws":                       {"GET"},
}

func routeExists(path, method string) bool {
	methods, ok := registeredRoutes[path]
	if !ok {
		return false
	}
	for _, m := range methods {
		if m == method {
			return true
		}
	}
	return false
}

func getAllowedMethods(path string) []string {
	methods, ok := registeredRoutes[path]
	if !ok {
		return nil
	}
	return methods
}

func handleCatchAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.URL.Path == "/" {
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "ok",
			"service":  "wsapi",
			"message":  "WhatsApp API running",
			"frontend": "Serve separately on port 3000 or configure nginx",
		})
		return
	}
	http.Error(w, "not found", http.StatusNotFound)
}
