# Deferred Work

## Deferred from: code review of 2-1-eliminar-select-empresa-usuario-admin (2026-05-05)

- `restoreEmpresa` añadida en `frontend/lib/api.ts:292` sin endpoint backend registrado — pertenece a story 2-2 (restaurar empresas desactivadas), revisar en su code review correspondiente.

## Deferred from: code review of 2-9-enriquecer-metricas-dashboard (2026-05-08)

- **Variable shadowing de `err`** — `backend/internal/http/router.go` — en la rama else de empresa-scoped, `empresa, err :=` introduce un scope local que oculta el `err` externo. Pre-existente, no introducido por esta story.
- **`json.NewEncoder.Encode` error ignorado** — `backend/internal/http/router.go` — patrón pervasivo en el handler: errores de escritura al cliente son descartados silenciosamente. Pre-existente.

## Deferred from: code review of spec-auth-cleanup-b2b — Grupo 4 (2026-05-25)

- **`ConnectCompanyPhoneWS` no llama `CanAccessEmpresa`** — `backend/internal/http/admin.go` L1382-1425 — el único handler WS del admin no verifica que el admin tenga acceso a la empresa del teléfono. Todos los handlers HTTP adyacentes sí verifican. Pre-existente, no introducido por este diff.
- **Sin test de acceso cross-empresa en WS admin** — no hay cobertura para el caso donde un admin intenta abrir WS a un teléfono de otra empresa. Pre-existente.
- **`writeEvent`/`writeWSEvent` duplicados** — `admin.go` y `v1_ws.go` tienen helpers de escritura WS con la misma lógica pero distintas firmas. Deuda técnica pre-existente.
- **Tests ahora solo root** — todas las operaciones admin se testean con `IsRoot: true`. No hay cobertura de admin no-root, cuya diferencia es solo `IsRoot` para operaciones privilegiadas.

## Deferred from: code review of spec-auth-cleanup-b2b — Grupo 3 (2026-05-25)

- **WS handler: `phone == nil` sin close frame WebSocket explícito** — `backend/internal/http/handlers/v1_ws.go` L62-72 — cuando el teléfono no se encuentra tras aceptar la conexión, se envía un evento `error` pero el cierre usa `defer c.CloseNow()` sin código de cierre WS. Cliente no puede distinguir cierre graceful de caída de red. Pre-existente.

## Deferred from: code review of spec-auth-cleanup-b2b — Grupo 2 (2026-05-25)

- **`registeredRoutes` con gaps pre-existentes** — `backend/internal/http/router.go` — el mapa `registeredRoutes` omite varias rutas activas (telefonos CRUD, admin users, webhooks B2B, etc.), causando 404 en preflight OPTIONS/CORS para esas rutas. No introducido por este diff.

## Deferred from: code review of spec-auth-cleanup-b2b — Grupo 1 (2026-05-25)

- **float64→int64 truncación silenciosa para IDs > 2^53** — `backend/internal/auth/qr_link_jwt.go` ~L50 — JSON decode de MapClaims usa float64; IDs > 9×10^15 pierden precisión. Pre-existente en el patrón JWT del proyecto.
- **Token WS en query param aparece en logs del servidor** — `GET /api/service/v1/ws?token=...` — el token QR-link llega por URL, queda en access logs. Diseño pre-existente del WS handler.
- **Sin claim `nbf`, ventana de replay completa de 10 minutos** — `backend/internal/auth/qr_link_jwt.go` — un token interceptado puede reutilizarse hasta su expiración sin restricción.

## Deferred from: code review of 2-6-fix-ws-timer-simplificar-ui-conexion (2026-05-08)

- **Token JWT en query param URL** — `backend/internal/http/admin.go` — el token ?token= aparece en access logs, browser history y referrer. Pre-existente, mejorar en hardening de seguridad posterior.
- **Claims del JWT no validados** — `backend/internal/http/admin.go` — después de ValidateToken, `_ = claims` descarta el resultado; rol y user_id no son chequeados. Pre-existente.
- **InsecureSkipVerify:true en WS upgrade** — `backend/internal/http/admin.go` — deshabilita chequeo de origen, permite CSWSH desde cualquier origen. Pre-existente.
- **StartSession fallo parcial y s.starting** — `backend/internal/whatsapp/service.go` — si openSQLiteContainer falla después de SetInitializing, el sessionStore queda en "initializing" sin limpiarse. Pre-existente en service.go, no introducido por esta story.
