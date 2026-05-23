package whatsapp

import (
	"strings"
	"sync"

	"go.mau.fi/whatsmeow"
)

type Manager struct {
	mu           sync.RWMutex
	clients      map[string]*whatsmeow.Client
	service      *Service
	outboundRefs map[string]string
}

func NewManager() *Manager {
	return &Manager{
		clients:      make(map[string]*whatsmeow.Client),
		outboundRefs: make(map[string]string),
	}
}

func (m *Manager) attachService(service *Service) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.service = service
}

func (m *Manager) Get(accountID string) (*whatsmeow.Client, bool) {
	accountID = NormalizeAccountID(accountID)
	if accountID == "" {
		return nil, false
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.clients[accountID]
	return c, ok
}

func (m *Manager) Set(accountID string, client *whatsmeow.Client) {
	accountID = NormalizeAccountID(accountID)
	if accountID == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[accountID] = client
}

func (m *Manager) Delete(accountID string) {
	accountID = NormalizeAccountID(accountID)
	if accountID == "" {
		return
	}

	m.mu.RLock()
	service := m.service
	m.mu.RUnlock()
	if service != nil {
		service.StopSession(accountID)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, accountID)
}

func (m *Manager) Exists(accountID string) bool {
	_, ok := m.Get(accountID)
	return ok
}

func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

func (m *Manager) ListKeys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.clients))
	for k := range m.clients {
		keys = append(keys, k)
	}
	return keys
}

func (m *Manager) clearClient(accountID string) {
	accountID = NormalizeAccountID(accountID)
	if accountID == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, accountID)
}

func (m *Manager) registerClient(accountID string, client *whatsmeow.Client) {
	accountID = NormalizeAccountID(accountID)
	if accountID == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[accountID] = client
}

func (m *Manager) getService() *Service {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.service
}

func (m *Manager) RegisterOutboundMessageReference(accountID, providerMessageID, referenceID string) {
	accountID = NormalizeAccountID(accountID)
	providerMessageID = strings.TrimSpace(providerMessageID)
	referenceID = strings.TrimSpace(referenceID)
	if accountID == "" || providerMessageID == "" || referenceID == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.outboundRefs[outboundReferenceKey(accountID, providerMessageID)] = referenceID
}

func (m *Manager) ResolveOutboundMessageReference(accountID, providerMessageID string) (string, bool) {
	accountID = NormalizeAccountID(accountID)
	providerMessageID = strings.TrimSpace(providerMessageID)
	if accountID == "" || providerMessageID == "" {
		return "", false
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	referenceID, ok := m.outboundRefs[outboundReferenceKey(accountID, providerMessageID)]
	return referenceID, ok
}

func outboundReferenceKey(accountID, providerMessageID string) string {
	return accountID + "|" + providerMessageID
}

// NormalizeAccountID normalizes account ID by trimming whitespace
func NormalizeAccountID(accountID string) string {
	accountID = strings.TrimSpace(accountID)
	accountID = strings.TrimPrefix(accountID, "+")
	return accountID
}
