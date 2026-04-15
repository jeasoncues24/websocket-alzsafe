package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"wsapi/internal/config"
	"wsapi/internal/domain"
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
	var msgRepo storage.MessagesRepository
	if cfg.DBHost != "" {
		db, err := storage.NewDB(cfg)
		if err != nil {
			fmt.Printf("[WARN] DB no disponible: %v\n", err)
		} else {
			msgRepo = storage.NewMessagesRepository(db)
			fmt.Printf("[INFO] DB conectada a %s:%s/%s\n", cfg.DBHost, cfg.DBPort, cfg.DBName)
		}
	}

	h := NewHandlerWithBroadcast(manager, sessionStore, msgRepo, broadcastWorker, broadcastStore)

	mux := http.NewServeMux()

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

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "ok",
			"service":  "wsapi",
			"message":  "WhatsApp API running",
			"frontend": "Serve separately on port 3000 or configure nginx",
		})
	})

	return LoggingMiddleware(CorrelationIDMiddleware(CORSMiddleware(mux)))
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
			msgs, _, err := msgRepo.GetByEmpresa(query.Get("account_id"), query.Get("status"), limit, 0)
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

type LoginResponse struct {
	Token string `json:"token"`
	User  string `json:"user"`
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
		json.NewEncoder(w).Encode(LoginResponse{
			Token: token,
			User:  "admin",
		})
		return
	}

	http.Error(w, "invalid credentials", http.StatusUnauthorized)
}
