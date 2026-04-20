# Epic 13: Mensajeria Confiable, Reintentos y Edicion (Backend -> Frontend)

## Objetivo

Resolver de forma integral la confiabilidad de envio de mensajes por API key, garantizando que:

1. El estado de conexion real de WhatsApp sea consistente entre runtime, backend y frontend.
2. Los errores de envio queden persistidos y visibles para operador/API.
3. Existan flujos de reintento y edicion controlada antes de reenviar.

## Requisito clave acordado (nuevo)

En **cada reinicio del backend**, cuando:

1. MySQL ya este conectado, y
2. El puerto HTTP ya este expuesto,

el sistema debe ejecutar automaticamente un **bootstrap de sesiones post-start** para:

- cargar sesiones candidatas desde DB,
- reconstruir clientes runtime en memoria,
- reconciliar `estado_db` vs `estado_runtime_real`,
- y persistir el estado reconciliado con trazabilidad.

Este comportamiento es mandatorio antes de crear stories de frontend.

## Problemas detectados (estado actual)

- El frontend puede mostrar telefono `active` por estado persistido en DB, aunque el cliente runtime de WhatsApp no este conectado en memoria.
- Despues de reiniciar backend, el `Manager` pierde clientes activos (estado en memoria) y un envio puede fallar con `cliente WhatsApp no conectado para este numero`.
- Faltan endpoints de ciclo de vida de mensajes (detalle, editar y reintentar).

## Alcance

- Fase 1 (Backend): bootstrap/reconciliacion post-start + observabilidad de sesion real + persistencia de errores + endpoints de reintento/edicion.
- Fase 2 (Frontend): UX para visualizar estado real, errores y ejecutar editar/reintentar.

---

## Fase 1: Backend (prioridad alta)

### 13.1 Bootstrap de sesiones post-start (DB lista + puerto expuesto)

- Definir secuencia de arranque por fases:
  - `startup` -> `db_ready` -> `port_exposed` -> `reconciliation_running` -> `reconciliation_done`.
- Ejecutar reconciliacion en goroutine post-start (sin bloquear el servidor).
- Cargar telefonos candidatos desde DB (ej. `status=active` o criterio equivalente).
- Intentar reconstruccion runtime con limite de concurrencia y timeout por cuenta.

**Criterios de aceptacion**
- Cada reinicio dispara el bootstrap automaticamente cuando DB y puerto estan listos.
- No hay doble arranque concurrente para la misma cuenta (idempotencia por `accountID`).

### 13.2 Reconciliacion de estado DB vs runtime real

- Calcular y exponer para cada telefono:
  - `status_db` (persistido)
  - `status_runtime` (`manager.Get + client.IsConnected`).
- Aplicar regla de reconciliacion: ante conflicto, persistir estado observado real y razon.
- Dejar trazabilidad de transicion (timestamp + reason).

**Criterios de aceptacion**
- Si DB dice `active` pero runtime no esta conectado, queda explicitamente marcado como inconsistente/reconciliado.
- El sistema no reporta "activo" sin cliente real conectado.

### 13.3 Persistencia fuerte de errores de envio

- Confirmar que todo error de `SendTextMessage` actualiza:
  - `estado = failed`
  - `error_reason`
- Agregar metadatos de intento (si faltan):
  - `retry_count`
  - `last_attempt_at`

**Criterios de aceptacion**
- Todo fallo de envio queda trazable por `reference_id` en DB y API.

### 13.4 Endpoints de ciclo de vida de mensajes

- `GET /api/mensajes/{reference_id}`: detalle con `estado`, `error_reason`, timestamps y trazabilidad de intentos.
- `PATCH /api/mensajes/{reference_id}`: editar contenido/destino solo si estado permite (pending/failed, no sent).
- `POST /api/mensajes/{reference_id}/reintentar`: reintento controlado con nueva transicion de estado.

**Criterios de aceptacion**
- Reintento y edicion respetan ownership API key por telefono.
- No se permite modificar mensajes ya enviados (`sent`/`delivered`).

### 13.5 Contrato API de errores consistente

- Estandarizar payload de error para mensajes y difusiones.
- Incluir codigo, mensaje y `reference_id` cuando aplique.

**Criterios de aceptacion**
- Frontend puede renderizar causa exacta de fallo sin parseos ambiguos.

### 13.6 Observabilidad y rollout seguro

- Agregar logs estructurados del bootstrap/reconciliacion.
- Agregar metricas minimas:
  - sesiones objetivo
  - sesiones reconciliadas
  - sesiones fallidas
  - duracion total/p95
- Habilitar estrategia de rollout con feature flag opcional:
  - `enabled`
  - `dry_run`
  - `max_concurrency`
  - `timeout_ms`

**Criterios de aceptacion**
- Existe kill switch operacional sin cambios de codigo.
- Se puede auditar por que una sesion quedo desconectada o en mismatch.

### 13.7 Testing backend

- Unit/integration para:
  - bootstrap post-start
  - idempotencia de `StartSession`
  - reconciliacion DB/runtime
  - persistencia de `error_reason`
  - editar/reintentar
  - bloqueo de estados no permitidos

---

## Fase 2: Frontend (despues de backend)

### 13.8 Vista de estado de sesion confiable

- Mostrar badge dual: `Conectado (runtime)` vs `Activo (DB)`.
- Mostrar alerta de inconsistencia y CTA para reconectar.

### 13.9 Bandeja de errores y detalle

- En listado de mensajes, mostrar fallidos con `error_reason`.
- Drawer/modal de detalle por `reference_id`.

### 13.10 Editar y reintentar

- UI para editar mensaje fallido.
- Boton reintentar con feedback de estado.

### 13.11 Telemetria UX

- Registrar eventos de reintento/edicion para auditoria operativa.

---

## Definition of Done

- Backend:
  - Bootstrap post-start implementado (DB conectada + puerto expuesto).
  - Reconciliacion de sesion real disponible y persistida.
  - Errores de envio persistidos y expuestos en API.
  - Endpoints de editar/reintentar operativos y con ownership.
  - Logs/metricas de reconciliacion disponibles.
- Frontend:
  - Estado real de sesion visible.
  - Error reason visible en listado/detalle.
  - Editar/reintentar funcional para mensajes failed/pending.

## Orden de ejecucion recomendado

1. 13.1 + 13.2 (bootstrap y reconciliacion real)
2. 13.6 + 13.7 (observabilidad y tests backend)
3. 13.3 + 13.5 (persistencia y contrato de errores)
4. 13.4 (ciclo de vida mensajes)
5. 13.8-13.11 (frontend)
