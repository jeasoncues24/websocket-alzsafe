package storage

import (
	"strings"
	"sync"
	"time"
)

type SessionState struct {
	AccountID string
	Status    string
	IsActive  bool
	QRString  string
	Reason    string
	UpdatedAt time.Time
	SessionID string
}

type SessionStore struct {
	mu    sync.RWMutex
	state map[string]SessionState
}

func NewSessionStore() *SessionStore {
	return &SessionStore{state: make(map[string]SessionState)}
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

func (s *SessionStore) Get(accountID string) (SessionState, bool) {
	accountID = normalizeSessionAccountID(accountID)
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.state[accountID]
	return v, ok
}

func (s *SessionStore) set(v SessionState) {
	v.AccountID = normalizeSessionAccountID(v.AccountID)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state[v.AccountID] = v
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
