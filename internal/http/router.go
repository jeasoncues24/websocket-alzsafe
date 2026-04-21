package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"wsapi/internal/config"
	"wsapi/internal/domain"
	httpHandlers "wsapi/internal/http/handlers"
	"wsapi/internal/http/middleware"
	"wsapi/internal/metrics"
	"wsapi/internal/storage"
	"wsapi/internal/whatsapp"
)

type startupAwareRouter struct {
	handler http.Handler
	startFn func(context.Context)
	once    sync.Once
}

func (r *startupAwareRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}

func (r *startupAwareRouter) RunStartupTasks(ctx context.Context) {
	if r == nil || r.startFn == nil {
		return
	}
	r.once.Do(func() {
		go r.startFn(ctx)
	})
}

func NewRouter() http.Handler {
	manager := whatsapp.NewManager()
	sessionStore := storage.NewSessionStore()

	broadcastWorker := whatsapp.NewBroadcastWorker(whatsapp.DefaultWorkerConfig, manager)
	broadcastWorker.Start(whatsapp.DefaultWorkerConfig.MaxWorkersGlobal)

	broadcastStore := storage.NewBroadcastStore()

	cfg := config.Load()
	jwtCfg := config.LoadJWT()

	whatsapp.ConfigureLogging(whatsapp.LoggingOptions{
		DebugLogDir:        cfg.WhatsAppDebugLogDir,
		DebugLogPerAccount: cfg.WhatsAppDebugLogPerAccount,
		DebugLogLevel:      cfg.WhatsAppDebugLogLevel,
		ConsoleLogLevel:    cfg.WhatsAppConsoleLogLevel,
	})

	var msgRepo storage.MessagesRepository
	var empresaStore domain.EmpresaStoreInterface
	var telefonoStore *storage.TelefonoStore
	var db *sql.DB
	if cfg.DBHost != "" {
		var err error
		db, err = storage.NewDB(cfg)
		if err != nil {
			fmt.Printf("[WARN] DB no disponible: %v\n", err)
		} else {
			msgRepo = storage.NewMessagesRepository(db)
			empresaStore = storage.NewEmpresaStore(db)
			telefonoStore = storage.NewTelefonoStore(db)
			fmt.Printf("[INFO] DB conectada a %s:%s/%s\n", cfg.DBHost, cfg.DBPort, cfg.DBName)
		}
	}

	whatsapp.NewService(manager, sessionStore, telefonoStore, cfg.WhatsAppSQLiteDir)

	h := NewHandlerWithBroadcast(manager, sessionStore, msgRepo, empresaStore, broadcastWorker, broadcastStore)

	// Dashboard handler
	var dashboardHandler *DashboardHandler
	if msgRepo != nil {
		dashboardHandler = NewDashboardHandler(msgRepo, sessionStore, empresaStore)
	}

	// Auth handler and middleware
	var authHandler *httpHandlers.AuthHandler
	var companiesHandler *httpHandlers.CompaniesHandler
	var apiKeysHandler *httpHandlers.ApiKeysHandler
	var empresaAuthMiddleware *middleware.EmpresaAuthMiddleware
	var apiKeyAuthMiddleware *middleware.ApiKeyAuthMiddleware
	var v1MessagesHandler *httpHandlers.V1MessagesHandler
	var v1BroadcastsHandler *httpHandlers.V1BroadcastsHandler
	var v1MetricsHandler *httpHandlers.V1MetricsHandler
	var authMiddleware *middleware.AuthMiddleware
	if db != nil {
		userStore := storage.NewAdminUserStore(db)
		blacklistStore := storage.NewTokenBlacklistStore(db)
		apiKeyStore := storage.NewApiKeyStore(db)
		authHandler = httpHandlers.NewAuthHandler(userStore, empresaStore, blacklistStore, jwtCfg)
		companiesHandler = httpHandlers.NewCompaniesHandler(empresaStore, sessionStore, jwtCfg)
		apiKeysHandler = httpHandlers.NewApiKeysHandler(apiKeyStore, telefonoStore, empresaStore, manager)
		empresaAuthMiddleware = middleware.NewEmpresaAuthMiddleware(jwtCfg, empresaStore, telefonoStore)
		apiKeyAuthMiddleware = middleware.NewApiKeyAuthMiddleware(apiKeyStore, empresaStore, telefonoStore)
		v1MessagesHandler = httpHandlers.NewV1MessagesHandler(msgRepo, telefonoStore, manager)
		v1BroadcastsHandler = httpHandlers.NewV1BroadcastsHandler(broadcastStore, telefonoStore, broadcastWorker)
		v1MetricsHandler = httpHandlers.NewV1MetricsHandler(msgRepo, telefonoStore)
		authMiddleware = middleware.NewAuthMiddleware(jwtCfg, blacklistStore)
	}

	mux := http.NewServeMux()

	// Admin users/roles/modules handler
	var adminHandler *AdminHandler
	if db != nil {
		adminHandler = NewAdminHandler(db, sessionStore, manager, jwtCfg)
		if adminHandler != nil {
			adminHandler.telefonoStore = telefonoStore
		}
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
	if empresaAuthMiddleware != nil && adminHandler != nil {
		protectedEmpresa := empresaAuthMiddleware.RequireEmpresaAuth()
		mux.Handle("GET /api/admin/usuario_admin", protectedEmpresa(http.HandlerFunc(adminHandler.ListUsuarioAdmins)))
		mux.Handle("GET /api/admin/usuario_admin/{id}", protectedEmpresa(http.HandlerFunc(adminHandler.GetUsuarioAdmin)))
		mux.Handle("POST /api/admin/usuario_admin", protectedEmpresa(http.HandlerFunc(adminHandler.CreateUsuarioAdmin)))
		mux.Handle("PUT /api/admin/usuario_admin/{id}", protectedEmpresa(http.HandlerFunc(adminHandler.UpdateUsuarioAdmin)))
		mux.Handle("DELETE /api/admin/usuario_admin/{id}", protectedEmpresa(http.HandlerFunc(adminHandler.DeleteUsuarioAdmin)))
		mux.Handle("GET /api/admin/usuario_admin/{id}/modulos", protectedEmpresa(http.HandlerFunc(adminHandler.GetUsuarioAdminModules)))
		mux.Handle("PUT /api/admin/usuario_admin/{id}/modulos", protectedEmpresa(http.HandlerFunc(adminHandler.AssignUsuarioAdminModules)))
		mux.Handle("GET /api/admin/roles", protectedEmpresa(http.HandlerFunc(adminHandler.ListRoles)))
		mux.Handle("GET /api/admin/roles/{id}", protectedEmpresa(http.HandlerFunc(adminHandler.GetRole)))
		mux.Handle("POST /api/admin/roles", protectedEmpresa(http.HandlerFunc(adminHandler.CreateRole)))
		mux.Handle("PUT /api/admin/roles/{id}", protectedEmpresa(http.HandlerFunc(adminHandler.UpdateRole)))
		mux.Handle("DELETE /api/admin/roles/{id}", protectedEmpresa(http.HandlerFunc(adminHandler.DeleteRole)))
		mux.Handle("GET /api/admin/modules", protectedEmpresa(http.HandlerFunc(adminHandler.ListModules)))
	}

	if authMiddleware != nil && adminHandler != nil {
		protected := authMiddleware.RequireAuth()
		mux.Handle("GET /api/admin/users", protected(http.HandlerFunc(adminHandler.ListUsers)))
		mux.Handle("GET /api/admin/users/", protected(http.HandlerFunc(adminHandler.GetUser)))
		mux.Handle("POST /api/admin/users", protected(http.HandlerFunc(adminHandler.CreateUser)))
		mux.Handle("PUT /api/admin/users/", protected(http.HandlerFunc(adminHandler.UpdateUser)))
		mux.Handle("DELETE /api/admin/users/", protected(http.HandlerFunc(adminHandler.DeleteUser)))
		mux.Handle("POST /api/admin/users/promote/", protected(http.HandlerFunc(adminHandler.PromoteUser)))
		mux.Handle("PUT /api/admin/users/modules/", protected(http.HandlerFunc(adminHandler.AssignUserModules)))
		mux.Handle("GET /api/admin/empresas/{id}/telefonos", protected(http.HandlerFunc(adminHandler.ListCompanyPhones)))
		mux.Handle("POST /api/admin/empresas/{id}/telefonos", protected(http.HandlerFunc(adminHandler.CreateCompanyPhone)))
		mux.Handle("GET /api/admin/telefonos/{id}", protected(http.HandlerFunc(adminHandler.GetCompanyPhone)))
		mux.Handle("PUT /api/admin/telefonos/{id}", protected(http.HandlerFunc(adminHandler.UpdateCompanyPhone)))
		mux.Handle("DELETE /api/admin/telefonos/{id}", protected(http.HandlerFunc(adminHandler.DeleteCompanyPhone)))
		mux.Handle("POST /api/admin/telefonos/{id}/connect", protected(http.HandlerFunc(adminHandler.StartCompanyPhoneConnection)))
		mux.Handle("GET /api/admin/telefonos/{id}/connect/ws", http.HandlerFunc(adminHandler.ConnectCompanyPhoneWS))
	}

	// Admin companies routes and company JWT endpoints kept for compatibility
	if authMiddleware != nil && companiesHandler != nil {
		adminProtected := authMiddleware.RequireAuth()
		mux.Handle("GET /api/admin/empresas", adminProtected(http.HandlerFunc(companiesHandler.List)))
		mux.Handle("GET /api/admin/empresas/", adminProtected(http.HandlerFunc(companiesHandler.Get)))
		mux.Handle("POST /api/admin/empresas", adminProtected(http.HandlerFunc(companiesHandler.Create)))
		mux.Handle("PUT /api/admin/empresas/", adminProtected(http.HandlerFunc(companiesHandler.Update)))
		mux.Handle("DELETE /api/admin/empresas/", adminProtected(http.HandlerFunc(companiesHandler.Delete)))
		mux.Handle("POST /api/admin/empresas/{id}/token", adminProtected(http.HandlerFunc(companiesHandler.GenerateToken)))
		mux.Handle("POST /api/admin/empresas/{id}/token/revoke", adminProtected(http.HandlerFunc(companiesHandler.RevokeToken)))
	}

	// Admin API keys routes
	if authMiddleware != nil && apiKeysHandler != nil {
		adminProtected := authMiddleware.RequireAuth()
		mux.Handle("GET /api/admin/telefonos/{id}/api-keys", adminProtected(http.HandlerFunc(apiKeysHandler.ListByTelefono)))
		mux.Handle("POST /api/admin/telefonos/{id}/api-keys", adminProtected(http.HandlerFunc(apiKeysHandler.CreateForTelefono)))
		mux.Handle("GET /api/admin/api-keys/{id}", adminProtected(http.HandlerFunc(apiKeysHandler.Get)))
		mux.Handle("POST /api/admin/api-keys/{id}/rotate", adminProtected(http.HandlerFunc(apiKeysHandler.Rotate)))
		mux.Handle("POST /api/admin/api-keys/{id}/revoke", adminProtected(http.HandlerFunc(apiKeysHandler.Revoke)))
		mux.Handle("GET /api/admin/api-keys/{id}/usage", adminProtected(http.HandlerFunc(apiKeysHandler.Usage)))
		mux.Handle("GET /api/admin/api-keys/{id}/audit", adminProtected(http.HandlerFunc(apiKeysHandler.Audit)))
	}

	// Admin operational routes
	if authMiddleware != nil {
		adminProtected := authMiddleware.RequireAuth()
		mux.Handle("GET /api/admin/mensajes", adminProtected(http.HandlerFunc(HandleGetAdminMessages)))
		mux.Handle("POST /api/admin/mensajes/", adminProtected(http.HandlerFunc(HandleAdminRetryMessage)))
		mux.Handle("GET /api/admin/sesiones", adminProtected(http.HandlerFunc(HandleGetAdminSessions)))
		mux.Handle("POST /api/admin/sesiones", adminProtected(http.HandlerFunc(HandlePostAdminSessions)))
		if adminHandler != nil {
			mux.Handle("GET /api/admin/sesiones/diagnostico", adminProtected(http.HandlerFunc(adminHandler.GetSessionsDiagnostics)))
		}
		mux.Handle("GET /api/admin/difusiones", adminProtected(http.HandlerFunc(HandleGetAdminBroadcasts)))
		if dashboardHandler != nil {
			mux.Handle("GET /api/admin/metricas", adminProtected(http.HandlerFunc(dashboardHandler.GetMetrics)))
		}
	}

	// Empresa auth validation
	if empresaAuthMiddleware != nil {
		empresaProtected := empresaAuthMiddleware.RequireEmpresaAuth()
		mux.Handle("POST /api/auth/empresa/validate", empresaProtected(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
		})))
	}

	// API key auth validation and public API routes
	if apiKeyAuthMiddleware != nil && apiKeysHandler != nil && v1MessagesHandler != nil && v1BroadcastsHandler != nil {
		apiKeyProtected := apiKeyAuthMiddleware.RequireApiKeyAuth()
		mux.Handle("GET /api/me", apiKeyProtected(http.HandlerFunc(apiKeysHandler.Me)))
		mux.Handle("GET /api/v1/me", apiKeyProtected(http.HandlerFunc(apiKeysHandler.Me)))
		mux.Handle("GET /api/sesion", apiKeyProtected(http.HandlerFunc(apiKeysHandler.Session)))
		mux.Handle("GET /api/mensajes", apiKeyProtected(http.HandlerFunc(v1MessagesHandler.GetMessages)))
		mux.Handle("POST /api/mensajes", apiKeyProtected(http.HandlerFunc(v1MessagesHandler.PostMessage)))
		mux.Handle("GET /api/mensajes/", apiKeyProtected(http.HandlerFunc(v1MessagesHandler.GetMessageByReference)))
		mux.Handle("PATCH /api/mensajes/", apiKeyProtected(http.HandlerFunc(v1MessagesHandler.UpdateMessage)))
		mux.Handle("POST /api/mensajes/", apiKeyProtected(http.HandlerFunc(v1MessagesHandler.RetryMessage)))
		mux.Handle("GET /api/difusiones", apiKeyProtected(http.HandlerFunc(v1BroadcastsHandler.GetBroadcasts)))
		mux.Handle("POST /api/difusiones", apiKeyProtected(http.HandlerFunc(v1BroadcastsHandler.PostBroadcast)))
		mux.Handle("GET /api/difusiones/", apiKeyProtected(http.HandlerFunc(v1BroadcastsHandler.GetBroadcast)))
	}

	// Empresa routes (/api/*) - empresa JWT protected
	if empresaAuthMiddleware != nil && v1MetricsHandler != nil && companiesHandler != nil {
		empresaProtected := empresaAuthMiddleware.RequireEmpresaAuth()
		mux.Handle("GET /api/empresas", empresaProtected(http.HandlerFunc(companiesHandler.GetCurrent)))
		mux.Handle("PUT /api/empresas", empresaProtected(http.HandlerFunc(companiesHandler.UpdateCurrent)))
		mux.Handle("GET /api/metricas", empresaProtected(http.HandlerFunc(v1MetricsHandler.GetMetrics)))
	}

	// Public endpoints (no auth required for now)
	mux.HandleFunc("/ws", h.HandleWS)
	mux.HandleFunc("GET /metrics", HandleGetMetrics)
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

	baseHandler := LoggingMiddleware(CorrelationIDMiddleware(CORSMiddleware(wrappedMux)))
	return &startupAwareRouter{
		handler: baseHandler,
		startFn: buildStartupBootstrap(cfg, manager, sessionStore, telefonoStore),
	}
}

func buildStartupBootstrap(cfg *config.Config, manager *whatsapp.Manager, sessionStore *storage.SessionStore, telefonoStore *storage.TelefonoStore) func(context.Context) {
	if cfg == nil || manager == nil || telefonoStore == nil || !cfg.WhatsAppBootstrapEnabled {
		return nil
	}

	return func(ctx context.Context) {
		if ctx == nil {
			ctx = context.Background()
		}
		timeout := time.Duration(cfg.WhatsAppBootstrapTimeoutSec) * time.Second
		if timeout <= 0 {
			timeout = 60 * time.Second
		}
		bootstrapCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		bootstrap := whatsapp.NewStartupBootstrapper(manager, sessionStore, telefonoStore, whatsapp.StartupBootstrapConfig{
			MaxConcurrency: cfg.WhatsAppBootstrapMaxConcurrency,
		})
		summary := bootstrap.Run(bootstrapCtx)

		fmt.Printf("[INFO] startup bootstrap sesiones: total=%d activos_db=%d runtime_activos=%d mismatches=%d intentos_start=%d errores_start=%d duracion=%s\n",
			summary.TotalTelefonos,
			summary.ActivosEnDB,
			summary.RuntimeActivos,
			summary.MismatchesDetectados,
			summary.IntentosStart,
			summary.ErroresStart,
			summary.Duracion,
		)
	}
}

type DashboardMetrics struct {
	ActiveCompanies                int     `json:"active_companies"`
	MessagesToday                  int     `json:"messages_today"`
	BroadcastsToday                int     `json:"broadcasts_today"`
	SuccessRate                    float64 `json:"success_rate"`
	LastUpdate                     string  `json:"last_update"`
	SessionsActive                 int     `json:"sessions_active"`
	MessagesSent                   int     `json:"messages_sent"`
	MessagesFailed                 int     `json:"messages_failed"`
	BroadcastsCreated              int     `json:"broadcasts_created"`
	StartupBootstrapRuns           int     `json:"startup_bootstrap_runs"`
	StartupBootstrapMismatches     int     `json:"startup_bootstrap_mismatches"`
	StartupBootstrapStartAttempts  int     `json:"startup_bootstrap_start_attempts"`
	StartupBootstrapStartErrors    int     `json:"startup_bootstrap_start_errors"`
	StartupBootstrapLastDurationMs int64   `json:"startup_bootstrap_last_duration_ms"`
	Alerts                         []Alert `json:"alerts,omitempty"`
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
		ActiveCompanies:                int(c.SessionsActive),
		MessagesToday:                  int(c.MessagesSent),
		BroadcastsToday:                int(c.BroadcastsCreated),
		SessionsActive:                 int(c.SessionsActive),
		MessagesSent:                   int(c.MessagesSent),
		MessagesFailed:                 int(c.MessagesFailed),
		BroadcastsCreated:              int(c.BroadcastsCreated),
		StartupBootstrapRuns:           int(c.StartupBootstrapRuns),
		StartupBootstrapMismatches:     int(c.StartupBootstrapMismatches),
		StartupBootstrapStartAttempts:  int(c.StartupBootstrapStartAttempts),
		StartupBootstrapStartErrors:    int(c.StartupBootstrapStartErrors),
		StartupBootstrapLastDurationMs: c.StartupBootstrapLastDurationMs,
		LastUpdate:                     time.Now().Format(time.RFC3339),
		Alerts:                         []Alert{},
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

	if c.StartupBootstrapStartErrors > 0 {
		m.Alerts = append(m.Alerts, Alert{
			Type:    "bootstrap",
			Level:   "warning",
			Message: fmt.Sprintf("bootstrap con %d errores de inicio de sesion", c.StartupBootstrapStartErrors),
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
	ID          int       `json:"id"`
	ReferenceID string    `json:"reference_id,omitempty"`
	AccountID   string    `json:"account_id"`
	To          string    `json:"to"`
	Content     string    `json:"content"`
	Status      string    `json:"status"`
	ErrorReason *string   `json:"error_reason,omitempty"`
	RetryCount  *int      `json:"retry_count,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
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
			empresaStore := storage.NewEmpresaStore(db)
			accountID := strings.TrimSpace(query.Get("account_id"))
			status := strings.TrimSpace(query.Get("status"))

			appendMessages := func(empresaID int64, ruc string) {
				items, _, err := msgRepo.GetByEmpresa(empresaID, status, "", limit, 0)
				if err != nil {
					return
				}
				for _, m := range items {
					msg := AdminMessage{
						ID:          int(m.ID),
						ReferenceID: m.ReferenceID,
						AccountID:   ruc,
						To:          m.Destino,
						Content:     m.Contenido,
						Status:      string(m.Estado),
						CreatedAt:   m.TiempoEnvio,
					}
					if m.ErrorReason != "" {
						msg.ErrorReason = &m.ErrorReason
					}
					if m.RetryCount > 0 {
						msg.RetryCount = &m.RetryCount
					}
					messages = append(messages, msg)
				}
			}

			if accountID != "" {
				if empresa, err := empresaStore.GetByRUC(accountID); err == nil && empresa != nil {
					appendMessages(empresa.ID, empresa.RUC)
				}
			} else {
				empresas, _, err := empresaStore.GetAll(1, 1000, "", nil)
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
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
		"total":    len(messages),
	})
}

func HandleAdminRetryMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := domain.GetAdminJWTClaims(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	refID := extractReferenceID(r.URL.Path)
	if refID == "" {
		http.Error(w, "missing reference_id", http.StatusBadRequest)
		return
	}

	cfg := config.Load()
	if cfg.DBHost == "" {
		http.Error(w, "database not configured", http.StatusInternalServerError)
		return
	}

	db, err := storage.NewDB(cfg)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	msgRepo := storage.NewMessagesRepository(db)
	telefonoStore := storage.NewTelefonoStore(db)

	msg, err := msgRepo.GetByReferenceID(refID)
	if err != nil || msg == nil {
		http.Error(w, "message not found", http.StatusNotFound)
		return
	}

	if claims.EmpresaID != nil && msg.EmpresaID != *claims.EmpresaID && !claims.IsRoot {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if msg.Estado == domain.MessageStateSent || msg.Estado == domain.MessageStateDelivered {
		http.Error(w, "message already sent", http.StatusBadRequest)
		return
	}
	if len(msg.Adjuntos) > 0 {
		http.Error(w, "media retry unsupported", http.StatusBadRequest)
		return
	}

	telefono, err := telefonoStore.GetByID(msg.TelefonoID)
	if err != nil || telefono == nil {
		http.Error(w, "telefono not found", http.StatusNotFound)
		return
	}

	if telefono.Status != domain.TelefonoStatusActive {
		http.Error(w, "session not active", http.StatusBadRequest)
		return
	}

	if err := msgRepo.IncrementRetryCount(refID); err != nil {
		http.Error(w, "error preparing retry", http.StatusInternalServerError)
		return
	}

	manager := whatsapp.NewManager()
	err = whatsapp.SendRichMessage(r.Context(), manager, telefono.NumeroCompleto, msg.Destino, msg.Contenido, nil)
	if err != nil {
		_ = msgRepo.UpdateEstado(refID, domain.MessageStateFailed, err.Error())
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":           false,
			"reference_id": refID,
			"estado":       string(domain.MessageStateFailed),
			"error":        err.Error(),
		})
		return
	}

	_ = msgRepo.UpdateEstado(refID, domain.MessageStateSent, "")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":           true,
		"reference_id": refID,
		"estado":       string(domain.MessageStateSent),
	})
}

func extractReferenceID(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	for i, p := range parts {
		if p == "mensajes" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

type SessionInfo struct {
	AccountID string    `json:"account_id"`
	Status    string    `json:"status"`
	QRString  string    `json:"qr_string,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

func HandleGetAdminSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	result := []SessionInfo{}
	cfg := config.Load()
	if cfg.DBHost != "" {
		if db, err := storage.NewDB(cfg); err == nil {
			empresaStore := storage.NewEmpresaStore(db)
			telefonoStore := storage.NewTelefonoStore(db)
			empresas, _, err := empresaStore.GetAll(1, 1000, "", nil)
			if err == nil {
				for i := range empresas {
					telefonos, err := telefonoStore.GetByEmpresa(empresas[i].ID)
					if err != nil {
						continue
					}
					for _, t := range telefonos {
						qr := ""
						if t.Status == domain.TelefonoStatusQRPending {
							qr = t.QRString
						}
						result = append(result, SessionInfo{
							AccountID: t.NumeroCompleto,
							Status:    string(t.Status),
							QRString:  qr,
							UpdatedAt: t.UpdatedAt,
						})
					}
				}
			}
		}
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

	if req.Action == "disconnect" {
		cfg := config.Load()
		if cfg.DBHost != "" {
			if db, err := storage.NewDB(cfg); err == nil {
				telefonoStore := storage.NewTelefonoStore(db)
				if telefono, err := telefonoStore.GetByNumeroCompleto(req.AccountID); err == nil && telefono != nil {
					_ = telefonoStore.SetDisconnected(telefono.ID)
				}
			}
		}
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
		cfg := config.Load()
		if cfg.DBHost != "" {
			if db, err := storage.NewDB(cfg); err == nil {
				empresaStore := storage.NewEmpresaStore(db)
				if empresa, err := empresaStore.GetByRUC(ruc); err == nil && empresa != nil {
					jobs = broadcastStore.ListByEmpresa(empresa.ID)
				}
			}
		}
	} else {
		cfg := config.Load()
		if cfg.DBHost != "" {
			if db, err := storage.NewDB(cfg); err == nil {
				empresaStore := storage.NewEmpresaStore(db)
				empresas, _, err := empresaStore.GetAll(1, 1000, "", nil)
				if err == nil {
					for i := range empresas {
						jobs = append(jobs, broadcastStore.ListByEmpresa(empresas[i].ID)...)
					}
				}
			}
		}
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
			RUCEmpresa:  fmt.Sprintf("%d", job.EmpresaID),
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
				metrics, err = h.msgRepo.GetMessageMetricsByEmpresa(empresa.ID)
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
		if empresaID, ok := domain.GetEmpresaIDFromContext(r.Context(), filter); ok {
			metrics, err = h.msgRepo.GetMessageMetricsByEmpresa(empresaID)
		} else {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(DashboardMetricsResponse{OK: false})
			return
		}
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
	"/api/auth/login":                 {"POST"},
	"/api/auth/logout":                {"POST"},
	"/api/auth/refresh":               {"POST"},
	"/api/auth/me":                    {"GET"},
	"/api/me":                         {"GET"},
	"/api/sesion":                     {"GET"},
	"/api/companies":                  {"GET", "POST"},
	"/api/companies/":                 {"GET", "PUT", "DELETE"},
	"/api/message":                    {"POST"},
	"/api/messages":                   {"GET"},
	"/api/broadcast":                  {"POST"},
	"/api/broadcast/":                 {"GET"},
	"/api/sessions":                   {"GET", "POST"},
	"/api/admin/messages":             {"GET"},
	"/api/admin/sesiones/diagnostico": {"GET"},
	"/api/admin/broadcasts":           {"GET"},
	"/api/admin/users":                {"GET", "POST"},
	"/api/admin/users/":               {"GET", "PUT", "DELETE"},
	"/api/admin/users/promote/":       {"POST"},
	"/api/admin/users/modules/":       {"PUT"},
	"/api/admin/roles":                {"GET"},
	"/api/admin/modules":              {"GET"},
	"/api/dashboard/metricas":         {"GET"},
	"/message":                        {"POST"},
	"/messages":                       {"GET"},
	"/broadcast":                      {"POST"},
	"/broadcast/":                     {"GET"},
	"/metrics":                        {"GET"},
	"/companies":                      {"GET"},
	"/admin/messages":                 {"GET"},
	"/admin/sessions":                 {"GET", "POST"},
	"/admin/broadcasts":               {"GET"},
	"/admin/login":                    {"POST"},
	"/ws":                             {"GET"},
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
