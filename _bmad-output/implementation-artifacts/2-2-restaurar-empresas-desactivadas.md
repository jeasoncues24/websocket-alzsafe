---
title: 'Story 2.2 — Restaurar empresas desactivadas (toggle activo)'
type: 'feature'
created: '2026-05-04'
status: 'review'
epic: 'epic-2-mejoras-post-revision'
baseline_commit: '69509dc'
context:
  - '{project-root}/_bmad-output/project-context.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** El flujo de empresas solo permite soft-delete (`activo = FALSE`). Una vez desactivada, no existe forma de recuperar la empresa desde el panel: ni endpoint en el backend ni acción en el frontend. Esto obliga a intervención directa en la BD.

**Approach:** Agregar un endpoint `POST /api/admin/empresas/{id}/restore` en el backend que reactive la empresa (`activo = TRUE`) y un botón "Restaurar" en la tabla del frontend que aparece solo cuando `activo === false`. El cambio es acotado: un método en el store, un handler, una ruta, una función en `lib/api.ts`, y un botón condicional en la tabla.

## Boundaries & Constraints

**Always:** solo usuarios con JWT admin válido pueden restaurar; validar que la empresa exista antes de restaurar; devolver la empresa actualizada en la respuesta; el botón "Restaurar" solo aparece en filas con `activo === false`; el botón "Eliminar" solo aparece en filas con `activo === true`.

**Never:** modificar la lógica de `Delete` existente; restaurar empresas que no existen (404); restaurar sin autenticación; tocar migraciones ni esquema de BD.

**Ask First:** si al restaurar la empresa deben reactivarse también sus teléfonos y api_keys asociados (fuera del alcance actual; solo reactivar la empresa).

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Restaurar empresa activa | `POST /api/admin/empresas/1/restore` con empresa ya activa | 200 OK, empresa devuelta con `activo: true` (idempotente) | No retornar error; operación idempotente |
| Restaurar empresa inexistente | `POST /api/admin/empresas/999/restore` | 404 Not Found | JSON `{"ok": false, "error": "empresa no encontrada"}` |
| Sin autenticación | `POST` sin JWT | 401 Unauthorized | Respuesta estándar del middleware |
| Admin no-root restaura empresa ajena | JWT con `empresa_id` distinto | 403 Forbidden | Mismo control que Delete |
| Frontend: empresa inactiva | Fila con `activo: false` | Muestra botón "Restaurar" (verde/outline), no muestra botón "Eliminar" | N/A |
| Frontend: empresa activa | Fila con `activo: true` | Muestra botón "Eliminar" (rojo/destructive), no muestra botón "Restaurar" | N/A |
| Restaurar desde frontend | Click en "Restaurar" → confirm → POST | Lista recarga; empresa aparece como activa | Mostrar alerta de error si falla la llamada |

</frozen-after-approval>

## Code Map

**Backend:**
- `backend/internal/storage/empresa.go` — agregar método `Restore(id int64) error` (UPDATE activo=TRUE)
- `backend/internal/domain/empresa_filter.go` o `backend/internal/domain/empresa.go` — agregar `Restore` a la interfaz `EmpresaStoreInterface` si existe
- `backend/internal/http/handlers/companies.go` — agregar handler `Restore(w, r)`
- `backend/internal/http/routes_admin.go` — registrar `POST /api/admin/empresas/{id}/restore`

**Frontend:**
- `frontend/lib/api.ts` — agregar función `restoreEmpresa(id: number)`
- `frontend/app/empresas/page.tsx` — botón condicional "Restaurar" en filas inactivas; botón "Eliminar" solo en filas activas

## Tasks & Acceptance

**Execution:**

- [x] `backend/internal/storage/empresa.go` — agregar método `Restore`:
  ```go
  func (s *EmpresaStore) Restore(id int64) error {
      _, err := s.db.Exec(`UPDATE empresas SET activo = TRUE, updated_at = NOW() WHERE id = ?`, id)
      return err
  }
  ```

- [x] `backend/internal/domain/` — verificar si `EmpresaStoreInterface` declara los métodos del store; si existe la interfaz, agregar `Restore(id int64) error`

- [x] `backend/internal/http/handlers/companies.go` — agregar handler `Restore`:
  - Autenticación: mismo guard que `Delete` (JWT admin, control de empresa para no-root)
  - Extraer `id` del path con `h.extractIDFromPath`
  - Verificar que la empresa existe con `h.empresaStore.GetByID(id)`; 404 si no existe
  - Llamar `h.empresaStore.Restore(id)`
  - Devolver empresa actualizada con `GetByID` + `writeJSON(200, {ok: true, empresa: ...})`

- [x] `backend/internal/http/routes_admin.go` — registrar la ruta nueva:
  ```go
  mux.Handle("POST /api/admin/empresas/{id}/restore", adminStack(http.HandlerFunc(c.CompaniesHandler.Restore)))
  ```

- [x] `frontend/lib/api.ts` — agregar función:
  ```ts
  export async function restoreEmpresa(id: number): Promise<{ ok: boolean; empresa: Empresa }> {
    return fetchAdmin(`/api/admin/empresas/${id}/restore`, { method: "POST" })
  }
  ```

- [x] `frontend/app/empresas/page.tsx` — modificar la columna de acciones:
  - Importar `restoreEmpresa` desde `lib/api`
  - Agregar estado `restoringId: number | null`
  - Mostrar botón "Restaurar" (variante `outline`, color verde, ícono `RotateCcw` de lucide) cuando `empresa.activo === false`
  - Mostrar botón "Eliminar" solo cuando `empresa.activo === true`
  - Al confirmar restauración: llamar `restoreEmpresa`, recargar lista, mostrar feedback inline

**Acceptance Criteria:**

- Given una empresa con `activo = false`, when el admin hace click en "Restaurar" y confirma, then la empresa aparece como activa en la tabla sin recargar la página completa.
- Given `POST /api/admin/empresas/{id}/restore` con empresa válida, when se llama, then responde `200 { ok: true, empresa: { ... activo: true } }`.
- Given `POST /api/admin/empresas/{id}/restore` con id inexistente, when se llama, then responde `404`.
- Given una fila con `activo: true`, when se visualiza la tabla, then no aparece el botón "Restaurar" en esa fila.
- Given una fila con `activo: false`, when se visualiza la tabla, then no aparece el botón "Eliminar" en esa fila.
- Given el binario compilado, when se ejecuta `go build ./...` en `backend/`, then compila sin errores.

## Spec Change Log

_(vacío al crear)_

## Design Notes

La operación es idempotente: restaurar una empresa ya activa no debe fallar, solo confirmar el estado. El soft-delete existente (`EmpresaStore.Delete`) no se toca. El control de acceso para no-root se aplica igual que en `Delete`: un admin con `empresa_id` en el JWT solo puede operar sobre su propia empresa.

Los teléfonos y api_keys de la empresa restaurada quedan en el estado en que estaban al desactivarse; reactivarlos está fuera del alcance de esta story.

## Verification

**Commands:**
- `go build ./...` desde `backend/` — expected: sin errores de compilación
- `curl -X POST http://localhost:8083/api/admin/empresas/1/restore -H "Authorization: Bearer <jwt>"` — expected: `{"ok":true,"empresa":{...,"activo":true}}`
- `grep -n "Restore\|restore" backend/internal/storage/empresa.go` — expected: método presente
- `grep -n "restoreEmpresa\|RotateCcw" frontend/app/empresas/page.tsx` — expected: función importada y botón presente

## Suggested Review Order

**Backend — store:**
- Verificar el método `Restore` y que la query es correcta.
  [`backend/internal/storage/empresa.go`](../../backend/internal/storage/empresa.go)

**Backend — handler y ruta:**
- Verificar el handler y que la ruta está registrada con el stack de auth correcto.
  [`backend/internal/http/handlers/companies.go`](../../backend/internal/http/handlers/companies.go)
  [`backend/internal/http/routes_admin.go`](../../backend/internal/http/routes_admin.go)

**Frontend:**
- Verificar el botón condicional y la función `restoreEmpresa`.
  [`frontend/app/empresas/page.tsx`](../../frontend/app/empresas/page.tsx)
  [`frontend/lib/api.ts`](../../frontend/lib/api.ts)
