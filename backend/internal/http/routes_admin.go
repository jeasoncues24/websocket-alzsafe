package http

import "net/http"

func RegisterAdminRoutes(mux *http.ServeMux, c *Container, k *Kernel) {
	adminStack := k.AdminAuth

	if c.AuthHandler != nil {
		mux.Handle("POST /api/auth/login", http.HandlerFunc(c.AuthHandler.Login))
		mux.Handle("POST /api/auth/logout", http.HandlerFunc(c.AuthHandler.Logout))
		mux.Handle("POST /api/auth/refresh", http.HandlerFunc(c.AuthHandler.Refresh))
		mux.Handle("GET /api/auth/me", adminStack(http.HandlerFunc(c.AuthHandler.Me)))
		mux.Handle("PUT /api/auth/me", adminStack(http.HandlerFunc(c.AuthHandler.UpdateMe)))
		mux.Handle("PUT /api/auth/me/password", adminStack(http.HandlerFunc(c.AuthHandler.UpdateMePassword)))
	}
	if c.AdminHandler != nil {
		mux.Handle("GET /api/admin/users", adminStack(http.HandlerFunc(c.AdminHandler.ListUsers)))
		mux.Handle("GET /api/admin/users/{id}", adminStack(http.HandlerFunc(c.AdminHandler.GetUser)))
		mux.Handle("POST /api/admin/users", adminStack(http.HandlerFunc(c.AdminHandler.CreateUser)))
		mux.Handle("PUT /api/admin/users/{id}", adminStack(http.HandlerFunc(c.AdminHandler.UpdateUser)))
		mux.Handle("DELETE /api/admin/users/{id}", adminStack(http.HandlerFunc(c.AdminHandler.DeleteUser)))
		mux.Handle("POST /api/admin/users/{id}/promote", adminStack(http.HandlerFunc(c.AdminHandler.PromoteUserByID)))
		mux.Handle("GET /api/admin/users/{id}/modulos", adminStack(http.HandlerFunc(c.AdminHandler.GetUserModules)))
		mux.Handle("PUT /api/admin/users/{id}/modulos", adminStack(http.HandlerFunc(c.AdminHandler.AssignUserModulesByID)))

		mux.Handle("GET /api/admin/usuario_admin", adminStack(http.HandlerFunc(c.AdminHandler.ListUsuarioAdmins)))
		mux.Handle("GET /api/admin/usuario_admin/{id}", adminStack(http.HandlerFunc(c.AdminHandler.GetUsuarioAdmin)))
		mux.Handle("POST /api/admin/usuario_admin", adminStack(http.HandlerFunc(c.AdminHandler.CreateUsuarioAdmin)))
		mux.Handle("PUT /api/admin/usuario_admin/{id}", adminStack(http.HandlerFunc(c.AdminHandler.UpdateUsuarioAdmin)))
		mux.Handle("DELETE /api/admin/usuario_admin/{id}", adminStack(http.HandlerFunc(c.AdminHandler.DeleteUsuarioAdmin)))
		mux.Handle("POST /api/admin/usuario_admin/{id}/promote", adminStack(http.HandlerFunc(c.AdminHandler.PromoteUsuarioAdmin)))
		mux.Handle("GET /api/admin/usuario_admin/{id}/modulos", adminStack(http.HandlerFunc(c.AdminHandler.GetUsuarioAdminModules)))
		mux.Handle("PUT /api/admin/usuario_admin/{id}/modulos", adminStack(http.HandlerFunc(c.AdminHandler.AssignUsuarioAdminModules)))

		mux.Handle("GET /api/admin/roles", adminStack(http.HandlerFunc(c.AdminHandler.ListRoles)))
		mux.Handle("GET /api/admin/roles/{id}", adminStack(http.HandlerFunc(c.AdminHandler.GetRole)))
		mux.Handle("POST /api/admin/roles", adminStack(http.HandlerFunc(c.AdminHandler.CreateRole)))
		mux.Handle("PUT /api/admin/roles/{id}", adminStack(http.HandlerFunc(c.AdminHandler.UpdateRole)))
		mux.Handle("DELETE /api/admin/roles/{id}", adminStack(http.HandlerFunc(c.AdminHandler.DeleteRole)))
		mux.Handle("GET /api/admin/modules", adminStack(http.HandlerFunc(c.AdminHandler.ListModules)))

		mux.Handle("GET /api/admin/empresas/{id}/telefonos", adminStack(http.HandlerFunc(c.AdminHandler.ListCompanyPhones)))
		mux.Handle("POST /api/admin/empresas/{id}/telefonos", adminStack(http.HandlerFunc(c.AdminHandler.CreateCompanyPhone)))
		mux.Handle("GET /api/admin/telefonos/{id}", adminStack(http.HandlerFunc(c.AdminHandler.GetCompanyPhone)))
		mux.Handle("PUT /api/admin/telefonos/{id}", adminStack(http.HandlerFunc(c.AdminHandler.UpdateCompanyPhone)))
		mux.Handle("DELETE /api/admin/telefonos/{id}", adminStack(http.HandlerFunc(c.AdminHandler.DeleteCompanyPhone)))
		mux.Handle("POST /api/admin/telefonos/{id}/connect", adminStack(http.HandlerFunc(c.AdminHandler.StartCompanyPhoneConnection)))
		mux.Handle("GET /api/admin/telefonos/{id}/connect/ws", http.HandlerFunc(c.AdminHandler.ConnectCompanyPhoneWS))
		mux.Handle("GET /api/admin/sesiones/diagnostico", adminStack(http.HandlerFunc(c.AdminHandler.GetSessionsDiagnostics)))
		mux.Handle("GET /api/admin/telefonos/{id}/webhooks", adminStack(http.HandlerFunc(c.AdminHandler.ListTelefonoWebhooks)))
	}

	if c.CompaniesHandler != nil {
		mux.Handle("GET /api/admin/empresas", adminStack(http.HandlerFunc(c.CompaniesHandler.List)))
		mux.Handle("GET /api/admin/empresas/{id}", adminStack(http.HandlerFunc(c.CompaniesHandler.Get)))
		mux.Handle("POST /api/admin/empresas", adminStack(http.HandlerFunc(c.CompaniesHandler.Create)))
		mux.Handle("PUT /api/admin/empresas/{id}", adminStack(http.HandlerFunc(c.CompaniesHandler.Update)))
		mux.Handle("DELETE /api/admin/empresas/{id}", adminStack(http.HandlerFunc(c.CompaniesHandler.Delete)))
		mux.Handle("POST /api/admin/empresas/{id}/restore", adminStack(http.HandlerFunc(c.CompaniesHandler.Restore)))
	}

	if c.ApiKeysHandler != nil {
		mux.Handle("GET /api/admin/telefonos/{id}/api-keys", adminStack(http.HandlerFunc(c.ApiKeysHandler.ListByTelefono)))
		mux.Handle("POST /api/admin/telefonos/{id}/api-keys", adminStack(http.HandlerFunc(c.ApiKeysHandler.CreateForTelefono)))
		mux.Handle("GET /api/admin/api-keys/{id}", adminStack(http.HandlerFunc(c.ApiKeysHandler.Get)))
		mux.Handle("POST /api/admin/api-keys/{id}/rotate", adminStack(http.HandlerFunc(c.ApiKeysHandler.Rotate)))
		mux.Handle("POST /api/admin/api-keys/{id}/revoke", adminStack(http.HandlerFunc(c.ApiKeysHandler.Revoke)))
		mux.Handle("GET /api/admin/api-keys/{id}/audit", adminStack(http.HandlerFunc(c.ApiKeysHandler.Audit)))
	}

	if c.ApiKeyMetricsHandler != nil {
		mux.Handle("GET /api/admin/api-keys/{id}/usage/stats", adminStack(http.HandlerFunc(c.ApiKeyMetricsHandler.UsageStats)))
		mux.Handle("GET /api/admin/api-keys/{id}/usage/timeseries", adminStack(http.HandlerFunc(c.ApiKeyMetricsHandler.UsageTimeSeries)))
		mux.Handle("GET /api/admin/api-keys/{id}/audit/stats", adminStack(http.HandlerFunc(c.ApiKeyMetricsHandler.AuditStats)))
	}

	if c.DashboardHandler != nil {
		mux.Handle("GET /api/admin/metricas", adminStack(http.HandlerFunc(c.DashboardHandler.GetMetrics)))
	}

	if c.AdminMessagesHandler != nil {
		mux.Handle("GET /api/admin/mensajes", adminStack(http.HandlerFunc(c.AdminMessagesHandler.GetMessages)))
		mux.Handle("POST /api/admin/mensajes/{id}", adminStack(http.HandlerFunc(c.AdminMessagesHandler.RetryMessage)))
	}
	if c.AdminSessionsHandler != nil {
		mux.Handle("GET /api/admin/sesiones", adminStack(http.HandlerFunc(c.AdminSessionsHandler.GetSessions)))
		mux.Handle("POST /api/admin/sesiones", adminStack(http.HandlerFunc(c.AdminSessionsHandler.PostSession)))
		mux.Handle("POST /api/admin/telefonos/{id}/qr-link", adminStack(http.HandlerFunc(c.AdminSessionsHandler.GenerateQRLink)))
	}
	if c.AdminClientsHandler != nil {
		mux.Handle("GET /api/admin/clientes/buscar", adminStack(http.HandlerFunc(c.AdminClientsHandler.BuscarCliente)))
	}
	mux.Handle("GET /api/admin/difusiones", adminStack(http.HandlerFunc(c.HandleGetAdminBroadcasts)))
	mux.Handle("GET /api/admin/difusiones/{id}", adminStack(http.HandlerFunc(c.HandleGetAdminBroadcastDetail)))

	mux.Handle("GET /metrics", http.HandlerFunc(HandleGetMetrics))
	mux.Handle("GET /health", http.HandlerFunc(HandleHealth))
}
