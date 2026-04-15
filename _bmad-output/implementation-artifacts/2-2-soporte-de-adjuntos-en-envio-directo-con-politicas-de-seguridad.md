# Story 2.2: Soporte de adjuntos en envío directo con políticas de seguridad

Status: done

## Story

As a operador de empresa,
I want adjuntar archivos permitidos en mensajes directos,
So that pueda enviar evidencia o documentos a mis destinatarios.

## Acceptance Criteria

1. Dado un adjunto en la solicitud con tipo MIME y tamaño dentro de políticas, cuando se valida, entonces se acepta el archivo y se procesa normalmente. ✓

2. Dado un adjunto cuyo tipo MIME no está en whitelist (por ejemplo: .exe, .bat, .sh), cuando se valida, entonces la API rechaza con error "ATTACHMENT_TYPE_NOT_ALLOWED". ✓

3. Dado un adjunto cuyo tamaño excede el límite máximo (5MB por archivo, 20MB por mensaje), cuando se valida, entonces la API rechaza con error "ATTACHMENT_SIZE_EXCEEDED". ✓

4. Dado múltiples adjuntos en una solicitud, cuando se validan, entonces se valida cada uno según políticas y se rechaza la solicitud si al menos uno no cumple. ✓

5. Dado un mensaje con adjuntos permitidos, cuando se envía, entonces el handler retorna información de los adjuntos procesados (nombre, hash, size) en la respuesta. ✓

## Tasks / Subtasks

- [x] Definir políticas de seguridad para adjuntos (AC: 1, 2, 3)
  - [x] Whitelist de tipos MIME permitidos: image/jpeg, image/png, application/pdf, application/msword, application/vnd.openxmlformats-officedocument.wordprocessingml.document
  - [x] Whitelist de extensiones: .jpg, .jpeg, .png, .pdf, .doc, .docx
  - [x] Límite por archivo: 5MB
  - [x] Límite por mensaje: 20MB (total de todos los adjuntos)
- [x] Extender MessageRequest para adjuntos (AC: 1, 4)
  - [x] Agregar campo Adjuntos: []AttachmentPayload
  - [x] AttachmentPayload: {nombre, mimeType, contenidoBase64, [hash]}
  - [x] Validación de estructura de adjuntos
- [x] Crear validador de adjuntos (AC: 2, 3, 4)
  - [x] ValidateAttachment(attachment): verifica MIME, extensión, tamaño
  - [x] ValidateAttachments(attachments): valida colección, límite total
- [x] Agregar seguridad contra exploits (AC: 2)
  - [x] Validar que el nombre no contiene path traversal (../, ..\)
  - [x] Validar que extensión coincide con MIME type
  - [x] Sanitizar nombre de archivo
- [x] Agregar pruebas unitarias (AC: 1, 2, 3, 4, 5)
  - [x] Test: valid attachment → accepted (TestValidateAttachmentValid)
  - [x] Test: invalid MIME type → 400 ATTACHMENT_TYPE_NOT_ALLOWED (TestValidateAttachmentInvalidMIMEType)
  - [x] Test: MIME-extension mismatch → error (TestValidateAttachmentMismatchExtensionMIME)
  - [x] Test: path traversal in filename → rejected (TestValidateAttachmentPathTraversal)
  - [x] Test: invalid base64 → error (TestValidateAttachmentInvalidBase64)
  - [x] Test: empty name → error (TestValidateAttachmentEmptyName)
  - [x] Test: multiple valid attachments → all processed (TestValidateAttachmentsMultipleValid)
  - [x] Test: mixed valid+invalid → all rejected (TestValidateAttachmentsMixedValidInvalid)
  - [x] Test: total size exceeded → error (TestValidateAttachmentsSizeExceededTotal)

## Dev Notes

- Los adjuntos se procesan en memoria; no se persisten todavía (Story 2.3).
- El contenido debe estar en base64 para transportarlo en JSON.
- El hash se calcula como SHA256 del contenido decodificado (para anti-duplicación futura).
- Extensión se valida contra el nombre de archivo, no se confía en MIME type del cliente.
- Path traversal: rechazar si nombre contiene "..", "/", "\", o caracteres de control.

### Technical Requirements

- Mantener estructura por capas:
  - [internal/domain](internal/domain) - Attachment, AttachmentInfo structs
  - [internal/http](internal/http) - ValidateAttachment, ValidateAttachments
  - [internal/http](internal/http) - Extender HandlePostMessage
- Uso de crypto/sha256 para hash
- Base64 decodification con encoding/base64
- HTTP status codes: 400 para validación fallida
- MaxMemory límites para evitar DoS (parser de multipart no aplicable aquí; JSON directo)

### Suggested File Targets

- [internal/domain/attachment.go](internal/domain/attachment.go) - NEW
- [internal/http/validator.go](internal/http/validator.go) - Extend con ValidateAttachment(s)
- [internal/http/handlers.go](internal/http/handlers.go) - Modify HandlePostMessage para procesar adjuntos
- [internal/http/handlers_test.go](internal/http/handlers_test.go) - Add 7+ tests

### Testing Requirements

- Casos principales:
  - PDF válido (5MB exactly) → accepted, hash incluido
  - DOCX válido (2MB) → accepted
  - EXE (1MB) → 400 ATTACHMENT_TYPE_NOT_ALLOWED
  - PDF (6MB) → 400 ATTACHMENT_SIZE_EXCEEDED
  - 4 × 5MB PDFs (20MB total) → 400 ATTACHMENT_SIZE_EXCEEDED
  - Nombre con "../../../etc/passwd" → rejected
  - Base64 malformado → 400 INVALID_ATTACHMENT_FORMAT
  - Múltiples adjuntos (1 válido + 1 inválido) → all rejected

### References

- Story source: [\_bmad-output/planning-artifacts/epics.md](_bmad-output/planning-artifacts/epics.md#L142)
- PRD source: [\_bmad-output/planning-artifacts/prd.md](_bmad-output/planning-artifacts/prd.md)
- Story 2.1: [\_bmad-output/implementation-artifacts/2-1-endpoint-de-envio-directo-con-validacion-de-payload.md](_bmad-output/implementation-artifacts/2-1-endpoint-de-envio-directo-con-validacion-de-payload.md)
- Policy inspiration: OWASP File Upload Cheat Sheet

## Implementation Notes

- AttachmentPayload: datos entrada (base64)
- AttachmentInfo: datos salida (nombre, hash, size)
- Message.Adjuntos: slice de AttachmentInfo
- ValidateAttachment es la función unit principal
- ValidateAttachments es el wrapper para múltiples (llamada desde handler)

## File List

- \_bmad-output/implementation-artifacts/2-2-soporte-de-adjuntos-en-envio-directo-con-politicas-de-seguridad.md
