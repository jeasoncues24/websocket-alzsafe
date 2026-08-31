package storage

import (
	"strings"
	"sync"
	"time"
)

type SessionEventEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Details   string    `json:"details,omitempty"`
}

type SessionState struct {
	AccountID string
	Status    string
	IsActive  bool
	QRString  string
	Reason    string
	UpdatedAt time.Time
	SessionID string
	Events    []SessionEventEntry
}

type SessionStore struct {
	mu    sync.RWMutex
	state map[string]SessionState

	cbMu     sync.RWMutex
	onChange func()
}

func NewSessionStore() *SessionStore {
	return &SessionStore{state: make(map[string]SessionState)}
}

// SetOnChange registra un callback que se invoca tras cada mutación del store
// (set / AppendEvent), fuera del lock. Lo usa el stream SSE del reporte de sesiones
// para recomputar el snapshot ante cualquier transición. El callback debe ser no
// bloqueante: aquí solo se dispara una señal.
func (s *SessionStore) SetOnChange(fn func()) {
	s.cbMu.Lock()
	s.onChange = fn
	s.cbMu.Unlock()
}

func (s *SessionStore) fireOnChange() {
	s.cbMu.RLock()
	fn := s.onChange
	s.cbMu.RUnlock()
	if fn != nil {
		fn()
	}
}

func (s *SessionStore) SetInitializing(accountID string) {
	accountID = normalizeSessionAccountID(accountID)
	s.set(SessionState{
		AccountID: accountID,
		Status:    "initializing",
		IsActive:  false,
		UpdatedAt: time.Now(),
	})
}

func (s *SessionStore) SetQRPending(accountID, qr string) {
	accountID = normalizeSessionAccountID(accountID)
	s.set(SessionState{
		AccountID: accountID,
		Status:    "qr_pending",
		IsActive:  false,
		QRString:  qr,
		UpdatedAt: time.Now(),
	})
}

func (s *SessionStore) SetActive(accountID string) {
	accountID = normalizeSessionAccountID(accountID)
	s.set(SessionState{
		AccountID: accountID,
		Status:    "active",
		IsActive:  true,
		UpdatedAt: time.Now(),
	})
}

func (s *SessionStore) SetDisconnected(accountID, reason string) {
	accountID = normalizeSessionAccountID(accountID)
	s.set(SessionState{
		AccountID: accountID,
		Status:    "disconnected",
		IsActive:  false,
		Reason:    reason,
		UpdatedAt: time.Now(),
	})
}

// SetClientOutdated marca la sesión como no conectable porque WhatsApp rechazó el
// handshake con "client outdated" (405). No se resuelve reintentando ni con un QR
// nuevo: hay que actualizar la librería whatsmeow y redesplegar el servidor.
func (s *SessionStore) SetClientOutdated(accountID string) {
	accountID = normalizeSessionAccountID(accountID)
	s.set(SessionState{
		AccountID: accountID,
		Status:    "client_outdated",
		IsActive:  false,
		Reason:    "client_outdated",
		UpdatedAt: time.Now(),
	})
}

func (s *SessionStore) Get(accountID string) (SessionState, bool) {
	accountID = normalizeSessionAccountID(accountID)
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.state[accountID]
	return v, ok
}

func (s *SessionStore) set(v SessionState) {
	defer s.fireOnChange()
	v.AccountID = normalizeSessionAccountID(v.AccountID)
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.state[v.AccountID]; ok {
		v.Events = existing.Events
	}
	s.state[v.AccountID] = v
}

func (s *SessionStore) AppendEvent(accountID, eventType, details string) {
	defer s.fireOnChange()
	accountID = normalizeSessionAccountID(accountID)
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.state[accountID]
	if !ok {
		state = SessionState{AccountID: accountID}
	}
	state.Events = append(state.Events, SessionEventEntry{
		Timestamp: time.Now(),
		Type:      eventType,
		Details:   details,
	})
	if len(state.Events) > 20 {
		state.Events = state.Events[len(state.Events)-20:]
	}
	s.state[accountID] = state
}

func (s *SessionStore) ActiveCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, v := range s.state {
		if v.IsActive {
			count++
		}
	}
	return count
}

func (s *SessionStore) CountByStatus() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]int)
	for _, v := range s.state {
		result[v.Status]++
	}
	return result
}

func normalizeSessionAccountID(accountID string) string {
	accountID = strings.TrimSpace(accountID)
	accountID = strings.TrimPrefix(accountID, "+")
	return accountID
}
