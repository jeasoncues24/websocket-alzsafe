# Story 8.6.5: Estados criticos y feedback de riesgo

Status: done

## Story

As a panel admin operator,
I want clear feedback for risky or blocked actions,
so that I understand when the system disabled, rejected, or confirmed an operation.

## Acceptance Criteria

1. [AC: destructive-confirmation] Delete actions require an explicit confirmation dialog.
2. [AC: disabled-vs-deleted] The UI explains when a user was disabled instead of removed.
3. [AC: role-block-feedback] The UI explains when a role delete is blocked because the role is in use.
4. [AC: inline-errors] Backend validation and integrity errors are shown in a visible, actionable way.
5. [AC: badges-states] Active, inactive, root, and in-use states are visually distinct.
6. [AC: empty-guidance] Empty states tell the admin what to do next.

## Tasks / Subtasks

- [ ] Add confirmation flows to destructive actions across the admin UI (AC: 1)
- [ ] Standardize error banners or inline alerts for backend failures (AC: 4)
- [ ] Add state badges for active, inactive, root, and in-use cases (AC: 2, 3, 5)
- [ ] Improve empty-state copy so it guides the next action (AC: 6)

## Dev Notes

- This story is a UX polish pass that spans the users, roles, and modules screens.
- Use shadcn-friendly patterns already present in the app instead of introducing a new component library.
- References: [Source: _bmad-output/planning-artifacts/epic-8.6-panel-admin-frontend-usuario-roles-modulos.md], [Source: frontend/components/layout/sidebar.tsx]

### Project Structure Notes

- Likely touch points: shared feedback components, admin tables, dialog components, and badge styles.
- Keep the feedback copy short and operational; this is an internal admin tool.

### References

- [Source: _bmad-output/planning-artifacts/epic-8.6-panel-admin-frontend-usuario-roles-modulos.md]
- [Source: frontend/components/layout/sidebar.tsx]

## Dev Agent Record

### Agent Model Used

BMAD Create Story

### Debug Log References

### Completion Notes List

### File List
