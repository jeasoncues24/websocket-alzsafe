# Story 8.5.4: Catalogo de modules y override por usuario

Status: done

## Story

As a panel admin operator,
I want to view the module catalog and override modules per user,
so that permissions stay explicit and editable from the admin panel.

## Acceptance Criteria

1. [AC: list-modules] GET `/api/admin/modules` returns the module catalog as read-only data.
2. [AC: get-user-modules] GET `/api/admin/usuario_admin/:id/modulos` returns the module IDs currently assigned to a user.
3. [AC: replace-user-modules] PUT `/api/admin/usuario_admin/:id/modulos` replaces the full module set atomically.
4. [AC: invalid-module] Invalid module IDs or slugs are rejected with a validation error.
5. [AC: read-only-catalog] Modules do not expose create, update, or delete handlers in this epic.
6. [AC: permissions-support] The module catalog can be used as the source of truth for role permissions and user overrides.

## Tasks / Subtasks

- [ ] Add read-only module handlers and storage methods (AC: 1, 5)
- [ ] Implement user-module lookup and replacement endpoints (AC: 2, 3)
- [ ] Make module replacement atomic so partial writes do not leak (AC: 3)
- [ ] Validate module IDs/slugs before persisting overrides (AC: 4)
- [ ] Add tests for read-only access and override replacement (AC: 1-6)

## Dev Notes

- `modules` has `name`, `slug`, `description`, `created_at`, and `updated_at` in the current schema.
- `user_modules` is the join table used for per-user overrides; replace the set, do not patch it item by item.
- Keep this story read-only for the module catalog itself; CRUD belongs nowhere in this epic.
- References: [Source: _bmad-output/planning-artifacts/epic-8.5-usuarios-roles-modulos.md], [Source: frontend/app/modules/page.tsx], [Source: frontend/app/users_admin/page.tsx]

### Project Structure Notes

- Likely touch points: `internal/storage/module.go`, `internal/storage/user_module.go`, admin handlers, and transaction helpers.
- Reuse the module catalog validation for the role permissions story.

### References

- [Source: _bmad-output/planning-artifacts/epic-8.5-usuarios-roles-modulos.md]
- [Source: frontend/app/modules/page.tsx]
- [Source: frontend/app/users_admin/page.tsx]

## Dev Agent Record

### Agent Model Used

BMAD Create Story

### Debug Log References

### Completion Notes List

### File List
