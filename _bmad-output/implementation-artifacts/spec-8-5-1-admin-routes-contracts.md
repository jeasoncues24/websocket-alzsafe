---
title: 'Rutas admin y contratos base'
type: 'feature'
created: '2026-04-21T06:01:39Z'
status: 'done'
baseline_commit: '6b874ba874d344543461898c75e8e35b9bb3b7fe'
context:
  - _bmad-output/project-context.md
  - _bmad-output/planning-artifacts/epic-8.5-usuarios-roles-modulos.md
  - _bmad-output/implementation-artifacts/story-8-5-1-admin-routes-contracts.md
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** El panel admin todavía expone el flujo de usuarios bajo `/api/admin/users` y lo protege con el auth de admin-token, pero el nuevo contrato debe vivir en `/api/admin/usuario_admin` y aceptar solo JWT de empresa.

**Approach:** Mantener la funcionalidad existente como base, añadir la nueva ruta admin con el prefijo correcto, y mover la protección al middleware de empresa para que el boundary quede explícito sin romper el resto de rutas.

## Boundaries & Constraints

**Always:** `/api/admin/usuario_admin`, `/api/admin/roles` y `/api/admin/modules` son el nuevo contrato público del panel admin; el token de teléfono nunca debe abrir estas rutas.

**Ask First:** Cualquier decisión de retirar por completo los aliases legacy `/api/admin/users` o cambiar el token storage key del frontend.

**Never:** Reescribir el panel completo o tocar el flujo de `/api/*` del token por teléfono en esta story.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| HAPPY_PATH | Empresa JWT válido contra `/api/admin/usuario_admin` | La request pasa y el handler ve claims de empresa | N/A |
| ERROR_CASE | Token por teléfono o ausencia de auth | La request se rechaza con 401/403 | No exponer datos ni fallback a legacy |

</frozen-after-approval>

## Code Map

- `internal/http/router.go` -- registra las rutas admin nuevas y su middleware
- `internal/http/middleware/empresa_auth.go` -- boundary de JWT de empresa
- `internal/http/admin.go` -- handlers de admin users/roles/modules y parsing de paths
- `internal/storage/admin_user.go` -- listado/lookup filtrable por empresa
- `internal/http/admin_test.go` -- cobertura del boundary y route parsing

## Tasks & Acceptance

**Execution:**
- [x] `internal/http/router.go` -- registrar `/api/admin/usuario_admin` y protegerlo con empresa JWT -- expone el contrato nuevo sin romper el resto
- [x] `internal/http/admin.go` -- aceptar el prefijo `usuario_admin` además del legacy en parsing -- la UI nueva puede consumir el nuevo path
- [x] `internal/storage/admin_user.go` -- permitir listado por empresa para el nuevo boundary -- evita filtrar datos de otras empresas
- [x] `internal/http/admin_test.go` -- cubrir 401/403 y parsing del nuevo prefijo -- evita regresiones de auth/routing

**Acceptance Criteria:**
- Given una request sin JWT de empresa, when entra a `/api/admin/usuario_admin`, then retorna 401/403.
- Given un JWT de empresa válido, when entra a `/api/admin/usuario_admin`, then el handler responde con contexto de empresa.
- Given la ruta nueva, when se parsea el ID de usuario, then acepta `/api/admin/usuario_admin/:id` y no depende de `/api/admin/users`.

## Spec Change Log

## Design Notes

La regla es simple: el contrato nuevo se expone por ruta nueva y se protege con empresa JWT; los handlers se adaptan para reconocer ambos contextos durante la transición. Esto permite migrar frontend después sin dejar un agujero de auth.

## Verification

**Commands:**
- `go test ./internal/http/...` -- expected: route/auth/admin tests pass
- `go test ./...` -- expected: repo still builds and unit tests remain green

## Suggested Review Order

**Route boundary**

- Admin route entry point and company JWT boundary
  [`router.go:138`](../../internal/http/router.go#L138)

- Admin route handlers for the new resource names
  [`admin.go:147`](../../internal/http/admin.go#L147)

**Data access and policy**

- Empresa-scoped listing and dependency-aware delete policy
  [`admin_user.go:172`](../../internal/storage/admin_user.go#L172)

- Route and policy regression coverage
  [`admin_test.go:157`](../../internal/http/admin_test.go#L157)
