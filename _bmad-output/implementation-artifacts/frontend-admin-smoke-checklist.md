# Frontend Admin Smoke Checklist

- Login with `admin_token` and open `/dashboard`.
- Open `/usuario_admin` and verify search, pagination, create, edit, delete, and module assignment.
- Open `/roles` and verify create, edit, delete-blocked, and permissions selection.
- Open `/modules` and confirm read-only catalog copy and slugs.
- Delete a user with dependencies and confirm the UI reports disabled vs deleted.
- Delete a role in use and confirm the UI explains the conflict.
