---
epic_id: epic-3
title: "Hardening de Seguridad y Calidad Frontend"
status: backlog
created: 2026-05-11
estimated_duration: "10 días hábiles (1 dev)"
react_doctor_score_before: 78
react_doctor_score_target: 90
---

# Epic 3: Hardening de Seguridad y Calidad Frontend

## 🚨 REGLA DE RAMA — IMPERATIVA

**Todo el código de este epic se implementa EXCLUSIVAMENTE en la rama `feature/security`.**

```bash
# Verificación obligatoria antes de tocar cualquier archivo de código:
git branch --show-current  # → debe mostrar: feature/security
```

- Si la rama activa no es `feature/security` → detener. No escribir código. Cambiar de rama primero.
- Esta rama es exclusiva del Epic 3. Ningún otro epic o tarea debe usar `feature/security`.
- Merge a `v1`/`main` solo cuando todas las stories estén `done` y el build pase.

---

## Objetivo

El equipo puede operar wsapi con la confianza de que el frontend protege los tokens de sesión admin contra XSS, usa comunicaciones WebSocket sin exponer credenciales en URLs, emite headers HTTP defensivos en toda la aplicación, y el código base está libre de bugs de correctness, estado inconsistente y deuda técnica acumulada — sin alterar ninguna lógica de negocio existente.

## Motivación

React Doctor escaneó el frontend el 2026-05-11 y encontró **312 issues en 37/46 archivos** (score 78/100). El audit manual de código confirmó 3 vulnerabilidades de seguridad críticas no detectadas por React Doctor:

1. **JWT en localStorage** — vector XSS directo (`frontend/lib/api.ts`)
2. **Token WebSocket en query param** — aparece en logs del servidor (`buildAdminWsUrl`)
3. **Cero security headers** — sin CSP, sin X-Frame-Options, sin protección clickjacking (`next.config.ts`)

## Alcance

### Incluido
- Migración JWT de localStorage a httpOnly cookie (frontend + ajuste mínimo backend Go)
- Seguridad WebSocket: eliminar token de query param
- Security headers HTTP en next.config.ts
- Fix de bugs de correctness: MetricCard anidado, hydration mismatches, array index keys
- Refactor de estado: useReducer + eliminación de cascading setState + effect chains
- Accesibilidad: labels htmlFor, key events, headings con contenido
- Limpieza de dead code: exports, types, archivos sin usar
- Tailwind shortcuts (size-N), font headings (semibold), useRouter destructuring

### Excluido
- Cambios a endpoints o payloads del backend Go
- Cambios a lógica de negocio (WhatsApp, empresas, teléfonos, API keys, difusiones)
- Cambios a la protección de rutas client-side existente (middleware Next.js)
- Nuevas features o cambios visuales significativos

## Dependencias

- Epic 2 puede seguir `in-progress` en paralelo — este epic no bloquea ni es bloqueado por Epic 2
- Story 3.2 depende de Story 3.1 (la cookie debe existir antes de eliminar token de WS query param)
- Stories 3.3–3.8 son independientes entre sí y pueden ejecutarse en cualquier orden

## Riesgos

| Riesgo | Mitigación |
|--------|-----------|
| Migración cookie httpOnly rompe login en dev sin HTTPS | Usar `Secure` solo en producción; en dev usar solo `HttpOnly; SameSite=Strict` |
| CSP demasiado estricta bloquea assets o WS | Probar en dev antes de deploy; ajustar `connect-src` para WS |
| Refactor useReducer introduce regresiones en lógica de teléfonos | Mantener comportamiento idéntico; verificar con lint + build + prueba manual del flujo QR |
| Codemod Tailwind (132 archivos) introduce errores | Revisar diff completo antes de commit; ejecutar build visual check |

## Criterios de Éxito

- [ ] JWT no persiste en `localStorage` en ningún path del frontend
- [ ] WebSocket no expone token en URL (`?token=` eliminado)
- [ ] `curl -I http://localhost:3001` muestra todos los security headers configurados
- [ ] `npx react-doctor@latest .` reporta score ≥ 90/100
- [ ] `npm run lint && npm run build` pasan sin errores en cada story
- [ ] `cd backend && go build ./...` pasa sin errores (para stories con cambio Go)
- [ ] Ninguna funcionalidad de negocio rota: login, dashboard, QR, sesiones, API keys, empresas

## Skills recomendadas

| Skill | Cuándo usarla |
|-------|--------------|
| `better-auth-best-practices` | Stories 3.1 y 3.2 — auth y JWT |
| `golang-security` | Story 3.1 — cambios en backend Go |
| `security-review` | Stories 3.1, 3.2, 3.3 — verificación de seguridad |
| `bmad-code-review` | Todas las stories — review antes de marcar done |
| `bmad-dev-story` | Para implementar cada story |

## Stories

| ID | Título | Área | Días est. | Prioridad |
|----|--------|------|-----------|-----------|
| 3.1 | Migrar JWT de localStorage a httpOnly Cookie | Seguridad crítica | 1-2 | 🔴 |
| 3.2 | Asegurar WebSocket — Eliminar Token de Query Param | Seguridad crítica | 1 | 🔴 |
| 3.3 | Security Headers HTTP en next.config.ts | Seguridad crítica | 1 | 🔴 |
| 3.4 | Fix Correctness — MetricCard, Hydration, Array Keys | Correctness | 1-2 | 🟡 |
| 3.5 | Refactor Estado — useReducer y Cascading setState | State/Effects | 1-2 | 🟡 |
| 3.6 | Fix Accesibilidad — Labels, Key Events, Headings | Accesibilidad | 1 | 🟡 |
| 3.7 | Limpieza de Dead Code | Deuda técnica | 1 | 🟢 |
| 3.8 | Tailwind Shortcuts, Font Headings y useRouter | Deuda técnica | 1-2 | 🟢 |

## Condición de Cierre

El epic se cierra cuando todas las stories están en estado `done`, `npm run build` pasa sin errores, `npx react-doctor@latest .` reporta ≥ 90/100, y ningún flujo de negocio fue alterado.
