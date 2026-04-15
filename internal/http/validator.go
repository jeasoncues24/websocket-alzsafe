package http

import (
	"encoding/base64"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"wsapi/internal/domain"
)

// ValidateMessageRequest validates a message request payload
func ValidateMessageRequest(req *domain.MessageRequest) *domain.ValidationError {
	// Validate ruc_empresa presence
	if strings.TrimSpace(req.RUCEmpresa) == "" {
		return &domain.ValidationError{
			Code:    domain.ErrorCodeMissingField,
			Message: "ruc_empresa is required",
		}
	}

	// Validate destino presence
	if strings.TrimSpace(req.Destino) == "" {
		return &domain.ValidationError{
			Code:    domain.ErrorCodeMissingField,
			Message: "destino is required",
		}
	}

	// Validate destino format (phone number: digits only, >= 11 chars)
	if err := validatePhoneNumber(req.Destino); err != nil {
		return err
	}

	// Validate mensaje presence and not empty
	if strings.TrimSpace(req.Mensaje) == "" {
		return &domain.ValidationError{
			Code:    domain.ErrorCodeEmptyMessage,
			Message: "mensaje cannot be empty",
		}
	}

	return nil
}

// validatePhoneNumber validates phone format
func validatePhoneNumber(phone string) *domain.ValidationError {
	phone = strings.TrimSpace(phone)

	// Check minimum length (11 digits for Peru mobile)
	if len(phone) < 11 {
		return &domain.ValidationError{
			Code:    domain.ErrorCodeInvalidPhoneFormat,
			Message: "phone number must be at least 11 digits",
		}
	}

	// Check only numeric characters
	if !regexp.MustCompile(`^\d+$`).MatchString(phone) {
		return &domain.ValidationError{
			Code:    domain.ErrorCodeInvalidPhoneFormat,
			Message: "phone number must contain only digits",
		}
	}

	return nil
}

// ValidateAttachments validates a slice of attachment payloads
func ValidateAttachments(attachments []domain.AttachmentPayload) *domain.ValidationError {
	if len(attachments) == 0 {
		return nil // No attachments is valid
	}

	totalSize := 0
	for _, att := range attachments {
		if validErr := ValidateAttachment(att); validErr != nil {
			return validErr
		}
		totalSize += att.TamanoBytes
	}

	// Check total size limit
	if totalSize > domain.MaxAttachmentSizePerMsg {
		return &domain.ValidationError{
			Code:    domain.ErrorCodeAttachmentSizeExceeded,
			Message: "total attachment size exceeds 20MB limit",
		}
	}

	return nil
}

// ValidateAttachment validates a single attachment payload
func ValidateAttachment(att domain.AttachmentPayload) *domain.ValidationError {
	// Validate filename presence
	if strings.TrimSpace(att.Nombre) == "" {
		return &domain.ValidationError{
			Code:    domain.ErrorCodeAttachmentNameInvalid,
			Message: "attachment nombre is required",
		}
	}

	// Validate filename against path traversal
	if err := validateAttachmentName(att.Nombre); err != nil {
		return err
	}

	// Validate MIME type
	if !domain.AllowedMIMETypes[att.MIMEType] {
		return &domain.ValidationError{
			Code:    domain.ErrorCodeAttachmentTypeNotAllowed,
			Message: "attachment MIME type not allowed: " + att.MIMEType,
		}
	}

	// Validate extension matches MIME type
	if err := validateMIMETypeExtensionMatch(att.Nombre, att.MIMEType); err != nil {
		return err
	}

	// Decode base64 to get actual size
	decodedData, err := base64.StdEncoding.DecodeString(att.ContenidoBase64)
	if err != nil {
		return &domain.ValidationError{
			Code:    domain.ErrorCodeAttachmentFormatInvalid,
			Message: "attachment contenido_base64 is not valid base64",
		}
	}

	actualSize := len(decodedData)
	// Update TamanoBytes after decoding
	att.TamanoBytes = actualSize

	// Validate individual size limit
	if actualSize > domain.MaxAttachmentSizePerFile {
		return &domain.ValidationError{
			Code:    domain.ErrorCodeAttachmentSizeExceeded,
			Message: "attachment exceeds 5MB limit",
		}
	}

	if actualSize == 0 {
		return &domain.ValidationError{
			Code:    domain.ErrorCodeAttachmentFormatInvalid,
			Message: "attachment is empty",
		}
	}

	return nil
}

// validateAttachmentName checks for path traversal and invalid characters in filename
func validateAttachmentName(name string) *domain.ValidationError {
	name = strings.TrimSpace(name)

	// Check for path traversal patterns
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return &domain.ValidationError{
			Code:    domain.ErrorCodeAttachmentNameInvalid,
			Message: "attachment name contains invalid path characters",
		}
	}

	// Check for control characters
	for _, ch := range name {
		if ch < 32 || ch == 127 {
			return &domain.ValidationError{
				Code:    domain.ErrorCodeAttachmentNameInvalid,
				Message: "attachment name contains control characters",
			}
		}
	}

	// Check for null bytes
	if strings.Contains(name, "\x00") {
		return &domain.ValidationError{
			Code:    domain.ErrorCodeAttachmentNameInvalid,
			Message: "attachment name contains null bytes",
		}
	}

	return nil
}

// validateMIMETypeExtensionMatch ensures MIME type matches file extension
func validateMIMETypeExtensionMatch(filename, mimeType string) *domain.ValidationError {
	ext := strings.ToLower(filepath.Ext(filename))

	// Check extension is whitelisted
	if !domain.AllowedExtensions[ext] {
		return &domain.ValidationError{
			Code:    domain.ErrorCodeAttachmentTypeNotAllowed,
			Message: "file extension not allowed: " + ext,
		}
	}

	// Verify MIME type matches extension
	expectedMIMETypes := map[string][]string{
		".jpg":  {"image/jpeg"},
		".jpeg": {"image/jpeg"},
		".png":  {"image/png"},
		".pdf":  {"application/pdf"},
		".doc":  {"application/msword"},
		".docx": {"application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
	}

	allowed := expectedMIMETypes[ext]
	found := false
	for _, m := range allowed {
		if m == mimeType {
			found = true
			break
		}
	}

	if !found {
		return &domain.ValidationError{
			Code:    domain.ErrorCodeAttachmentTypeNotAllowed,
			Message: "MIME type does not match file extension",
		}
	}

	return nil
}

const MaxBroadcastItems = 500

// ValidateBroadcastRequest validates a broadcast request payload.
// Checks ruc_empresa, that lista_difusion is a non-empty array, and that
// every item has a valid destino (phone) and non-empty mensaje.
func ValidateBroadcastRequest(req *domain.BroadcastRequest) *domain.ValidationError {
	if strings.TrimSpace(req.RUCEmpresa) == "" {
		return &domain.ValidationError{
			Code:    domain.ErrorCodeMissingField,
			Message: "ruc_empresa is required",
		}
	}

	if len(req.ListaDifusion) == 0 {
		return &domain.ValidationError{
			Code:    domain.ErrorCodeValidation,
			Message: "lista_difusion must be a non-empty array",
		}
	}

	if len(req.ListaDifusion) > MaxBroadcastItems {
		return &domain.ValidationError{
			Code:    domain.ErrorCodeValidation,
			Message: fmt.Sprintf("lista_difusion exceeds maximum of %d items", MaxBroadcastItems),
		}
	}

	for i, item := range req.ListaDifusion {
		if err := validatePhoneNumber(item.Destino); err != nil {
			return &domain.ValidationError{
				Code:    domain.ErrorCodeInvalidPhoneFormat,
				Message: fmt.Sprintf("item[%d]: %s", i, err.Message),
			}
		}
		if strings.TrimSpace(item.Mensaje) == "" {
			return &domain.ValidationError{
				Code:    domain.ErrorCodeEmptyMessage,
				Message: fmt.Sprintf("item[%d]: mensaje cannot be empty", i),
			}
		}
	}

	return nil
}
