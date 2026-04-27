package domain

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

// AttachmentPayload represents an attachment in the incoming request (base64 encoded)
type AttachmentPayload struct {
	Nombre          string `json:"nombre"`
	MIMEType        string `json:"mime_type"`
	ContenidoBase64 string `json:"contenido_base64"`
	TamanoBytes     int    `json:"tamano_bytes,omitempty"`
	TamanoValidado  int    `json:"-"`
}

// AttachmentInfo represents processed attachment info in response
type AttachmentInfo struct {
	Nombre      string `json:"nombre"`
	SHA256Hash  string `json:"sha256_hash"`
	TamanoBytes int    `json:"tamano_bytes"`
}

// Attachment security policy constants
const (
	MaxAttachmentSizePerFile = 5 * 1024 * 1024  // 5MB per file
	MaxAttachmentSizePerMsg  = 20 * 1024 * 1024 // 20MB per message
	MaxAttachmentsPerMsg     = 1

	// Error codes for attachment validation
	ErrorCodeAttachmentTypeNotAllowed = "ATTACHMENT_TYPE_NOT_ALLOWED"
	ErrorCodeAttachmentSizeExceeded   = "ATTACHMENT_SIZE_EXCEEDED"
	ErrorCodeAttachmentFormatInvalid  = "INVALID_ATTACHMENT_FORMAT"
	ErrorCodeAttachmentNameInvalid    = "INVALID_ATTACHMENT_NAME"
	ErrorCodeAttachmentCountExceeded  = "ATTACHMENT_COUNT_EXCEEDED"
)

// AllowedMIMETypes whitelist of permitted MIME types
var AllowedMIMETypes = map[string]bool{
	"image/jpeg":         true,
	"image/png":          true,
	"video/mp4":          true,
	"audio/mpeg":         true,
	"audio/mp4":          true,
	"audio/ogg":          true,
	"audio/wav":          true,
	"application/pdf":    true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
}

// AllowedExtensions whitelist of permitted file extensions
var AllowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".mp4":  true,
	".mp3":  true,
	".m4a":  true,
	".ogg":  true,
	".wav":  true,
	".pdf":  true,
	".doc":  true,
	".docx": true,
}

// CalculateSHA256 calculates SHA256 hash of binary data and returns hex string
func CalculateSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// NewAttachmentInfo creates AttachmentInfo from binary payload
func NewAttachmentInfo(nombre string, data []byte) *AttachmentInfo {
	return &AttachmentInfo{
		Nombre:      nombre,
		SHA256Hash:  CalculateSHA256(data),
		TamanoBytes: len(data),
	}
}

// Decode returns the raw attachment bytes from base64.
func (p AttachmentPayload) Decode() ([]byte, error) {
	return base64.StdEncoding.DecodeString(strings.TrimSpace(p.ContenidoBase64))
}

// Info builds the stored attachment metadata from the payload bytes.
func (p AttachmentPayload) Info() (*AttachmentInfo, error) {
	data, err := p.Decode()
	if err != nil {
		return nil, err
	}
	return NewAttachmentInfo(p.Nombre, data), nil
}
