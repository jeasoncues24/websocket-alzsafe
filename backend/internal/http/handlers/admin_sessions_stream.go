package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// flush vacía el buffer de la respuesta. Un ResponseWriter que no soporta flush
// (p. ej. un wrapper de middleware sin Flush/Unwrap) no es fatal para el stream:
// net/http acabará vaciando por su cuenta. Solo se propaga un error real de
// escritura (conexión cerrada).
func flush(rc *http.ResponseController) error {
	if err := rc.Flush(); err != nil && !errors.Is(err, http.ErrNotSupported) {
		return err
	}
	return nil
}

// StartStream inicializa el hub de tiempo real del reporte de sesiones y arranca su
// productor. Idempotente. Debe llamarse una vez durante el arranque, pasando el
// contexto de vida de la aplicación. Conecta el hook de SessionStore para que
// cualquier transición de sesión dispare un recomputo del snapshot.
func (h *AdminSessionsHandler) StartStream(ctx context.Context) {
	if h.hub != nil {
		return
	}
	h.hub = newFleetHub(h.buildSessionsSnapshot)
	if h.sessionStore != nil {
		h.sessionStore.SetOnChange(h.hub.Trigger)
	}
	go h.hub.Run(ctx)
}

// StreamSessions expone el reporte de sesiones como stream SSE. Cada actualización
// se emite como un evento `snapshot` con el mismo payload que el GET REST. Se envía
// un comentario `: ping` cada 20 s para mantener viva la conexión a través de
// proxies con idle timeout.
func (h *AdminSessionsHandler) StreamSessions(w http.ResponseWriter, r *http.Request) {
	if h.hub == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "stream no disponible")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	rc := http.NewResponseController(w)

	ch, unsub := h.hub.Subscribe()
	defer unsub()

	ctx := r.Context()

	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-ch:
			if !ok {
				return
			}
			if _, err := fmt.Fprintf(w, "event: snapshot\ndata: %s\n\n", data); err != nil {
				return
			}
			if err := flush(rc); err != nil {
				return
			}
		case <-heartbeat.C:
			if _, err := io.WriteString(w, ": ping\n\n"); err != nil {
				return
			}
			if err := flush(rc); err != nil {
				return
			}
		}
	}
}
