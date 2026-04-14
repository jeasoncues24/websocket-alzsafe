package whatsapp

import (
	"sync"

	"go.mau.fi/whatsmeow"
)

type Manager struct {
	mu      sync.RWMutex
	clients map[string]*whatsmeow.Client
}

func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]*whatsmeow.Client),
	}
}

func (m *Manager) Get(accountID string) (*whatsmeow.Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.clients[accountID]
	return c, ok
}

func (m *Manager) Set(accountID string, client *whatsmeow.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[accountID] = client
}
