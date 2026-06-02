package whatsapp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waEvents "go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"wsapi/internal/storage"
)

var logger waLog.Logger = NewModuleLogger("Service")

// qrExpiresInSec es la validez aproximada de cada código QR emitido por whatsmeow.
// El cliente usa este valor para el countdown visual; el timeout real lo maneja el servidor.
const qrExpiresInSec = 60

type SessionEvent struct {
	Event string
	Data  map[string]any
}

type Service struct {
	manager        *Manager
	sessionStore   *storage.SessionStore
	telefonoStore  *storage.TelefonoStore
	webhookEmitter *WebhookEmitter
	baseDir        string

	mu       sync.Mutex
	runtimes map[string]*sessionRuntime
}

// sessionRuntime representa UNA sola conexión WhatsApp (un cliente whatsmeow) que
// puede ser observada por múltiples WebSockets simultáneamente vía fan-out. Los
// eventos producidos por runSession se difunden a todos los suscriptores; nunca
// se abre más de una conexión WhatsApp por accountID.
type sessionRuntime struct {
	ctx     context.Context
	cancel  context.CancelFunc
	client  *whatsmeow.Client
	storage *sqlstore.Container
	dbPath  string

	mu          sync.Mutex
	subscribers map[chan SessionEvent]struct{}
	last        *SessionEvent
	closed      bool
	everActive  bool // true una vez que la sesión llegó a conectarse al menos una vez
}

// markEverActive marca que la sesión llegó a estar activa. Es sticky: una sesión
// que ya se conectó debe sobrevivir aunque no queden observadores (cerrar la
// pestaña del panel no debe desconectar el WhatsApp del cliente).
func (rt *sessionRuntime) markEverActive() {
	rt.mu.Lock()
	rt.everActive = true
	rt.mu.Unlock()
}

// subscribe registra un nuevo observador y devuelve su canal junto a una función
// para darse de baja. Si ya hay un último evento conocido, se entrega de inmediato
// como snapshot para que el observador vea el estado actual sin esperar.
func (rt *sessionRuntime) subscribe() (<-chan SessionEvent, func()) {
	ch := make(chan SessionEvent, 16)

	rt.mu.Lock()
	if rt.closed {
		if rt.last != nil {
			ch <- *rt.last
		}
		close(ch)
		rt.mu.Unlock()
		return ch, func() {}
	}
	if rt.subscribers == nil {
		rt.subscribers = make(map[chan SessionEvent]struct{})
	}
	rt.subscribers[ch] = struct{}{}
	if rt.last != nil {
		select {
		case ch <- *rt.last:
		default:
		}
	}
	rt.mu.Unlock()

	var once sync.Once
	return ch, func() {
		once.Do(func() {
			rt.mu.Lock()
			if _, ok := rt.subscribers[ch]; ok {
				delete(rt.subscribers, ch)
				close(ch)
			}
			// Si se fue el último observador y la sesión nunca llegó a estar
			// activa (QR abandonado), cancelamos el runtime para no dejar una
			// sesión QR colgada. Las sesiones ya conectadas se mantienen vivas.
			abandon := len(rt.subscribers) == 0 && !rt.everActive
			rt.mu.Unlock()
			if abandon && rt.cancel != nil {
				rt.cancel()
			}
		})
	}
}

// broadcast difunde un evento a todos los suscriptores de forma no bloqueante y
// guarda el último estado para futuros suscriptores. Nunca bloquea al productor.
func (rt *sessionRuntime) broadcast(evt SessionEvent) {
	rt.mu.Lock()
	rt.last = &evt
	for ch := range rt.subscribers {
		select {
		case ch <- evt:
		default:
			// Observador lento: descartamos para no bloquear runSession.
		}
	}
	rt.mu.Unlock()
}

// closeAll cierra todos los canales de suscriptores y marca el runtime terminado.
func (rt *sessionRuntime) closeAll() {
	rt.mu.Lock()
	rt.closed = true
	for ch := range rt.subscribers {
		close(ch)
	}
	rt.subscribers = nil
	rt.mu.Unlock()
}

func NewService(manager *Manager, sessionStore *storage.SessionStore, telefonoStore *storage.TelefonoStore, webhookStore *storage.WebhookStore, baseDir string) *Service {
	logger = NewModuleLogger("Service")

	s := &Service{
		manager:        manager,
		sessionStore:   sessionStore,
		telefonoStore:  telefonoStore,
		webhookEmitter: NewWebhookEmitter(webhookStore, telefonoStore, WebhookEmitterConfig{}),
		baseDir:        baseDir,
		runtimes:       make(map[string]*sessionRuntime),
	}
	if manager != nil {
		manager.attachService(s)
	}
	return s
}

// StartSession inicia (o se une a) la sesión WhatsApp de accountID y devuelve un
// canal de eventos junto a una función de baja (unsubscribe). Es idempotente: si
// ya existe un runtime para el accountID, el llamador se suscribe como observador
// adicional al MISMO runtime — no se abre una segunda conexión WhatsApp. El canal
// permanece vivo hasta que el runtime termina o el llamador llama a unsubscribe.
func (s *Service) StartSession(accountID string) (<-chan SessionEvent, func(), error) {
	accountID = NormalizeAccountID(accountID)
	if accountID == "" {
		return nil, nil, ErrInvalidAccountID
	}

	s.mu.Lock()
	if existing, ok := s.runtimes[accountID]; ok {
		s.mu.Unlock()
		ch, unsub := existing.subscribe()
		return ch, unsub, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	runtime := &sessionRuntime{
		ctx:         ctx,
		cancel:      cancel,
		dbPath:      sqliteDBPath(s.baseDir, accountID),
		subscribers: make(map[chan SessionEvent]struct{}),
	}
	s.runtimes[accountID] = runtime
	s.mu.Unlock()

	if s.sessionStore != nil {
		s.sessionStore.SetInitializing(accountID)
	}

	container, err := openSQLiteContainer(context.Background(), s.baseDir, accountID)
	if err != nil {
		s.abortRuntime(accountID, runtime)
		return nil, nil, err
	}

	device, err := container.GetFirstDevice(context.Background())
	if err != nil {
		_ = container.Close()
		s.abortRuntime(accountID, runtime)
		return nil, nil, fmt.Errorf("no se pudo cargar device whatsapp: %w", err)
	}
	if device == nil {
		device = container.NewDevice()
	}

	clientLog := NewWhatsAppClientLogger(accountID)
	client := whatsmeow.NewClient(device, clientLog)
	runtime.client = client
	runtime.storage = container

	if s.manager != nil {
		s.manager.registerClient(accountID, client)
	}

	// Suscribir al llamador antes de arrancar runSession para no perder eventos.
	ch, unsub := runtime.subscribe()
	go s.runSession(accountID, runtime)
	return ch, unsub, nil
}

// abortRuntime libera un runtime que falló durante su preparación (antes de
// arrancar runSession): cancela el contexto, cierra a los suscriptores y lo
// elimina del registro.
func (s *Service) abortRuntime(accountID string, runtime *sessionRuntime) {
	runtime.cancel()
	runtime.closeAll()
	if s.manager != nil {
		s.manager.clearClient(accountID)
	}
	s.mu.Lock()
	delete(s.runtimes, accountID)
	s.mu.Unlock()
}

func (s *Service) StopSession(accountID string) {
	accountID = NormalizeAccountID(accountID)
	if accountID == "" {
		return
	}

	s.mu.Lock()
	runtime := s.runtimes[accountID]
	s.mu.Unlock()
	if runtime == nil {
		return
	}
	runtime.cancel()
}

func (s *Service) runSession(accountID string, runtime *sessionRuntime) {
	// purgeStore se activa cuando WhatsApp revoca la sesión (logged_out). En ese
	// caso las credenciales en SQLite quedan inservibles y deben eliminarse: de lo
	// contrario el siguiente intento carga un device enrolado, GetQRChannel falla y
	// la sesión entra en bucle sin poder mostrar un QR nuevo.
	purgeStore := false
	defer func() {
		logger.Infof("[SESSION] %s terminó (purge_sqlite=%v)", accountID, purgeStore)
		runtime.cancel()
		if runtime.client != nil {
			runtime.client.Disconnect()
		}
		if runtime.storage != nil {
			_ = runtime.storage.Close()
		}
		if purgeStore && runtime.dbPath != "" {
			if err := removeSQLiteArtifacts(runtime.dbPath); err != nil {
				logger.Warnf("no se pudo eliminar sqlite tras logout para %s: %v", accountID, err)
			} else {
				logger.Infof("sqlite eliminado tras logout para %s; el proximo inicio mostrara un QR nuevo", accountID)
			}
		}
		if s.manager != nil {
			s.manager.clearClient(accountID)
		}
		s.mu.Lock()
		delete(s.runtimes, accountID)
		s.mu.Unlock()
		runtime.closeAll()
	}()

	emit := func(event string, data map[string]any) {
		runtime.broadcast(SessionEvent{Event: event, Data: data})
	}

	emitActive := func(message string, isActive bool, extra map[string]any) {
		if isActive {
			runtime.markEverActive()
		}
		data := map[string]any{"message": message, "isActive": isActive}
		for k, v := range extra {
			data[k] = v
		}
		emit("active-"+accountID, data)
	}

	logger.Infof("[SESSION] %s iniciando", accountID)
	if s.sessionStore != nil {
		s.sessionStore.SetInitializing(accountID)
		s.sessionStore.AppendEvent(accountID, "initializing", "")
	}
	emitActive("Sesion en proceso de inicializacion", false, nil)

	type disconnectEvent struct {
		reason    string
		detail    string
		permanent bool
	}
	connectedCh := make(chan struct{}, 1)
	disconnectCh := make(chan disconnectEvent, 1)
	handlerID := runtime.client.AddEventHandler(func(evt interface{}) {
		switch evt.(type) {
		case *waEvents.Message, *waEvents.Receipt:
			go s.handleWhatsAppEvent(accountID, evt)
			return
		case *waEvents.Connected:
			logger.Infof("[SESSION] %s ← WA:Connected", accountID)
			select {
			case connectedCh <- struct{}{}:
			default:
			}
			return
		}

		disconnect := disconnectEvent{}
		switch v := evt.(type) {
		case *waEvents.Disconnected:
			logger.Warnf("[SESSION] %s ← WA:Disconnected", accountID)
			disconnect.reason = "disconnect"
		case *waEvents.StreamReplaced:
			logger.Warnf("[SESSION] %s ← WA:StreamReplaced", accountID)
			disconnect.reason = "stream_replaced"
			disconnect.permanent = true
		case *waEvents.LoggedOut:
			logger.Warnf("[SESSION] %s ← WA:LoggedOut onConnect=%v reason=%s", accountID, v.OnConnect, v.Reason.String())
			disconnect.reason = "logged_out"
			disconnect.permanent = true
			if v.OnConnect {
				disconnect.detail = v.Reason.String()
			}
		case *waEvents.TemporaryBan:
			logger.Warnf("[SESSION] %s ← WA:TemporaryBan %s", accountID, v.String())
			disconnect.reason = "temporary_ban"
			disconnect.permanent = true
			disconnect.detail = v.String()
		case *waEvents.ConnectFailure:
			logger.Warnf("[SESSION] %s ← WA:ConnectFailure reason=%s", accountID, v.Reason.String())
			disconnect.reason = "connect_failure"
			disconnect.detail = v.Reason.String()
		default:
			return
		}
		select {
		case disconnectCh <- disconnect:
		default:
		}
	})
	defer runtime.client.RemoveEventHandler(handlerID)

	const reconnectGracePeriod = 75 * time.Second
	const maxReconnectAttempts = 5

	waitForDisconnect := func() {
		for {
			select {
			case disconnect := <-disconnectCh:
				if runtime.ctx.Err() != nil {
					return
				}
				reason := disconnect.reason
				if reason == "" {
					reason = "disconnect"
				}
				extra := map[string]any{"reason": reason, "requiresNewQR": true}
				if disconnect.detail != "" {
					extra["detail"] = disconnect.detail
				}
				if disconnect.permanent {
					if disconnect.reason == "logged_out" {
						purgeStore = true
					}
					s.markDisconnected(accountID, reason)
					emitActive("Sesion desconectada", false, extra)
					return
				}

				// Desconexión temporal: whatsmeow lanza autoReconnect en paralelo.
				// Primero, vaciar cualquier evento obsoleto del canal connectedCh para evitar bypass inmediato.
				for len(connectedCh) > 0 {
					<-connectedCh
				}

				// Esperar a que confirme reconexión. Si el timer expira, reintentamos hasta
				// maxReconnectAttempts veces antes de declarar la sesión muerta. Esto cubre
				// el caso donde whatsmeow usa backoff exponencial y su primer intento exitoso
				// tarda más de reconnectGracePeriod.
				reconnectAttempt := 0
				timer := time.NewTimer(reconnectGracePeriod)
				reconnected := false

				for !reconnected {
					select {
					case <-connectedCh:
						timer.Stop()
						reconnectAttempt = 0
						s.markConnected(accountID)
						emitActive("Sesion activa", true, nil)
						reconnected = true
					case d := <-disconnectCh:
						if d.permanent {
							if d.reason == "logged_out" {
								purgeStore = true
							}
							timer.Stop()
							s.markDisconnected(accountID, d.reason)
							extraPermanent := map[string]any{"reason": d.reason, "requiresNewQR": true}
							if d.detail != "" {
								extraPermanent["detail"] = d.detail
							}
							emitActive("Sesion desconectada", false, extraPermanent)
							return
						}
						reason = d.reason
						if d.detail != "" {
							extra["detail"] = d.detail
						}
					case <-timer.C:
						reconnectAttempt++
						if reconnectAttempt < maxReconnectAttempts {
							logger.Warnf("reconexion intento %d/%d para %s, esperando...", reconnectAttempt, maxReconnectAttempts, accountID)
							for len(connectedCh) > 0 {
								<-connectedCh
							}
							timer = time.NewTimer(reconnectGracePeriod)
						} else {
							logger.Warnf("reconexion fallida tras %d intentos para %s", maxReconnectAttempts, accountID)
							s.markDisconnected(accountID, reason)
							emitActive("Sesion desconectada", false, extra)
							return
						}
					case <-runtime.ctx.Done():
						timer.Stop()
						return
					}
				}
			case <-runtime.ctx.Done():
				return
			}
		}
	}

	qrChan, qrErr := runtime.client.GetQRChannel(runtime.ctx)
	connectErr := runtime.client.ConnectContext(runtime.ctx)
	if connectErr != nil {
		if runtime.ctx.Err() == nil {
			errMsg := connectErr.Error()
			logger.Warnf("ConnectContext failed for %s: %v", accountID, connectErr)
			s.markDisconnected(accountID, "connect_error")
			emitActive("No se pudo conectar", false, map[string]any{"reason": "connect_error", "requiresNewQR": true, "detail": errMsg})
		}
		return
	}

	if qrErr != nil {
		logger.Infof("Device already logged in for %s, waiting for connection...", accountID)
		if waitForConnection(runtime.client, 30*time.Second) {
			s.markConnected(accountID)
			emitActive("Sesion activa", true, nil)
		} else if runtime.ctx.Err() == nil {
			s.markDisconnected(accountID, "connect_timeout")
			emitActive("No se pudo conectar", false, map[string]any{"reason": "connect_timeout", "requiresNewQR": true})
			return
		}
		waitForDisconnect()
		return
	} else {
		activeEmitted := false
		for {
			select {
			case evt, ok := <-qrChan:
				if !ok {
					if !activeEmitted && runtime.ctx.Err() == nil {
						s.markDisconnected(accountID, "qr_channel_closed")
						emitActive("Sesion cerrada antes de completar el QR", false, map[string]any{"reason": "qr_channel_closed", "requiresNewQR": true})
						return
					}
					waitForDisconnect()
					return
				}
				switch evt.Event {
				case "code":
					if s.sessionStore != nil {
						s.sessionStore.SetQRPending(accountID, evt.Code)
						s.sessionStore.AppendEvent(accountID, "qr_generated", "")
					}
					s.syncTelefonoQR(accountID, evt.Code)
					emit("qr-"+accountID, map[string]any{
						"message":    "Escanee el codigo QR para iniciar sesion.",
						"qrString":   evt.Code,
						"expires_in": qrExpiresInSec,
					})
				case "timeout":
					s.markDisconnected(accountID, "qr_timeout")
					emitActive("Sesion QR expirada", false, map[string]any{"reason": "qr_timeout", "requiresNewQR": true})
					runtime.cancel()
					return
				case "error":
					s.markDisconnected(accountID, "qr_error")
					extra := map[string]any{"reason": "qr_error", "requiresNewQR": true}
					if evt.Error != nil {
						extra["detail"] = evt.Error.Error()
					}
					emitActive("Error generando QR", false, extra)
					runtime.cancel()
					return
				case "success":
					s.markConnected(accountID)
					if !activeEmitted {
						emitActive("Sesion activa", true, nil)
						activeEmitted = true
					}
					waitForDisconnect()
					return
				}
			case <-runtime.ctx.Done():
				return
			}
		}
	}
}

func (s *Service) markConnected(accountID string) {
	logger.Infof("[SESSION] %s → CONECTADO", accountID)
	if s.sessionStore != nil {
		s.sessionStore.SetActive(accountID)
		s.sessionStore.AppendEvent(accountID, "connected", "")
	}
	s.syncTelefonoConnected(accountID)
	if s.webhookEmitter != nil {
		if err := s.webhookEmitter.EmitSessionConnectedByAccount(accountID); err != nil {
			logger.Warnf("webhook emit session.connected failed for %s: %v", accountID, err)
		}
	}
}

func (s *Service) markDisconnected(accountID, reason string) {
	logger.Warnf("[SESSION] %s → DESCONECTADO reason=%s", accountID, reason)
	if s.sessionStore != nil {
		s.sessionStore.SetDisconnected(accountID, reason)
		s.sessionStore.AppendEvent(accountID, "disconnected", reason)
	}
	s.syncTelefonoDisconnected(accountID)
	if s.webhookEmitter != nil {
		if err := s.webhookEmitter.EmitSessionDisconnectedByAccount(accountID, reason); err != nil {
			logger.Warnf("webhook emit session.disconnected failed for %s: %v", accountID, err)
		}
	}
}

func (s *Service) syncTelefonoQR(accountID, qr string) {
	if s.telefonoStore == nil {
		return
	}
	phone, err := s.telefonoStore.GetByNumeroCompletoNormalized(accountID)
	if err != nil || phone == nil {
		return
	}
	_ = s.telefonoStore.UpdateQRString(phone.ID, qr)
}

func (s *Service) syncTelefonoConnected(accountID string) {
	if s.telefonoStore == nil {
		return
	}
	phone, err := s.telefonoStore.GetByNumeroCompletoNormalized(accountID)
	if err != nil || phone == nil {
		return
	}
	_ = s.telefonoStore.SetConnected(phone.ID)
}

func (s *Service) syncTelefonoDisconnected(accountID string) {
	if s.telefonoStore == nil {
		return
	}
	phone, err := s.telefonoStore.GetByNumeroCompletoNormalized(accountID)
	if err != nil || phone == nil {
		return
	}
	_ = s.telefonoStore.SetDisconnected(phone.ID)
}

func (s *Service) handleWhatsAppEvent(accountID string, evt interface{}) {
	if s == nil || s.webhookEmitter == nil {
		return
	}

	switch v := evt.(type) {
	case *waEvents.Message:
		if err := s.webhookEmitter.EmitMessageReceivedByAccount(accountID, v); err != nil {
			logger.Warnf("webhook emit message.received failed for %s: %v", accountID, err)
		}
	case *waEvents.Receipt:
		for _, messageID := range v.MessageIDs {
			referenceID := ""
			if s.manager != nil {
				if resolved, ok := s.manager.ResolveOutboundMessageReference(accountID, string(messageID)); ok {
					referenceID = resolved
				}
			}
			if err := s.webhookEmitter.EmitMessageStatusUpdateByAccount(accountID, string(messageID), referenceID, v.Type, v.Timestamp); err != nil {
				logger.Warnf("webhook emit message.status_update failed for %s message=%s: %v", accountID, string(messageID), err)
			}
		}
	}
}
