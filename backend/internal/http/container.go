package http

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"strconv"
	"wsapi/internal/config"
	"wsapi/internal/domain"
	handlers "wsapi/internal/http/handlers"
	"wsapi/internal/http/middleware"
	"wsapi/internal/storage"
	"wsapi/internal/telemetry"
	"wsapi/internal/whatsapp"
)

type Container struct {
	Manager               *whatsapp.Manager
	LegacyWSHandler       *Handler
	SessionStore          *storage.SessionStore
	BroadcastWorker       *whatsapp.BroadcastWorker
	BroadcastStore        *storage.BroadcastStore
	MsgRepo               storage.MessagesRepository
	EmpresaStore          domain.EmpresaStoreInterface
	TelefonoStore         *storage.TelefonoStore
	DB                    *sql.DB
	StartupTasks          func(context.Context)
	AuthHandler           *handlers.AuthHandler
	CompaniesHandler      *handlers.CompaniesHandler
	ApiKeysHandler        *handlers.ApiKeysHandler
	ApiKeyMetricsHandler  *handlers.ApiKeyMetricsHandler
	AdminHandler          *AdminHandler
	V1MessagesHandler     *handlers.V1MessagesHandler
	V1BroadcastsHandler   *handlers.V1BroadcastsHandler
	V1MetricsHandler      *handlers.V1MetricsHandler
	V1PhonesHandler       *handlers.V1PhonesHandler
	V1SessionsHandler     *handlers.V1SessionsHandler
	V1HealthHandler       *handlers.V1HealthHandler
	V1WSHandler           *handlers.V1WSHandler
	V1WebhooksHandler     *handlers.V1WebhooksHandler
	AdminMessagesHandler  *handlers.AdminMessagesHandler
	AdminSessionsHandler  *handlers.AdminSessionsHandler
	AdminClientsHandler   *handlers.AdminClientsHandler
	AuthMiddleware        *middleware.AuthMiddleware
	EmpresaAuthMiddleware *middleware.EmpresaAuthMiddleware
	ApiKeyAuthMiddleware  *middleware.ApiKeyAuthMiddleware
	DashboardHandler      *DashboardHandler
	Config                *config.Config
	JWTCfg                *config.JWTConfig
	TelemetryMW           func(http.Handler) http.Handler
}

// NewContainer inicializa todas las dependencias y handlers
func NewContainer() *Container {
	cfg := config.Load()
	jwtCfg := config.LoadJWT()
	whatsapp.ConfigureLogging(whatsapp.LoggingOptions{
		DebugLogDir:        cfg.WhatsAppDebugLogDir,
		DebugLogPerAccount: cfg.WhatsAppDebugLogPerAccount,
		DebugLogLevel:      cfg.WhatsAppDebugLogLevel,
		ConsoleLogLevel:    cfg.WhatsAppConsoleLogLevel,
	})

	manager := whatsapp.NewManager()
	sessionStore := storage.NewSessionStore()
	broadcastWorker := whatsapp.NewBroadcastWorker(whatsapp.DefaultWorkerConfig, manager)
	broadcastWorker.Start(whatsapp.DefaultWorkerConfig.MaxWorkersGlobal)
	broadcastStore := storage.NewBroadcastStore()

	var msgRepo storage.MessagesRepository
	var empresaStore domain.EmpresaStoreInterface
	var telefonoStore *storage.TelefonoStore
	var webhookStore *storage.WebhookStore
	var db *sql.DB
	if cfg.DBHost != "" {
		var err error
		db, err = storage.NewDB(cfg)
		if err != nil {
			config.GetLogger().Warn().Err(err).Msg("DB no disponible")
		} else {
			msgRepo = storage.NewMessagesRepository(db)
			empresaStore = storage.NewEmpresaStore(db)
			telefonoStore = storage.NewTelefonoStore(db)
			webhookStore = storage.NewWebhookStore(db)
			config.GetLogger().Info().Str("host", cfg.DBHost).Str("port", cfg.DBPort).Str("db", cfg.DBName).Msg("DB conectada")
		}
	}

	whatsapp.NewService(manager, sessionStore, telefonoStore, webhookStore, cfg.WhatsAppSQLiteDir)
	legacyWSHandler := NewHandlerWithBroadcast(manager, sessionStore, msgRepo, empresaStore, broadcastWorker, broadcastStore)

	userStore := storage.NewAdminUserStore(db)
	blacklistStore := storage.NewTokenBlacklistStore(db)
	apiKeyStore := storage.NewApiKeyStore(db)

	authHandler := handlers.NewAuthHandler(userStore, empresaStore, blacklistStore, jwtCfg)
	companiesHandler := handlers.NewCompaniesHandler(empresaStore, sessionStore, jwtCfg)
	apiKeysHandler := handlers.NewApiKeysHandler(apiKeyStore, telefonoStore, empresaStore, manager)
	v1MessagesHandler := handlers.NewV1MessagesHandler(msgRepo, telefonoStore, manager)
	v1BroadcastsHandler := handlers.NewV1BroadcastsHandler(broadcastStore, telefonoStore, broadcastWorker)
	v1MetricsHandler := handlers.NewV1MetricsHandler(msgRepo, telefonoStore)
	v1PhonesHandler := handlers.NewV1PhonesHandler(telefonoStore, sessionStore)
	v1SessionsHandler := handlers.NewV1SessionsHandler(telefonoStore, sessionStore, manager)
	v1WSHandler := handlers.NewV1WSHandler(manager, jwtCfg, telefonoStore, sessionStore)

	var v1WebhooksHandler *handlers.V1WebhooksHandler
	var webhookStartupTask func(context.Context)
	if webhookStore != nil {
		maxWebhooks := 10
		if v := os.Getenv("WEBHOOKS_MAX_PER_EMPRESA"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				maxWebhooks = n
			}
		}
		v1WebhooksHandler = handlers.NewV1WebhooksHandler(webhookStore, maxWebhooks)

		webhookWorker := whatsapp.NewWebhookDeliveryWorker(webhookStore, nil, whatsapp.WebhookDeliveryWorkerConfig{})
		webhookStartupTask = func(ctx context.Context) {
			webhookWorker.Run(ctx)
		}
	}

	healthRate := 60
	if v := os.Getenv("HEALTH_RATE_LIMIT_PER_MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			healthRate = n
		}
	}
	v1HealthHandler := handlers.NewV1HealthHandler(cfg.Version, healthRate)
	adminMessagesHandler := handlers.NewAdminMessagesHandler(msgRepo, empresaStore, telefonoStore, manager)
	adminSessionsHandler := handlers.NewAdminSessionsHandler(empresaStore, telefonoStore, manager, sessionStore, jwtCfg)
	adminClientsHandler := handlers.NewAdminClientsHandler(db)
	dashboardHandler := NewDashboardHandler(msgRepo, sessionStore, empresaStore, telefonoStore, db)

	authMiddleware := middleware.NewAuthMiddleware(jwtCfg, blacklistStore)
	empresaAuthMiddleware := middleware.NewEmpresaAuthMiddleware(jwtCfg, empresaStore, telefonoStore)
	apiKeyAuthMiddleware := middleware.NewApiKeyAuthMiddleware(apiKeyStore, empresaStore, telefonoStore)

	var telemetryMW func(http.Handler) http.Handler
	var apiKeyMetricsHandler *handlers.ApiKeyMetricsHandler
	if db != nil {
		teleCfg := telemetry.DefaultConfig()
		teleStore := telemetry.NewMySQLStore(db, teleCfg)
		teleMiddleware := telemetry.NewTelemetryMiddleware(teleStore, teleCfg)
		telemetryMW = teleMiddleware.Capture

		teleStoreSvc := storage.NewTelemetryStore(db)
		apiKeyMetricsHandler = handlers.NewApiKeyMetricsHandler(teleStoreSvc)
	}

	// AdminHandler requiere tipos propios, ajústalo según tu implementación
	adminHandler := NewAdminHandler(db, sessionStore, manager, jwtCfg)
	if adminHandler != nil {
		adminHandler.telefonoStore = telefonoStore
	}

	return &Container{
		Manager:               manager,
		LegacyWSHandler:       legacyWSHandler,
		SessionStore:          sessionStore,
		BroadcastWorker:       broadcastWorker,
		BroadcastStore:        broadcastStore,
		MsgRepo:               msgRepo,
		EmpresaStore:          empresaStore,
		TelefonoStore:         telefonoStore,
		DB:                    db,
		StartupTasks:          composeStartupTasks(buildStartupBootstrap(cfg, manager, sessionStore, telefonoStore), webhookStartupTask),
		AuthHandler:           authHandler,
		CompaniesHandler:      companiesHandler,
		ApiKeysHandler:        apiKeysHandler,
		ApiKeyMetricsHandler:  apiKeyMetricsHandler,
		AdminHandler:          adminHandler,
		V1HealthHandler:       v1HealthHandler,
		V1MessagesHandler:     v1MessagesHandler,
		V1BroadcastsHandler:   v1BroadcastsHandler,
		V1MetricsHandler:      v1MetricsHandler,
		V1PhonesHandler:       v1PhonesHandler,
		V1SessionsHandler:     v1SessionsHandler,
		V1WSHandler:           v1WSHandler,
		V1WebhooksHandler:     v1WebhooksHandler,
		AdminMessagesHandler:  adminMessagesHandler,
		AdminSessionsHandler:  adminSessionsHandler,
		AdminClientsHandler:   adminClientsHandler,
		AuthMiddleware:        authMiddleware,
		EmpresaAuthMiddleware: empresaAuthMiddleware,
		ApiKeyAuthMiddleware:  apiKeyAuthMiddleware,
		DashboardHandler:      dashboardHandler,
		Config:                cfg,
		JWTCfg:                jwtCfg,
		TelemetryMW:           telemetryMW,
	}
}
