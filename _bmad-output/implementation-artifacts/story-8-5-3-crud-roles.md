# Story 8.5.3: CRUD de roles

Status: done

## Story

As a panel admin operator,
I want to manage roles from the admin backend,
so that role definitions and permissions stay editable without touching code.

## Acceptance Criteria

1. [AC: list-roles] GET `/api/admin/roles` returns all roles with `name`, `description`, `is_root`, and `permissions`.
2. [AC: create-role] POST `/api/admin/roles` creates a new role and validates `name` uniqueness.
3. [AC: update-role] PUT `/api/admin/roles/:id` updates role metadata and permissions without breaking users already assigned to it.
4. [AC: delete-role] DELETE `/api/admin/roles/:id` deletes only when the role is not used by any `usuario_admin`.
5. [AC: in-use-block] If a role is referenced by users, the delete request returns a conflict and the role remains intact.
6. [AC: permissions-validation] `permissions` must be valid JSON and only reference real module slugs.
7. [AC: root-guard] Root-role semantics remain protected by backend validation.

## Tasks / Subtasks

- [ ] Implement role storage CRUD and usage checks (AC: 1-5)
- [ ] Add permissions JSON validation against module slugs (AC: 6)
- [ ] Protect root role semantics in create/update/delete flows (AC: 7)
- [ ] Wire `/api/admin/roles` handlers and responses (AC: 1-7)
- [ ] Add tests for conflict, validation, and root guard behavior (AC: 1-7)

## Dev Notes

- The roles table includes `name`, `description`, `is_root`, `permissions`, `created_at`, and `updated_at`.
- Do not degrade delete into disable for roles; if the role is in use, reject the operation.
- Keep the contract consistent with the user-admin story so role selectors and permission editors can consume it directly.
- References: [Source: _bmad-output/planning-artifacts/epic-8.5-usuarios-roles-modulos.md], [Source: frontend/app/roles/page.tsx], [Source: frontend/lib/api.ts]

### Project Structure Notes

- Likely touch points: `internal/storage/role.go`, admin handlers, validation helpers, and test files for storage and handlers.
- Reuse the module catalog when validating permissions instead of hardcoding slugs.

### References

- [Source: _bmad-output/planning-artifacts/epic-8.5-usuarios-roles-modulos.md]
- [Source: frontend/app/roles/page.tsx]
- [Source: frontend/lib/api.ts]

## Dev Agent Record

### Agent Model Used

BMAD Create Story

### Debug Log References

### Completion Notes List

### File List
