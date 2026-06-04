# Deferred Work

## Deferred from: code review of spec-v1-ws-connect-apikey (2026-06-03)

- **Exposición a Cross-Site WebSocket Hijacking (CSWSH) por `InsecureSkipVerify: true`** — `backend/internal/http/handlers/v1_ws.go:159` — El handshake WebSocket se realiza configurando `InsecureSkipVerify: true` en `AcceptOptions`, omitiendo por completo la verificación de `Origin` y permitiendo conexiones maliciosas de otros sitios web en el navegador del usuario. Preexistente en el handler existente `HandleWS`.
- **Falta de bucle de lectura (`Read`) en la conexión WebSocket** — `backend/internal/http/handlers/v1_ws.go:189-213` — El manejador abre la conexión pero nunca lee de ella. Para detectar desconexiones inmediatas del cliente y procesar tramas de control, es necesario tener un loop que llame a `c.Read(ctx)`. Preexistente en el handler existente `HandleWS`.
- **Cierre abrupto de la conexión WebSocket con `CloseNow()`** — `backend/internal/http/handlers/v1_ws.go:167` — El manejador ejecuta `defer c.CloseNow()`, cerrando la conexión TCP abruptamente sin enviar una trama de cierre grácil. Preexistente en el handler existente `HandleWS`.
- **Registro de PII (número de teléfono) en los logs en texto plano** — `backend/internal/http/handlers/v1_ws.go:171` — Se imprime en consola `fmt.Printf("[INFO] V1 WS connect opened phone=%d account=%s empresa=%d\n", ...)` expone el número de teléfono del usuario en texto plano en la consola. Preexistente en `HandleWS`.
- **Ausencia de límite de tiempo de escritura (Write Timeout) al enviar eventos** — `backend/internal/http/handlers/v1_ws.go:193` — `writeWSEvent` no tiene un context con límite de tiempo de escritura, lo que podría bloquear indefinidamente a la goroutine de conexión. Preexistente en `HandleWS`.
- **Falta de propagación del contexto de la petición en `StartSession`** — `backend/internal/http/handlers/v1_ws.go:173` — `whatsapp.StartSession` no recibe un context y por tanto no se detiene si la petición se cancela. Comportamiento preexistente.
- **Fuga de detalles de errores internos de infraestructura en `StartSession`** — `backend/internal/http/handlers/v1_ws.go:175` — Si falla la inicialización de la sesión de WhatsApp, se envía el error crudo `err.Error()` directamente al cliente. Preexistente en `HandleWS`.
- **Falta de rate limiting y control de conexiones WebSocket concurrentes** — `backend/internal/http/handlers/v1_ws.go:142` — El handler permite conexiones WebSocket ilimitadas para una misma API Key, lo que expone el servidor a agotamiento de recursos.

## Deferred from: code review of spec-difusiones-anti-ban-job-queue (2026-05-25)

- **Omisión del WebSocket administrativo `/admin/ws/difusiones/{reference_id}` para progreso en tiempo real** — Se difiere debido a que el mecanismo de *polling* (sondeo de 3 segundos) ya está implementado y es plenamente funcional como *fallback* en el panel administrativo del frontend, de acuerdo con las notas de integración del *spec*.


- `restoreEmpresa` añadida en `frontend/lib/api.ts:292` sin endpoint backend registrado — pertenece a story 2-2 (restaurar empresas desactivadas), revisar en su code review correspondiente.

## Deferred from: code review of 2-9-enriquecer-metricas-dashboard (2026-05-08)

- **Variable shadowing de `err`** — `backend/internal/http/router.go` — en la rama else de empresa-scoped, `empresa, err :=` introduce un scope local que oculta el `err` externo. Pre-existente, no introducido por esta story.
- **`json.NewEncoder.Encode` error ignorado** — `backend/internal/http/router.go` — patrón pervasivo en el handler: errores de escritura al cliente son descartados silenciosamente. Pre-existente.

## Deferred from: code review of spec-auth-cleanup-b2b — Grupo 4 (2026-05-25)

- **✅ Resuelto:** `ConnectCompanyPhoneWS` no llama `CanAccessEmpresa` — `backend/internal/http/admin.go` L1382-1425 — el único handler WS del admin no verifica que el admin tenga acceso a la empresa del teléfono. Todos los handlers HTTP adyacentes sí verifican. Pre-existente, no introducido por este diff.
- **✅ Resuelto:** Sin test de acceso cross-empresa en WS admin — no hay cobertura para el caso donde un admin intenta abrir WS a un teléfono de otra empresa. Pre-existente.
- **✅ Resuelto:** `writeEvent`/`writeWSEvent` duplicados — `admin.go` y `v1_ws.go` tienen helpers de escritura WS con la misma lógica pero distintas firmas. Deuda técnica pre-existente.
- **✅ Resuelto:** Tests ahora solo root — todas las operaciones admin se testean con `IsRoot: true`. No hay cobertura de admin no-root, cuya diferencia es solo `IsRoot` para operaciones privilegiadas.

## Deferred from: code review of spec-auth-cleanup-b2b — Grupo 3 (2026-05-25)

- **✅ Resuelto:** WS handler: `phone == nil` sin close frame WebSocket explícito — `backend/internal/http/handlers/v1_ws.go` L62-72 — cuando el teléfono no se encuentra tras aceptar la conexión, se envía un evento `error` pero el cierre usa `defer c.CloseNow()` sin código de cierre WS. Cliente no puede distinguir cierre graceful de caída de red. Pre-existente.

## Deferred from: code review of spec-auth-cleanup-b2b — Grupo 2 (2026-05-25)

- **✅ Resuelto:** `registeredRoutes` con gaps pre-existentes — `backend/internal/http/router.go` — el mapa `registeredRoutes` omite varias rutas activas (telefonos CRUD, admin users, webhooks B2B, etc.), causando 404 en preflight OPTIONS/CORS para esas rutas. No introducido por este diff.

## Deferred from: code review of spec-auth-cleanup-b2b — Grupo 1 (2026-05-25)

- **✅ Resuelto:** float64→int64 truncación silenciosa para IDs > 2^53 — `backend/internal/auth/qr_link_jwt.go` ~L50 — JSON decode de MapClaims usa float64; IDs > 9×10^15 pierden precisión. Pre-existente en el patrón JWT del proyecto.
- **✅ Resuelto:** Token WS en query param aparece en logs del servidor — `GET /api/service/v1/ws?token=...` — el token QR-link llega por URL, queda en access logs. Diseño pre-existente del WS handler.
- **✅ Resuelto:** Sin claim `nbf`, ventana de replay completa de 10 minutos — `backend/internal/auth/qr_link_jwt.go` — un token interceptado puede reutilizarse hasta su expiración sin restricción.

## Deferred from: code review of 2-6-fix-ws-timer-simplificar-ui-conexion (2026-05-08)

- **Token JWT en query param URL** — `backend/internal/http/admin.go` — el token ?token= aparece en access logs, browser history y referrer. Pre-existente, mejorar en hardening de seguridad posterior.
- **Claims del JWT no validados** — `backend/internal/http/admin.go` — después de ValidateToken, `_ = claims` descarta el resultado; rol y user_id no son chequeados. Pre-existente.
- **InsecureSkipVerify:true en WS upgrade** — `backend/internal/http/admin.go` — deshabilita chequeo de origen, permite CSWSH desde cualquier origen. Pre-existente.
- **StartSession fallo parcial y s.starting** — `backend/internal/whatsapp/service.go` — si openSQLiteContainer falla después de SetInitializing, el sessionStore queda en "initializing" sin limpiarse. Pre-existente en service.go, no introducido por esta story.
