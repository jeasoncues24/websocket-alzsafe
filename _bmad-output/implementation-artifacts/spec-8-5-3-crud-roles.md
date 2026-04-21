---
title: 'CRUD de roles'
type: 'feature'
created: '2026-04-21T06:01:39Z'
status: 'in-progress'
baseline_commit: '6b874ba874d344543461898c75e8e35b9bb3b7fe'
context:
  - _bmad-output/project-context.md
  - _bmad-output/planning-artifacts/epic-8.5-usuarios-roles-modulos.md
  - _bmad-output/implementation-artifacts/story-8-5-3-crud-roles.md
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The admin panel can list roles, but it cannot yet create, update, or delete them safely with the new schema fields and protection rules.

**Approach:** Extend the existing roles store and admin handlers so the CRUD is available under `/api/admin/roles`, permissions are stored as JSON, and the root role stays protected from invalid mutations or deletion.

## Boundaries & Constraints

**Always:** Keep `roles` read/write aligned to the current schema (`name`, `description`, `is_root`, `permissions`). Reject deletions when a role is in use.

**Ask First:** Any change that alters the root-role policy beyond preventing unsafe delete/disable or any schema migration.

**Never:** Add a CRUD path for `modules`; this story only consumes the module catalog to validate role permissions.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| HAPPY_PATH | Valid create/update role payload | Role is persisted with JSON permissions and returned via `/api/admin/roles` | N/A |
| ERROR_CASE | Role in use, duplicate name, invalid permissions, or root mutation | Request is rejected with a clear conflict/validation error | No partial writes |

</frozen-after-approval>

## Code Map

- `internal/http/admin.go` -- roles CRUD handlers and permission validation
- `internal/storage/role.go` -- role persistence, permissions JSON handling, delete guard support
- `internal/domain/role.go` -- role contract with permissions field
- `internal/http/router.go` -- register `/api/admin/roles/{id}` routes
- `internal/http/admin_test.go` -- CRUD and root/delete validation coverage

## Tasks & Acceptance

**Execution:**
- [ ] `internal/domain/role.go` -- add permissions to the role contract -- handlers can serialize the schema completely
- [ ] `internal/storage/role.go` -- load/save permissions JSON and support CRUD helpers -- persistence matches migrations
- [ ] `internal/http/admin.go` -- add create/get/update/delete handlers for roles with validation and root guard -- exposes the new admin contract
- [ ] `internal/http/router.go` -- register role detail routes under `/api/admin/roles/{id}` -- the new CRUD is reachable
- [ ] `internal/http/admin_test.go` -- cover create/update/delete, duplicate names, root protection, and in-use conflicts -- prevents regressions

**Acceptance Criteria:**
- Given valid role data, when POST `/api/admin/roles`, then the role is created with permissions stored as JSON.
- Given a role in use by a user, when DELETE `/api/admin/roles/:id`, then the request returns conflict and the role remains intact.
- Given a root role mutation or delete request, when the backend validates it, then the unsafe change is rejected.

## Spec Change Log

## Design Notes

Validate permissions against the module catalog rather than hardcoding names, but keep the rule permissive enough to accept the existing default role payloads already seeded in the database.

## Verification

**Commands:**
- `go test ./internal/http/...` -- expected: CRUD and role-policy tests pass
- `go test ./...` -- expected: repo remains green
