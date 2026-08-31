package http

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// fleetHub difunde el snapshot del reporte de sesiones a todos los clientes SSE
// conectados. Un único productor (Run) construye el snapshot y lo publica; el
// coste no crece con el número de observadores.
//
// Invariante: asume un solo proceso backend. Con varias réplicas el estado en
// memoria (Manager, SessionStore) es parcial y haría falta un pub/sub compartido.
type fleetHub struct {
	build func() sessionsSnapshotDTO

	mu     sync.Mutex
	subs   map[chan []byte]struct{}
	latest []byte

	notify chan struct{}
}

func newFleetHub(build func() sessionsSnapshotDTO) *fleetHub {
	return &fleetHub{
		build:  build,
		subs:   make(map[chan []byte]struct{}),
		notify: make(chan struct{}, 1),
	}
}

// Trigger señala que el estado cambió. No bloquea: si ya hay una señal pendiente
// se descarta, porque el próximo recomputo ya reflejará el último estado.
func (h *fleetHub) Trigger() {
	select {
	case h.notify <- struct{}{}:
	default:
	}
}

// Subscribe registra un cliente SSE. Devuelve el canal por el que llegan snapshots
// serializados (JSON) y una función idempotente para darse de baja. Si ya hay un
// snapshot conocido se entrega de inmediato para que el primer render no espere.
func (h *fleetHub) Subscribe() (<-chan []byte, func()) {
	ch := make(chan []byte, 1)

	h.mu.Lock()
	h.subs[ch] = struct{}{}
	if h.latest != nil {
		ch <- h.latest
	}
	h.mu.Unlock()

	var once sync.Once
	unsub := func() {
		once.Do(func() {
			h.mu.Lock()
			delete(h.subs, ch)
			h.mu.Unlock()
			close(ch)
		})
	}
	return ch, unsub
}

// publish construye el snapshot una vez y lo entrega a cada suscriptor con
// semántica "gana el más reciente": si el buffer del cliente está lleno se
// descarta el intermedio.
func (h *fleetHub) publish() {
	data, err := json.Marshal(h.build())
	if err != nil {
		return
	}

	h.mu.Lock()
	h.latest = data
	for ch := range h.subs {
		select {
		case ch <- data:
		default:
			select {
			case <-ch:
			default:
			}
			select {
			case ch <- data:
			default:
			}
		}
	}
	h.mu.Unlock()
}

// Run es el único productor de snapshots. Publica ante señales de Trigger (con
// debounce para agrupar ráfagas, p. ej. durante el bootstrap) y, como red de
// seguridad, cada 5 s. Termina al cancelarse ctx; los handlers SSE salen solos
// por su propio contexto de request.
func (h *fleetHub) Run(ctx context.Context) {
	const debounce = 250 * time.Millisecond

	safety := time.NewTicker(5 * time.Second)
	defer safety.Stop()

	debounceTimer := time.NewTimer(debounce)
	if !debounceTimer.Stop() {
		<-debounceTimer.C
	}
	pending := false

	h.publish()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.notify:
			if !pending {
				debounceTimer.Reset(debounce)
				pending = true
			}
		case <-debounceTimer.C:
			pending = false
			h.publish()
		case <-safety.C:
			h.publish()
		}
	}
}
