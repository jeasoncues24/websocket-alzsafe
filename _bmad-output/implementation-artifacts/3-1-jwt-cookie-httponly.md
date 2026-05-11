---
story_id: "3.1"
epic: "epic-3"
title: "Migrar JWT de localStorage a httpOnly Cookie"
status: backlog
estimated_days: 2
priority: critical
skills: ["better-auth-best-practices", "golang-security", "security-review"]
affects:
  - frontend/lib/api.ts
  - frontend/app/login/page.tsx
  - backend/internal/http/handlers.go (o admin.go — handler de login)
  - backend/internal/http/middleware.go
---

# Story 3.1: Migrar JWT de localStorage a httpOnly Cookie

## Contexto

**Vulnerabilidad confirmada** en `frontend/lib/api.ts`:

```typescript
// INSEGURO — actual en authHeaders() ~línea 180:
const token = typeof window !== "undefined" ? localStorage.getItem("admin_token") : null;

// INSEGURO — actual en fetchWithAuth():
const token = localStorage.getItem("admin_token");

// INSEGURO — actual en generateQRLink():
const token = localStorage.getItem("admin_token");
```

Un script XSS en cualquier parte del panel puede ejecutar `localStorage.getItem("admin_token")` y robar el JWT del administrador. La solución es migrar a `httpOnly cookie` — los scripts no pueden acceder a cookies httpOnly.

## User Story

Como administrador del sistema,
quiero que mi sesión admin esté protegida por una cookie httpOnly en lugar de localStorage,
para que un eventual script XSS en el panel no pueda robar el token de sesión.

## Scope

### Incluido
- Backend Go: agregar `Set-Cookie` httpOnly en respuesta de `/api/admin/login`
- Backend Go: middleware de auth lee token desde cookie como fallback cuando `Authorization` header está ausente
- Frontend: eliminar `localStorage.getItem("admin_token")` de `authHeaders()`, `fetchWithAuth()`, `generateQRLink()`
- Frontend: agregar `credentials: 'include'` a todas las llamadas fetch para que el browser envíe la cookie
- Frontend: en login exitoso, no guardar token en localStorage
- Frontend: en logout, limpiar la cookie (llamar endpoint de logout o expirar cookie)

### Excluido
- Cambios a los endpoints de negocio (empresas, teléfonos, API keys)
- Cambios al esquema JWT (payload, expiración, algoritmo)
- Implementar refresh tokens (puede ser epic futuro)

## Acceptance Criteria

**AC1 — Backend emite cookie en login:**
**Dado** que el handler Go de `/api/admin/login` procesa credenciales válidas
**Cuando** retorna la respuesta HTTP 200
**Entonces** incluye el header: `Set-Cookie: admin_token=<jwt>; HttpOnly; Path=/; SameSite=Strict`
**Y** en producción (variable de entorno `APP_ENV=production`), agrega el flag `Secure`
**Y** el body JSON puede mantener el token para compatibilidad transitoria (o retirarlo si no hay otros consumidores)

**AC2 — Backend middleware lee cookie:**
**Dado** que llega un request a cualquier ruta admin protegida sin header `Authorization`
**Cuando** el middleware de auth busca el token
**Entonces** lee la cookie `admin_token` y la valida
**Y** si la cookie es válida, el request continúa autenticado
**Y** si no hay ni header ni cookie, retorna 401

**AC3 — Frontend deja de usar localStorage:**
**Dado** que `frontend/lib/api.ts` contiene las funciones `authHeaders()`, `fetchWithAuth()`, `generateQRLink()`
**Cuando** se actualiza el archivo
**Entonces** ninguna de las tres funciones llama `localStorage.getItem("admin_token")`
**Y** todas las llamadas `fetch()` en el archivo incluyen `credentials: 'include'`

**AC4 — Login no guarda en localStorage:**
**Dado** que `app/login/page.tsx` maneja el submit del formulario de login
**Cuando** el login es exitoso (response 200)
**Entonces** no ejecuta `localStorage.setItem("admin_token", ...)`
**Y** la sesión se mantiene a través de recargas de página gracias a la cookie

**AC5 — Build y lint pasan:**
**Dado** que todos los cambios están aplicados
**Cuando** se ejecuta `cd frontend && npm run lint && npm run build`
**Entonces** no hay errores de TypeScript ni de ESLint
**Y** `cd backend && go build ./...` también pasa sin errores

## Notas de Implementación

- El middleware Go debe verificar primero el header `Authorization: Bearer` (para compatibilidad con herramientas CLI/API), y solo si está ausente, buscar la cookie.
- Para dev local sin HTTPS, omitir el flag `Secure` en la cookie (condicionarlo a `APP_ENV`).
- La función `buildAdminWsUrl` en `api.ts` puede quedar simplificada: ya no necesita el parámetro `token` si el WS también recibe la cookie (ver Story 3.2).
- Usar la skill `better-auth-best-practices` para validar el enfoque de cookie httpOnly.
- Usar la skill `golang-security` para revisar el manejo seguro de cookies en Go.

## Verificación Manual

1. Abrir DevTools → Application → Local Storage → verificar que NO hay `admin_token`
2. DevTools → Application → Cookies → verificar que SÍ hay cookie `admin_token` con flags `HttpOnly`
3. En consola del browser: `document.cookie` no debe mostrar `admin_token`
4. Recargar página — la sesión debe persistir
5. Verificar que todas las llamadas API en Network tab muestran cookie enviada automáticamente
