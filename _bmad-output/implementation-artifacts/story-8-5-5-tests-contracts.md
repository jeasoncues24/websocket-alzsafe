# Story 8.5.5: Tests de integridad y contratos

Status: done

## Story

As a backend maintainer,
I want automated coverage for the admin refactor,
so that the new contracts and delete rules do not regress.

## Acceptance Criteria

1. [AC: route-tests] Router and auth boundary tests verify that `/api/admin/*` only accepts the empresa JWT.
2. [AC: user-delete-tests] Tests cover hard delete vs disable behavior for `usuario_admin`.
3. [AC: role-delete-tests] Tests verify that deleting a role in use returns a conflict and does not mutate data.
4. [AC: module-readonly-tests] Tests confirm modules remain read-only and only expose GET behavior.
5. [AC: contract-tests] CRUD handlers return the expected payload shape for users, roles, and modules.
6. [AC: validation-tests] Invalid payloads for roles, users, and module overrides fail with clear validation errors.

## Tasks / Subtasks

- [ ] Add handler tests for the admin router and JWT boundary (AC: 1)
- [ ] Add storage/service tests for user delete paths (AC: 2)
- [ ] Add storage/service tests for role delete conflict handling (AC: 3)
- [ ] Add module catalog tests to prove read-only access (AC: 4)
- [ ] Add contract and validation coverage for create/update flows (AC: 5, 6)
- [ ] Run `go test ./...` and fix any regressions introduced by the refactor (AC: 1-6)

## Dev Notes

- Use the current Go testing style already present in the repo and keep tests close to the code they protect.
- This story is the safety net for the entire backend epic, so it should cover router, service, and storage behavior.
- Validate the admin schema/migrations before locking the test expectations.
- References: [Source: _bmad-output/planning-artifacts/epic-8.5-usuarios-roles-modulos.md], [Source: _bmad-output/project-context.md]

### Project Structure Notes

- Likely touch points: `_test.go` files alongside handlers/storage/services under `internal/`.
- Prefer table-driven tests for boundary and validation cases.

### References

- [Source: _bmad-output/planning-artifacts/epic-8.5-usuarios-roles-modulos.md]
- [Source: _bmad-output/project-context.md]

## Dev Agent Record

### Agent Model Used

BMAD Create Story

### Debug Log References

### Completion Notes List

### File List
