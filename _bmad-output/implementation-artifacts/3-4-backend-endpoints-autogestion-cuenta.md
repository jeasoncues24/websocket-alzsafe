# Story 3-4: Backend — Endpoints de auto-gestión de cuenta

**Estado:** review
**Epic padre:** Epic 3 — Módulos Dinámicos desde BD y Perfil de Usuario
**Story ID:** 3.4

---

## Story

Como **Usuario Autenticado**,
quiero **poder actualizar mi propia información de perfil y mi contraseña a través de endpoints seguros**,
para **poder gestionar mi cuenta personal en el panel de control de forma autónoma**.

---

## Acceptance Criteria

**AC1 — `PUT /api/auth/me` actualiza username y email**
Dado un usuario autenticado vía JWT de administración,
cuando realiza una petición `PUT /api/auth/me` con un JSON conteniendo `username` y `email`,
entonces:
- El sistema valida que los campos tengan formato correcto y no estén vacíos.
- El sistema valida que el nuevo username o email no pertenezcan ya a otro usuario diferente en base de datos (retornando `409 Conflict` si existiese colisión).
- El sistema actualiza el registro únicamente del usuario en sesión (determinado por el ID en los claims del JWT).
- Retorna `200 OK` con un JSON `{ "ok": true }`.

**AC2 — `PUT /api/auth/me/password` realiza el cambio seguro de contraseña**
Dado un usuario autenticado,
cuando realiza una petición `PUT /api/auth/me/password` con un JSON conteniendo `current_password` y `new_password`,
entonces:
- El sistema valida que ambos campos no estén vacíos.
- El sistema recupera el hash actual de la base de datos y utiliza `bcrypt.CompareHashAndPassword` para confirmar que `current_password` sea correcta (retornando `401 Unauthorized` si es incorrecta).
- El sistema genera un nuevo hash usando bcrypt a partir de `new_password` y lo persiste en la base de datos.
- Retorna `200 OK` con un JSON `{ "ok": true }`.

**AC3 — Mitigación de Timing Attacks**
El endpoint de cambio de contraseña debe mitigar ataques de temporización (timing attacks). Esto significa que ante un usuario inexistente (si bien está protegido por auth) o contraseña incorrecta, el proceso computacional debe ejecutar un trabajo criptográfico similar (ej. correr `CompareHashAndPassword` siempre) de modo que el atacante no pueda inferir fallas rápidas de lógica.

**AC4 — Seguridad estricta a nivel de JWT**
El ID del usuario a modificar **nunca** se lee del body ni de parámetros de ruta del request (`/api/auth/me` no tiene ID en la URL). Se lee única y exclusivamente a partir del contexto inyectado por el middleware de autenticación de JWT (`GetAdminUserFromContext` o equivalente).

**AC5 — Tests y Compilación**
`cd backend && go test ./...` y `cd backend && go build ./...` se ejecutan perfectamente sin errores. Se incluyen tests unitarios que cubren:
- Cambio de datos personales exitoso y fallido por colisión de correo.
- Cambio de contraseña exitoso, contraseña actual errónea y contraseñas vacías.

---

## Tasks / Subtasks

- [ ] **T1 — Desarrollar los handlers en `auth.go`**
  - **Archivo:** `backend/internal/http/handlers/auth.go`
  - [ ] Definir los structs de petición: `UpdateMeRequest` (username, email) y `UpdateMePasswordRequest` (current_password, new_password).
  - [ ] Implementar la función de handler `UpdateMe`:
    - Obtener usuario en sesión del contexto.
    - Parsear y validar entrada.
    - Verificar unicidad en BD.
    - Llamar a `h.userStore.Update` o crear un método específico en el store si fuese necesario.
  - [ ] Implementar la función de handler `UpdateMePassword`:
    - Obtener usuario de sesión y su hash de la BD.
    - Validar entrada.
    - Comparar contraseña actual usando bcrypt.
    - Hashear nueva contraseña usando bcrypt.
    - Actualizar contraseña en la BD (`h.userStore.UpdatePassword` o método de store equivalente).

- [ ] **T2 — Registrar las nuevas rutas REST**
  - **Archivos:** `backend/internal/http/routes_admin.go` o `backend/internal/http/routes_api.go` (donde se definan las rutas autenticadas del panel de administración)
  - [ ] Registrar `PUT /api/auth/me` asignando `authHandler.UpdateMe`.
  - [ ] Registrar `PUT /api/auth/me/password` asignando `authHandler.UpdateMePassword`.
  - [ ] Asegurar que ambas rutas estén protegidas bajo el middleware de JWT correspondiente (`AdminJWTAuthMiddleware`).

- [ ] **T3 — Añadir tests unitarios robustos**
  - **Archivo:** `backend/internal/http/handlers/auth_me_self_test.go` (o agregarlos a `auth_modules_test.go` / `auth_test.go`)
  - [ ] Testear peticiones exitosas de actualización de datos de perfil.
  - [ ] Testear error 409 cuando se intenta cambiar email a uno que ya pertenece a otro usuario en BD.
  - [ ] Testear peticiones de contraseña exitosa y fallida por contraseña incorrecta.

---

## Dev Notes

- **Mitigación Timing Attacks:** Si la autenticación falla, corre un hash bcrypt simulado (o una comparación contra un hash ficticio de estructura válida) para asegurar latencias uniformes.
- **Acceso a Stores:** `AuthHandler` ya dispone de `userStore *storage.AdminUserStore` para realizar consultas y escrituras directas sobre la tabla `admin_users`.

---

## Dev Agent Record

### Agent Model Used
Gemini 1.5 Pro (Antigravity Coordinator)

### Debug Log References

### Completion Notes List

### File List
