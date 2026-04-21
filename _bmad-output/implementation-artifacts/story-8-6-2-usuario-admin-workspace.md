# Story 8.6.2: Usuario Admin workspace

Status: done

## Story

As a panel admin operator,
I want a workspace for `usuario_admin`,
so that I can manage admin users with a clear and safe UI.

## Acceptance Criteria

1. [AC: table-columns] The user admin table shows `username`, `email`, `empresa`, `rol`, `root`, `activo`, and row actions.
2. [AC: create-edit] Create and edit open in a modal or drawer and map to the backend contract.
3. [AC: search-pagination] The workspace keeps search and pagination for long user lists.
4. [AC: delete-feedback] Delete actions clearly show whether the backend deleted the user or disabled it.
5. [AC: module-override] Module overrides can be assigned from the workspace and replace the full set.
6. [AC: loading-error] Loading and error states are explicit and do not leave the page blank.

## Tasks / Subtasks

- [ ] Refactor the current users page into a `usuario_admin` workspace (AC: 1-6)
- [ ] Replace the custom overlay form with a shadcn dialog or drawer pattern (AC: 2)
- [ ] Map create/update/delete actions to the new API client (AC: 2, 4, 5)
- [ ] Keep search, pagination, and empty states clear for large lists (AC: 3, 6)
- [ ] Add visible badges for root and active/inactive state (AC: 1, 4)

## Dev Notes

- The current page already has the core table and modal logic in `frontend/app/users_admin/page.tsx`; this story is a refactor, not a brand-new screen.
- The module override behavior must replace the set, not patch it piecemeal, to stay consistent with the backend story.
- References: [Source: frontend/app/users_admin/page.tsx], [Source: frontend/lib/api.ts], [Source: _bmad-output/planning-artifacts/epic-8.6-panel-admin-frontend-usuario-roles-modulos.md]

### Project Structure Notes

- Likely touch points: `frontend/app/usuario_admin/page.tsx`, form/dialog components, and the admin API hooks in `frontend/lib/api.ts`.
- Keep the page dense and operational; this is an internal admin tool, not a marketing UI.

### References

- [Source: frontend/app/users_admin/page.tsx]
- [Source: frontend/lib/api.ts]
- [Source: _bmad-output/planning-artifacts/epic-8.6-panel-admin-frontend-usuario-roles-modulos.md]

## Dev Agent Record

### Agent Model Used

BMAD Create Story

### Debug Log References

### Completion Notes List

### File List
