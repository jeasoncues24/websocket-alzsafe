package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"wsapi/internal/config"
	"wsapi/internal/domain"
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

// NewRouter inicializa el runtime HTTP y delega el registro a archivos por dominio.
func NewRouter() http.Handler {
	mux := http.NewServeMux()

	c := NewContainer()
	k := NewKernel(c.AuthMiddleware, c.EmpresaAuthMiddleware, c.ApiKeyAuthMiddleware)

	RegisterAdminRoutes(mux, c, k)
	RegisterAPIRoutes(mux, c, k)
	mux.Handle("GET /", http.HandlerFunc(handleRoot))

	return &startupAwareRouter{
		handler: k.Apply(mux),
		startFn: c.StartupTasks,
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

	writeJSON(w, http.StatusOK, m)
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

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":        true,
		"companies": result,
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

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"broadcasts": result,
	})
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

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		writeAPIError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"status":  "ok",
		"message": "API is running",
	})
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"status":   "ok",
		"service":  "wsapi",
		"message":  "WhatsApp API running",
		"frontend": "Serve separately on port 3000 or configure nginx",
	})
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Path

	if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/admin/") {
		if r.Method == "GET" || r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" {
			if !routeExists(path, r.Method) {
				writeAPIError(w, http.StatusNotFound, "not found")
				return
			}
		}
		if !routeExists(path, r.Method) {
			allowedMethods := getAllowedMethods(path)
			if len(allowedMethods) == 0 {
				writeAPIError(w, http.StatusNotFound, "not found")
				return
			}
			w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
			writeAPIError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
	}

	writeAPIError(w, http.StatusNotFound, "not found")
}

func handleOtherMethods(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Path
	allowedMethods := getAllowedMethods(path)

	if len(allowedMethods) == 0 {
		writeAPIError(w, http.StatusNotFound, "not found")
		return
	}

	w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
	writeAPIError(w, http.StatusMethodNotAllowed, "method not allowed")
}

var registeredRoutes = map[string][]string{
	// Admin panel — JWT auth
	"/api/auth/login":                          {"POST"},
	"/api/auth/logout":                         {"POST"},
	"/api/auth/refresh":                        {"POST"},
	"/api/auth/me":                             {"GET"},
	"/api/admin/users":                         {"GET", "POST"},
	"/api/admin/usuarios_admin":                {"GET", "POST"},
	"/api/admin/roles":                         {"GET", "POST"},
	"/api/admin/modules":                       {"GET"},
	"/api/admin/empresas":                      {"GET", "POST"},
	"/api/admin/empresas/{id}":                 {"GET", "PUT", "DELETE"},
	"/api/admin/empresas/{id}/restore":         {"POST"},
	"/api/admin/empresas/{id}/token":           {"POST"},
	"/api/admin/empresas/{id}/token/revoke":    {"POST"},
	"/api/admin/telefonos":                     {"GET", "POST"},
	"/api/admin/telefonos/{id}/connect":        {"POST"},
	"/api/admin/telefonos/{id}/connect/ws":     {"GET"},
	"/api/admin/telefonos/{id}/api-keys":       {"GET", "POST"},
	"/api/admin/api-keys/{id}":                 {"GET"},
	"/api/admin/api-keys/{id}/rotate":          {"POST"},
	"/api/admin/api-keys/{id}/revoke":          {"POST"},
	"/api/admin/api-keys/{id}/usage":           {"GET"},
	"/api/admin/api-keys/{id}/audit":           {"GET"},
	"/api/admin/sesiones":                      {"GET", "POST"},
	"/api/admin/sesiones/diagnostico":          {"GET"},
	"/api/admin/mensajes":                      {"GET", "POST"},
	"/api/admin/metricas":                      {"GET"},
	"/api/admin/difusiones":                    {"GET"},
	// Service API — API token por teléfono
	"/api/service/v1/auth/empresa/validate":    {"POST"},
	"/api/service/v1/empresas":                 {"GET", "PUT"},
	"/api/service/v1/metricas":                 {"GET"},
	"/api/service/v1/telefonos":                {"GET"},
	"/api/service/v1/telefonos/{id}/qr":        {"POST"},
	"/api/service/v1/telefonos/{id}/estado":    {"GET"},
	"/api/service/v1/sesiones":                 {"GET", "POST"},
	"/api/service/v1/sesiones/{id}":            {"GET", "DELETE"},
	"/api/service/v1/sesiones/{id}/connect":    {"POST"},
	"/api/service/v1/me":                       {"GET"},
	"/api/service/v1/sesion":                   {"GET"},
	"/api/service/v1/mensajes":                 {"GET", "POST"},
	"/api/service/v1/mensajes/{id}":            {"GET", "PATCH", "POST"},
	"/api/service/v1/difusiones":               {"GET", "POST"},
	"/api/service/v1/difusiones/{id}":          {"GET"},
	"/api/service/v1/ws":                       {"GET"},
	// Infraestructura
	"/metrics":                                 {"GET"},
	"/health":                                  {"GET"},
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
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":       true,
			"status":   "ok",
			"service":  "wsapi",
			"message":  "WhatsApp API running",
			"frontend": "Serve separately on port 3000 or configure nginx",
		})
		return
	}
	writeAPIError(w, http.StatusNotFound, "not found")
}
