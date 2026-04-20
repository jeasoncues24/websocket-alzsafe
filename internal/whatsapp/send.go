package whatsapp

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// ErrClientNotConnected is returned when the WhatsApp client is not found or not connected.
var ErrClientNotConnected = errors.New("cliente WhatsApp no conectado para este número")

// SendTextMessage sends a plain-text WhatsApp message via the client registered for accountID.
// accountID must be the NumeroCompleto of the sending phone (e.g. "51999888777").
// destino must be a full phone number, digits only (e.g. "51912345678" or "+51912345678").
func SendTextMessage(ctx context.Context, manager *Manager, accountID string, destino string, contenido string) error {
	sendLogger := NewModuleLogger("WA-SEND")

	accountID = NormalizeAccountID(accountID)
	client, ok := manager.Get(accountID)
	if !ok || client == nil {
		sendLogger.Warnf("send.fail account=%s reason=client_not_found", accountID)
		return ErrClientNotConnected
	}
	if !client.IsConnected() {
		sendLogger.Warnf("send.fail account=%s reason=client_not_connected", accountID)
		return ErrClientNotConnected
	}

	// Normalize destination: strip leading '+' and whitespace
	destino = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(destino), "+"))
	if destino == "" {
		return errors.New("destino vacío")
	}

	toJID := types.NewJID(destino, types.DefaultUserServer)

	resp, err := client.SendMessage(ctx, toJID, &waE2E.Message{
		Conversation: proto.String(contenido),
	})
	if err != nil {
		sendLogger.Errorf("send.fail account=%s to=%s error=%v", accountID, destino, err)
		return err
	}
	sendLogger.Infof("send.ok account=%s to=%s message_id=%s ts=%s", accountID, destino, string(resp.ID), resp.Timestamp.Format(time.RFC3339))
	return nil
}
