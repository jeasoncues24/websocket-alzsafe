# Story 8.6.4: Modules catalog

Status: done

## Story

As a panel admin operator,
I want a read-only modules catalog,
so that I can understand available permissions without editing the catalog itself.

## Acceptance Criteria

1. [AC: modules-table] The modules page shows `name`, `slug`, and `description`.
2. [AC: readonly] There are no create, edit, or delete actions for modules in this epic.
3. [AC: permissions-reference] The catalog can be used as a reference when editing role permissions or user overrides.
4. [AC: loading-empty] Loading and empty states are clear.

## Tasks / Subtasks

- [ ] Keep the modules page read-only and aligned with the backend catalog (AC: 1, 2)
- [ ] Improve the table copy and empty state guidance (AC: 4)
- [ ] Ensure the page can serve as a permission reference in nearby flows (AC: 3)

## Dev Notes

- The module catalog is intentionally read-only in both backend and frontend.
- This story should not introduce any write affordances, even temporarily.
- References: [Source: frontend/app/modules/page.tsx], [Source: _bmad-output/planning-artifacts/epic-8.6-panel-admin-frontend-usuario-roles-modulos.md]

### Project Structure Notes

- Likely touch points: `frontend/app/modules/page.tsx` and any shared table/empty-state components.
- Keep the screen simple and fast; it is support tooling for permissions work.

### References

- [Source: frontend/app/modules/page.tsx]
- [Source: _bmad-output/planning-artifacts/epic-8.6-panel-admin-frontend-usuario-roles-modulos.md]

## Dev Agent Record

### Agent Model Used

BMAD Create Story

### Debug Log References

### Completion Notes List

### File List
