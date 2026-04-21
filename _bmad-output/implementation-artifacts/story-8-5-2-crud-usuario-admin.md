# Story 8.5.2: CRUD de usuario_admin

Status: done

## Story

As a panel admin operator,
I want to create, update, disable, and delete `usuario_admin` records,
so that the backend admin users match the real migration schema.

## Acceptance Criteria

1. [AC: list-user-admins] GET `/api/admin/usuario_admin` returns paginated users with `username`, `email`, `empresa_id`, `rol`, `role_id`, `is_root`, `activo`, and `last_login_at`.
2. [AC: get-user-admin] GET `/api/admin/usuario_admin/:id` returns one user and preserves legacy `rol` plus normalized `role_id`.
3. [AC: create-user-admin] POST `/api/admin/usuario_admin` validates `username` and `password`, stores the password hash, and links the role/company correctly.
4. [AC: update-user-admin] PUT `/api/admin/usuario_admin/:id` updates allowed fields without corrupting `last_login_at`.
5. [AC: delete-user-admin] DELETE `/api/admin/usuario_admin/:id` hard-deletes only when no blocking dependency exists; otherwise it disables the user with `activo = 0`.
6. [AC: dependency-check] Dependency checks respect the real schema and migration state before deciding between delete and disable.
7. [AC: validation] Duplicate usernames, invalid role IDs, and malformed payloads return clear validation errors.

## Tasks / Subtasks

- [ ] Read the current admin_users migration before coding the handler/service changes (AC: 1-7)
- [ ] Update storage/repository methods for list/get/create/update/delete (AC: 1-5)
  - [ ] Preserve `rol` while writing `role_id` as the normalized relation
  - [ ] Keep `last_login_at` read-only from this story
- [ ] Implement delete decision logic: hard delete vs disable (AC: 5, 6)
- [ ] Wire handler endpoints under `/api/admin/usuario_admin` (AC: 1-5)
- [ ] Add unit and integration tests for the CRUD and delete rules (AC: 1-7)

## Dev Notes

- Table fields to honor: `username`, `password_hash`, `email`, `empresa_id`, `rol`, `role_id`, `is_root`, `activo`, `created_at`, `updated_at`, `last_login_at`.
- `user_modules` has a cascading FK to `admin_users`; the soft-delete path is for non-cascading references or operational safety rules defined in the migration set.
- Keep the contract aligned with the current frontend admin page shape so the UI story can land on top of this work.
- References: [Source: _bmad-output/planning-artifacts/epic-8.5-usuarios-roles-modulos.md], [Source: frontend/app/users_admin/page.tsx], [Source: frontend/lib/api.ts]

### Project Structure Notes

- Likely touch points: `internal/http/handlers.go` or admin handler files, `internal/storage/`, and domain/service structs for admin users.
- Prefer small focused functions for dependency checks and payload validation.

### References

- [Source: _bmad-output/planning-artifacts/epic-8.5-usuarios-roles-modulos.md]
- [Source: frontend/app/users_admin/page.tsx]
- [Source: frontend/lib/api.ts]

## Dev Agent Record

### Agent Model Used

BMAD Create Story

### Debug Log References

### Completion Notes List

### File List
