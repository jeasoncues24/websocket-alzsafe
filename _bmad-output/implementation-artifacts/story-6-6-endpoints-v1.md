# Story 6.6: Endpoints v1/* + elimina legacy

Status: done

## Story

As a developer implementing the API,
I want to create HTTP handlers for the `/v1/*` endpoints protected with empresa JWT,
so that empresas can manage their phones, messages, broadcasts, and metrics securely.

## Acceptance Criteria

1. [AC: v1-sessions] GET /v1/sessions returns list of phones (sessions) filtered by empresaJWT claims
2. [AC: v1-sessions] POST /v1/sessions creates new phone session and returns QR code
3. [AC: v1-sessions] GET /v1/sessions/{telefono_id} returns phone status (ownership validated)
4. [AC: v1-sessions] DELETE /v1/sessions/{telefono_id} disconnects phone (ownership validated)
5. [AC: v1-messages] GET /v1/messages returns messages filtered by empresa_id from JWT + optional filters
6. [AC: v1-messages] POST /v1/message sends a message using phone ownership
7. [AC: v1-broadcasts] GET /v1/broadcasts returns broadcasts filtered by empresa_id
8. [AC: v1-broadcasts] POST /v1/broadcast creates new broadcast
9. [AC: v1-broadcasts] GET /v1/broadcast/{id} returns broadcast status
10. [AC: v1-metrics] GET /v1/metrics returns empresa metrics (messages sent, success rate, etc.)
11. [AC: v1-phones] GET /v1/phones returns list of registered phones
12. [AC: v1-phones] POST /v1/phones/{telefono_id}/qr regenerates QR code (ownership validated)
13. [AC: response-format] All v1 responses follow format: {ok, data, meta}
14. [AC: response-format] All v1 errors follow format: {ok: false, error, message}
15. [AC: middleware] Protected by empresaAuthMiddleware.RequireEmpresaAuth()
16. [AC: middleware] Endpoints with telefono_id use additional empresaAuthMiddleware.RequireOwnership()

## Tasks / Subtasks

- [x] Task 1: Create handler files structure (AC: all)
  - [x] Subtask 1.1: internal/http/handlers/v1_sessions.go
  - [x] Subtask 1.2: internal/http/handlers/v1_messages.go
  - [x] Subtask 1.3: internal/http/handlers/v1_broadcasts.go
  - [x] Subtask 1.4: internal/http/handlers/v1_metrics.go
  - [x] Subtask 1.5: internal/http/handlers/v1_phones.go
  - [x] Subtask 1.6: internal/http/handlers/v1_helpers.go
- [x] Task 2: Register v1 routes in router.go (AC: 15-16)
  - [x] Subtask 2.1: Add requiresEmpresa := empresaAuthMiddleware.RequireEmpresaAuth()
  - [x] Subtask 2.2: Add requireOwnership := empresaAuthMiddleware.RequireOwnership()
  - [x] Subtask 2.3: Route /v1/sessions, /v1/sessions/{telefono_id}
  - [x] Subtask 2.4: Route /v1/message, /v1/messages
  - [x] Subtask 2.5: Route /v1/broadcast, /v1/broadcast/{id}
  - [x] Subtask 2.6: Route /v1/broadcasts
  - [x] Subtask 2.7: Route /v1/metrics
  - [x] Subtask 2.8: Route /v1/phones, /v1/phones/{telefono_id}/qr
  - [ ] Subtask 2.9: Remove or mark deprecated legacy /api/* endpoints
- [x] Task 3: Implement sessions handler v1_sessions.go (AC: 1-4)
  - [x] Task 3.1: ListGetSessions - GET /v1/sessions - list empresa phones
  - [x] Task 3.2: PostSessions - POST /v1/sessions - create phone + generate QR
  - [x] Task 3.3: GetSession - GET /v1/sessions/{telefono_id} - get phone status
  - [x] Task 3.4: DeleteSession - DELETE /v1/sessions/{telefono_id} - disconnect
- [x] Task 4: Implement messages handler v1_messages.go (AC: 5-6)
  - [x] Task 4.1: GetMessages - list messages with filters
  - [x] Task 4.2: PostMessage - send message (validate telefono ownership)
- [x] Task 5: Implement broadcasts handler v1_broadcasts.go (AC: 7-9)
  - [x] Task 5.1: GetBroadcasts - list broadcasts
  - [x] Task 5.2: GetBroadcast - get broadcast by ID
  - [x] Task 5.3: PostBroadcast - create new broadcast job
- [x] Task 6: Implement metrics handler v1_metrics.go (AC: 10)
  - [x] Task 6.1: GetMetrics - return empresa metrics
- [x] Task 7: Implement phones handler v1_phones.go (AC: 11-12)
  - [x] Task 7.1: GetPhones - list phones (alias for sessions)
  - [x] Task 7.2: PostPhoneQr - regenerate QR code
- [x] Task 8: Implement response helpers (AC: 13-14)
  - [x] Task 8.1: writeV1Success - wrapper for success response
  - [x] Task 8.2: writeV1Error - wrapper for error response

## Dev Notes

### Architecture Patterns

1. **Response format** (spec-6-6 spec):
```go
// Success
{
  "ok": true,
  "data": { ... },
  "meta": {
    "empresa_id": 123,
    "timestamp": "2026-04-16T12:00:00Z"
  }
}

// Error
{
  "ok": false,
  "error": "CODE",
  "message": "Descripción"
}
```

2. **Middleware usage** (from empresa_auth.go line 37-76):
   - `RequireEmpresaAuth()` → validates JWT empresa, injects claims in context
   - `RequireOwnership()` → validates telefono_id belongs to empresa

3. **Get claims from context** (see domain/):
```go
claims, ok := domain.GetEmpresaTokenClaims(r.Context())
// claims.EmpresaID
```

### Project Structure Notes

- Handlers go in: `internal/http/handlers/`
- Existing handlers pattern: auth.go, companies.go
- Middleware already exists: `internal/http/middleware/empresa_auth.go`
- Domain types: `internal/domain/` (EmpresaTokenClaims)

### Dependencies (all completed in previous stories)

- S-6.0: Schema DB multi-tenant (done)
- S-6.2: Modelo Empresa + Teléfono (done)
- S-6.4: Generación JWT empresa (done)
- S-6.5: Middleware auth JWT + ownership (done)

### Stores/Repositories to use

- `storage.TelefonoStore` - manage phones
- `storage.EmpresaStore` - get empresa info
- `storage.MessagesRepository` - messages
- `storage.BroadcastStore` - broadcasts

### Reference Files

- Spec: `_bmad-output/implementation-artifacts/spec-6-6-endpoints-v1.md`
- Middleware: `internal/http/middleware/empresa_auth.go` (lines 37-110)
- Router registration: `internal/http/router.go` (lines 130-142 shows existing placeholder)
- Response helpers: existing in handlers/auth.go for format patterns

### Testing Standards

- Create unit tests in handler files using `_test.go` suffix
- Test valid and invalid JWT scenarios
- Test ownership validation scenarios
- Test response format compliance

## References

- [Source: spec-6-6-endpoints-v1.md]
- [Source: internal/http/middleware/empresa_auth.go]
- [Source: internal/http/router.go lines 130-142]

## Dev Agent Record

### Agent Model Used

BMAD Dev Story Workflow + quick-dev

### Debug Log References

### Completion Notes List

### File List