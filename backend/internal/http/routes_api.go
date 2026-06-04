package http

import "net/http"

func RegisterAPIRoutes(mux *http.ServeMux, c *Container, k *Kernel) {
	clientStack := k.ServiceStack

	if c.V1HealthHandler != nil {
		mux.Handle("GET /api/service/v1/health", http.HandlerFunc(c.V1HealthHandler.GetHealth))
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
		mux.Handle("GET /api/service/v1/ws/connect", http.HandlerFunc(c.V1WSHandler.HandleConnectWS))
	}

	if c.V1WebhooksHandler != nil {
		mux.Handle("POST /api/service/v1/webhooks", clientStack(http.HandlerFunc(c.V1WebhooksHandler.Create)))
		mux.Handle("GET /api/service/v1/webhooks", clientStack(http.HandlerFunc(c.V1WebhooksHandler.List)))
		mux.Handle("DELETE /api/service/v1/webhooks/{id}", clientStack(http.HandlerFunc(c.V1WebhooksHandler.Delete)))
	}
}
