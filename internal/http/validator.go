package http

import "wsapi/internal/domain"

func ValidateMessageRequest(req *domain.MessageRequest) *domain.ValidationError {
	return domain.ValidateMessageRequest(req)
}

func ValidateAttachments(attachments []domain.AttachmentPayload) *domain.ValidationError {
	return domain.ValidateAttachments(attachments)
}

func ValidateAttachment(att domain.AttachmentPayload) *domain.ValidationError {
	return domain.ValidateAttachment(att)
}

func ValidateBroadcastRequest(req *domain.BroadcastRequest) *domain.ValidationError {
	return domain.ValidateBroadcastRequest(req)
}
