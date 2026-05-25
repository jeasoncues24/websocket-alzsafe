# Limpieza de Auth — B2B solo API Key, eliminar EmpresaJWT

El objetivo de esta historia es purgar por completo el concepto obsoleto de `EmpresaJWT` del sistema `wsapi`, ya que el verdadero mecanismo de autenticación para integraciones B2B es mediante `ApiKey` vinculada a un teléfono.

Actualmente, existe un `empresaStack` y middleware que inyecta claims sintéticos de `EmpresaJWT`, lo que genera deuda técnica y ambigüedad de identidad. Además, la funcionalidad de generar un token temporal para el enlace QR de WhatsApp (QR-link) está acoplada al `EmpresaJWT`, por lo que debe extraerse antes de eliminarlo.

## User Review Required

> [!WARNING]
> **Archivos de Tests Omitidos en el Draft Original**
> Durante el análisis del código, encontré que los archivos `admin_test.go` y `companies_test.go` también contienen referencias a `EmpresaJWT` (probablemente testeando las rutas admin de generación de tokens de empresa que vamos a eliminar). Estos tests deberán ser eliminados/modificados para que el build pase con éxito. He añadido esto a la sección de archivos a modificar.

> [!IMPORTANT]
> **Compatibilidad Frontend**
> El frontend (`frontend/lib/api.ts`) consume `GET /api/service/v1/ws` usando el token QR-link. Se garantiza que este endpoint continuará aceptando un JWT firmado con el mismo `secret`, solo que ahora sus claims serán validados estrictamente bajo la estructura `QRLinkClaims`.

## Open Questions

- ¿Deberíamos agregar cobertura de tests unitarios para las nuevas funciones `GenerateQRLinkToken` y `ParseQRLinkToken` en el paquete `auth`, o basta con los tests de integración existentes que levantan el websocket? (Por defecto añadiré tests unitarios para `qr_link_jwt.go`).

## Proposed Changes

### Componente de Autenticación (Tokens Provisionales)

Extracción de la lógica del token del QR-link hacia sus propias estructuras independientes.

#### [NEW] [domain/qr_link_claims.go](file:///home/fulanito/development/wsapi/backend/internal/domain/qr_link_claims.go)
- Crear el struct `QRLinkClaims` que representa los claims del token provisional de QR link (10 min).
- Contiene `EmpresaID`, `PhoneID` y `Scope` (siempre "qr_link").
- Funciones `WithQRLinkClaims` y `GetQRLinkClaims` para el contexto (opcional, pero buena práctica).

#### [NEW] [auth/qr_link_jwt.go](file:///home/fulanito/development/wsapi/backend/internal/auth/qr_link_jwt.go)
- Extraer `GenerateQRLinkToken` y crear `ParseQRLinkToken` validando estrictamente que el scope sea "qr_link" y contenga `phone_id`.
- Mismo `secret` y mecanismo HS256, pero sin relación con `EmpresaJWTClaims`.

---

### Eliminación de Código Muerto y Obsoleto (EmpresaStack)

Se procede a borrar de raíz los handlers y middlewares que dependían de la autenticación por token de larga duración B2B.

#### [DELETE] `backend/internal/http/handlers.go`
- Contiene un struct `Handler` muerto que referencia `EmpresaJWTClaims`.

#### [DELETE] `backend/internal/http/v1_handler.go`
- Contiene un struct `V1Handler` muerto que referencia `EmpresaJWTClaims`.

#### [DELETE] `backend/internal/http/handlers/v1_sessions.go`
- Endpoint obsoleto. Depende de `GetEmpresaJWTClaims`.

#### [DELETE] `backend/internal/http/handlers/v1_phones.go`
- Endpoint obsoleto. Depende de `GetEmpresaJWTClaims`.

#### [DELETE] `backend/internal/http/handlers/v1_metrics.go`
- Endpoint obsoleto. Depende de `GetEmpresaJWTClaims`.

#### [DELETE] `backend/internal/http/middleware/empresa_auth.go`
- Middleware `EmpresaAuthMiddleware`. Ya no existe este stack.

#### [DELETE] `backend/internal/auth/empresa_jwt.go`
- Una vez extraído el QR-Link al paso anterior, este archivo desaparece.

#### [DELETE] `backend/internal/domain/empresa_token.go`
- Define los claims y helpers de contexto.

---

### Refactorización del Router, Kernel y Handlers Activos

Limpieza de las referencias restantes en la infraestructura HTTP y simplificación de middlewares.

#### [MODIFY] [kernel.go](file:///home/fulanito/development/wsapi/backend/internal/http/kernel.go)
- Eliminar la inyección de `EmpresaAuth` en la construcción del Kernel y borrar la interfaz `EmpresaAuthProvider`.

#### [MODIFY] [router.go](file:///home/fulanito/development/wsapi/backend/internal/http/router.go)
- Eliminar la inyección de `c.EmpresaAuthMiddleware` al llamar a `NewKernel`.

#### [MODIFY] [routes_api.go](file:///home/fulanito/development/wsapi/backend/internal/http/routes_api.go)
- Borrar por completo el bloque de asignación de `empresaStack` y todas las rutas asociadas (`/api/service/v1/empresas`, `/telefonos`, `/sesiones`, etc).

#### [MODIFY] [routes_admin.go](file:///home/fulanito/development/wsapi/backend/internal/http/routes_admin.go)
- Eliminar las rutas de admin para generar y revocar tokens JWT de empresas: `POST /api/admin/empresas/{id}/token` y `POST /api/admin/empresas/{id}/token/revoke`.

#### [MODIFY] [middleware/api_key_auth.go](file:///home/fulanito/development/wsapi/backend/internal/http/middleware/api_key_auth.go)
- **Crítico:** Eliminar la inyección engañosa de `EmpresaJWTClaims` y `EmpresaID` en el contexto. Ahora la API Key solo inyecta `ApiKeyClaims`.

#### [MODIFY] [handlers/companies.go](file:///home/fulanito/development/wsapi/backend/internal/http/handlers/companies.go)
- Eliminar métodos que pertenecían a rutas eliminadas: `GetCurrent`, `UpdateCurrent`, `GenerateToken`, `RevokeToken`.

#### [MODIFY] [handlers/v1_webhooks.go](file:///home/fulanito/development/wsapi/backend/internal/http/handlers/v1_webhooks.go)
- Eliminar el método `ListByEmpresa`.

#### [MODIFY] [handlers/v1_helpers.go](file:///home/fulanito/development/wsapi/backend/internal/http/handlers/v1_helpers.go)
- Eliminar funciones de utilidad que parsean el token eliminado: `getEmpresaIDFromContext` y `getAccessClaims`.

#### [MODIFY] [domain/panel_access.go](file:///home/fulanito/development/wsapi/backend/internal/domain/panel_access.go)
- Quitar el fallback hacia `EmpresaJWTClaims` en la resolución de permisos y simplificar el struct `PanelAccess`.

#### [MODIFY] [container.go](file:///home/fulanito/development/wsapi/backend/internal/http/container.go)
- Remover `EmpresaAuthMiddleware` y los Handlers eliminados (`V1SessionsHandler`, `V1PhonesHandler`, `V1MetricsHandler`) de la inyección de dependencias.

#### [MODIFY] [handlers/admin_sessions.go](file:///home/fulanito/development/wsapi/backend/internal/http/handlers/admin_sessions.go)
- Actualizar import para usar la nueva ruta de generación de token del QR-link si es necesario (ambos en el paquete `auth`, pero verificar compilación).

---

### Actualización del Endpoint WebSockets (B2B)

#### [MODIFY] [handlers/v1_ws.go](file:///home/fulanito/development/wsapi/backend/internal/http/handlers/v1_ws.go)
- Reemplazar el handler completo para que **solo** acepte `ParseQRLinkToken`.
- Eliminar el flujo de suscripción JSON (`c.Read(ctx)`) que permitía a los tokens de larga duración de empresa conectarse. El token QR tiene directamente el `phone_id`.
- Remover el uso de `telefonoStore` que ya no es necesario para validar la pertenencia del teléfono a la empresa (eso ya lo garantiza el token QR firmado por nosotros mismos).

---

### Actualización de Tests y Documentación

#### [MODIFY] [admin_test.go](file:///home/fulanito/development/wsapi/backend/internal/http/admin_test.go) y [companies_test.go](file:///home/fulanito/development/wsapi/backend/internal/http/handlers/companies_test.go)
- Eliminar o adaptar los unit tests que verificaban la generación, revocación y uso del token JWT de empresa.

#### [MODIFY] `docs/routes/contrato-b2b/`
- Eliminar archivos documentando endpoints muertos (`sesiones.md`, `telefonos.md`, `metricas.md`, `empresa.md`).
- Actualizar `README.md` indicando que toda autenticación B2B es por encabezado `X-API-Key`.

## Verification Plan

### Automated Tests
- `cd backend && go build ./...` (Debe compilar sin errores de dependencias cíclicas ni variables no declaradas).
- `cd backend && go test ./...` (Todos los tests, habiendo modificado los de `admin` y `companies`, deben pasar en verde).
- Ejecución de comandos `grep` asegurando la completa extinción de referencias a `EmpresaJWTClaims` y `empresaStack`.

### Manual Verification
- Levantar el entorno de desarrollo y entrar al panel web `/qr`. Validar que el token temporal generado mediante la API de admin sirva para conectar existosamente el WebSocket B2B.
