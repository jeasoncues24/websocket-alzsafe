---
story_id: "5.2"
epic: "epic-5"
title: "Tabla y modelo de webhooks salientes"
status: review
estimated_days: 2
priority: high
branch: "feature/integracion-loyo"
skills: ["sql-optimization", "bmad-code-review"]
affects:
  - backend/internal/storage/migrations/017_create_webhooks_outbound.up.sql
  - backend/internal/storage/migrations/017_create_webhooks_outbound.down.sql
  - backend/internal/storage/migrations/018_create_webhooks_outbound_queue.up.sql
  - backend/internal/storage/migrations/018_create_webhooks_outbound_queue.down.sql
  - backend/internal/domain/webhook.go
  - backend/internal/storage/webhook_store.go
  - backend/internal/storage/webhook_store_test.go
---

# Story 5.2: Tabla y modelo de webhooks salientes

Status: in-progress

## Change Log

- **2026-05-23** — Reabierta tras evaluación de ownership (`spec-webhook-ownership-evaluacion.md`). Cambio de grain: el webhook pasa de pertenecer a **empresa** a pertenecer a **api_key/teléfono**. Razones:
  - El api_key ya está atado a un teléfono (`api_keys.telefono_id`) — el grain natural del contrato B2B es teléfono, no empresa.
  - Meta permite una sola callback URL por número de WhatsApp; emparejar webhook por teléfono evita mezclar eventos de números distintos.
  - El admin sigue teniendo visibilidad por empresa (read-only) para soporte.
  - Migración 017 actualizada in-place (no se crea 019) por convención del proyecto: las CREATE TABLE de feature branch portan el schema final. La trazabilidad queda en este changelog.

### AC actualizados por el cambio

- **AC1** ahora incluye campos `telefono_id BIGINT NOT NULL` y `api_key_id BIGINT NOT NULL`, con FK a `telefonos` y `api_keys`, e índices `idx_webhooks_telefono`, `idx_webhooks_api_key`.
- **AC3** `domain.Webhook` ahora incluye `TelefonoID int64` y `ApiKeyID int64`.
- **AC4** Store agrega `ListByApiKey(apiKeyID int64)` y `ListByTelefono(telefonoID int64)`; `ListByEmpresa` se conserva para uso admin read-only.

## Story

Como integrador B2B que usa wsapi como provider de WhatsApp,
quiero que el sistema pueda persistir la configuración de mis webhooks y encolar eventos para entregar,
para que wsapi pueda notificarme en tiempo real cuando ocurran eventos (mensajes, sesiones) sin mantener una conexión WebSocket persistente.

## Acceptance Criteria

**AC1 — Migración `webhooks_outbound` corre limpio:**
**Dado** el migration runner de wsapi con la migración 017,
**Cuando** se ejecuta `up`,
**Entonces** la tabla `webhooks_outbound` existe con los campos: `id`, `empresa_id`, `url`, `secret`, `eventos`, `activo`, `failure_count`, `last_error`, `last_success_at`, `created_at`, `updated_at`
**Y** la migración `down` elimina la tabla sin efectos colaterales.

**AC2 — Migración `webhooks_outbound_queue` corre limpio:**
**Dado** el migration runner con la migración 018 (ejecutada después de 017),
**Cuando** se ejecuta `up`,
**Entonces** la tabla `webhooks_outbound_queue` existe con los campos: `id`, `webhook_id`, `payload`, `intentos`, `proximo_intento_at`, `estado`, `last_error`, `created_at`
**Y** la FK `webhook_id → webhooks_outbound(id) ON DELETE CASCADE` está activa
**Y** la migración `down` elimina la tabla sin tocar `webhooks_outbound`.

**AC3 — Modelos Go definidos:**
**Dado** el archivo `backend/internal/domain/webhook.go`,
**Cuando** se compila con `go build ./...`,
**Entonces** existen los tipos: `Webhook`, `WebhookEvent` (enum de eventos permitidos), `WebhookQueueItem`, `WebhookQueueEstado` (enum: `pending`, `sending`, `done`, `failed`)
**Y** `Webhook.Eventos` es `[]WebhookEvent` con serialización JSON correcta
**Y** `Webhook.Secret` está marcado `json:"-"` (nunca se serializa en respuestas de API).

**AC4 — Store CRUD implementado:**
**Dado** `backend/internal/storage/webhook_store.go`,
**Cuando** se compilan y ejecutan los tests de store,
**Entonces** el store provee al menos: `Create(w *domain.Webhook) error`, `ListByEmpresa(empresaID int64) ([]domain.Webhook, error)`, `GetByID(id int64) (*domain.Webhook, error)`, `Delete(id int64) error`, `IncrementFailureCount(id int64) error`, `Deactivate(id int64) error`
**Y** el store provee para la cola: `EnqueueEvent(item *domain.WebhookQueueItem) error`, `PollPending(limit int) ([]domain.WebhookQueueItem, error)`, `MarkSending(id int64) error`, `MarkDone(id int64) error`, `MarkFailed(id int64, err string, nextRetryAt time.Time) error`.

**AC5 — Tests de store CRUD pasan:**
**Dado** `backend/internal/storage/webhook_store_test.go`,
**Cuando** se ejecuta `cd backend && go test ./internal/storage/... -run TestWebhook`,
**Entonces** los tests cubren Create, ListByEmpresa, Delete de `webhooks_outbound` y EnqueueEvent + PollPending de `webhooks_outbound_queue`
**Y** todos los tests pasan sin errores.

**AC6 — Skill sql-optimization ejecutada:**
**Dado** que las tablas involucran `empresa_id`, `estado`, `proximo_intento_at` y FK,
**Cuando** se diseña el SQL final de las migraciones,
**Entonces** la skill `/sql-optimization` fue invocada sobre el SQL propuesto antes de escribir los archivos `.up.sql`
**Y** los índices y tipos de columna reflejan las recomendaciones de la skill.

**AC7 — Verificación de build:**
**Dado** los cambios aplicados,
**Cuando** se ejecutan `cd backend && go build ./...` y `cd backend && go test ./...`,
**Entonces** ambos comandos terminan sin errores ni regresiones en otros paquetes.

## Tasks / Subtasks

- [x] **T1 — Ejecutar sql-optimization** (AC6)
  - [x] Invocar skill `/sql-optimization` con el SQL propuesto del epic
  - [x] Registrar recomendaciones en Dev Notes de esta story

- [x] **T2 — Migración 017: `webhooks_outbound`** (AC1)
  - [x] Crear `backend/internal/storage/migrations/017_create_webhooks_outbound.up.sql`
  - [x] Crear `backend/internal/storage/migrations/017_create_webhooks_outbound.down.sql`
  - [x] Verificar con `cd backend && go test ./internal/storage/... -run TestMigrationsLayout`

- [x] **T3 — Migración 018: `webhooks_outbound_queue`** (AC2)
  - [x] Crear `backend/internal/storage/migrations/018_create_webhooks_outbound_queue.up.sql`
  - [x] Crear `backend/internal/storage/migrations/018_create_webhooks_outbound_queue.down.sql`
  - [x] Verificar layout test de nuevo

- [x] **T4 — Modelo de dominio** (AC3)
  - [x] Crear `backend/internal/domain/webhook.go` con tipos: `Webhook`, `WebhookEvent`, `WebhookQueueItem`, `WebhookQueueEstado`
  - [x] `Webhook.Secret` → `json:"-"`
  - [x] `Webhook.Eventos []WebhookEvent` con serialización JSON
  - [x] Compilar sin errores

- [x] **T5 — Store de webhooks** (AC4)
  - [x] Crear `backend/internal/storage/webhook_store.go` con `WebhookStore` struct e inyección de `*sql.DB`
  - [x] Implementar CRUD de `webhooks_outbound`: `Create`, `ListByEmpresa`, `ListByApiKey`, `ListByTelefono`, `GetByID`, `Delete`, `IncrementFailureCount`, `Deactivate`
  - [x] Implementar CRUD de cola: `EnqueueEvent`, `PollPending`, `MarkSending`, `MarkDone`, `MarkFailed`, `MarkDeliverySucceeded`, `MarkDeliveryFailed`
  - [x] Constructor `NewWebhookStore(db *sql.DB) *WebhookStore`

- [x] **T6 — Tests de store** (AC5)
  - [x] Crear `backend/internal/storage/webhook_store_test.go`
  - [x] Tests: Create+ListByApiKey, ListByEmpresa, Delete, EnqueueEvent+PollPending, MarkSending doble proceso
  - [x] `go test ./internal/storage/... -run TestWebhook` → todos pasan

- [x] **T7 — Verificación final** (AC7)
  - [x] `cd backend && go build ./...` sin errores
  - [x] `cd backend && go test ./...` sin regresiones

## Dev Notes

### ⚠️ REGLA OBLIGATORIA: sql-optimization PRIMERO

El `project-context.md` establece explícitamente:
> **OBLIGATORIO:** Para cualquier tarea que involucre MySQL (escribir o modificar un CREATE TABLE, ALTER TABLE, consulta con JOIN/GROUP BY/agregaciones, decisión de índices), invocar primero la skill `/sql-optimization`.

**No escribir los `.up.sql` antes de invocar la skill y registrar sus recomendaciones aquí.**

### Convención de migraciones

- Patrón de nombres: `NNN_descripcion_accion.up.sql` / `.down.sql`
- El último número usado es `016_seeds`. Las dos nuevas migraciones son **017** y **018**.
- **Cada par define exactamente UNA acción** (una tabla). Por eso esta story tiene dos pares, no uno.
- El `.down.sql` revierte exactamente el `.up.sql` sin efectos en otras tablas.
- Ejemplo de down correcto: `DROP TABLE IF EXISTS webhooks_outbound;` — sin tocar `webhooks_outbound_queue`.
- Verificar el layout test tras crear cada par: `go test ./internal/storage/... -run TestMigrationsLayout`

### SQL de referencia del epic (a revisar con sql-optimization)

```sql
-- Para 017_create_webhooks_outbound.up.sql
CREATE TABLE webhooks_outbound (
  id            BIGINT AUTO_INCREMENT PRIMARY KEY,
  empresa_id    BIGINT NOT NULL,
  url           TEXT NOT NULL,
  secret        VARCHAR(255) NOT NULL,
  eventos       JSON NOT NULL,
  activo        TINYINT(1) DEFAULT 1,
  failure_count INT DEFAULT 0,
  last_error    TEXT,
  last_success_at DATETIME,
  created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_webhooks_empresa (empresa_id, activo),
  FOREIGN KEY (empresa_id) REFERENCES empresas(id) ON DELETE CASCADE
);

-- Para 018_create_webhooks_outbound_queue.up.sql
CREATE TABLE webhooks_outbound_queue (
  id            BIGINT AUTO_INCREMENT PRIMARY KEY,
  webhook_id    BIGINT NOT NULL,
  payload       JSON NOT NULL,
  intentos      INT DEFAULT 0,
  proximo_intento_at DATETIME NOT NULL,
  estado        ENUM('pending','sending','done','failed') DEFAULT 'pending',
  last_error    TEXT,
  created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_queue_due (estado, proximo_intento_at),
  FOREIGN KEY (webhook_id) REFERENCES webhooks_outbound(id) ON DELETE CASCADE
);
```

**Preguntas para sql-optimization:** ¿`url TEXT` o `VARCHAR(2048)`? ¿`CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci` en ambas (como el resto del proyecto — ver `015_create_audit_log_table.up.sql`)? ¿FK explícita o índice en `webhook_id` en la tabla de cola?

### Patrón de dominio

Copiar patrón de `backend/internal/domain/api_key.go`:
- Struct con tags `json:` y `db:` donde aplique
- Campos opcionales como punteros (`*time.Time`, `*string`)
- Tipos de respuesta separados (con y sin secret)
- `json:"-"` en `Secret` — el secret nunca viaja en JSON de respuesta de listado

```go
// Ejemplo de tipos a definir
type WebhookEvent string

const (
    WebhookEventMessageReceived   WebhookEvent = "message.received"
    WebhookEventMessageStatus     WebhookEvent = "message.status_update"
    WebhookEventSessionConnected  WebhookEvent = "session.connected"
    WebhookEventSessionDisconnected WebhookEvent = "session.disconnected"
)

type WebhookQueueEstado string

const (
    WebhookQueuePending  WebhookQueueEstado = "pending"
    WebhookQueueSending  WebhookQueueEstado = "sending"
    WebhookQueueDone     WebhookQueueEstado = "done"
    WebhookQueueFailed   WebhookQueueEstado = "failed"
)

type Webhook struct {
    ID            int64         `json:"id"`
    EmpresaID     int64         `json:"empresa_id"`
    URL           string        `json:"url"`
    Secret        string        `json:"-"`   // nunca serializar
    Eventos       []WebhookEvent `json:"eventos"`
    Activo        bool          `json:"activo"`
    FailureCount  int           `json:"failure_count"`
    LastError     *string       `json:"last_error,omitempty"`
    LastSuccessAt *time.Time    `json:"last_success_at,omitempty"`
    CreatedAt     time.Time     `json:"created_at"`
    UpdatedAt     time.Time     `json:"updated_at"`
}

type WebhookQueueItem struct {
    ID               int64              `json:"id"`
    WebhookID        int64              `json:"webhook_id"`
    Payload          json.RawMessage    `json:"payload"`
    Intentos         int                `json:"intentos"`
    ProximoIntentoAt time.Time          `json:"proximo_intento_at"`
    Estado           WebhookQueueEstado `json:"estado"`
    LastError        *string            `json:"last_error,omitempty"`
    CreatedAt        time.Time          `json:"created_at"`
}
```

### Patrón de store

Copiar patrón de `backend/internal/storage/api_key.go`:
- Struct `WebhookStore` con `db *sql.DB`
- Constructor `NewWebhookStore(db *sql.DB) *WebhookStore`
- Imports: `wsapi/internal/domain`, `database/sql`, `encoding/json`, `fmt`, `time`
- Manejo de errores: siempre wrappear con `fmt.Errorf("...: %w", err)`
- Para `Eventos []WebhookEvent`: serializar/deserializar con `json.Marshal`/`json.Unmarshal` (igual que `Scopes []string` en api_key.go)

**PollPending** — consulta crítica para Story 5.4 (worker):
```sql
SELECT id, webhook_id, payload, intentos, proximo_intento_at, estado, last_error, created_at
FROM webhooks_outbound_queue
WHERE estado = 'pending' AND proximo_intento_at <= NOW()
ORDER BY proximo_intento_at ASC
LIMIT ?
```

**MarkSending** — update optimista para evitar doble procesamiento en Story 5.4:
```sql
UPDATE webhooks_outbound_queue
SET estado = 'sending'
WHERE id = ? AND estado = 'pending'
```
Verificar `rows.RowsAffected() == 1`; si es 0, otro worker tomó el item.

### Tests de store — estrategia

Los tests en `backend/internal/storage/` que requieren DB usan test con DB real (ver `migrations_layout_test.go` como referencia de cómo el proyecto maneja tests de storage). Si los tests de store existentes usan mocks, copiar ese patrón. Si usan DB real, necesitarás variables de entorno de DB.

Alternativamente, tests de compilación son suficientes para esta story; los tests de integración reales se validan en Story 5.4 (worker con httptest).

**Mínimo aceptable para AC5:** tests que al menos compilen y cubran los caminos felices. Si no hay BD de test disponible en CI, usar `t.Skip("requires DB")` con la condición habitual del proyecto.

### Learnings de Story 5.1

- El proyecto usa `package http` en `backend/internal/http/handlers/` (no sub-package) — para el store, usar `package storage` coherente con los vecinos.
- Imports internos siempre con prefijo `wsapi/internal/...`
- No introducir librerías nuevas — para serialización de `[]WebhookEvent` usar `encoding/json` estándar, igual que `Scopes` en `api_key.go`.
- `container.go` instancia los stores dentro del bloque `if cfg.DBHost != "" { ... }` — el `WebhookStore` seguirá el mismo patrón cuando sea cableado en Story 5.3.
- Las migraciones se embeben automáticamente vía `//go:embed migrations/*.sql` en `migrations_embed.go` — no hay nada extra que hacer para que el runner las detecte.

### Project Structure Notes

- Alineado con `backend/internal/storage/` (stores) y `backend/internal/domain/` (modelos).
- Sin cambios en frontend, handlers HTTP, ni rutas — eso viene en Story 5.3.
- Sin cambios en `container.go` en esta story — el store se instanciará cuando Story 5.3 añada los endpoints.
- Rama obligatoria: `feature/integracion-loyo`. Verificar antes de codear:
  ```bash
  git branch --show-current  # debe ser feature/integracion-loyo
  ```

### References

- Patrón de migración SQL: `backend/internal/storage/migrations/015_create_audit_log_table.up.sql`
- Patrón de domain model: `backend/internal/domain/api_key.go`
- Patrón de store: `backend/internal/storage/api_key.go`
- Embed de migraciones: `backend/internal/storage/migrations_embed.go`
- Número de última migración: `016_seeds` → próximas son `017` y `018`
- Epic origen: `_bmad-output/planning-artifacts/epic-integracion-loyo.md` (Story 5.2)
- Project context: `_bmad-output/project-context.md` (regla sql-optimization, convención migraciones)
- Regla de rama: `_bmad-output/project-context.md` (tabla de ramas por epic)

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6 (bmad-create-story)

### Debug Log References

### Completion Notes List

### File List
