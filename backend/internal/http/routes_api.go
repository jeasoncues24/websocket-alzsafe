package http

import "net/http"

func RegisterAPIRoutes(mux *http.ServeMux, c *Container, k *Kernel) {
	clientStack := k.ServiceStack
	empresaStack := k.EmpresaAuth

	if c.V1HealthHandler != nil {
		mux.Handle("GET /api/service/v1/health", http.HandlerFunc(c.V1HealthHandler.GetHealth))
	}

	if c.CompaniesHandler != nil {
		mux.Handle("POST /api/service/v1/auth/empresa/validate", empresaStack(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"ok":true}`))
		})))
		mux.Handle("GET /api/service/v1/empresas", empresaStack(http.HandlerFunc(c.CompaniesHandler.GetCurrent)))
		mux.Handle("PUT /api/service/v1/empresas", empresaStack(http.HandlerFunc(c.CompaniesHandler.UpdateCurrent)))
	}

	if c.V1MetricsHandler != nil {
		mux.Handle("GET /api/service/v1/metricas", empresaStack(http.HandlerFunc(c.V1MetricsHandler.GetMetrics)))
	}

	if c.V1PhonesHandler != nil {
		mux.Handle("GET /api/service/v1/telefonos", empresaStack(http.HandlerFunc(c.V1PhonesHandler.GetPhones)))
		mux.Handle("POST /api/service/v1/telefonos/{id}/qr", empresaStack(http.HandlerFunc(c.V1PhonesHandler.PostPhoneQr)))
		mux.Handle("GET /api/service/v1/telefonos/{id}/estado", empresaStack(http.HandlerFunc(c.V1PhonesHandler.GetPhoneStatus)))
	}

	if c.V1SessionsHandler != nil {
		mux.Handle("GET /api/service/v1/sesiones", empresaStack(http.HandlerFunc(c.V1SessionsHandler.GetSessions)))
		mux.Handle("POST /api/service/v1/sesiones", empresaStack(http.HandlerFunc(c.V1SessionsHandler.PostSessions)))
		mux.Handle("GET /api/service/v1/sesiones/{id}", empresaStack(http.HandlerFunc(c.V1SessionsHandler.GetSession)))
		mux.Handle("DELETE /api/service/v1/sesiones/{id}", empresaStack(http.HandlerFunc(c.V1SessionsHandler.DeleteSession)))
		mux.Handle("POST /api/service/v1/sesiones/{id}/connect", empresaStack(http.HandlerFunc(c.V1SessionsHandler.StartPhoneConnection)))
	}

	if c.ApiKeysHandler != nil {
		mux.Handle("GET /api/service/v1/me", clientStack(http.HandlerFunc(c.ApiKeysHandler.Me)))
		mux.Handle("GET /api/service/v1/sesion", clientStack(http.HandlerFunc(c.ApiKeysHandler.Session)))
	}

	if c.V1MessagesHandler != nil {
		mux.Handle("GET /api/service/v1/mensajes", clientStack(http.HandlerFunc(c.V1MessagesHandler.GetMessages)))
		mux.Handle("POST /api/service/v1/mensajes", clientStack(http.HandlerFunc(c.V1MessagesHandler.PostMessage)))
		mux.Handle("GET /api/service/v1/mensajes/{id}", clientStack(http.HandlerFunc(c.V1MessagesHandler.GetMessageByReference)))
		mux.Handle("PATCH /api/service/v1/mensajes/{id}", clientStack(http.HandlerFunc(c.V1MessagesHandler.UpdateMessage)))
		mux.Handle("POST /api/service/v1/mensajes/{id}", clientStack(http.HandlerFunc(c.V1MessagesHandler.RetryMessage)))
	}

	if c.V1BroadcastsHandler != nil {
		mux.Handle("GET /api/service/v1/difusiones", clientStack(http.HandlerFunc(c.V1BroadcastsHandler.GetBroadcasts)))
		mux.Handle("POST /api/service/v1/difusiones", clientStack(http.HandlerFunc(c.V1BroadcastsHandler.PostBroadcast)))
		mux.Handle("GET /api/service/v1/difusiones/{id}", clientStack(http.HandlerFunc(c.V1BroadcastsHandler.GetBroadcast)))
	}

	if c.V1WSHandler != nil {
		mux.Handle("GET /api/service/v1/ws", http.HandlerFunc(c.V1WSHandler.HandleWS))
	}
}
