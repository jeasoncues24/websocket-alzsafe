# Epic 4: Seguridad Avanzada del Panel — 2FA, Verificación de Email y Auditoría de Sesiones

Status: backlog

## Objetivo

Elevar el nivel de seguridad del sistema de autenticación del panel administrativo incorporando verificación de email, autenticación de dos factores (TOTP), notificaciones de acceso y auditoría persistente de sesiones. Toda la configuración de seguridad quedará accesible desde la sección "Mi cuenta" (story 3-5 del Epic 3), que debe estar completa antes de iniciar el frontend de este epic.

---

## Evaluación actual

### Resumen del estado de auth en el backend Go

El sistema de autenticación actual (`backend/internal/http/handlers/auth.go`) implementa:
- Login con bcrypt para verificación de contraseña (`golang.org/x/crypto/bcrypt`)
- JWT firmado con HMAC-SHA256 (`github.com/golang-jwt/jwt/v5`)
- Blacklist de tokens invalidados en memoria (`storage.TokenBlacklistStore`)
- Refresh de token invalidando el anterior vía blacklist
- Registro de `last_login_at` en `admin_users` tras login exitoso
- Generación de JTI con `crypto/rand` (correcto — no usa `math/rand`)

### Gaps de seguridad identificados

1. **No existe auditoría persistente de sesiones.** `last_login_at` solo guarda el último acceso; no hay historial, ni IP, ni user-agent registrado. Un compromiso de cuenta no deja rastro consultable.

2. **No existe verificación de email.** El campo `email` en `admin_users` no tiene mecanismo de confirmación. Un usuario con email incorrecto no puede recibir notificaciones ni verificaciones de seguridad.

3. **No hay notificación de nuevo login.** No se alerta al usuario cuando se abre una sesión desde una IP o dispositivo nuevo. Un acceso no autorizado puede pasar desapercibido.

4. **No existe segundo factor de autenticación.** El login solo requiere username + password. Un credential stuffing o password leak compromete la cuenta inmediatamente.

5. **No hay opciones de seguridad por usuario.** Timeout de sesión, notificaciones y 2FA son globales o inexistentes; el usuario no puede configurarlos desde su perfil.

6. **La blacklist de tokens vive en memoria.** Un reinicio del proceso invalida la blacklist. Tokens que debieron estar invalidados quedan activos tras restart. *(Riesgo Medium — documentado para evaluar persistencia futura.)*

---

## Contexto técnico Go para implementación

### Patrones de seguridad requeridos (golang-security skill)

**Generación de tokens seguros** — siempre `crypto/rand`, nunca `math/rand`:
```go
b := make([]byte, 32)
if _, err := rand.Read(b); err != nil {
    return "", fmt.Errorf("generar token: %w", err)
}
token := hex.EncodeToString(b)
```

**TOTP (RFC 6238)** — librería recomendada: `github.com/pquerna/otp/totp`
```go
key, err := totp.Generate(totp.GenerateOpts{
    Issuer:      "wsapi",
    AccountName: user.Username,
    SecretSize:  20,
    Algorithm:   otp.AlgorithmSHA1, // RFC 6238 estándar
})
// Secret se almacena cifrado en BD (AES-GCM con clave de env)
// NO en texto plano, NO solo base64
```

**Comparación segura de tokens** — evitar timing attacks:
```go
// CORRECTO: constant-time
if !hmac.Equal([]byte(provided), []byte(stored)) { ... }
// INCORRECTO: == compara byte a byte y fuga timing
```

**Secretos TOTP en reposo** — cifrar el `totp_secret` antes de guardar en BD:
```go
// AES-256-GCM con clave derivada de APP_SECRET via HKDF
// La clave nunca toca la BD, solo el ciphertext
```

**Backup codes** — hashear con bcrypt (no reversibles, igual que passwords):
```go
// Generar 8 códigos de 8 dígitos; hashear cada uno
// Al verificar: bcrypt.CompareHashAndPassword para cada code
```

**Emails de verificación** — tokens de un solo uso con expiración corta:
```go
// Token de 32 bytes hex, expiración 24h
// Invalidar inmediatamente tras uso (marcar `used_at`)
// Rate limit: máximo 3 envíos por hora por usuario
```

### Modificación del flujo de login para 2FA

El login actual retorna el JWT directamente. Con 2FA debe ser en dos pasos:

```
Paso 1: POST /api/auth/login
  - Valida username + password (igual que ahora)
  - Si usuario tiene 2FA habilitado → retorna { requires_2fa: true, session_token: "<temp>" }
  - Si no tiene 2FA → retorna JWT (comportamiento actual)

Paso 2: POST /api/auth/2fa/verify
  - Recibe session_token + totp_code
  - Valida TOTP o backup code
  - Si válido → retorna JWT definitivo
  - session_token: JWT temporal de corta vida (5 min), scope restringido
```

### Esquema de BD propuesto

```sql
-- Auditoría de sesiones (indexar por user_id + created_at)
CREATE TABLE session_audit (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id     BIGINT NOT NULL,
    action      VARCHAR(20) NOT NULL,  -- 'login', 'logout', 'refresh', '2fa_success', '2fa_failed'
    ip_address  VARCHAR(45),           -- IPv6 compatible
    user_agent  VARCHAR(512),
    created_at  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    FOREIGN KEY (user_id) REFERENCES admin_users(id) ON DELETE CASCADE,
    INDEX idx_session_audit_user_created (user_id, created_at)
);

-- Tokens de verificación de email (un solo uso)
CREATE TABLE email_verification_tokens (
    id         BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id    BIGINT NOT NULL UNIQUE,  -- un token pendiente por usuario
    token_hash VARCHAR(64) NOT NULL,    -- SHA-256 del token real
    expires_at DATETIME NOT NULL,
    used_at    DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES admin_users(id) ON DELETE CASCADE,
    INDEX idx_evt_token_hash (token_hash)
);

-- Campos adicionales en admin_users (ALTER TABLE):
-- totp_secret_enc   VARCHAR(512) NULL   -- AES-GCM ciphertext del secret TOTP
-- totp_enabled      BOOLEAN NOT NULL DEFAULT FALSE
-- totp_enabled_at   DATETIME NULL
-- backup_codes_json TEXT NULL           -- JSON array de bcrypt hashes
-- email_verified_at DATETIME NULL
-- notify_on_login   BOOLEAN NOT NULL DEFAULT FALSE
-- session_timeout_m INT NOT NULL DEFAULT 480  -- minutos, default 8h
```

> **Nota para sql-optimization:** Los índices de `session_audit` deben evaluarse con EXPLAIN para queries de historial paginado. Considerar particionamiento por rango de fecha si el volumen crece. La columna `token_hash` en `email_verification_tokens` es VARCHAR(64) para SHA-256 hex; evaluar si BINARY(32) es más eficiente.

---

## Alcance incluido

- Tabla `session_audit` con registro automático en login, logout y refresh.
- Endpoint `GET /api/auth/me/sessions` para consultar historial desde el frontend.
- Verificación de email: generación de token seguro, envío por SMTP configurable, confirmación de un solo uso.
- Notificación de nuevo login por email (requiere email verificado).
- TOTP 2FA completo: generación de secret cifrado, QR code, verificación, desactivación con contraseña, backup codes hasheados.
- Flujo de login en dos pasos cuando 2FA está activo.
- Tabla de opciones de seguridad por usuario: notify_on_login, session_timeout configurable.
- Frontend: sección "Seguridad" integrada en Settings > Mi cuenta (extiende story 3-5).

## Fuera de alcance

- SMS / push como segundo factor (solo TOTP en este epic).
- WebAuthn / passkeys.
- Detección de anomalías o geolocalización de IPs.
- Persistencia de la token blacklist en BD (evaluación posterior).
- Sistema de roles de seguridad (ya cubierto por Epic 3).

## Dependencias

- **Epic 3 completo** antes de iniciar story 7-6 (frontend requiere que "Mi cuenta" exista).
- Backend stories 7-1 a 7-5 pueden desarrollarse en paralelo con Epic 3 backend.
- Configuración SMTP: requiere variables de entorno `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM`.
- Clave de cifrado TOTP: variable de entorno `APP_ENCRYPTION_KEY` (32 bytes en hex).
- Librería Go: `github.com/pquerna/otp` para TOTP RFC 6238.

## Riesgos conocidos

- **Brute force en verificación TOTP:** el endpoint de verificación de código debe tener rate limit estricto (5 intentos / 10 min por usuario) para evitar ataques de fuerza bruta sobre los 6 dígitos.
- **Replay de TOTP:** el código debe marcarse como usado en una ventana de tiempo para prevenir replay en la misma ventana de 30s. `pquerna/otp` maneja esto con drift de ±1 ventana.
- **Pérdida de acceso sin backup codes:** el flujo de setup debe forzar que el usuario guarde los backup codes antes de activar 2FA. Sin ellos, un teléfono perdido bloquea la cuenta.
- **Email no verificado bloqueando funciones:** la verificación de email debe ser opcional en primera fase; no bloquear el acceso si el email no está verificado, solo limitar las notificaciones.

## Criterios de éxito

- Cada login y logout queda registrado en `session_audit` con IP y user-agent.
- Un usuario puede verificar su email desde "Mi cuenta" y recibir notificaciones de login.
- Un usuario puede activar 2FA desde "Mi cuenta"; el siguiente login le solicita el código TOTP.
- Los backup codes funcionan como fallback cuando no hay acceso al TOTP.
- `cd backend && go test ./...` y `go build ./...` pasan sin errores.
- `cd frontend && npm run lint` y `npm run build` pasan sin errores.
- `go tool gosec ./...` no reporta issues nuevos en el código de auth.

## Skills por story

Invocar únicamente las skills indicadas al crear o implementar cada story. No aplicar todas a todas.

| Story | Skills a invocar | Motivo |
|---|---|---|
| 7-1 | `golang-security`, `sql-optimization` | Nueva tabla `session_audit`: diseño de índices, query de historial paginado; registro en cada request de login/logout con IP y user-agent |
| 7-2 | `golang-security`, `better-auth-best-practices` | Tokens de un solo uso con expiración, servicio SMTP, comparación constant-time; flujo de verificación de email sensible a replay y brute force |
| 7-3 | `golang-security` | Trigger de envío de email en login; evitar information disclosure sobre si el email existe; rate limit de notificaciones |
| 7-4 | `golang-security`, `better-auth-best-practices`, `two-factor-authentication-best-practices` | Story más crítica del epic: cifrado de secret TOTP en reposo (AES-GCM), replay prevention, backup codes con bcrypt, modificación del flujo de login en dos pasos |
| 7-5 | `golang-security`, `sql-optimization` | ALTER TABLE sobre `admin_users` con nuevos campos; evaluar índices si se consultan por filtros; validación de rangos para session_timeout |
| 7-6 | `bmad-agent-ux-designer`, `ui-ux-pro-max` | Pantalla de seguridad compleja: QR code setup para 2FA, lista de sesiones, toggle de notificaciones, formulario de verificación de email — requiere diseño UX cuidadoso para no confundir al usuario |

## Stories propuestas

| ID  | Nombre | Tipo | Prioridad | Estado |
|-----|--------|------|-----------|--------|
| 7-1 | Backend: auditoría de sesiones (tabla + registro automático) | Backend | Alta | backlog |
| 7-2 | Backend: verificación de email y servicio SMTP | Backend | Alta | backlog |
| 7-3 | Backend: notificación de nuevo login por email | Backend | Media | backlog |
| 7-4 | Backend: TOTP 2FA — setup, verificación y flujo de login | Backend | Alta | backlog |
| 7-5 | Backend: opciones de seguridad por usuario | Backend | Media | backlog |
| 7-6 | Frontend: sección "Seguridad" en Mi cuenta | Frontend | Alta | backlog |

## Orden recomendado de implementación

```text
7-1 → independiente, prioritario (auditoría sin dependencias externas)
7-2 → independiente (SMTP + verificación, prerrequisito para 7-3)
7-3 → depende de 7-2 (necesita email verificado para notificar)
7-4 → independiente del grupo anterior (puede correr en paralelo con 7-1/7-2)
7-5 → puede desarrollarse junto con 7-4 (comparten la tabla admin_users)
7-6 → depende de 7-1..7-5 completados + Epic 3 story 3-5 completada
```

## Condición de cierre del epic

El epic se cierra cuando: toda sesión queda auditada, el usuario puede verificar su email, activar 2FA desde "Mi cuenta", y configurar sus preferencias de seguridad de sesión desde el panel, con tests pasando y gosec limpio.
