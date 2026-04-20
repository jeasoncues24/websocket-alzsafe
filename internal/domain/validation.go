package domain

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidateMessageRequest validates a message request payload.
func ValidateMessageRequest(req *MessageRequest) *ValidationError {
	if req.TelefonoID <= 0 {
		return &ValidationError{Code: ErrorCodeMissingField, Message: "telefono_id is required and must be greater than 0"}
	}

	if strings.TrimSpace(req.Destino) == "" {
		return &ValidationError{Code: ErrorCodeMissingField, Message: "destino is required"}
	}

	if err := validatePhoneNumber(req.Destino); err != nil {
		return err
	}

	if validErr := ValidateAttachments(req.Adjuntos); validErr != nil {
		return validErr
	}

	if strings.TrimSpace(req.Mensaje) == "" && len(req.Adjuntos) == 0 {
		return &ValidationError{Code: ErrorCodeEmptyMessage, Message: "mensaje cannot be empty"}
	}

	if len(req.Adjuntos) == 1 && !attachmentAllowsCaption(req.Adjuntos[0].MIMEType) && strings.TrimSpace(req.Mensaje) != "" {
		return &ValidationError{Code: ErrorCodeValidation, Message: "audio attachments cannot include mensaje; send the text separately"}
	}

	return nil
}

// ValidateAttachments validates a slice of attachment payloads.
func ValidateAttachments(attachments []AttachmentPayload) *ValidationError {
	if len(attachments) == 0 {
		return nil
	}
	if len(attachments) > MaxAttachmentsPerMsg {
		return &ValidationError{Code: ErrorCodeAttachmentCountExceeded, Message: "only one attachment per message is supported"}
	}

	totalSize := 0
	for i := range attachments {
		if validErr := validateAttachment(&attachments[i]); validErr != nil {
			return validErr
		}
		totalSize += attachments[i].TamanoValidado
	}

	if totalSize > MaxAttachmentSizePerMsg {
		return &ValidationError{Code: ErrorCodeAttachmentSizeExceeded, Message: "total attachment size exceeds 20MB limit"}
	}

	return nil
}

// ValidateAttachment validates a single attachment payload.
func ValidateAttachment(att AttachmentPayload) *ValidationError {
	return validateAttachment(&att)
}

func validateAttachment(att *AttachmentPayload) *ValidationError {
	if att == nil {
		return &ValidationError{Code: ErrorCodeAttachmentFormatInvalid, Message: "attachment cannot be nil"}
	}

	if strings.TrimSpace(att.Nombre) == "" {
		return &ValidationError{Code: ErrorCodeAttachmentNameInvalid, Message: "attachment nombre is required"}
	}

	if err := validateAttachmentName(att.Nombre); err != nil {
		return err
	}
	mimeType := strings.TrimSpace(strings.ToLower(att.MIMEType))

	if !AllowedMIMETypes[mimeType] {
		return &ValidationError{Code: ErrorCodeAttachmentTypeNotAllowed, Message: "attachment MIME type not allowed: " + att.MIMEType}
	}

	if err := validateMIMETypeExtensionMatch(att.Nombre, mimeType); err != nil {
		return err
	}

	decodedData, err := att.Decode()
	if err != nil {
		return &ValidationError{Code: ErrorCodeAttachmentFormatInvalid, Message: "attachment contenido_base64 is not valid base64"}
	}

	actualSize := len(decodedData)
	att.TamanoValidado = actualSize

	if actualSize > MaxAttachmentSizePerFile {
		return &ValidationError{Code: ErrorCodeAttachmentSizeExceeded, Message: "attachment exceeds 5MB limit"}
	}
	if actualSize == 0 {
		return &ValidationError{Code: ErrorCodeAttachmentFormatInvalid, Message: "attachment is empty"}
	}

	return nil
}

// ValidateBroadcastRequest validates a broadcast request payload.
func ValidateBroadcastRequest(req *BroadcastRequest) *ValidationError {
	if req.TelefonoID <= 0 {
		return &ValidationError{Code: ErrorCodeMissingField, Message: "telefono_id is required and must be greater than 0"}
	}

	if len(req.ListaDifusion) == 0 {
		return &ValidationError{Code: ErrorCodeValidation, Message: "lista_difusion must be a non-empty array"}
	}

	if len(req.ListaDifusion) > MaxBroadcastItems {
		return &ValidationError{Code: ErrorCodeValidation, Message: fmt.Sprintf("lista_difusion exceeds maximum of %d items", MaxBroadcastItems)}
	}

	if validErr := ValidateAttachments(req.Adjuntos); validErr != nil {
		return validErr
	}

	for i, item := range req.ListaDifusion {
		if err := validatePhoneNumber(item.Destino); err != nil {
			return &ValidationError{Code: ErrorCodeInvalidPhoneFormat, Message: fmt.Sprintf("item[%d]: %s", i, err.Message)}
		}
		if strings.TrimSpace(item.Mensaje) == "" && len(req.Adjuntos) == 0 {
			return &ValidationError{Code: ErrorCodeEmptyMessage, Message: fmt.Sprintf("item[%d]: mensaje cannot be empty", i)}
		}
		if len(req.Adjuntos) == 1 && !attachmentAllowsCaption(req.Adjuntos[0].MIMEType) && strings.TrimSpace(item.Mensaje) != "" {
			return &ValidationError{Code: ErrorCodeValidation, Message: fmt.Sprintf("item[%d]: audio attachments cannot include mensaje", i)}
		}
	}

	return nil
}

func validatePhoneNumber(phone string) *ValidationError {
	phone = strings.TrimSpace(phone)
	if len(phone) < 11 {
		return &ValidationError{Code: ErrorCodeInvalidPhoneFormat, Message: "phone number must be at least 11 digits"}
	}
	if !regexp.MustCompile(`^\d+$`).MatchString(phone) {
		return &ValidationError{Code: ErrorCodeInvalidPhoneFormat, Message: "phone number must contain only digits"}
	}
	return nil
}

func attachmentAllowsCaption(mimeType string) bool {
	switch strings.TrimSpace(strings.ToLower(mimeType)) {
	case "audio/mpeg", "audio/mp4", "audio/ogg", "audio/wav":
		return false
	default:
		return true
	}
}

func validateAttachmentName(name string) *ValidationError {
	name = strings.TrimSpace(name)
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return &ValidationError{Code: ErrorCodeAttachmentNameInvalid, Message: "attachment name contains invalid path characters"}
	}
	for _, ch := range name {
		if ch < 32 || ch == 127 {
			return &ValidationError{Code: ErrorCodeAttachmentNameInvalid, Message: "attachment name contains control characters"}
		}
	}
	if strings.Contains(name, "\x00") {
		return &ValidationError{Code: ErrorCodeAttachmentNameInvalid, Message: "attachment name contains null bytes"}
	}
	return nil
}

func validateMIMETypeExtensionMatch(filename, mimeType string) *ValidationError {
	ext := strings.ToLower(filepath.Ext(filename))
	if !AllowedExtensions[ext] {
		return &ValidationError{Code: ErrorCodeAttachmentTypeNotAllowed, Message: "file extension not allowed: " + ext}
	}

	expectedMIMETypes := map[string][]string{
		".jpg":  {"image/jpeg"},
		".jpeg": {"image/jpeg"},
		".png":  {"image/png"},
		".mp4":  {"video/mp4"},
		".mp3":  {"audio/mpeg"},
		".m4a":  {"audio/mp4"},
		".ogg":  {"audio/ogg"},
		".wav":  {"audio/wav"},
		".pdf":  {"application/pdf"},
		".doc":  {"application/msword"},
		".docx": {"application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
	}

	allowed := expectedMIMETypes[ext]
	for _, m := range allowed {
		if m == mimeType {
			return nil
		}
	}

	return &ValidationError{Code: ErrorCodeAttachmentTypeNotAllowed, Message: "MIME type does not match file extension"}
}
