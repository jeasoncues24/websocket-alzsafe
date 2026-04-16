---
title: 'Frontend Bug Fixes and Token Consistency'
type: 'bugfix'
created: '2026-04-16'
status: 'in-review'
baseline_commit: '2e4c3843d281b796245e6a9ce4c5aeab5633e1df'
context: ['_bmad-output/project-context.md']
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** El frontend presenta un error crítico de compilación porque el componente `DashboardPage` está definido dos veces en el mismo archivo. Además, existe una inconsistencia en el uso de las claves del localStorage para el token de autenticación (`token` vs `admin_token`) y las llamadas a la API del dashboard están usando rutas públicas antiguas en lugar de las nuevas rutas protegidas.

**Approach:** Eliminar la definición duplicada en el dashboard, estandarizar el uso de `admin_token` en todo el frontend, y actualizar los endpoints de la API para que coincidan con la nueva arquitectura de seguridad del backend.

## Boundaries & Constraints

**Always:** Usar `admin_token` como única clave de autenticación en el frontend. Mantener la estética premium y el uso de componentes de shadcn/ui.

**Ask First:** Cambiar otros endpoints de `/admin/` a `/api/` si no están explitamente listados.

**Never:** Eliminar lógica de las métricas del dashboard; solo debemos limpiar el código duplicado manteniendo la versión más completa.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Login Exitoso | Credenciales correctas | Guarda `admin_token` y redirige a `/dashboard` | Muestra error del backend si falla |
| Carga Dashboard | JWT válido en `admin_token` | Llama a `/api/dashboard/metricas` y muestra datos | Muestra ceros/error si falla |
| Acceso Usuarios | JWT válido | Llama a `/admin/users` usando header `Authorization: Bearer <admin_token>` | Redirige a login si no hay token |

</frozen-after-approval>

## Code Map

- `frontend/app/dashboard/page.tsx` -- Página de dashboard con duplicación de componente.
- `frontend/lib/api.ts` -- Definiciones de interfaces y funciones de llamada a la API.
- `frontend/app/login/page.tsx` -- Lógica de inicio de sesión y persistencia del token.
- `frontend/app/users/page.tsx` -- Gestión de usuarios con llamada manual a fetch.

## Tasks & Acceptance

**Execution:**
- [x] `frontend/app/dashboard/page.tsx` -- Eliminar el segundo bloque duplicado de `DashboardPage` (líneas 264-474) -- Resuelve error de compilación.
- [x] `frontend/lib/api.ts` -- Cambiar `token` por `admin_token` en `fetchWithAuth` y actualizar `getMetrics` para usar el endpoint `/api/dashboard/metricas` -- Alineación con seguridad backend.
- [x] `frontend/app/login/page.tsx` -- Cambiar endpoint a `/api/auth/login` y mejorar visualización de errores -- Uso de auth real basada en DB.
- [x] `frontend/app/users/page.tsx` -- Actualizar header de Authorization para usar `admin_token` en el fetch manual de roles/módulos -- Consistencia de tokens.

**Acceptance Criteria:**
- Given session started, when navigating to `/dashboard`, then the metrics load without duplicate component errors.
- Given login form, when submitting correct credentials, then the token is stored as `admin_token`.
- Given any protected page, when the token is present, then headers include `Bearer <admin_token>`.

## Design Notes

La primera versión de `DashboardPage` en `frontend/app/dashboard/page.tsx` es más completa ya que incluye el componente `MetricCard` y la estructura de `Tabs`. La segunda versión parece ser un remanente de una implementación anterior o fallida.

## Verification

**Commands:**
- `grep -r "DashboardPage" frontend/app/dashboard/page.tsx | wc -l` -- expected: total occurrences should decrease (only 1 export default).
- `grep "admin_token" frontend/lib/api.ts` -- expected: find usages of admin_token.
