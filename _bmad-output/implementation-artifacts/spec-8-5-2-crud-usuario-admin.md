---
title: 'CRUD de usuario_admin'
type: 'feature'
created: '2026-04-21T06:01:39Z'
status: 'done'
baseline_commit: '6b874ba874d344543461898c75e8e35b9bb3b7fe'
context:
  - _bmad-output/project-context.md
  - _bmad-output/planning-artifacts/epic-8.5-usuarios-roles-modulos.md
  - _bmad-output/implementation-artifacts/story-8-5-2-crud-usuario-admin.md
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The admin user management still needs the final CRUD contract to match the new `usuario_admin` domain, including password hashing, empresa scoping, and delete-vs-disable policy.

**Approach:** Reuse the new admin route boundary and storage primitives, then finish the CRUD paths so create/update/delete/get all speak the new contract and obey the dependency rules from the data model.

## Boundaries & Constraints

**Always:** Preserve `username`, `email`, `empresa_id`, `rol`, `role_id`, `is_root`, `activo`, and `last_login_at` semantics from the current schema.

**Ask First:** Any schema migration or removal of the legacy `/api/admin/users` alias.

**Never:** Introduce a second user entity or rewrite role/module management in this story.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| HAPPY_PATH | Valid create/update payload for `/api/admin/usuario_admin` | User is created or updated with bcrypt password storage on create | N/A |
| ERROR_CASE | Duplicate username, invalid role, or cross-company access | Request is rejected clearly | No partial writes |

</frozen-after-approval>

## Code Map

- `internal/http/admin.go` -- create/get/update/delete handlers for `usuario_admin`
- `internal/storage/admin_user.go` -- CRUD store and dependency-aware delete policy
- `internal/http/admin_test.go` -- CRUD and delete policy coverage
- `internal/domain/admin_user.go` -- request/response contract fields

## Tasks & Acceptance

**Execution:**
- [ ] `internal/http/admin.go` -- finish the `usuario_admin` CRUD paths and preserve the new response shape -- the frontend can consume a stable contract
- [ ] `internal/storage/admin_user.go` -- keep create/update/list/get aligned with schema fields and dependency-aware delete policy -- no data loss or dangling state
- [ ] `internal/http/admin_test.go` -- cover create/update, company scoping, and delete policy outcomes -- protects the refactor

**Acceptance Criteria:**
- Given valid create data, when POST `/api/admin/usuario_admin`, then a new admin user is stored with bcrypt password hash.
- Given a user with references, when DELETE `/api/admin/usuario_admin/:id`, then the user is disabled instead of removed.
- Given a user without references, when DELETE `/api/admin/usuario_admin/:id`, then the user is removed.

## Spec Change Log

## Design Notes

The handler should stay thin: normalize input, enforce company access, and delegate persistence and delete policy to the store so the rule stays testable.

## Verification

**Commands:**
- `go test ./internal/http/...` -- expected: CRUD and policy tests pass
- `go test ./...` -- expected: repo remains green

## Suggested Review Order

**CRUD entry points**

- Usuario admin list/create/update/delete handlers
  [`admin.go:147`](../../internal/http/admin.go#L147)

- Empresa-scoped storage and delete policy
  [`admin_user.go:172`](../../internal/storage/admin_user.go#L172)

**Coverage**

- Create/update and dependency behavior tests
  [`admin_test.go:256`](../../internal/http/admin_test.go#L256)

- Delete policy tests for hard delete and disable
  [`admin_test.go:209`](../../internal/http/admin_test.go#L209)
