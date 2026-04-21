# Story 8.6.1: Migracion de API client y navegacion admin

Status: done

## Story

As a panel admin user,
I want the frontend API client and navigation to match the new admin contracts,
so that the UI points to the right backend routes and naming.

## Acceptance Criteria

1. [AC: api-client] `frontend/lib/api.ts` uses `/api/admin/usuario_admin`, `/api/admin/roles`, and `/api/admin/modules` for the admin panel.
2. [AC: route-rename] The admin navigation points to `/usuario_admin` instead of `/users_admin`.
3. [AC: naming-alignment] Frontend admin labels use `usuario_admin` where the screen or function name is user-facing.
4. [AC: token-flow] The admin panel keeps using `admin_token` for Authorization headers.
5. [AC: no-phone-token] No phone-token endpoints are introduced or modified in this story.

## Tasks / Subtasks

- [ ] Update `frontend/lib/api.ts` method names and endpoint URLs for the admin client (AC: 1, 4)
- [ ] Move or rename the user admin page route to `/usuario_admin` (AC: 2, 3)
- [ ] Update the sidebar navigation labels and links for the admin section (AC: 2, 3)
- [ ] Keep the admin auth guard using `admin_token` unchanged (AC: 4)
- [ ] Verify the refactor does not touch phone-token flows (AC: 5)

## Dev Notes

- Current mismatch to remove: `frontend/lib/api.ts` still points to `/api/admin/users` and `frontend/components/layout/sidebar.tsx` still links to `/users_admin`.
- This story is the integration seam between the new backend admin contract and the existing panel UI.
- References: [Source: frontend/lib/api.ts], [Source: frontend/components/layout/sidebar.tsx], [Source: frontend/app/users_admin/page.tsx]

### Project Structure Notes

- Likely touch points: `frontend/lib/api.ts`, `frontend/components/layout/sidebar.tsx`, and the user admin route folder under `frontend/app/`.
- Keep the token storage flow intact; only the endpoint contract and admin route naming change here.

### References

- [Source: frontend/lib/api.ts]
- [Source: frontend/components/layout/sidebar.tsx]
- [Source: frontend/app/users_admin/page.tsx]

## Dev Agent Record

### Agent Model Used

BMAD Create Story

### Debug Log References

### Completion Notes List

### File List
