---
story_id: "3.3"
epic: "epic-3"
title: "Security Headers HTTP en next.config.ts"
status: backlog
estimated_days: 1
priority: critical
skills: ["security-review"]
affects:
  - frontend/next.config.ts
---

# Story 3.3: Security Headers HTTP en next.config.ts

## Contexto

`frontend/next.config.ts` actualmente solo define `rewrites()` y una variable de entorno de versión. No tiene ningún header de seguridad HTTP, exponiendo la aplicación a:

- **Clickjacking**: sin `X-Frame-Options` o CSP `frame-ancestors`, el panel puede embeberse en un iframe malicioso
- **MIME sniffing**: sin `X-Content-Type-Options: nosniff`, browsers pueden interpretar respuestas incorrectamente
- **XSS amplificado**: sin CSP, scripts inyectados pueden cargar recursos externos
- **Referrer leaks**: sin `Referrer-Policy`, URLs internas del admin se filtran a terceros

## User Story

Como operador de la plataforma,
quiero que el panel admin emita headers HTTP de seguridad en todas las respuestas,
para proteger contra clickjacking, XSS, sniffing de contenido e información de referrer.

## Scope

### Incluido
- Agregar `async headers()` en `next.config.ts` con headers de seguridad para todas las rutas
- CSP adaptada al uso real del panel (WS, assets internos, sin CDNs externos)
- Verificar que los headers no rompen funcionalidad existente

### Excluido
- HSTS (`Strict-Transport-Security`) — se configura mejor en el proxy/nginx en producción, no en Next.js
- Headers específicos por ruta (todos aplican a `source: '/(.*)'`)

## Acceptance Criteria

**AC1 — Headers configurados en next.config.ts:**
**Dado** que `frontend/next.config.ts` exporta la configuración de Next.js
**Cuando** se agrega la función `async headers()`
**Entonces** retorna el siguiente conjunto de headers para `source: '/(.*)'`:
```
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: camera=(), microphone=(), geolocation=()
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; connect-src 'self' ws: wss:; font-src 'self'; frame-ancestors 'none'
```

**AC2 — WebSocket no bloqueado por CSP:**
**Dado** que el dashboard usa WebSocket hacia el mismo origen del backend (via proxy Next.js `/ws`)
**Cuando** el browser conecta el WebSocket con la CSP activa
**Entonces** la consola del browser no muestra ningún error de CSP bloqueando la conexión WS
**Y** `connect-src 'self' ws: wss:` permite la conexión

**AC3 — Assets del panel cargan correctamente:**
**Dado** que el panel usa imágenes, fuentes e iconos internos (lucide-react, Next.js assets)
**Cuando** se carga cualquier página del panel con la CSP activa
**Entonces** la consola del browser no muestra violaciones de CSP para recursos legítimos
**Y** la UI renderiza completamente sin recursos bloqueados

**AC4 — Headers verificables:**
**Dado** que el servidor Next.js está corriendo en dev (`npm run dev`)
**Cuando** se hace `curl -I http://localhost:3001/`
**Entonces** la respuesta incluye todos los headers de seguridad configurados

**AC5 — Build y lint pasan:**
**Dado** que el cambio solo modifica `next.config.ts`
**Cuando** se ejecuta `cd frontend && npm run build`
**Entonces** no hay errores de TypeScript ni de configuración

## Notas de Implementación

- La CSP con `'unsafe-inline'` en `script-src` y `style-src` es un compromiso inicial razonable para no romper el panel. En un epic futuro de seguridad avanzada se puede migrar a nonces o hashes.
- `frame-ancestors 'none'` en CSP hace redundante `X-Frame-Options: DENY` pero se mantiene para browsers que no soportan CSP completamente.
- Si el panel carga fuentes desde Google Fonts u otros CDN externos, agregar el dominio a `font-src`.
- Usar la skill `security-review` para validar la CSP antes de hacer merge.
