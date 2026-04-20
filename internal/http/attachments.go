package http

import "wsapi/internal/domain"

func buildAttachmentInfos(payloads []domain.AttachmentPayload) ([]domain.AttachmentInfo, error) {
	if len(payloads) == 0 {
		return nil, nil
	}

	infos := make([]domain.AttachmentInfo, 0, len(payloads))
	for _, payload := range payloads {
		info, err := payload.Info()
		if err != nil {
			return nil, err
		}
		infos = append(infos, *info)
	}

	return infos, nil
}
