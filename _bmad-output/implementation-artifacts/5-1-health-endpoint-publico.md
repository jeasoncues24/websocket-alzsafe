---
story_id: "5.1"
epic: "epic-5"
title: "Endpoint público /api/service/v1/health"
status: done
estimated_days: 1
priority: high
branch: "feature/integracion-loyo"
skills: ["golang-security", "bmad-code-review"]
affects:
  - backend/internal/http/handlers/v1_health.go
  - backend/internal/http/handlers/v1_health_test.go
  - backend/internal/http/routes_api.go
  - backend/internal/http/container.go
  - backend/internal/config/config.go
  - docs/integracion-b2b.md
---

# Story 5.1: Endpoint público `/api/service/v1/health`

Status: review

## Contexto

Epic 5 (Integración Loyo) habilita a wsapi como provider B2B para integraciones serverless. Loyalty-app necesita validar que la URL del servicio efectivamente apunta a un wsapi vivo **antes** de pedirle credenciales (API key + secret) al usuario que está configurando la integración. Hoy no existe un endpoint público — todos los endpoints `/api/service/v1/*` requieren auth (`empresaStack` o `clientStack`), por lo que un integrador no puede probar la URL sin antes tener credenciales válidas, lo cual es un huevo-gallina.

Esta story añade un único endpoint no autenticado, mínimo y rate-limited, que devuelve identidad del servicio + versión + timestamp. Es el bloque más barato del epic y desbloquea Loyo Story 1.6 (validar URL en panel).

## User Story

Como integrador B2B que está configurando wsapi como provider de WhatsApp en su panel,
quiero hacer `GET /api/service/v1/health` antes de guardar credenciales,
para validar que la URL apunta efectivamente a un wsapi vivo y obtener su versión.

## Objetivo

Exponer un healthcheck público idempotente y rate-limited que sirva como sonda de identidad del servicio, sin filtrar información sensible (sin DB status, sin counts, sin métricas).

## Scope

### Incluido
- Nuevo handler `V1HealthHandler` en `backend/internal/http/handlers/v1_health.go`.
- Cableo en `routes_api.go` y `container.go`.
- Lectura de versión desde `config.Config` (env `APP_VERSION`, default `dev`), inyectable también vía `-ldflags`.
- Rate limiting in-process por IP (bucket simple, sin librerías nuevas).
- Test de integración con `net/http/httptest` en `v1_health_test.go`.
- Documento `docs/integracion-b2b.md` (sección Health endpoint).

### Excluido
- No tocar otros endpoints `/api/service/v1/*`.
- No introducir librerías nuevas (nada de `golang.org/x/time/rate`, `chi`, etc.).
- No exponer estado de DB, WhatsApp, sesiones ni métricas en la respuesta.
- No incorporar autenticación opcional ni headers especiales para Loyo — el endpoint es genérico para cualquier integrador.
- No mover handlers existentes ni refactorizar `v1_helpers.go`.

## Acceptance Criteria

**AC1 — Endpoint responde 200 sin autenticación:**
**Dado** un request `GET /api/service/v1/health` sin header `Authorization` ni API key,
**Cuando** el servidor está corriendo,
**Entonces** responde `HTTP 200` con `Content-Type: application/json`
**Y** el body es exactamente `{"ok": true, "service": "wsapi", "version": "<APP_VERSION>", "timestamp": "<RFC3339>"}`.

**AC2 — Métodos distintos a GET rechazados:**
**Dado** un request `POST`, `PUT`, `DELETE` o `PATCH` a `/api/service/v1/health`,
**Cuando** llega al handler,
**Entonces** responde `HTTP 405` con body JSON `{"ok": false, "error": "METHOD_NOT_ALLOWED", "message": "Método no permitido"}` (formato consistente con `writeV1Error`).

**AC3 — Versión inyectable y trazable:**
**Dado** que se levanta el binario con `APP_VERSION=v1.2.3` en `.env` o entorno,
**Cuando** se hace `GET /api/service/v1/health`,
**Entonces** el campo `version` del response es exactamente `v1.2.3`
**Y** si no se setea `APP_VERSION`, el valor es `dev`.

**AC4 — Timestamp en RFC3339 UTC:**
**Dado** un request al endpoint,
**Cuando** se serializa la respuesta,
**Entonces** `timestamp` es la hora actual en formato `time.RFC3339` (ej. `2026-05-17T14:23:01Z`).

**AC5 — Rate limit por IP:**
**Dado** que un mismo IP remoto hace más de `HEALTH_RATE_LIMIT_PER_MIN` requests (default `60`) en una ventana móvil de 60s,
**Cuando** se reciben más requests,
**Entonces** responde `HTTP 429` con body `{"ok": false, "error": "RATE_LIMITED", "message": "Demasiados requests"}`
**Y** incluye header `Retry-After: 60`
**Y** la métrica de IP se limpia automáticamente para no crecer indefinidamente (ver Notas de implementación).

**AC6 — No filtra información sensible:**
**Dado** cualquier request al endpoint,
**Cuando** se inspecciona el body de respuesta,
**Entonces** NO aparece ningún dato de: empresas, teléfonos, sesiones WhatsApp, DB host/port/name, paths internos, ni stack traces.

**AC7 — Test de integración pasa:**
**Dado** el archivo `v1_health_test.go`,
**Cuando** se ejecuta `cd backend && go test ./internal/http/handlers/... -run TestV1Health`,
**Entonces** los tests cubren AC1, AC2, AC3 (con `t.Setenv`), AC4 (parseo del timestamp) y AC5 (saturar el bucket) sin errores.

**AC8 — Documentación:**
**Dado** un integrador externo,
**Cuando** consulta `docs/integracion-b2b.md`,
**Entonces** encuentra: URL del endpoint, método, ejemplo de request (`curl`), ejemplo de respuesta, política de rate-limit, y nota explícita de que es público sin auth.

**AC9 — Verificación de build:**
**Dado** los cambios aplicados,
**Cuando** se ejecutan `cd backend && go build ./...` y `cd backend && go test ./...`,
**Entonces** ambos comandos terminan sin errores.

## Tareas técnicas

- [x] **T1 — Versión en config** (AC3)
  - [x] Añadir campo `Version string` a `Config` en `backend/internal/config/config.go`.
  - [x] En `Load()`, asignar `Version: getEnv("APP_VERSION", "dev")`.
  - [x] (Opcional, no bloqueante) documentado en `docs/integracion-b2b.md` — se dejó solo `APP_VERSION` por ser suficiente.

- [x] **T2 — Handler** (AC1, AC2, AC3, AC4, AC6)
  - [x] Crear `backend/internal/http/handlers/v1_health.go`, package `http`.
  - [x] Definir `V1HealthHandler` con `version`, `limiter`, `nowFunc` inyectable.
  - [x] `GetHealth` valida método, aplica rate-limit con `Retry-After` header manual (antes de WriteHeader), devuelve shape plano sin `writeV1Success`.

- [x] **T3 — Rate limiter in-process** (AC5)
  - [x] `healthLimiter` con ventana fija 60s por IP, `sync.Mutex`, limpieza inline cada 256 llamadas.
  - [x] `clientIP` con soporte `X-Forwarded-For` y `net.SplitHostPort`.

- [x] **T4 — Cableo Container** (AC1)
  - [x] Campo `V1HealthHandler` añadido al struct.
  - [x] Instanciado con `cfg.Version` y `HEALTH_RATE_LIMIT_PER_MIN` (default 60).
  - [x] Imports `os`, `strconv` añadidos.

- [x] **T5 — Cableo Routes** (AC1, AC2)
  - [x] Ruta `GET /api/service/v1/health` registrada sin middleware de auth al inicio de `RegisterAPIRoutes`.

- [x] **T6 — Test de integración** (AC7)
  - [x] 7 tests cubriendo AC1–AC6: shape, 405, versión, rate-limit 429, Retry-After header, reset de ventana, y ausencia de campos sensibles.
  - [x] Todos pasan: `go test ./internal/http/handlers/... -run TestV1Health` → 7/7 PASS.

- [x] **T7 — Documentación** (AC8)
  - [x] `docs/integracion-b2b.md` creado con tabla de spec, ejemplos curl, campos, rate limit, nota de liveness probe.

- [x] **T8 — Verificación final** (AC9)
  - [x] `go build ./...` → sin errores.
  - [x] `go test ./...` → todos los paquetes pasan, sin regresiones.

## Archivos afectados

| Archivo | Acción | Notas |
|---------|--------|-------|
| `backend/internal/http/handlers/v1_health.go` | NEW | Handler + limiter inline |
| `backend/internal/http/handlers/v1_health_test.go` | NEW | Tests AC1–AC5 |
| `backend/internal/http/routes_api.go` | UPDATE | Añadir 1 ruta pública |
| `backend/internal/http/container.go` | UPDATE | Añadir campo + instanciar handler |
| `backend/internal/config/config.go` | UPDATE | Añadir `Version string` y env `APP_VERSION` |
| `docs/integracion-b2b.md` | NEW | Doc de la sección Health |

## Pruebas requeridas

- **Unit/integration**: `v1_health_test.go` con `httptest` siguiendo el patrón de `v1_messages_test.go:1-40`.
- **Build**: `cd backend && go build ./...`.
- **Manual smoke**: `curl -i $URL/api/service/v1/health` desde fuera del cluster.
- **Negative**: `curl -i -X POST $URL/api/service/v1/health` → 405.

## Riesgos y edge cases

- **R1 — Leak de memoria del limiter**: si nunca se limpia el mapa `buckets`, crece con cada IP nueva. Mitigación: limpieza inline cada N llamadas o por timestamp; ver T3.
- **R2 — IP detrás de proxy**: `r.RemoteAddr` puede ser siempre la IP del reverse proxy (Caddy/Nginx). Mitigación: leer `X-Forwarded-For` si está presente. **Asumir** que el proxy lo setea correctamente — si no, el rate-limit aplica al proxy en bloque, lo cual es aceptable como primera versión.
- **R3 — Versión vacía**: si `APP_VERSION` está seteado a string vacío explícitamente, `getEnv` devuelve `defaultValue` ("dev"), así que está cubierto. Verificar.
- **R4 — Concurrencia del limiter**: `sync.Mutex` cubre. No usar map sin lock.
- **R5 — Información sensible**: AC6 obliga a no exponer DB/path/etc. Confirmar en el handler que no se logean detalles internos en caso de error (no debería haber errores en este handler tan simple).
- **R6 — `writeV1Error` setea WriteHeader**: para AC5, `Retry-After` debe setearse en `w.Header()` antes de invocar `writeV1Error`. Validar leyendo `v1_helpers.go:13-21`.

## Dependencias y bloqueos

- **Dependencias**: ninguna técnica previa. Es la story más barata del Epic 5.
- **Bloquea a**: Loyo Story 1.6 (validar URL real en panel de Loyo).
- **Branch obligatoria**: `feature/integracion-loyo`. Verificar antes de codear:
  ```bash
  git branch --show-current  # debe ser feature/integracion-loyo
  ```
  Nota: `docs/bmad-project-rules.md` solo lista oficialmente Epic 3 ↔ `feature/security`. Epic 5 fija su rama en el frontmatter del epic file (`epic-integracion-loyo.md`) y en el sprint plan. Si se actualiza `docs/bmad-project-rules.md` con la entrada Epic 5 ↔ `feature/integracion-loyo` en una story de doc separada, mejor; no es bloqueante para esta story.

## Notas de implementación (Dev Notes)

### Patrón de handlers del proyecto

Los handlers de `/api/service/v1/*` siguen este patrón (verificado en `v1_metrics.go`):

1. Archivo en `backend/internal/http/handlers/`, declarado `package http` (sí, el directorio `handlers/` usa el mismo `package http` que el padre — NO confundir con un sub-package).
2. Struct con dependencias inyectadas + constructor `NewXxxHandler(...)`.
3. Métodos con firma `func (h *Xxx) Action(w http.ResponseWriter, r *http.Request)`.
4. Validar método HTTP arriba con `writeV1Error`.
5. Para responses no-empresa, **NO** usar `writeV1Success` (ese helper inyecta `meta.empresa_id`); usar `json.NewEncoder(w).Encode(...)` directamente, seteando `Content-Type` antes.

### Versión: dos rutas posibles

- **Simple (recomendada)**: env `APP_VERSION`, default `dev`. Suficiente para esta story.
- **Avanzada (opcional)**: variable de paquete `var versionLDFlag string` inyectada con `-ldflags "-X 'wsapi/internal/config.versionLDFlag=v1.2.3'"`. Si se setea, prevalece sobre env. Útil para Docker build reproducible. No es obligatoria para AC3.

### Rate limiter: por qué ventana fija y no token bucket

Token bucket es más justo pero pide más estado por IP. La ventana fija de 60s con contador entero es trivial, sin dependencias, suficiente contra abuso accidental. AC5 solo exige rechazar > N en 60s. No optimizar prematuramente.

### Forma de la respuesta — exacta

```json
{
  "ok": true,
  "service": "wsapi",
  "version": "dev",
  "timestamp": "2026-05-17T14:23:01Z"
}
```

Orden de keys no garantizado por `encoding/json`. El cliente Loyo solo lee `ok` y `service` para validar. **No** añadir campos extra (no `db_ok`, no `uptime`, no `whatsapp_status`) — eso filtra detalle interno y rompe el principio de healthcheck público (AC6).

### Anti-patrones a evitar

- ❌ No usar `writeV1Success` (mete `meta.empresa_id`).
- ❌ No leer DB ni manager de WhatsApp para "verificar salud" — esto es un *liveness probe* de identidad, no un *readiness probe* completo. Esa distinción está en la sección de docs.
- ❌ No introducir `chi`, `gorilla/mux`, `golang.org/x/time/rate` ni middleware HTTP de terceros.
- ❌ No autenticar el endpoint (es el propósito de la story).
- ❌ No agregar logging por request en este endpoint — sería ruido (puede ser pulleado cada 30s por monitores).

### Project Structure Notes

- Alineado con `backend/internal/http/handlers/v1_*.go`.
- Sin migraciones de DB → no aplica `/sql-optimization`.
- Sin frontend → no aplica `npm run lint`/`build`.
- Imports: usar `wsapi/internal/...` (módulo `wsapi`).

### References

- Patrón handler V1: `backend/internal/http/handlers/v1_metrics.go:1-60`
- Helpers JSON: `backend/internal/http/handlers/v1_helpers.go:13-35`
- Cableo de rutas: `backend/internal/http/routes_api.go:5-58`
- Container: `backend/internal/http/container.go:51-161`
- Config + env loading: `backend/internal/config/config.go:31-83`
- Pattern de tests httptest: `backend/internal/http/handlers/v1_messages_test.go:1-40`
- Epic origen: `_bmad-output/planning-artifacts/epic-integracion-loyo.md:41-66`
- Regla de rama: `docs/bmad-project-rules.md:5-26` (extender mentalmente a Epic 5)

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6 (bmad-dev-story)

### Debug Log References

- Import `strconv` removido de `v1_health.go` (no se usa en handler; solo en container).

### Completion Notes List

- AC1–AC9 satisfechos.
- `writeV1Success` descartado (inyecta `meta.empresa_id`); respuesta plana con `json.NewEncoder`.
- Rate-limit: `Retry-After` seteado en `w.Header()` antes de `WriteHeader` para evitar el bug de header-after-write.
- `nowFunc` inyectable permite tests deterministas sin sleeps.
- Limpieza de buckets inline cada 256 llamadas → sin goroutine de fondo, sin leaks.
- 7 tests añadidos; suite completa sin regresiones (`go test ./...` → todos OK).

### File List

- `backend/internal/config/config.go` — añadido campo `Version` y env `APP_VERSION`
- `backend/internal/http/handlers/v1_health.go` — nuevo handler + limiter
- `backend/internal/http/handlers/v1_health_test.go` — 7 tests
- `backend/internal/http/container.go` — campo + instancia + imports `os`/`strconv`
- `backend/internal/http/routes_api.go` — ruta `GET /api/service/v1/health`
- `docs/integracion-b2b.md` — documentación del endpoint
