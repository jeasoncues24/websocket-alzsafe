# Story 8.6.3: Roles workspace

Status: done

## Story

As a panel admin operator,
I want to manage roles from the UI,
so that permissions stay editable without leaving the panel.

## Acceptance Criteria

1. [AC: role-list] The roles page shows `name`, `description`, `is_root`, and usage status.
2. [AC: create-edit-role] Create and edit role flows update the backend contract correctly.
3. [AC: delete-block] If a role is in use, the UI blocks or explains the delete failure clearly.
4. [AC: root-badge] Root roles are visually distinguished from normal roles.
5. [AC: permissions-editor] Permissions can be edited with a module-aware control and validated before save.
6. [AC: loading-error] The screen includes loading, empty, and error states.

## Tasks / Subtasks

- [ ] Replace the read-only roles table with a CRUD workspace (AC: 1-6)
- [ ] Add create/edit dialog or drawer for roles (AC: 2, 5)
- [ ] Show root and in-use badges in the table (AC: 1, 4)
- [ ] Wire delete confirmation and in-use error handling (AC: 3)
- [ ] Connect the permissions editor to the module catalog (AC: 5)

## Dev Notes

- The current roles page is read-only; this story upgrades it to a full management workspace.
- Keep `permissions` aligned with the module slugs returned by the backend catalog.
- References: [Source: frontend/app/roles/page.tsx], [Source: frontend/lib/api.ts], [Source: _bmad-output/planning-artifacts/epic-8.6-panel-admin-frontend-usuario-roles-modulos.md]

### Project Structure Notes

- Likely touch points: `frontend/app/roles/page.tsx`, shared form controls, and the admin API client.
- Reuse existing shadcn components to keep the UI consistent with the rest of the panel.

### References

- [Source: frontend/app/roles/page.tsx]
- [Source: frontend/lib/api.ts]
- [Source: _bmad-output/planning-artifacts/epic-8.6-panel-admin-frontend-usuario-roles-modulos.md]

## Dev Agent Record

### Agent Model Used

BMAD Create Story

### Debug Log References

### Completion Notes List

### File List
