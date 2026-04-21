# Story 8.5.1: Rutas admin y contratos base

Status: done

## Story

As a admin JWT consumer,
I want all admin routes under `/api/admin/*`,
so that the panel admin stays isolated from the phone-token API.

## Acceptance Criteria

1. [AC: route-prefix] All admin endpoints required by this epic are registered only under `/api/admin/*`.
2. [AC: auth-boundary] Requests authenticated with the telefono JWT are rejected from `/api/admin/*`.
3. [AC: admin-jwt] Requests authenticated with the empresa JWT are accepted by the admin router and preserve claims in context.
4. [AC: user-resource] The user-admin resource is exposed as `usuario_admin`, not `usuarios` or `users`.
5. [AC: no-legacy-leak] No new admin contract depends on `/api/admin/users`.
6. [AC: regression-safety] Existing non-admin routes keep working unchanged.

## Tasks / Subtasks

- [ ] Update router registrations to use `/api/admin/usuario_admin`, `/api/admin/roles`, and `/api/admin/modules` (AC: 1, 4)
  - [ ] Remove legacy users route exposure from the admin router
- [ ] Enforce empresa JWT middleware on admin routes (AC: 2, 3)
  - [ ] Reject telefono-token claims early with 401/403
- [ ] Normalize admin response shape for the new resource names (AC: 4)
- [ ] Add router and middleware tests for admin boundary cases (AC: 1-3, 6)

## Dev Notes

- Use the existing auth/middleware patterns in `internal/http/middleware/` and router registration in `internal/http/router.go`.
- Keep the phone-token API untouched under `/api/*`.
- This story is the entry point for the rest of the backend admin refactor.
- References: [Source: _bmad-output/planning-artifacts/epic-8.5-usuarios-roles-modulos.md], [Source: _bmad-output/project-context.md]

### Project Structure Notes

- Likely touch points: `internal/http/router.go`, admin handlers, auth middleware, and any response helpers used by admin endpoints.
- Do not introduce a second admin auth path; reuse the empresa JWT flow already in the project.

### References

- [Source: frontend/lib/api.ts]
- [Source: frontend/components/layout/sidebar.tsx]
- [Source: _bmad-output/planning-artifacts/epic-8.5-usuarios-roles-modulos.md]

## Dev Agent Record

### Agent Model Used

BMAD Create Story

### Debug Log References

### Completion Notes List

### File List
