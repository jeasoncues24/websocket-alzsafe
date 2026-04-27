package http

import (
	"context"
	"database/sql"
	"fmt"
	"wsapi/internal/config"
	"wsapi/internal/domain"
	handlers "wsapi/internal/http/handlers"
	"wsapi/internal/http/middleware"
	"wsapi/internal/storage"
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
	AdminHandler          *AdminHandler
	V1MessagesHandler     *handlers.V1MessagesHandler
	V1BroadcastsHandler   *handlers.V1BroadcastsHandler
	V1MetricsHandler      *handlers.V1MetricsHandler
	V1PhonesHandler       *handlers.V1PhonesHandler
	V1SessionsHandler     *handlers.V1SessionsHandler
	V1WSHandler           *handlers.V1WSHandler
	AuthMiddleware        *middleware.AuthMiddleware
	EmpresaAuthMiddleware *middleware.EmpresaAuthMiddleware
	ApiKeyAuthMiddleware  *middleware.ApiKeyAuthMiddleware
	DashboardHandler      *DashboardHandler
	Config                *config.Config
	JWTCfg                *config.JWTConfig
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
	v1WSHandler := handlers.NewV1WSHandler(manager, jwtCfg)
	dashboardHandler := NewDashboardHandler(msgRepo, sessionStore, empresaStore)

	authMiddleware := middleware.NewAuthMiddleware(jwtCfg, blacklistStore)
	empresaAuthMiddleware := middleware.NewEmpresaAuthMiddleware(jwtCfg, empresaStore, telefonoStore)
	apiKeyAuthMiddleware := middleware.NewApiKeyAuthMiddleware(apiKeyStore, empresaStore, telefonoStore)

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
		StartupTasks:          buildStartupBootstrap(cfg, manager, sessionStore, telefonoStore),
		AuthHandler:           authHandler,
		CompaniesHandler:      companiesHandler,
		ApiKeysHandler:        apiKeysHandler,
		AdminHandler:          adminHandler,
		V1MessagesHandler:     v1MessagesHandler,
		V1BroadcastsHandler:   v1BroadcastsHandler,
		V1MetricsHandler:      v1MetricsHandler,
		V1PhonesHandler:       v1PhonesHandler,
		V1SessionsHandler:     v1SessionsHandler,
		V1WSHandler:           v1WSHandler,
		AuthMiddleware:        authMiddleware,
		EmpresaAuthMiddleware: empresaAuthMiddleware,
		ApiKeyAuthMiddleware:  apiKeyAuthMiddleware,
		DashboardHandler:      dashboardHandler,
		Config:                cfg,
		JWTCfg:                jwtCfg,
	}
}
