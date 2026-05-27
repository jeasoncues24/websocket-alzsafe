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
	starting map[string]bool
}

type sessionRuntime struct {
	ctx     context.Context
	cancel  context.CancelFunc
	client  *whatsmeow.Client
	storage *sqlstore.Container
	events  chan SessionEvent
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
		starting:       make(map[string]bool),
	}
	if manager != nil {
		manager.attachService(s)
	}
	return s
}

func (s *Service) StartSession(accountID string) (<-chan SessionEvent, error) {
	accountID = NormalizeAccountID(accountID)
	if accountID == "" {
		return nil, ErrInvalidAccountID
	}

	s.mu.Lock()
	if existing, ok := s.runtimes[accountID]; ok {
		ch := make(chan SessionEvent, 2)
		if s.sessionStore != nil {
			if state, ok := s.sessionStore.Get(accountID); ok {
				ch <- s.eventFromState(accountID, state)
			} else {
				ch <- SessionEvent{
					Event: "active-" + accountID,
					Data: map[string]any{
						"message":  "Sesion en proceso de inicializacion",
						"isActive": false,
					},
				}
			}
		} else if existing.client != nil && existing.client.IsConnected() {
			ch <- SessionEvent{
				Event: "active-" + accountID,
				Data: map[string]any{
					"message":  "Sesion activa",
					"isActive": true,
				},
			}
		}
		close(ch)
		s.mu.Unlock()
		return ch, nil
	}
	if s.starting[accountID] {
		ch := make(chan SessionEvent, 2)
		if s.sessionStore != nil {
			if state, ok := s.sessionStore.Get(accountID); ok {
				ch <- s.eventFromState(accountID, state)
			} else {
				ch <- SessionEvent{
					Event: "active-" + accountID,
					Data: map[string]any{
						"message":  "Sesion en proceso de inicializacion",
						"isActive": false,
					},
				}
			}
		} else {
			ch <- SessionEvent{
				Event: "active-" + accountID,
				Data: map[string]any{
					"message":  "Sesion en proceso de inicializacion",
					"isActive": false,
				},
			}
		}
		close(ch)
		s.mu.Unlock()
		return ch, nil
	}
	s.starting[accountID] = true
	s.mu.Unlock()
	defer s.clearStarting(accountID)

	container, err := openSQLiteContainer(context.Background(), s.baseDir, accountID)
	if err != nil {
		return nil, err
	}

	device, err := container.GetFirstDevice(context.Background())
	if err != nil {
		_ = container.Close()
		return nil, fmt.Errorf("no se pudo cargar device whatsapp: %w", err)
	}
	if device == nil {
		device = container.NewDevice()
	}

	clientLog := NewWhatsAppClientLogger(accountID)
	client := whatsmeow.NewClient(device, clientLog)
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan SessionEvent, 8)
	runtime := &sessionRuntime{
		ctx:     ctx,
		cancel:  cancel,
		client:  client,
		events:  ch,
		storage: container,
	}

	s.mu.Lock()
	s.runtimes[accountID] = runtime
	s.mu.Unlock()

	if s.manager != nil {
		s.manager.registerClient(accountID, client)
	}
	if s.sessionStore != nil {
		s.sessionStore.SetInitializing(accountID)
	}

	go s.runSession(accountID, runtime)
	return ch, nil
}

func (s *Service) clearStarting(accountID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.starting, accountID)
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
	defer func() {
		runtime.cancel()
		if runtime.client != nil {
			runtime.client.Disconnect()
		}
		if runtime.storage != nil {
			_ = runtime.storage.Close()
		}
		if s.manager != nil {
			s.manager.clearClient(accountID)
		}
		s.mu.Lock()
		delete(s.runtimes, accountID)
		s.mu.Unlock()
		close(runtime.events)
	}()

	emit := func(event string, data map[string]any) {
		select {
		case runtime.events <- SessionEvent{Event: event, Data: data}:
		case <-runtime.ctx.Done():
		}
	}

	emitActive := func(message string, isActive bool, extra map[string]any) {
		data := map[string]any{"message": message, "isActive": isActive}
		for k, v := range extra {
			data[k] = v
		}
		emit("active-"+accountID, data)
	}

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
			select {
			case connectedCh <- struct{}{}:
			default:
			}
			return
		}

		disconnect := disconnectEvent{}
		switch v := evt.(type) {
		case *waEvents.Disconnected:
			disconnect.reason = "disconnect"
		case *waEvents.StreamReplaced:
			disconnect.reason = "stream_replaced"
			disconnect.permanent = true
		case *waEvents.LoggedOut:
			disconnect.reason = "logged_out"
			disconnect.permanent = true
			if v.OnConnect {
				disconnect.detail = v.Reason.String()
			}
		case *waEvents.TemporaryBan:
			disconnect.reason = "temporary_ban"
			disconnect.permanent = true
			disconnect.detail = v.String()
		case *waEvents.ConnectFailure:
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
					s.markDisconnected(accountID, reason)
					emitActive("Sesion desconectada", false, extra)
					return
				}
				
				// Desconexión temporal: whatsmeow lanza autoReconnect en paralelo.
				// Primero, vaciar cualquier evento obsoleto del canal connectedCh para evitar bypass inmediato.
				for len(connectedCh) > 0 {
					<-connectedCh
				}
				
				// Esperar a que confirme reconexión antes de declarar la sesión muerta o recibir desconexión permanente.
				timer := time.NewTimer(reconnectGracePeriod)
				reconnected := false
				expired := false
				
				for !reconnected && !expired {
					select {
					case <-connectedCh:
						timer.Stop()
						s.markConnected(accountID)
						emitActive("Sesion activa", true, nil)
						reconnected = true
					case d := <-disconnectCh:
						if d.permanent {
							timer.Stop()
							s.markDisconnected(accountID, d.reason)
							extraPermanent := map[string]any{"reason": d.reason, "requiresNewQR": true}
							if d.detail != "" {
								extraPermanent["detail"] = d.detail
							}
							emitActive("Sesion desconectada", false, extraPermanent)
							return
						}
						// Si es otra desconexión temporal, actualizamos la razón/detalle pero seguimos esperando la reconexión.
						reason = d.reason
						if d.detail != "" {
							extra["detail"] = d.detail
						}
					case <-timer.C:
						s.markDisconnected(accountID, reason)
						emitActive("Sesion desconectada", false, extra)
						expired = true
						return
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

func (s *Service) eventFromState(accountID string, state storage.SessionState) SessionEvent {
	switch state.Status {
	case "active":
		return SessionEvent{Event: "active-" + accountID, Data: map[string]any{"message": "Sesion activa", "isActive": true}}
	case "qr_pending":
		return SessionEvent{Event: "qr-" + accountID, Data: map[string]any{"message": "Escanee el codigo QR para iniciar sesion.", "qrString": state.QRString}}
	default:
		return SessionEvent{Event: "active-" + accountID, Data: map[string]any{"message": "Sesion desconectada", "isActive": false, "reason": state.Reason}}
	}
}
