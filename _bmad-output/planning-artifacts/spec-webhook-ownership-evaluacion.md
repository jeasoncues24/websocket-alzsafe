---
title: 'Evaluación de modelo de ownership de webhooks'
type: 'spec'
status: 'ready-for-review'
source: 'party-mode roundtable — Mary (analyst), Winston (architect), John (PM), Amelia (dev)'
created: '2026-05-22'
reviewer: 'next-model'
---

# Evaluación: Modelo de ownership de webhooks (per-API-key vs per-empresa)

## Problema

Los webhooks salientes de wsapi (`webhooks_outbound`) se registran en la capa `clientStack` (API key auth). Existen dos stacks de autenticación:

- **`clientStack`** — API key, usado por mensajes, difusiones y webhooks
- **`empresaStack`** — JWT empresa, usado por teléfonos, sesiones, métricas

La tabla `webhooks_outbound` tiene FK a `empresa_id`. El handler `v1_webhooks.go` obtiene el `empresa_id` desde `domain.GetApiKeyClaims`. Esto ata el webhook al lifecycle de una API key específica, no a la empresa como unidad de negocio.

**Pregunta central:** ¿Debe el webhook pertenecer a la API key o a la empresa?

## Contexto técnico

- Meta Webhooks Cloud API permite **solo una callback URL por app/número** — no una por API key
- El delivery worker (`webhook_delivery_worker.go`) ya lee por `empresa_id`, no discrimina por API key
- El emitter (`webhook_events.go`) dispara eventos desde whatsmeow core, también a nivel de empresa
- El panel admin (frontend) aún no expone webhooks — solo existen via API REST
- El admin JWT (empresa_id en body/path) y API key (empresa_id en claim) son dos orígenes distintos para el mismo `empresa_id`

## Análisis (Party Mode — 4 agentes)

### Mary (Business Analyst)

**Por empresa.** El webhook pertenece al negocio, no a una llave. Una empresa puede tener múltiples API keys (dev, prod, integración), pero los mensajes entrantes de su número deben ir a un solo endpoint. Si cada API key tiene su propio webhook: ¿cuál se dispara cuando llega un mensaje? ¿Todos? Confusión y duplicidad.

**Dual visibility:** deseable si se maneja bien — admin panel para soporte/debug, API para clientes. Riesgo: exponer signing key del webhook en admin panel sin RBAC.

**Meta risk:** si un cliente usa wsapi para spam, Meta puede marcar la plataforma, no solo al cliente. Mitigación: términos de uso con transferencia de responsabilidad.

### Winston (Architect)

**Por empresa**, no hay discusión. API key es auth, empresa es configuración de negocio. Meta solo permite una callback URL por número.

**Dual visibility — 3 riesgos técnicos:**

1. **Conflictos de actualización:** admin cambia URL mientras API la usa. Sin transacción atómica, datos inconsistentes.
2. **Race conditions:** actualización concurrente sin versioning u optimistic locking pierde cambios.
3. **Seguridad:** webhook visible desde API key eleva el daño potencial de un token comprometido.

**Recomendación:** webhook configurable desde admin siempre (configuración crítica). API solo si hay caso de uso real (automatización de provisioning). Si se habilita API, usar recurso separado con permisos explícitos + mecanismo de bloqueo/versión.

### John (Product Manager)

**Por empresa.** El job-to-be-done del integrador B2B: "Recibir eventos de mi cliente sin cruzarme con eventos de otro cliente." API key es un mecanismo de auth, la empresa es la unidad de negocio.

**Meta risks:**
- Meta exige que proceses webhooks tú mismo si actúas como BSP. Re-expedir sin transformación limpia viola términos.
- Meta penaliza latencia — si tu webhook tarda, te marcan.
- Webhooks caídos constantemente → Meta bloquea el número.

**Pregunta clave sin resolver:** *¿Quién es dueño del webhook? ¿El integrador B2B o tu equipo de ops? Si son los dos, ¿quién gana en conflicto?*

### Amelia (Developer)

**Por empresa, con `created_by` para auditoría.** El webhook es de la empresa, no del API key. El API key delega porque el claim contiene `empresa_id`. Admin ve los mismos registros.

**Cambios necesarios en código actual:**
- `v1_webhooks.go:43` — validar `empresa_id` contra claims; admin usa `empresa_id` de otro origen (path param o body)
- `webhooks_outbound` tabla — agregar columna `created_by VARCHAR(20)` + `updated_by` para auditoría
- Middleware dual: `clientStack` + nuevo `adminStack` en `routes_api.go`
- Handler `AdminWebhooksCreate` — replica lógica con source admin
- Worker `webhook_delivery_worker.go` — ya lee por `empresa_id`, no distingue origen. Sin cambios.

**Riesgo principal:** dual management sin coordinación → race conditions, source of truth duplicada, eventos perdidos sin notificación.

## Consenso

Los 4 agentes coinciden: **el webhook debe pertenecer a la empresa, no a la API key.**

| Aspecto | Consenso |
|---------|----------|
| ¿Per-API-key? | ❌ Descartado — imposible técnicamente (Meta: 1 callback URL por número) |
| ¿Per-empresa? | ✅ Sí — es la unidad de negocio correcta |
| ¿Admin + API? | ⚠️ Sí, pero con condiciones: locking, created_by, autorización explícita |
| ¿Riesgo Meta? | ⚠️ Real — mitigable con ToS claros, rate limits, monitoreo de 5xx |

## Pregunta abierta (John)

> *"¿Quién es dueño del webhook? ¿El integrador B2B o tu equipo de ops? Si son los dos, ¿quién gana en conflicto?"*

Esta decisión debe resolverse antes de implementar dual visibility. Las opciones:

1. **Dueño: integrador (API key).** Admin puede leer pero no escribir. El cliente controla su webhook.
2. **Dueño: ops (admin).** API key puede leer pero no escribir. Ops controla.
3. **Co-dueño con prioridad.** El último escritor gana (con auditoría). Riesgo: conflictos silenciosos.
4. **Co-dueño con locking.** Un admin lock impide escritura vía API hasta release. Más seguro, más complejo.

## Archivos relevantes

- `backend/internal/http/routes_api.go:64-66` — webhook routes en `clientStack`
- `backend/internal/http/handlers/v1_webhooks.go` — handler que lee `empresa_id` de claims
- `backend/internal/http/kernel.go` — define `EmpresaAuth` (admin) y `ServiceStack` (client)
- `backend/internal/storage/webhook_store.go` — store con FK `empresa_id`
- `backend/internal/whatsapp/webhook_delivery_worker.go` — delivery worker
- `backend/internal/whatsapp/webhook_events.go` — event emitter
- `backend/internal/http/container.go:186,137` — wiring de middleware

## Referencias

- Meta Webhooks API: URL única por app, verify_token + challenge, 5s timeout, retry backoff ~24h
- Meta Business Messaging policy: opt-in consent requerido para procesamiento automatizado
- Epic 5 stories: `_bmad-output/implementation-artifacts/5-2-*` a `5-6-*`
