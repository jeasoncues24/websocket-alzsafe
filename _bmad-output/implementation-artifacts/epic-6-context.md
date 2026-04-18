# Epic 6 Context: Sistema de Autenticación JWT por Empresa + Multi-Tenant

<!-- Compiled from planning artifacts. Edit freely. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Añadir un segundo sistema de autenticación paralelo al admin existente. Las empresas obtienen un JWT de larga duración (5 años) y acceden a sus propios teléfonos WhatsApp bajo `/v1/*`. El admin actual (`/api/admin/*`) no se toca. Al finalizar, los endpoints legacy públicos quedan eliminados y cada recurso está aislado por `empresa_id`.

## Stories

- S-6.1: Schema DB — ALTER empresas (token_version, permissions) + CREATE telefonos
- S-6.2: Modelo Go Empresa + Teléfono (dominio, storage CRUD)
- S-6.3: Migrar messages/broadcasts de ruc_empresa → empresa_id/telefono_id
- S-6.4: Generación de JWT empresa (long-lived, claims: sub=empresa_id, ver=token_version)
- S-6.5: Middleware auth JWT empresa + ownership sobre telefonos
- S-6.6: Endpoints `/v1/*` + eliminar endpoints legacy
- S-6.7: WebSocket `/v1/ws` con eventos QR/connected/disconnected/message_status
- S-6.8: Flujo QR completo por teléfono
- S-6.9: Tests integración Epic 6
- S-6.10: Documentación API `/v1/*`

## Requirements & Constraints

- El sistema admin (`/api/admin/*`, JWT sesión corta, tabla `admin_users`) NO se modifica.
- Los JWT empresa usan claims mínimos: `sub` (empresa_id), `ver` (token_version), `exp` (5 años).
- `token_version` permite revocación: si la versión del claim < versión en DB → token inválido.
- Cada teléfono pertenece a exactamente una empresa (`empresa_id FK`).
- El ownership se valida en CADA request: el telefono_id del path debe pertenecer a la empresa del JWT.
- Endpoints legacy (`/message`, `/messages`, `/broadcast`, `/broadcast/{id}`, `/sessions`, `/metrics`, `/companies`) deben eliminarse en S-6.6.
- `session_data` en telefonos = LONGBLOB (datos whatsmeow serializado).
- `permissions` en empresas = JSON array, e.g. `["send","broadcast","sessions"]`.

## Technical Decisions

- **JWT empresa firmado con el mismo `JWT_SECRET`** del config existente (env var), pero claims distintos al admin.
- **Migration numbering:** continúa desde 013. S-6.1 usa 014 (ALTER empresas) y 015 (CREATE telefonos).
- **`numero_completo`** es una columna virtual generada: `CONCAT(codigo_pais, numero)`. Indexable.
- **Sessions en memoria** actualmente usan RUC como key. En S-6.8 se migran a telefono_id (int64).
- **Worker pool / broadcast** actualmente referencia `ruc_empresa`; S-6.3 lo migrará a `empresa_id`.
- El framework de migraciones ya existe en `internal/storage/` — seguir el patrón de archivos `NNN_name.up.sql` / `NNN_name.down.sql`.

## Cross-Story Dependencies

- S-6.2 depende de S-6.1 (necesita las tablas para implementar storage CRUD).
- S-6.3 depende de S-6.1 (necesita empresa_id/telefono_id como FK).
- S-6.5, S-6.6, S-6.7, S-6.8 dependen de S-6.4 (JWT empresa).
- S-6.8 depende de S-6.7 (WebSocket para emitir eventos QR).
