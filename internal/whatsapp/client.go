package whatsapp

import (
	"errors"
	"time"

	"go.mau.fi/whatsmeow"
)

var ErrInvalidAccountID = errors.New("invalid account id")

// StartSession starts a WhatsApp session and returns the stream of session events.
// When no service is attached, it falls back to a minimal in-memory QR event for compatibility.
func StartSession(manager *Manager, accountID string) (<-chan SessionEvent, error) {
	if manager == nil {
		return nil, errors.New("manager is required")
	}

	if NormalizeAccountID(accountID) == "" {
		return nil, ErrInvalidAccountID
	}

	service := manager.getService()
	if service == nil {
		events := make(chan SessionEvent, 2)
		manager.registerClient(accountID, nil)
		events <- SessionEvent{
			Event: "qr-" + NormalizeAccountID(accountID),
			Data: map[string]any{
				"message":  "Escanee el codigo QR para iniciar sesion.",
				"qrString": GenerateQRCode(accountID),
			},
		}
		events <- SessionEvent{
			Event: "active-" + NormalizeAccountID(accountID),
			Data: map[string]any{
				"message":  "Sesion en proceso de inicializacion",
				"isActive": false,
			},
		}
		close(events)
		return events, nil
	}

	return service.StartSession(accountID)
}

func waitForConnection(client *whatsmeow.Client, timeout time.Duration) bool {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if client.IsConnected() {
			return true
		}
		select {
		case <-deadline.C:
			return client.IsConnected()
		case <-ticker.C:
		}
	}
}
