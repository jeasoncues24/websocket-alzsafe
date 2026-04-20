package whatsapp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	wa "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	"wsapi/internal/domain"
)

// ErrClientNotConnected is returned when the WhatsApp client is not found or not connected.
var ErrClientNotConnected = errors.New("cliente WhatsApp no conectado para este número")

type preparedAttachment struct {
	Payload   domain.AttachmentPayload
	Upload    wa.UploadResponse
	MediaType wa.MediaType
}

// SendTextMessage sends a plain-text WhatsApp message via the client registered for accountID.
// accountID must be the NumeroCompleto of the sending phone (e.g. "51999888777").
// destino must be a full phone number, digits only (e.g. "51912345678" or "+51912345678").
func SendTextMessage(ctx context.Context, manager *Manager, accountID string, destino string, contenido string) error {
	return SendRichMessage(ctx, manager, accountID, destino, contenido, nil)
}

// SendRichMessage sends either plain text or a single media attachment with optional text.
func SendRichMessage(ctx context.Context, manager *Manager, accountID string, destino string, contenido string, attachments []domain.AttachmentPayload) error {
	sendLogger := NewModuleLogger("WA-SEND")
	if manager == nil {
		sendLogger.Warnf("send.fail account=%s reason=manager_nil", accountID)
		return ErrClientNotConnected
	}

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

	prepared, err := prepareAttachments(ctx, client, attachments)
	if err != nil {
		sendLogger.Errorf("send.fail account=%s reason=attachment_prepare_error error=%v", accountID, err)
		return err
	}

	return sendPreparedMessage(ctx, sendLogger, client, accountID, destino, contenido, prepared)
}

func prepareAttachments(ctx context.Context, client *wa.Client, attachments []domain.AttachmentPayload) ([]preparedAttachment, error) {
	if len(attachments) == 0 {
		return nil, nil
	}
	if len(attachments) > 1 {
		return nil, errors.New("solo se permite un adjunto por mensaje")
	}

	mediaType, err := mediaTypeForMIME(attachments[0].MIMEType)
	if err != nil {
		return nil, err
	}

	data, err := attachments[0].Decode()
	if err != nil {
		return nil, err
	}

	upload, err := client.Upload(ctx, data, mediaType)
	if err != nil {
		return nil, err
	}

	return []preparedAttachment{{
		Payload:   attachments[0],
		Upload:    upload,
		MediaType: mediaType,
	}}, nil
}

func sendPreparedMessage(ctx context.Context, sendLogger waLog.Logger, client *wa.Client, accountID string, destino string, contenido string, prepared []preparedAttachment) error {
	// Normalize destination: strip leading '+' and whitespace
	destino = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(destino), "+"))
	if destino == "" {
		return errors.New("destino vacío")
	}

	toJID := types.NewJID(destino, types.DefaultUserServer)

	msg, err := buildMessage(contenido, prepared)
	if err != nil {
		return err
	}

	resp, err := client.SendMessage(ctx, toJID, msg)
	if err != nil {
		sendLogger.Errorf("send.fail account=%s to=%s error=%v", accountID, destino, err)
		return err
	}

	mediaLabel := "text"
	if len(prepared) == 1 {
		mediaLabel = string(prepared[0].MediaType)
	}
	sendLogger.Infof("send.ok account=%s to=%s media=%s message_id=%s ts=%s", accountID, destino, mediaLabel, string(resp.ID), resp.Timestamp.Format(time.RFC3339))
	return nil
}

func buildMessage(contenido string, prepared []preparedAttachment) (*waE2E.Message, error) {
	caption := strings.TrimSpace(contenido)
	if len(prepared) == 0 {
		return &waE2E.Message{Conversation: proto.String(caption)}, nil
	}

	att := prepared[0]
	mimeType := strings.TrimSpace(att.Payload.MIMEType)
	fileName := strings.TrimSpace(att.Payload.Nombre)

	switch att.MediaType {
	case wa.MediaImage:
		return &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				Caption:       optionalString(caption),
				Mimetype:      proto.String(mimeType),
				URL:           &att.Upload.URL,
				DirectPath:    &att.Upload.DirectPath,
				MediaKey:      att.Upload.MediaKey,
				FileEncSHA256: att.Upload.FileEncSHA256,
				FileSHA256:    att.Upload.FileSHA256,
				FileLength:    proto.Uint64(att.Upload.FileLength),
			},
		}, nil
	case wa.MediaVideo:
		return &waE2E.Message{
			VideoMessage: &waE2E.VideoMessage{
				Caption:       optionalString(caption),
				Mimetype:      proto.String(mimeType),
				URL:           &att.Upload.URL,
				DirectPath:    &att.Upload.DirectPath,
				MediaKey:      att.Upload.MediaKey,
				FileEncSHA256: att.Upload.FileEncSHA256,
				FileSHA256:    att.Upload.FileSHA256,
				FileLength:    proto.Uint64(att.Upload.FileLength),
			},
		}, nil
	case wa.MediaAudio:
		if caption != "" {
			return nil, errors.New("audio attachments do not support mensaje")
		}
		return &waE2E.Message{
			AudioMessage: &waE2E.AudioMessage{
				Mimetype:      proto.String(mimeType),
				URL:           &att.Upload.URL,
				DirectPath:    &att.Upload.DirectPath,
				MediaKey:      att.Upload.MediaKey,
				FileEncSHA256: att.Upload.FileEncSHA256,
				FileSHA256:    att.Upload.FileSHA256,
				FileLength:    proto.Uint64(att.Upload.FileLength),
				PTT:           proto.Bool(false),
			},
		}, nil
	default:
		return &waE2E.Message{
			DocumentMessage: &waE2E.DocumentMessage{
				Title:         optionalString(fileName),
				FileName:      optionalString(fileName),
				Caption:       optionalString(caption),
				Mimetype:      proto.String(mimeType),
				URL:           &att.Upload.URL,
				DirectPath:    &att.Upload.DirectPath,
				MediaKey:      att.Upload.MediaKey,
				FileEncSHA256: att.Upload.FileEncSHA256,
				FileSHA256:    att.Upload.FileSHA256,
				FileLength:    proto.Uint64(att.Upload.FileLength),
			},
		}, nil
	}
}

func optionalString(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return proto.String(s)
}

func mediaTypeForMIME(mimeType string) (wa.MediaType, error) {
	switch strings.TrimSpace(strings.ToLower(mimeType)) {
	case "image/jpeg", "image/png":
		return wa.MediaImage, nil
	case "video/mp4":
		return wa.MediaVideo, nil
	case "audio/mpeg", "audio/mp4", "audio/ogg", "audio/wav":
		return wa.MediaAudio, nil
	case "application/pdf", "application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return wa.MediaDocument, nil
	default:
		return "", fmt.Errorf("attachment MIME type not allowed: %s", mimeType)
	}
}
