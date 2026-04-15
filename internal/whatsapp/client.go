package whatsapp

import "errors"

var ErrInvalidAccountID = errors.New("invalid account id")

// StartSession registers a company session in initializing state.
// A real WhatsApp client is attached in later stories.
func StartSession(manager *Manager, accountID string) error {
	if manager == nil {
		return errors.New("manager is required")
	}

	if NormalizeAccountID(accountID) == "" {
		return ErrInvalidAccountID
	}

	// nil client means "session requested/initializing" for now.
	manager.Set(accountID, nil)
	return nil
}
