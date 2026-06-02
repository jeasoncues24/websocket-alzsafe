package whatsapp

import (
	"context"
	"testing"
	"time"
)

// recvWithin lee un evento del canal o falla si no llega dentro del timeout.
func recvWithin(t *testing.T, ch <-chan SessionEvent, d time.Duration) (SessionEvent, bool) {
	t.Helper()
	select {
	case evt, ok := <-ch:
		return evt, ok
	case <-time.After(d):
		t.Fatal("timeout esperando evento del canal")
		return SessionEvent{}, false
	}
}

// TestRuntimeFanoutDeliversToAllSubscribers verifica que un broadcast llega a
// todos los observadores simultáneos (caso del enlace compartido: panel admin +
// página QR pública observando la misma sesión).
func TestRuntimeFanoutDeliversToAllSubscribers(t *testing.T) {
	rt := &sessionRuntime{subscribers: make(map[chan SessionEvent]struct{})}

	ch1, unsub1 := rt.subscribe()
	ch2, unsub2 := rt.subscribe()
	defer unsub1()
	defer unsub2()

	rt.broadcast(SessionEvent{Event: "qr-123", Data: map[string]any{"qrString": "ABC"}})

	for i, ch := range []<-chan SessionEvent{ch1, ch2} {
		evt, ok := recvWithin(t, ch, time.Second)
		if !ok {
			t.Fatalf("subscriber %d: canal cerrado inesperadamente", i)
		}
		if evt.Event != "qr-123" {
			t.Fatalf("subscriber %d: evento = %q, want qr-123", i, evt.Event)
		}
	}
}

// TestRuntimeSubscribeReceivesSnapshot verifica que un observador que se une
// tarde recibe de inmediato el último estado conocido (snapshot).
func TestRuntimeSubscribeReceivesSnapshot(t *testing.T) {
	rt := &sessionRuntime{subscribers: make(map[chan SessionEvent]struct{})}

	rt.broadcast(SessionEvent{Event: "qr-123", Data: map[string]any{"qrString": "XYZ"}})

	ch, unsub := rt.subscribe()
	defer unsub()

	evt, ok := recvWithin(t, ch, time.Second)
	if !ok || evt.Event != "qr-123" {
		t.Fatalf("snapshot = %q ok=%v, want qr-123", evt.Event, ok)
	}
}

// TestRuntimeUnsubscribeClosesChannel verifica que darse de baja cierra el canal
// del observador y deja de recibir eventos.
func TestRuntimeUnsubscribeClosesChannel(t *testing.T) {
	rt := &sessionRuntime{subscribers: make(map[chan SessionEvent]struct{})}

	ch, unsub := rt.subscribe()
	unsub()

	if _, ok := <-ch; ok {
		t.Fatal("se esperaba canal cerrado tras unsubscribe")
	}

	// Doble unsubscribe no debe entrar en pánico.
	unsub()

	// Un broadcast posterior no debe entregar nada al canal dado de baja.
	rt.broadcast(SessionEvent{Event: "active-123"})
}

// TestRuntimeCloseAllClosesEverySubscriber verifica que closeAll cierra todos
// los canales y que suscribirse luego devuelve el snapshot y un canal cerrado.
func TestRuntimeCloseAllClosesEverySubscriber(t *testing.T) {
	rt := &sessionRuntime{subscribers: make(map[chan SessionEvent]struct{})}

	ch1, _ := rt.subscribe()
	rt.broadcast(SessionEvent{Event: "active-123", Data: map[string]any{"isActive": true}})
	// Drenar el evento difundido.
	recvWithin(t, ch1, time.Second)

	rt.closeAll()

	if _, ok := <-ch1; ok {
		t.Fatal("se esperaba canal cerrado tras closeAll")
	}

	// Suscribirse a un runtime ya cerrado entrega snapshot y luego cierra.
	ch2, _ := rt.subscribe()
	evt, ok := recvWithin(t, ch2, time.Second)
	if !ok || evt.Event != "active-123" {
		t.Fatalf("snapshot tras cierre = %q ok=%v, want active-123", evt.Event, ok)
	}
	if _, ok := <-ch2; ok {
		t.Fatal("se esperaba cierre del canal tras el snapshot")
	}
}

// TestRuntimeAbandonCancelsWhenNeverActive verifica que cuando se va el último
// observador y la sesión nunca estuvo activa, el runtime se cancela (limpieza de
// sesión QR abandonada).
func TestRuntimeAbandonCancelsWhenNeverActive(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	rt := &sessionRuntime{
		ctx:         ctx,
		cancel:      cancel,
		subscribers: make(map[chan SessionEvent]struct{}),
	}

	_, unsub := rt.subscribe()
	unsub()

	select {
	case <-ctx.Done():
		// correcto: el runtime fue cancelado al quedarse sin observadores.
	case <-time.After(time.Second):
		t.Fatal("se esperaba cancelación del runtime al abandonar el QR")
	}
}

// TestRuntimeAbandonKeepsAliveWhenEverActive verifica que una sesión que llegó a
// estar activa NO se cancela aunque se vayan todos los observadores (cerrar la
// pestaña no debe desconectar el WhatsApp del cliente).
func TestRuntimeAbandonKeepsAliveWhenEverActive(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	rt := &sessionRuntime{
		ctx:         ctx,
		cancel:      cancel,
		subscribers: make(map[chan SessionEvent]struct{}),
	}

	_, unsub := rt.subscribe()
	rt.markEverActive()
	unsub()

	select {
	case <-ctx.Done():
		t.Fatal("el runtime no debía cancelarse: la sesión ya estuvo activa")
	case <-time.After(200 * time.Millisecond):
		// correcto: sigue vivo.
	}
}
