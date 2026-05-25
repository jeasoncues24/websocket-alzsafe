# Spec: Sistema de Difusiones Anti-Ban con Cola de Jobs Reutilizable

**Estado:** Implementado — backend y frontend completados  
**Fecha:** 2026-05-25  
**Rama de trabajo:** `v1` (verificar con project rules si aplica epic separado)

---

## Contexto y motivación

El sistema actual de difusiones tiene tres problemas críticos:

1. **El estado vive en memoria pura** (`storage/broadcast.go` usa un `map` en RAM). Si el proceso se reinicia, todos los jobs activos se pierden sin trazabilidad.
2. **No hay delays entre mensajes** — el loop de `processJob()` envía a máxima velocidad, patrón que WhatsApp detecta como spam.
3. **La arquitectura de persistencia no es extensible** — no sirve para mensajes programados ni otros tipos de jobs futuros.

---

## Alcance de este spec

### Qué cambia

| Área | Cambio |
|------|--------|
| Dominio | Nuevo límite de 30 destinatarios. Nuevo campo `estimated_seconds` en respuesta. Tipos de cola genérica. |
| Almacenamiento | Dos nuevas tablas MySQL (`job_queue`, `job_items`). Migración 019. `BroadcastStore` reemplazado por persistencia real. |
| Worker | `processJob()` refactorizado con batching aleatorio + delays anti-ban + context cancelable. `WorkerConfig` con nuevos parámetros. |
| Handler HTTP | `POST /difusiones` devuelve `estimated_seconds`. `GET /difusiones/{id}` devuelve resultados por item desde DB. |
| Frontend admin | Nueva vista de progreso en tiempo real por destinatario. WebSocket para jobs activos. |
| Documentación | `docs/routes/contrato-b2b/difusiones.md` actualizado. |

### Qué NO cambia

- El contrato del endpoint (mismos paths, mismos campos existentes — solo se añaden campos nuevos).
- La lógica de `sendPreparedMessage()` ni los adjuntos.
- El sistema de API keys y autenticación.

---

## 1. Dominio — `backend/internal/domain/`

### 1.1 Cambio en `broadcast.go`

```go
// MaxBroadcastItems limits broadcast fan-out. Anti-ban: 30 max.
const MaxBroadcastItems = 30
```

**También:** añadir campo `EstimatedSeconds int` a `BroadcastResponse`:
```go
type BroadcastResponse struct {
    OK               bool   `json:"ok"`
    ReferenceID      string `json:"reference_id,omitempty"`
    Total            int    `json:"total,omitempty"`
    EstimatedSeconds int    `json:"estimated_seconds,omitempty"` // NUEVO
    Estado           string `json:"estado,omitempty"`
    Error            string `json:"error,omitempty"`
    Details          string `json:"details,omitempty"`
}
```

### 1.2 Nuevo archivo `job_queue.go`

```go
package domain

import "time"

// JobType identifica el tipo de trabajo en la cola.
type JobType string

const (
    JobTypeBroadcast         JobType = "broadcast"
    JobTypeScheduledMessage  JobType = "scheduled_message"
)

// JobStatus ciclo de vida de un job en la cola.
type JobStatus string

const (
    JobStatusPending   JobStatus = "pending"
    JobStatusRunning   JobStatus = "running"
    JobStatusCompleted JobStatus = "completed"
    JobStatusFailed    JobStatus = "failed"
    JobStatusCancelled JobStatus = "cancelled"
)

// JobItemStatus estado de un item individual dentro de un job.
type JobItemStatus string

const (
    JobItemPending   JobItemStatus = "pending"
    JobItemSent      JobItemStatus = "sent"
    JobItemFailed    JobItemStatus = "failed"
    JobItemSkipped   JobItemStatus = "skipped"
)

// Job representa un trabajo encolado genérico.
type Job struct {
    ID             int64      `json:"id"`
    Type           JobType    `json:"type"`
    EntityID       string     `json:"entity_id"`       // reference_id del broadcast u otro
    Status         JobStatus  `json:"status"`
    Priority       int        `json:"priority"`
    EmpresaID      int64      `json:"empresa_id"`
    AttemptCount   int        `json:"attempt_count"`
    MaxAttempts    int        `json:"max_attempts"`
    LastHeartbeat  *time.Time `json:"last_heartbeat,omitempty"`
    NextRetryAt    *time.Time `json:"next_retry_at,omitempty"`
    Metadata       string     `json:"metadata"`        // JSON arbitrario por tipo
    CreatedAt      time.Time  `json:"created_at"`
    StartedAt      *time.Time `json:"started_at,omitempty"`
    CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

// JobItem representa un destinatario/item individual dentro de un job.
type JobItem struct {
    ID             int64         `json:"id"`
    JobID          int64         `json:"job_id"`
    SequenceOrder  int           `json:"sequence_order"`
    Payload        string        `json:"payload"`         // JSON: {destino, mensaje}
    Status         JobItemStatus `json:"status"`
    AttemptCount   int           `json:"attempt_count"`
    ErrorText      string        `json:"error_text,omitempty"`
    ProcessedAt    *time.Time    `json:"processed_at,omitempty"`
    CreatedAt      time.Time     `json:"created_at"`
}
```

### 1.3 Función de estimación de tiempo

Ubicar en `domain/broadcast.go` o en el handler:

```go
// EstimateBroadcastSeconds calcula el tiempo estimado en segundos para N destinatarios.
// Usa los valores medios de los rangos de delay configurados.
func EstimateBroadcastSeconds(n int, cfg BroadcastTimingConfig) int {
    if n <= 0 { return 0 }
    
    avgIntra  := avg(cfg.IntraBatchDelayMin, cfg.IntraBatchDelayMax)   // ms
    avgInter  := avg(cfg.InterBatchDelayMin, cfg.InterBatchDelayMax)   // ms
    avgMacro  := avg(cfg.MacroPauseMin, cfg.MacroPauseMax)            // ms
    batchSize := float64(cfg.BatchSizeMin+cfg.BatchSizeMax) / 2.0
    
    batches       := math.Ceil(float64(n) / batchSize)
    msgsInBatches := float64(n) * float64(avgIntra)
    interPauses   := (batches - 1) * float64(avgInter)
    macroPauses   := math.Floor(float64(n)/float64(cfg.MacroPauseEvery)) * float64(avgMacro)
    
    totalMs := msgsInBatches + interPauses + macroPauses
    return int(math.Ceil(totalMs / 1000.0))
}

type BroadcastTimingConfig struct {
    BatchSizeMin, BatchSizeMax   int
    IntraBatchDelayMin, IntraBatchDelayMax time.Duration
    InterBatchDelayMin, InterBatchDelayMax time.Duration
    MacroPauseEvery              int
    MacroPauseMin, MacroPauseMax time.Duration
}
```

---

## 2. Migración de base de datos — `backend/internal/storage/migrations/`

### Archivo: `019_create_job_queue.up.sql`

```sql
-- Cola genérica de jobs reutilizable para broadcast, mensajes programados, etc.
CREATE TABLE IF NOT EXISTS job_queue (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    type            VARCHAR(50)     NOT NULL,
    entity_id       VARCHAR(100)    NOT NULL,
    status          VARCHAR(20)     NOT NULL DEFAULT 'pending',
    priority        TINYINT         NOT NULL DEFAULT 5,
    empresa_id      BIGINT UNSIGNED NOT NULL,
    attempt_count   SMALLINT        NOT NULL DEFAULT 0,
    max_attempts    SMALLINT        NOT NULL DEFAULT 3,
    last_heartbeat  DATETIME        NULL,
    next_retry_at   DATETIME        NULL,
    metadata        JSON            NULL,
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at      DATETIME        NULL,
    completed_at    DATETIME        NULL,

    INDEX idx_empresa_status       (empresa_id, status),
    INDEX idx_type_status_retry    (type, status, next_retry_at),
    INDEX idx_entity               (entity_id),
    INDEX idx_heartbeat_running    (status, last_heartbeat)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Items individuales de cada job (destinatarios en broadcast, mensajes en scheduled, etc.)
CREATE TABLE IF NOT EXISTS job_items (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    job_id          BIGINT UNSIGNED NOT NULL,
    sequence_order  INT UNSIGNED    NOT NULL,
    payload         JSON            NOT NULL,
    status          VARCHAR(20)     NOT NULL DEFAULT 'pending',
    attempt_count   SMALLINT        NOT NULL DEFAULT 0,
    error_text      TEXT            NULL,
    processed_at    DATETIME        NULL,
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_job_status           (job_id, status),
    INDEX idx_job_sequence         (job_id, sequence_order),
    CONSTRAINT fk_job_items_job    FOREIGN KEY (job_id) REFERENCES job_queue(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**⚠️ NOTA PARA /sql-optimization:** Revisar esta migración específicamente:
- ¿Los índices compuestos están en el orden correcto para las queries de recovery de startup?
- ¿`JSON` es portable con la versión de MariaDB del proyecto?
- ¿El campo `last_heartbeat` necesita índice separado o es suficiente el compuesto con `status`?
- ¿`ON DELETE CASCADE` en `job_items` es correcto o debería ser `RESTRICT`?

### Archivo: `019_create_job_queue.down.sql`

```sql
DROP TABLE IF EXISTS job_items;
DROP TABLE IF EXISTS job_queue;
```

---

## 3. Storage — `backend/internal/storage/`

### 3.1 Nuevo archivo `job_queue.go`

Interfaz y repositorio genérico. Métodos requeridos:

```go
type JobQueueRepository interface {
    // Crear un job y sus items en una transacción
    CreateJobWithItems(ctx context.Context, job *domain.Job, items []domain.JobItem) error

    // Obtener job por entity_id (reference_id del broadcast)
    GetByEntityID(ctx context.Context, entityID string) (*domain.Job, error)

    // Actualizar estado del job
    UpdateStatus(ctx context.Context, jobID int64, status domain.JobStatus, completedAt *time.Time) error

    // Heartbeat — actualizar last_heartbeat para jobs running
    Heartbeat(ctx context.Context, jobID int64) error

    // Actualizar estado de un item individual
    UpdateItemStatus(ctx context.Context, itemID int64, status domain.JobItemStatus, errText string) error

    // Listar items pendientes de un job (para recovery)
    GetPendingItems(ctx context.Context, jobID int64) ([]domain.JobItem, error)

    // Listar todos los items de un job (para GET /difusiones/{id})
    GetAllItems(ctx context.Context, jobID int64) ([]domain.JobItem, error)

    // Recovery al startup: jobs stuck en running → pending
    RecoverStuckJobs(ctx context.Context, stuckThreshold time.Duration) (int, error)
}
```

### 3.2 Deprecar `storage/broadcast.go`

El `BroadcastStore` en memoria debe eliminarse. Sus callers en los handlers deben migrar al nuevo `JobQueueRepository`.

---

## 4. Worker — `backend/internal/whatsapp/broadcast.go`

### 4.1 WorkerConfig actualizado

```go
var DefaultWorkerConfig = WorkerConfig{
    MaxWorkersPerEmpresa: 3,
    MaxWorkersGlobal:     20,
    MaxRetries:           2,
    RetryDelay:           2 * time.Second,

    // Batching
    BatchSizeMin: 3,
    BatchSizeMax: 4,

    // Delay entre mensajes dentro del mismo batch
    IntraBatchDelayMin: 1500 * time.Millisecond,
    IntraBatchDelayMax: 4000 * time.Millisecond,

    // Delay entre batches
    InterBatchDelayMin: 3000 * time.Millisecond,
    InterBatchDelayMax: 8000 * time.Millisecond,

    // Macro-pausa cada N mensajes enviados
    MacroPauseEvery: 10,
    MacroPauseMin:   15 * time.Second,
    MacroPauseMax:   30 * time.Second,
}
```

### 4.2 Nuevas funciones helper (exportables para tests)

```go
// SplitIntoBatches divide items en grupos de tamaño [minSize, maxSize] usando rng local.
func SplitIntoBatches(items []domain.BroadcastItem, minSize, maxSize int, rng *rand.Rand) [][]domain.BroadcastItem

// RandDuration retorna duración aleatoria en [min, max] usando rng local.
func RandDuration(rng *rand.Rand, min, max time.Duration) time.Duration

// SleepWithContext duerme d o retorna ctx.Err() si el contexto se cancela.
func SleepWithContext(ctx context.Context, d time.Duration) error
```

### 4.3 processJob refactorizado (pseudocódigo)

```
processJob(ctx, job):
  rng = rand.New(rand.NewSource(time.Now().UnixNano()))
  batches = SplitIntoBatches(job.Items, cfg.BatchSizeMin, cfg.BatchSizeMax, rng)
  msgCount = 0

  for batchIdx, batch in batches:
    if ctx.Err() != nil: return

    for itemIdx, item in batch:
      if ctx.Err() != nil: return

      err = processItemWithRetry(ctx, item, ...)
      → UpdateItemStatus en DB (sent/failed)
      → enviar resultado por ResultChan

      msgCount++

      // Macro-pausa cada MacroPauseEvery mensajes
      if msgCount % cfg.MacroPauseEvery == 0 && msgCount < len(job.Items):
        d = RandDuration(rng, cfg.MacroPauseMin, cfg.MacroPauseMax)
        SleepWithContext(ctx, d)

      // Delay intra-batch (no después del último item del batch)
      if itemIdx < len(batch)-1:
        d = RandDuration(rng, cfg.IntraBatchDelayMin, cfg.IntraBatchDelayMax)
        SleepWithContext(ctx, d)

    // Delay inter-batch (no después del último batch)
    if batchIdx < len(batches)-1:
      d = RandDuration(rng, cfg.InterBatchDelayMin, cfg.InterBatchDelayMax)
      SleepWithContext(ctx, d)

  → UpdateStatus job: completed/failed en DB
```

### 4.4 Heartbeat goroutine

Dentro de `processJob`, lanzar goroutine interna que actualiza `last_heartbeat` cada 30 segundos mientras el job corre. Cancelar al terminar el job.

---

## 5. Handler HTTP — `backend/internal/http/handlers/v1_broadcasts.go`

### POST /api/service/v1/difusiones

Cambios:
1. Persistir job + items en DB via `JobQueueRepository.CreateJobWithItems()` antes de encolar en el worker.
2. Calcular `estimated_seconds` con `domain.EstimateBroadcastSeconds(n, timingConfig)`.
3. Respuesta añade `estimated_seconds`.

```json
{
  "ok": true,
  "data": {
    "reference_id": "bc550e8400-...",
    "total": 25,
    "estado": "pending",
    "estimated_seconds": 480
  }
}
```

### GET /api/service/v1/difusiones/{id}

Cambios:
1. Leer job desde DB via `JobQueueRepository.GetByEntityID()`.
2. Leer items desde DB via `JobQueueRepository.GetAllItems()`.
3. `results` ahora viene de DB (no de memoria), siempre disponible si el job fue creado.

---

## 6. WebSocket para panel admin

### Nuevo endpoint: `GET /admin/ws/difusiones/{reference_id}`

- Solo accesible con sesión admin (cookie JWT existente).
- El servidor emite eventos JSON mientras el job está en estado `running`:
  ```json
  { "type": "item_update", "destino": "51987000001", "status": "sent", "index": 5, "total": 25 }
  { "type": "job_complete", "status": "completed", "sent": 23, "failed": 2 }
  ```
- El worker escribe en un `broadcast hub` en memoria (similar al hub WebSocket existente del proyecto).
- Si el cliente conecta cuando el job ya completó, recibe el estado final desde DB inmediatamente.
- El frontend hace polling de fallback (5s interval) si el WS no está disponible.

**Archivo nuevo:** `backend/internal/http/handlers/ws_broadcasts.go`

---

## 7. Frontend — `frontend/app/broadcasts/page.tsx`

**⚠️ Esta sección debe ser diseñada por /ui-ux-pro-max + /bmad-agent-ux-designer**

### Requerimientos funcionales del frontend:

1. **Vista de lista** (ya existe): añadir columna "Progreso" con barra de progreso `sent/total`.
2. **Sheet de detalle** (ya existe): reemplazar con vista completa que incluya:
   - Tabla de resultados por destinatario: `destino | estado | timestamp | error`
   - Filtros por estado (todos / enviados / fallidos / pendientes)
   - Indicador de tiempo estimado restante (calculado en cliente desde `estimated_seconds` y `created_at`)
   - Barra de progreso animada
3. **Modo tiempo real** (nuevo): si el job está `pending` o `running`, conectar WS y actualizar la tabla en vivo.
4. **Estado vacío mejorado**: distinguir "cargando", "procesando", "completado", "fallido".

### Tipos TypeScript necesarios (en `lib/api.ts`):

```typescript
export interface BroadcastItemResult {
  id: number
  sequence_order: number
  destino: string
  status: "pending" | "sent" | "failed" | "skipped"
  error_text?: string
  processed_at?: string
}

export interface BroadcastDetail extends BroadcastInfo {
  estimated_seconds?: number
  items: BroadcastItemResult[]
}

// Evento WS
export type BroadcastWSEvent =
  | { type: "item_update"; destino: string; status: string; index: number; total: number }
  | { type: "job_complete"; status: string; sent: number; failed: number }
```

---

## 8. Recovery al startup

En `main.go` o en la inicialización del container, después de conectar la DB y antes de arrancar el servidor:

```go
stuck, err := jobQueueRepo.RecoverStuckJobs(ctx, 5*time.Minute)
if err != nil {
    log.Error().Err(err).Msg("error recovering stuck jobs")
} else if stuck > 0 {
    log.Warn().Int("count", stuck).Msg("jobs recovered from running state")
}
```

---

## 9. Documentación a actualizar

### `docs/routes/contrato-b2b/difusiones.md`

Cambios:
- Límite: de 500 a **30 destinatarios**.
- Respuesta 202: añadir campo `estimated_seconds` con descripción.
- Nuevo código de error: `MAX_BROADCAST_EXCEEDED` (400) cuando `destinos > 30`.
- Sección de buenas prácticas: mencionar que el envío tarda ~5-15 minutos por diseño anti-ban.

---

## 10. Tests requeridos

### Backend (`backend/internal/whatsapp/broadcast_test.go`)

| Test | Qué verifica |
|------|-------------|
| `TestSplitIntoBatches_NoItemLost` | Todos los items cubiertos sin pérdida ni duplicados |
| `TestSplitIntoBatches_SizeRange` | Tamaño de batch dentro de [BatchSizeMin, BatchSizeMax] |
| `TestRandDuration_Bounds` | 1000 iteraciones siempre en [min, max] |
| `TestSleepWithContext_Cancelled` | Retorna `context.Canceled` sin colgar |
| `TestProcessJob_ContextCancellation` | Context cancelado detiene el job limpiamente |
| `TestEstimateBroadcastSeconds` | Estimación dentro de rango esperado para N destinatarios |

### Backend (`backend/internal/storage/job_queue_test.go`)

| Test | Qué verifica |
|------|-------------|
| `TestCreateJobWithItems` | Job + items persisten correctamente en DB |
| `TestRecoverStuckJobs` | Jobs stuck se resetean a pending |
| `TestUpdateItemStatus` | Estado del item se actualiza correctamente |

---

## 11. Frontend implementado

### Archivos modificados

| Archivo | Qué hace |
|---------|----------|
| `frontend/lib/api.ts` | Nuevos tipos + función `getAdminBroadcastDetail()` |
| `frontend/app/broadcasts/page.tsx` | Página completa reescrita con progreso en tiempo real |

### Tipos TypeScript añadidos (`lib/api.ts`)

```typescript
export interface BroadcastItemResult {
  id: number
  sequence_order: number
  destino: string
  status: "pending" | "sent" | "failed" | "skipped"
  error_text?: string
  processed_at?: string
}

export interface BroadcastDetail extends BroadcastInfo {
  items: BroadcastItemResult[]
}

// Nueva función de API
export async function getAdminBroadcastDetail(
  referenceId: string,
): Promise<{ ok: boolean; data: BroadcastDetail }>
```

### Componentes implementados en `broadcasts/page.tsx`

#### `<ProgressBar sent total />`
Barra visual compacta en la columna "Progreso" de la tabla de lista. Muestra `sent/total` con porcentaje. Transición CSS de 500ms.

#### `<StatusBadge status />`
Badge con icono por cada estado del job:
- `pending` → gris, ícono Clock
- `running` → secondary, ícono Loader2 con `animate-spin`
- `completed` → default (verde), ícono CheckCircle2
- `failed` → destructive, ícono XCircle
- `cancelled` → outline, ícono XCircle

#### `<ItemBadge status />`
Badge compacto por item individual: Enviado / Fallido / Pendiente / Omitido.

#### `useBroadcastDetail(referenceId)` — hook
- Llama a `getAdminBroadcastDetail(id)` al montar.
- Inicia polling cada **3 segundos** si el job está en `pending` o `running`.
- Detiene el polling automáticamente cuando el job pasa a `completed`, `failed` o `cancelled`.
- Estado expuesto: `{ detail: BroadcastDetail | null, loading: boolean }`.

#### `<BroadcastDetailSheet .../>` — Sheet ampliado (`sm:max-w-3xl`)

Secciones:
1. **Empresa + Estado** — grid 2 columnas con nombre de empresa y badge de estado. Si el job está activo, muestra tiempo estimado restante calculado en cliente desde `estimated_seconds` y `created_at`.
2. **Barra de progreso** — barra ancha con porcentaje, contadores Enviados / Fallidos / Pendientes en 3 columnas.
3. **Adjuntos** — lista de archivos con nombre, tamaño, hash (solo si hay adjuntos).
4. **Tabla de destinatarios** — con tabs de filtro:
   - Todos (N) / Enviados (N) / Fallidos (N) / Pendientes (N)
   - Columnas: `#` orden | Destino (enmascarado, últimos 4 dígitos) | Estado (badge) | Hora | Error
   - Overflow horizontal en mobile.
5. **Fecha de creación** — footer del sheet.

#### Privacidad de números de teléfono
```
+51987·····1234  →  muestra últimos 4 dígitos, resto enmascarado con ·
```

#### Tiempo estimado restante
```typescript
function formatSecondsRemaining(estimatedSeconds: number, createdAt: string): string {
  const elapsed = Math.floor((Date.now() - new Date(createdAt).getTime()) / 1000)
  const remaining = Math.max(0, estimatedSeconds - elapsed)
  // "~3 min 24s restantes" | "completando..."
}
```

### Notas de integración

- El frontend funciona con **polling** (ya implementado). El WebSocket admin (`/admin/ws/difusiones/{id}`) es una mejora futura opcional — cuando se implemente, solo hay que conectar el hook al WS y dejar de hacer polling.
- `getAdminBroadcastDetail` llama al endpoint de admin (`/api/admin/difusiones/{id}`), no al endpoint B2B. Si se quiere usar desde el panel del cliente, debe apuntar a `/api/service/v1/difusiones/{id}`.
- El campo `adjuntos` del frontend usa el tipo `AttachmentInfo` existente — sin cambios en el tipo.

---

## 12. Orden de implementación recomendado

1. **Migración 019** — sin ella nada funciona. Correr `/sql-optimization` primero.
2. **`domain/job_queue.go`** — tipos base.
3. **`storage/job_queue.go`** — repositorio con MySQL.
4. **`whatsapp/broadcast.go`** — refactor con delays + heartbeat + persistencia.
5. **`http/handlers/v1_broadcasts.go`** — actualizar handler para usar nueva persistencia.
6. **`http/handlers/ws_broadcasts.go`** — WebSocket admin.
7. **Frontend** — diseño por ui-ux-pro-max/Sally, implementación posterior.
8. **Documentación** — última, cuando el contrato HTTP esté cerrado.
9. **Eliminar `storage/broadcast.go`** — cuando todo migrado y tests verdes.

---

## 12. Notas de riesgo

- **El `BroadcastStore` en memoria no tiene persistencia MySQL hoy.** Si se reinicia el proceso, todos los jobs activos desaparecen. La migración resuelve esto estructuralmente.
- **whatsmeow es cliente no oficial.** Ningún algoritmo elimina el riesgo de ban al 100%. El límite de 30 destinatarios y los delays largos minimizan el perfil de riesgo.
- **El campo `last_heartbeat` requiere que el worker actualice DB cada ~30s.** Si el worker es muy lento por los delays, el heartbeat previene falsos recovery.
- **MariaDB y JSON:** verificar versión exacta que soporta tipo JSON nativo antes de la migración.

---

### Review Findings

#### [Decision Needed]
- [x] [Review][Defer] Omisión del WebSocket administrativo `/admin/ws/difusiones/{reference_id}` para progreso en tiempo real [backend/internal/http/handlers/ws_broadcasts.go:1] — diferido, el sondeo de 3s (polling) ya está implementado y es funcional.

#### [Patches]
- [x] [Review][Patch] Falta de registro del endpoint administrativo de detalle de difusión (`/api/admin/difusiones/{id}`) [backend/internal/http/routes_admin.go:91]
- [x] [Review][Patch] Error de índice en `globalIdx` al procesar lotes de difusiones con tamaño variable [backend/internal/whatsapp/broadcast.go:1209]
- [x] [Review][Patch] Precedencia de evaluación en cláusula `SET` en MySQL impide actualizar `started_at` [backend/internal/storage/job_queue.go:108]
- [x] [Review][Patch] Vulnerabilidad de reinicio prematuro en `RecoverStuckJobs` por `last_heartbeat` nulo inicial [backend/internal/storage/job_queue.go:187]
- [x] [Review][Patch] Incompatibilidad en formato de respuesta JSON para detalles de difusión (`results` vs `items`) [backend/internal/http/handlers/v1_broadcasts.go:112]
- [x] [Review][Patch] Omisión completa de pruebas unitarias y de integración especificadas para el Worker y Repositorio [backend/internal/whatsapp/broadcast_test.go:1]
- [x] [Review][Patch] Trabajos recuperados a `pending` en el arranque se quedan bloqueados por falta de un queue runner activo [backend/internal/http/router.go:305]
- [x] [Review][Patch] Cuello de botella por consultas N+1 en el listado de difusiones para administración [backend/internal/http/router.go:277]
- [x] [Review][Patch] Pánico en Go (enviar a canal cerrado) si se encola una difusión durante el apagado del servidor [backend/internal/whatsapp/broadcast.go:380]
- [x] [Review][Patch] Desfase de Zona Horaria (*Timezone Mismatch*) en la comparación de `last_heartbeat` [backend/internal/storage/job_queue.go:187]
- [x] [Review][Patch] Condición de carrera en `useBroadcastDetail` (React) detiene el polling de progreso en conexiones lentas [frontend/app/broadcasts/page.tsx:118]
- [x] [Review][Patch] Omisión del campo `Metadata` en el struct `Job` y consultas SQL del repositorio [backend/internal/domain/job_queue.go:38]
- [x] [Review][Patch] Contradicción de tipo de datos en la columna `status` (`ENUM` en DB vs `VARCHAR` en spec) [backend/internal/storage/migrations/019_create_job_queue.up.sql:6]
- [x] [Review][Patch] Omisión de manejo de errores en `CreateWithItems` dentro de `PostBroadcast` [backend/internal/http/handlers/v1_broadcasts.go:219]
- [x] [Review][Patch] Truncamiento de errores en el listado de destinatarios sin visibilidad de tooltip [frontend/app/broadcasts/page.tsx:315]

