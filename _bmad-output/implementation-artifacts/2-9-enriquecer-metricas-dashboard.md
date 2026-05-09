---
title: 'Story 2.9 — Enriquecer métricas + rediseñar dashboard'
type: 'feature'
created: '2026-05-08'
status: 'done'
epic: 'epic-2-mejoras-post-revision'
baseline_commit: '3c41157'
context:
  - '{project-root}/_bmad-output/project-context.md'
---

## Story

Como administrador del panel de wsapi,
quiero ver un dashboard con métricas reales, correctas y útiles (empresas activas, mensajes hoy, broadcasts hoy, tasa de éxito, sesiones conectadas, alertas),
para evaluar el estado del servicio sin pedirle datos manuales a desarrollo.

## Acceptance Criteria

**AC1 — Backend: campo names en inglés alineados al frontend:**
`GET /api/admin/metricas` responde con los campos que el frontend ya espera:
```json
{
  "ok": true,
  "active_companies": 5,
  "sessions_active": 3,
  "messages_today": 120,
  "messages_sent": 980,
  "messages_failed": 12,
  "broadcasts_today": 4,
  "broadcasts_created": 87,
  "success_rate": 98.8,
  "last_update": "2026-05-08T14:32:00Z",
  "alerts": []
}
```

**AC2 — Backend: `active_companies` cuenta empresas activas:**
El campo `active_companies` contiene el count de empresas con `activo = true` en DB. Si la DB no está disponible retorna `0` sin error.

**AC3 — Backend: `broadcasts_today` cuenta broadcasts de hoy:**
El campo `broadcasts_today` cuenta broadcasts creados desde el inicio del día local (midnight UTC o server timezone). El campo `broadcasts_created` sigue siendo el total histórico.

**AC4 — Backend: `success_rate` calculado correctamente:**
`success_rate = (messages_sent / max(messages_sent + messages_failed, 1)) * 100`. Si no hay mensajes, retorna `0.0`.

**AC5 — Backend: `last_update` con timestamp RFC3339:**
El campo `last_update` contiene `time.Now().UTC().Format(time.RFC3339)`.

**AC6 — Backend: `alerts` con alertas de sesiones con mismatch:**
Si hay teléfonos con `status = 'active'` en DB pero desconectados en runtime (`mismatch`), se genera una alerta:
```json
{"type": "session_mismatch", "level": "warning", "message": "N teléfonos marcados activos pero desconectados"}
```
Si no hay mismatches, `alerts = []`. La alerta se agrega solo si el handler tiene `telefonoStore`.

**AC7 — Frontend: "Empresas Activas" muestra `active_companies`:**
La card "Empresas Activas" en `dashboard/page.tsx` muestra el valor de `metrics.active_companies` (no `sessions_active`). El sub-texto dice "empresas registradas".

**AC8 — Frontend: "Mensajes Hoy" muestra `messages_today`:**
La card "Mensajes Hoy" muestra `metrics.messages_today`.

**AC9 — Frontend: "Broadcasts Hoy" muestra `broadcasts_today`:**
La card "Broadcasts Hoy" muestra `metrics.broadcasts_today`.

**AC10 — Frontend: "Tasa de Éxito" muestra `success_rate` desde el backend:**
La card "Tasa de Éxito" muestra `metrics.success_rate` (ya calculado en backend, no recalculado en frontend).

**AC11 — Frontend: tab Resumen muestra sesiones activas y detalles completos:**
La sección "Métricas detalladas" del tab Resumen incluye:
- Mensajes enviados (total): `messages_sent`
- Mensajes fallidos: `messages_failed`
- Broadcasts completados (total): `broadcasts_created`
- Empresas registradas: `active_companies`
- Sesiones activas: `sessions_active`

**AC12:** `cd backend && go build ./...` pasa sin errores.

**AC13:** `cd backend && go test ./...` pasa sin nuevas regresiones.

**AC14:** `cd frontend && npm run lint` pasa sin nuevos errores (los 3 pre-existentes en `api.ts` son aceptables).

## Tasks / Subtasks

- [x] **Tarea 1: Actualizar `DashboardMetricsResponse` y agregar `DashboardAlert`** (AC: 1, 5)
  - [x] Agregar struct `DashboardAlert` en `backend/internal/http/router.go`
  - [x] Actualizar `DashboardMetricsResponse`: cambiar JSON tags a inglés, agregar campos nuevos

- [x] **Tarea 2: Ampliar `DashboardHandler` con telefonoStore y db** (AC: 2, 3, 6)
  - [x] Agregar `telefonoStore *storage.TelefonoStore` y `db *sql.DB` a `DashboardHandler` struct
  - [x] Actualizar `NewDashboardHandler` para aceptar los nuevos parámetros
  - [x] Actualizar `container.go` para pasar `telefonoStore` y `db`

- [x] **Tarea 3: Actualizar `GetMetrics` con datos enriquecidos** (AC: 1–6)
  - [x] Contar `active_companies` (query directa a DB: `SELECT COUNT(*) FROM empresas WHERE activo = true`)
  - [x] Contar `broadcasts_today` (query directa a DB con `timestamp_created >= todayStart`)
  - [x] Calcular `success_rate` desde mensajes exitosos/fallidos
  - [x] Agregar `last_update` = `time.Now().UTC().Format(time.RFC3339)`
  - [x] Detectar mismatches y generar `alerts`
  - [x] Actualizar el `json.NewEncoder(w).Encode(...)` para usar los nuevos campos

- [x] **Tarea 4: Corregir `frontend/app/dashboard/page.tsx`** (AC: 7–11)
  - [x] Corregir card "Empresas Activas": `value={fmt(metrics?.active_companies || 0)}`, `sub="empresas registradas"`
  - [x] Corregir card "Mensajes Hoy": `value={fmt(metrics?.messages_today || 0)}`
  - [x] Corregir card "Broadcasts Hoy": `value={fmt(metrics?.broadcasts_today || 0)}`
  - [x] Corregir card "Tasa de Éxito": `value={fmtPct(metrics?.success_rate || 0)}`
  - [x] Agregar "Sesiones activas" en sección de métricas detalladas

- [x] **Tarea 5: Build, tests y lint** (AC: 12, 13, 14)
  - [x] `cd backend && go build ./...`
  - [x] `cd backend && go test ./...`
  - [x] `cd frontend && npm run lint`

### Review Findings

- [x] [Review][Patch] nil dereference en `msgMetrics` sin nil-guard antes de acceder a `.MensajesExitosos` / `.MensajesFallidos` [router.go ~line 414]
- [x] [Review][Defer] Variable shadowing de `err` en rama else del filtro de empresa [router.go ~line 344] — deferred, pre-existente
- [x] [Review][Defer] `json.NewEncoder(w).Encode(...)` no verifica el error de escritura — patrón pre-existente en todo el handler [router.go] — deferred, pre-existente

## Dev Notes

### 🚨 Problema actual: mismatch de field names backend ↔ frontend

Este es el bug principal de esta story. El backend retorna JSON con nombres en español:
```json
{"mensajes_hoy": 120, "mensajes_exitosos": 980, "sesiones_activas": 3, ...}
```

Pero el frontend en `frontend/lib/api.ts` define `DashboardMetrics` con nombres en inglés:
```ts
interface DashboardMetrics {
  active_companies: number;
  messages_today: number;
  broadcasts_today: number;
  success_rate: number;
  last_update: string;
  sessions_active: number;
  messages_sent: number;
  messages_failed: number;
  broadcasts_created: number;
  alerts: Alert[];
}
```

Como resultado, **el dashboard muestra ceros en todas las métricas** aunque haya datos en DB. El fix es cambiar los JSON tags del backend para que coincidan con la interfaz TypeScript.

**NO modificar la interfaz TypeScript** — el backend debe adaptarse a los nombres que el frontend ya espera.

### 🚨 Bug adicional en dashboard/page.tsx

La card "Empresas Activas" usa el valor incorrecto:
```tsx
// ACTUAL (bug):
value={fmt(metrics?.sessions_active || 0)}  // muestra sesiones, no empresas
sub="Sesiones activas"

// CORRECTO:
value={fmt(metrics?.active_companies || 0)}
sub="empresas registradas"
```

### 📐 Tarea 1 — Código exacto: router.go

Agregar antes de `DashboardMetricsResponse`:
```go
type DashboardAlert struct {
    Type    string `json:"type"`
    Level   string `json:"level"`
    Message string `json:"message"`
}
```

Reemplazar `DashboardMetricsResponse`:
```go
type DashboardMetricsResponse struct {
    OK               bool             `json:"ok"`
    ActiveCompanies  int              `json:"active_companies"`
    SessionsActive   int              `json:"sessions_active"`
    MessagesToday    int64            `json:"messages_today"`
    MessagesSent     int64            `json:"messages_sent"`
    MessagesFailed   int64            `json:"messages_failed"`
    BroadcastsToday  int64            `json:"broadcasts_today"`
    BroadcastsCreated int64           `json:"broadcasts_created"`
    SuccessRate      float64          `json:"success_rate"`
    LastUpdate       string           `json:"last_update"`
    Alerts           []DashboardAlert `json:"alerts"`
}
```

### 📐 Tarea 2 — Código exacto: DashboardHandler struct y constructor

```go
type DashboardHandler struct {
    msgRepo       storage.MessagesRepository
    sessionStore  *storage.SessionStore
    empresaStore  domain.EmpresaStoreInterface
    telefonoStore *storage.TelefonoStore // para detectar mismatches
    db            *sql.DB               // para queries directas de conteo
}

func NewDashboardHandler(
    msgRepo storage.MessagesRepository,
    sessionStore *storage.SessionStore,
    empresaStore domain.EmpresaStoreInterface,
    telefonoStore *storage.TelefonoStore,
    db *sql.DB,
) *DashboardHandler {
    return &DashboardHandler{
        msgRepo:       msgRepo,
        sessionStore:  sessionStore,
        empresaStore:  empresaStore,
        telefonoStore: telefonoStore,
        db:            db,
    }
}
```

**Import a agregar en router.go:** `"database/sql"` (si no está ya importado)

### 📐 Tarea 2 — container.go

Cambiar:
```go
dashboardHandler := NewDashboardHandler(msgRepo, sessionStore, empresaStore)
```
Por:
```go
dashboardHandler := NewDashboardHandler(msgRepo, sessionStore, empresaStore, telefonoStore, db)
```

> `telefonoStore` y `db` ya existen en `NewContainer` — no hay que crearlos.

### 📐 Tarea 3 — GetMetrics completo

```go
func (h *DashboardHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    filter, ok := domain.GetEmpresaFilter(r.Context(), r.Header.Get("X-Empresa-ID"))
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(DashboardMetricsResponse{OK: false})
        return
    }

    if h.msgRepo == nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(DashboardMetricsResponse{OK: false})
        return
    }

    var metrics *storage.MessageMetrics
    var err error

    if filter.IsRoot && filter.EmpresaID == nil {
        empresaIDStr := strings.TrimSpace(r.URL.Query().Get("empresa_id"))
        if empresaIDStr != "" {
            if empresaID, parseErr := strconv.ParseInt(empresaIDStr, 10, 64); parseErr == nil && empresaID > 0 {
                empresa, err := h.empresaStore.GetByID(empresaID)
                if err != nil || empresa == nil {
                    w.WriteHeader(http.StatusNotFound)
                    json.NewEncoder(w).Encode(DashboardMetricsResponse{OK: false})
                    return
                }
                metrics, err = h.msgRepo.GetMessageMetricsByEmpresa(empresa.ID)
                if err != nil {
                    w.WriteHeader(http.StatusInternalServerError)
                    json.NewEncoder(w).Encode(DashboardMetricsResponse{OK: false})
                    return
                }
            } else {
                metrics, err = h.msgRepo.GetAllMessageMetrics()
            }
        } else {
            metrics, err = h.msgRepo.GetAllMessageMetrics()
        }
    } else {
        empresa, err := domain.GetRUCFromContext(r.Context(), filter, h.empresaStore)
        if err != nil || empresa == "" {
            w.WriteHeader(http.StatusForbidden)
            json.NewEncoder(w).Encode(DashboardMetricsResponse{OK: false})
            return
        }
        if empresaID, ok := domain.GetEmpresaIDFromContext(r.Context(), filter); ok {
            metrics, err = h.msgRepo.GetMessageMetricsByEmpresa(empresaID)
        } else {
            w.WriteHeader(http.StatusForbidden)
            json.NewEncoder(w).Encode(DashboardMetricsResponse{OK: false})
            return
        }
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(DashboardMetricsResponse{OK: false})
            return
        }
    }

    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(DashboardMetricsResponse{OK: false})
        return
    }

    // — Sesiones activas —
    sessionCount := 0
    if filter.IsRoot && filter.EmpresaID == nil {
        sessionCount = h.sessionStore.ActiveCount()
    } else {
        empresa, _ := domain.GetRUCFromContext(r.Context(), filter, h.empresaStore)
        if empresa != "" {
            if state, ok := h.sessionStore.Get(empresa); ok && state.Status == "active" && state.IsActive {
                sessionCount = 1
            }
        }
    }

    // — Active companies —
    activeCompanies := 0
    if h.db != nil {
        _ = h.db.QueryRow("SELECT COUNT(*) FROM empresas WHERE activo = TRUE").Scan(&activeCompanies)
    }

    // — Broadcasts hoy —
    var broadcastsToday int64
    if h.db != nil {
        now := time.Now()
        todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
        _ = h.db.QueryRow("SELECT COUNT(*) FROM broadcasts WHERE created_at >= ?", todayStart).Scan(&broadcastsToday)
    }

    // — Success rate —
    var successRate float64
    total := metrics.MensajesExitosos + metrics.MensajesFallidos
    if total > 0 {
        successRate = float64(metrics.MensajesExitosos) / float64(total) * 100
    }

    // — Alerts —
    alerts := []DashboardAlert{}
    if h.telefonoStore != nil && h.sessionStore != nil && filter.IsRoot && filter.EmpresaID == nil {
        telefonos, err := h.telefonoStore.ListAll()
        if err == nil {
            mismatchCount := 0
            for _, phone := range telefonos {
                if phone.Status == domain.TelefonoStatusActive {
                    accountID := whatsapp.NormalizeAccountID(phone.NumeroCompleto)
                    _ = accountID // solo para referencia; la detección usa sessionStore
                    if state, ok := h.sessionStore.Get(phone.NumeroCompleto); !ok || state.Status != "active" {
                        mismatchCount++
                    }
                }
            }
            if mismatchCount > 0 {
                alerts = append(alerts, DashboardAlert{
                    Type:    "session_mismatch",
                    Level:   "warning",
                    Message: fmt.Sprintf("%d teléfonos marcados activos pero desconectados", mismatchCount),
                })
            }
        }
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(DashboardMetricsResponse{
        OK:                true,
        ActiveCompanies:   activeCompanies,
        SessionsActive:    sessionCount,
        MessagesToday:     metrics.MensajesHoy,
        MessagesSent:      metrics.MensajesExitosos,
        MessagesFailed:    metrics.MensajesFallidos,
        BroadcastsToday:   broadcastsToday,
        BroadcastsCreated: metrics.BroadcastsEjecutados,
        SuccessRate:       math.Round(successRate*10) / 10,
        LastUpdate:        time.Now().UTC().Format(time.RFC3339),
        Alerts:            alerts,
    })
}
```

**Imports necesarios en router.go:** `"math"` y `"wsapi/internal/whatsapp"` (verificar si ya están).

### 📐 Tarea 3 — broadcasts tabla: verificar campo created_at

Antes de implementar `broadcasts_today`, verificar qué campo de timestamp existe en la tabla `broadcasts`. Las opciones más probables son `created_at` o `timestamp_created`. Hacer:
```bash
grep -r "broadcasts" /home/fulanito/development/wsapi/backend/internal/storage/migrations/ | grep "CREATE TABLE" -A 20
```
O bien buscar en el storage de broadcasts:
```go
// backend/internal/whatsapp/broadcast.go o storage/broadcast*.go
```
Usar el campo correcto. Si no existe un campo de fecha, omitir `broadcasts_today` y retornar `0` en su lugar (AC3 sería parcialmente N/A).

### 📐 Tarea 4 — dashboard/page.tsx cambios exactos

**Cards superiores (línea ~123–150):**
```tsx
// Card 1: Empresas
<MetricCard
  href="/empresas"
  title="Empresas Activas"
  icon={Building2}
  value={fmt(metrics?.active_companies || 0)}
  sub="empresas registradas"
/>

// Card 2: Mensajes (sin cambio en nombre — ya usa messages_today)
<MetricCard
  href="/messages"
  title="Mensajes Hoy"
  icon={MessageSquare}
  value={fmt(metrics?.messages_today || 0)}
  sub="Enviados hoy"
/>

// Card 3: Broadcasts
<MetricCard
  href="/broadcasts"
  title="Broadcasts Hoy"
  icon={Send}
  value={fmt(metrics?.broadcasts_today || 0)}
  sub="Creados hoy"
/>

// Card 4: Tasa de éxito (ya correcto — success_rate viene del backend ahora)
<MetricCard
  title="Tasa de Éxito"
  icon={CheckCircle2}
  value={fmtPct(metrics?.success_rate || 0)}
  sub="Mensajes entregados"
/>
```

**Tab Resumen — añadir sesiones activas:**
```tsx
<div className="flex items-center justify-between text-sm">
  <span className="text-muted-foreground">Sesiones conectadas</span>
  <span className="font-medium">{fmt(metrics?.sessions_active || 0)}</span>
</div>
```
Agregar este bloque junto a los otros existentes (mensajes_sent, messages_failed, etc.).

### ⚠️ Campos de MessageMetrics en español (NO cambiar)

Los campos de `MessageMetrics` en `storage/messages.go` usan nombres en español internamente:
```go
type MessageMetrics struct {
    TotalMensajes        int64
    MensajesHoy          int64
    MensajesSemana       int64
    MensajesExitosos     int64
    MensajesFallidos     int64
    BroadcastsEjecutados int64
}
```
**No modificar estos campos** — son internos. Solo cambian los JSON tags del `DashboardMetricsResponse`.

### ⚠️ Lo que NO cambia esta story

- `storage/messages.go` — `MessageMetrics` struct y métodos no se tocan
- `frontend/lib/api.ts` — `DashboardMetrics` interface NO se modifica (ya tiene los nombres correctos)
- `MessagesRepository` interface — no se modifica
- `EmpresaStoreInterface` — no se modifica
- Ningún handler distinto al `DashboardHandler`
- `sessionStore` y su uso en la lógica de sesiones activas — no cambia la lógica, solo el JSON tag

### ⚠️ Verificar import `math` en router.go

El `math.Round(successRate*10) / 10` requiere `"math"`. Verificar si está en el import block de `router.go`. Si no está, agregarlo.

### ⚠️ `filter.IsRoot` para alertas

Las alertas de mismatch solo se generan cuando el usuario es root (ve todos los teléfonos). Para usuarios de empresa, `filter.IsRoot` es false y `filter.EmpresaID` tiene valor, por lo que el bloque de alertas no se ejecuta y `alerts` queda como `[]`.

### 🧠 Cosas a verificar antes de implementar

1. **Campo fecha en tabla `broadcasts`**: usar `grep -r "CREATE TABLE broadcasts" backend/` para ver el schema. Si el campo es `created_at`, usar `created_at`; si es `timestamp_created`, usar `timestamp_created`.

2. **Import de `sql` en router.go**: El archivo ya importa muchas cosas. Verificar si `"database/sql"` ya está en el bloque de imports.

3. **Import de `whatsapp` en router.go**: Verificar si ya está — se usa en `ConnectCompanyPhoneWS` etc. Lo más probable es que sí.

4. **`math.Round` vs truncar**: Usar `math.Round(successRate*10) / 10` para una decimal. Requiere import `"math"`.

5. **`telefonoStore.ListAll()` vs `GetByEmpresa`**: `ListAll()` ya existe en `storage/telefono.go:99`. Usarlo directamente.

6. **sessionStore.Get(phone.NumeroCompleto)**: El sessionStore usa `NumeroCompleto` como key (no accountID). Esto ya está validado en el codebase.

### Learnings de stories anteriores

- Package de `router.go` es `package http` (mismo package que `admin.go`, `container.go`)
- `fmt.Printf` para logs (no zerolog)
- `nil-guard` en stores es obligatorio — siempre verificar `if h.db != nil` antes de queries directas
- `_ = h.db.QueryRow(...)` para queries auxiliares que no deben romper el handler si fallan
- `writeAdminJSON` y `writeAdminError` disponibles, pero en `GetMetrics` se usa `json.NewEncoder` directamente — mantener el patrón existente
- El package de handlers en `internal/http/handlers/` es `package http` también — mismo package que `router.go`
- En Go, `time.Date(now.Year()...)` para calcular inicio de día — ya establecido en `messages.go`
- `whatsapp.NormalizeAccountID` ya disponible en package `whatsapp`

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

### Completion Notes List

- `DashboardMetricsResponse` actualizado con JSON tags en inglés y campos nuevos (`active_companies`, `sessions_active`, `messages_today`, `messages_sent`, `messages_failed`, `broadcasts_today`, `broadcasts_created`, `success_rate`, `last_update`, `alerts`). Se reutilizó el struct `Alert` existente en lugar de crear `DashboardAlert` duplicado.
- `DashboardHandler` ampliado con `telefonoStore` y `db`. `NewDashboardHandler` actualizado con nuevos parámetros. `container.go` actualizado para pasar ambos.
- `GetMetrics` reescrito: cuenta empresas activas vía query DB, broadcasts_today con `created_at >= todayStart`, `success_rate` con `math.Round`, `last_update` RFC3339, alertas de mismatch usando `telefonoStore.ListAll()` + `sessionStore.Get()`.
- `dashboard/page.tsx`: card "Empresas Activas" corregida para usar `active_companies` (antes usaba `sessions_active`). Sub-texto actualizado a "empresas registradas". Añadida fila "Sesiones activas" en métricas detalladas.
- `go build ./...` OK, `go test ./...` OK (todos los tests pasan), `npm run lint` solo 3 errores pre-existentes aceptados por AC14.

### File List

- backend/internal/http/router.go
- backend/internal/http/container.go
- frontend/app/dashboard/page.tsx

### Change Log

- 2026-05-08: Story creada — fix mismatch nombres JSON backend↔frontend, enriquecer métricas dashboard
- 2026-05-08: Implementación completa — DashboardMetricsResponse con campos en inglés, DashboardHandler con telefonoStore+db, GetMetrics enriquecido, dashboard/page.tsx corregido
