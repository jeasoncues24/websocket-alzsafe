---
story_id: "3.2"
epic: "epic-3"
title: "Asegurar WebSocket — Eliminar Token de Query Param"
status: backlog
estimated_days: 1
priority: critical
depends_on: ["3.1"]
skills: ["security-review", "golang-security"]
affects:
  - frontend/lib/api.ts (buildAdminWsUrl)
  - frontend/app/dashboard/page.tsx (o donde se use buildAdminWsUrl)
  - backend/internal/http/middleware.go (WS upgrade handler)
---

# Story 3.2: Asegurar WebSocket — Eliminar Token de Query Param

## Contexto

**Vulnerabilidad confirmada** en `frontend/lib/api.ts`:

```typescript
export function buildAdminWsUrl(path: string, token?: string) {
  const url = new URL(API_BASE);
  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  url.pathname = path;
  url.search = "";
  if (token) {
    url.searchParams.set("token", token);  // ← INSEGURO: JWT en URL
  }
  return url.toString();
}
```

El JWT en la URL aparece en:
- Logs del servidor Go (zerolog registra la URL completa del request)
- Historial del browser
- Headers `Referer` de requests subsiguientes
- Herramientas de monitoreo y APM

**Prerequisito:** Story 3.1 debe estar completada — la cookie httpOnly es el mecanismo de auth que reemplaza el query param.

## User Story

Como administrador del sistema,
quiero que la conexión WebSocket admin no transmita el JWT en la URL,
para que el token no aparezca en logs del servidor ni en el historial del navegador.

## Scope

### Incluido
- Simplificar `buildAdminWsUrl` para eliminar el parámetro `token`
- Actualizar todos los call sites de `buildAdminWsUrl` que pasan el token
- Verificar que el middleware Go de WS upgrade lee la cookie (igual que requests HTTP)

### Excluido
- Cambios al protocolo WebSocket (mensajes, eventos)
- Cambios a la lógica del dashboard que usa el WS
- Implementar WebSocket tickets de corta duración (puede ser mejora futura si el backend no soporta cookies en WS)

## Acceptance Criteria

**AC1 — buildAdminWsUrl no incluye token en URL:**
**Dado** que `buildAdminWsUrl` en `frontend/lib/api.ts` acepta `token?: string`
**Cuando** se actualiza la función
**Entonces** elimina el parámetro `token` de su firma
**Y** no añade ningún query param con el JWT a la URL generada
**Y** la URL resultante tiene la forma `ws://host/path` sin query string de auth

**AC2 — Call sites actualizados:**
**Dado** que los componentes que usan WebSocket llaman `buildAdminWsUrl(path, token)`
**Cuando** se actualiza la firma de la función
**Entonces** todos los call sites se actualizan para no pasar token
**Y** TypeScript no reporta errores de tipo en ningún call site

**AC3 — Backend acepta WS sin token en query param:**
**Dado** que el handler Go de upgrade WebSocket en `/ws` (o la ruta equivalente)
**Cuando** llega un WebSocket upgrade request con cookie `admin_token` válida
**Entonces** el middleware valida el token desde la cookie
**Y** acepta el upgrade y establece la conexión WS normalmente

**AC4 — Logs del servidor limpios:**
**Dado** que el servidor Go registra las URLs de las conexiones WebSocket con zerolog
**Cuando** un cliente se conecta al WS
**Entonces** los logs no contienen ningún JWT en la URL de conexión

**AC5 — Build y lint pasan:**
**Dado** que todos los cambios están aplicados
**Cuando** se ejecuta `cd frontend && npm run lint && npm run build`
**Entonces** no hay errores

## Notas de Implementación

- Si el backend Go ya valida la cookie en el middleware general y el WS upgrade pasa por ese middleware, no se necesita cambio en el backend para esta story.
- Verificar en `backend/internal/http/routes_admin.go` o el router principal cómo se registra la ruta `/ws` y si pasa por el middleware de auth.
- Si las cookies no se envían en el WebSocket upgrade (algunos browsers no las envían en `wss://` cross-origin), evaluar enviar el token en el primer mensaje del protocolo como alternativa.
