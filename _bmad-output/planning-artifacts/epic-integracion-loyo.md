---
project_name: 'wsapi'
user_name: 'Fulanito'
date: '2026-05-17'
source: 'auditoría cruzada con loyalty-app (ver wsapi-action-plan.md)'
status: 'listo para copiar a wsapi/_bmad-output/planning-artifacts/epics.md'
target_branch: 'feature/integracion-loyo'
---

# Épicas — wsapi (sección a integrar)

_Esta sección debe **añadirse** al `epics.md` existente de wsapi (no reemplazarlo). Si no existe, este es el primer epic. Numerar según corresponda con épicas previas del proyecto._

---

## Epic 5: Habilitar wsapi como provider B2B para integraciones serverless

### Contexto

Loyalty-app (proyecto Loyo) integra wsapi como tercer provider de WhatsApp junto con Meta y QR (Evolution API). Loyo corre en Vercel (serverless) y **no puede mantener conexiones WebSocket persistentes**, lo que bloquea automatizaciones bidireccionales y tracking de estado de mensajes.

Este epic cierra dos gaps detectados en auditoría que habilitan a wsapi como proveedor de calidad enterprise para integradores B2B:

1. **Webhooks salientes** — sustituye al WebSocket para clientes serverless
2. **Healthcheck público** — habilita el flujo de validación de URL en clientes integradores

**Referencia completa**: `loyalty-app/_bmad-output/planning-artifacts/wsapi-action-plan.md`

### Definición de done del epic

- Loyalty-app puede registrar un webhook desde su panel y recibir eventos en tiempo real (mensajes entrantes, cambios de estado, desconexiones).
- Un integrador puede hacer un `GET /api/service/v1/health` antes de pedirle credenciales al usuario, validando que la URL apunta efectivamente a un wsapi vivo.
- Sin regresión en endpoints existentes ni en flujo del Admin WS.

### Rama obligatoria

Según `docs/bmad-project-rules.md` de wsapi: crear la rama `feature/integracion-loyo` antes de tocar `backend/` o `frontend/`. Verificar con `git branch --show-current`.

---

### Story 5.1: Endpoint público `/api/service/v1/health`

**Objetivo**: Permitir a integradores B2B validar la URL del servicio antes de guardar credenciales.

**Implementación**:
- Nuevo handler en `backend/internal/http/handlers/v1_health.go`
- Sin auth (público)
- Response: `{"ok": true, "service": "wsapi", "version": "X.Y.Z", "timestamp": "..."}`
- La versión se inyecta en build time vía `-ldflags "-X main.Version=..."` o se lee de `internal/config`

**Cambios en `routes_api.go`**:
```go
if c.HealthHandler != nil {
    mux.Handle("GET /api/service/v1/health", http.HandlerFunc(c.HealthHandler.GetHealth))
}
```

**Criterios de aceptación**:
- [ ] Endpoint responde 200 sin auth
- [ ] Incluye nombre del servicio, versión, timestamp
- [ ] Rate-limited (evitar abuso) — usar middleware existente si lo hay
- [ ] Documentado en `docs/` (sección API)
- [ ] Test de integración con `httptest`

**Esfuerzo**: pequeño (1 sesión de dev)

---

### Story 5.2: Tabla y modelo de webhooks salientes

**Objetivo**: Persistir configuración de webhooks por empresa.

**Cambios**:
- Migración SQL en `backend/internal/storage/migrations/`:
  ```sql
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
- Modelos Go en `backend/internal/domain/webhook.go`
- Store en `backend/internal/storage/webhook_store.go`

**Criterios de aceptación**:
- [ ] Migración corre limpio (up + down)
- [ ] Recordatorio: invocar skill `/sql-optimization` antes de aprobar (regla de proyecto)
- [ ] Tests de store CRUD

---

### Story 5.3: Endpoints REST para gestión de webhooks

**Objetivo**: Permitir al integrador registrar/listar/eliminar sus webhooks vía API.

**Endpoints** (auth: API key):
- `POST /api/service/v1/webhooks` — Body: `{url, eventos: ["message.received", ...]}` → Response: `{id, secret}` (el secret solo se muestra una vez)
- `GET /api/service/v1/webhooks` — Lista los webhooks de la empresa (sin exponer el secret)
- `DELETE /api/service/v1/webhooks/{id}` — Elimina

**Validaciones**:
- URL debe ser HTTPS (rechazar HTTP en producción)
- Lista de eventos debe ser subset de los permitidos: `message.received`, `message.status_update`, `session.connected`, `session.disconnected`
- Máximo N webhooks por empresa (límite configurable)

**Criterios de aceptación**:
- [ ] Handler en `backend/internal/http/handlers/v1_webhooks.go`
- [ ] Registrado en `routes_api.go` con `clientStack` (API key auth)
- [ ] Secret generado con `crypto/rand` (32 bytes hex)
- [ ] Tests cubriendo validaciones

---

### Story 5.4: Worker de envío con retry exponencial

**Objetivo**: Procesar la cola `webhooks_outbound_queue` y entregar eventos con confiabilidad.

**Diseño**:
- Goroutine arrancada en `Container.StartupTasks` (patrón ya existente para bootstrap de whatsmeow)
- Polling cada 5s a `webhooks_outbound_queue WHERE estado='pending' AND proximo_intento_at <= NOW()`
- Por cada item:
  - Marcar `estado='sending'`
  - POST al webhook con:
    - Header `Content-Type: application/json`
    - Header `X-Wsapi-Signature: sha256=<hmac>` (HMAC-SHA256 del body con `webhooks_outbound.secret`)
    - Header `X-Wsapi-Event: <event_type>`
    - Header `X-Wsapi-Delivery: <queue.id>` (idempotencia para el receptor)
    - Body: el payload JSON
  - Si 2xx → marcar `done`, reset failure_count del webhook
  - Si no → incrementar `intentos`, calcular `proximo_intento_at` (1m, 5m, 30m, 2h, 6h, 24h)
  - Tras 6 intentos fallidos → marcar `estado='failed'`, incrementar `webhook.failure_count`
  - Si `webhook.failure_count >= 20` → desactivar webhook automáticamente

**Criterios de aceptación**:
- [ ] Worker es interrumpible vía context (para shutdown limpio)
- [ ] No procesa el mismo item dos veces (lock optimista con `WHERE estado='pending'` en el UPDATE)
- [ ] Timeout HTTP de 10s por entrega
- [ ] Logs estructurados con `zerolog` (evento, webhook_id, status, latencia)
- [ ] Tests con servidor `httptest` simulando 2xx/5xx/timeout

---

### Story 5.5: Emisión de eventos desde el core de whatsmeow

**Objetivo**: Conectar los handlers de whatsmeow con la cola de webhooks.

**Eventos a emitir**:

| Trigger en whatsmeow | event_type | Payload |
|---------------------|------------|---------|
| `*events.Message` (incoming) | `message.received` | `{telefono_id, from, message_id, content, type, timestamp}` |
| Update de estado de mensaje enviado | `message.status_update` | `{telefono_id, reference_id, status, timestamp}` |
| `*events.Connected` | `session.connected` | `{telefono_id, phone, timestamp}` |
| `*events.Disconnected` / `*events.LoggedOut` | `session.disconnected` | `{telefono_id, reason, timestamp}` |

**Implementación**:
- En `backend/internal/whatsapp/` localizar los handlers existentes
- Añadir una función `emitWebhookEvent(empresaID, eventType, payload)` que:
  1. Lea de `webhooks_outbound WHERE empresa_id = ? AND activo = 1 AND JSON_CONTAINS(eventos, ?)`
  2. Por cada webhook activo, inserte en `webhooks_outbound_queue`

**Criterios de aceptación**:
- [ ] No bloquea el handler de whatsmeow (encolado async)
- [ ] Si la BD falla, log de error pero NO crashea el handler
- [ ] Test de integración: simular `*events.Message` → verificar que se inserta en la cola

---

### Story 5.6: Documentación de la integración por webhooks

**Objetivo**: Documentar el sistema de webhooks para que cualquier integrador (no solo Loyo) pueda usarlo.

**Cambios**:
- Nuevo doc `docs/webhooks-integracion.md`:
  - Lista de eventos disponibles + payloads de ejemplo
  - Cómo registrar un webhook (cURL example)
  - Cómo verificar la firma HMAC (snippet en Node.js + Go)
  - Política de retry y backoff
  - Comportamiento ante fallos (desactivación automática)

**Criterios de aceptación**:
- [ ] El doc cubre los 4 eventos mínimos
- [ ] Ejemplos de payload reales (no inventados)
- [ ] Linkeado desde el README principal de wsapi

---

### Resumen del Epic

| Story | Cambio principal | Bloquea a | Riesgo |
|-------|------------------|-----------|--------|
| 5.1 | Health endpoint | Loyo Story 1.6 (validar URL) | Bajo |
| 5.2 | Migración + modelo | 5.3, 5.4, 5.5 | Medio (DB) |
| 5.3 | API CRUD webhooks | Loyo registro de webhook | Bajo |
| 5.4 | Worker retry | Confiabilidad | Medio |
| 5.5 | Hooks de whatsmeow | Eventos reales | Medio (riesgo de regresión) |
| 5.6 | Docs | Adopción | Bajo |

---

## Cómo proceder en wsapi

1. **Copiar este archivo** a `_bmad-output/planning-artifacts/epics.md` (anexar al existente)
2. **Renombrar** "Epic 5" → número correcto según las épicas previas de wsapi
3. **Crear rama**: `git checkout -b feature/integracion-loyo` (regla del proyecto)
4. **Correr en wsapi**:
   - `bmad-sprint-planning` (regenera `sprint-status.yaml` con las nuevas stories)
   - `bmad-create-story` para la primera story (`5.1: Health endpoint`) — la más barata
   - `bmad-dev-story` para implementarla
5. **Orden recomendado de ejecución**: 5.1 → 5.2 → 5.3 → 5.4 → 5.5 → 5.6 (cada una desbloquea la siguiente)

---

## Sincronización con loyalty-app

Cuando termines 5.1 (health endpoint) → Story 1.6 de Loyo se desbloquea (puede usar healthcheck real en lugar de probar `/sesion`).

Cuando termines 5.5 (eventos) → Story 1.8 de Loyo se desbloquea (puede implementar el webhook receiver).

Volver a este proyecto (loyalty-app) y actualizar el `epics.md` marcando los unblocks.
