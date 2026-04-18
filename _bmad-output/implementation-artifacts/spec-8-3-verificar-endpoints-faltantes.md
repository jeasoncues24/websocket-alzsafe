# S-8.3: Verificar/Implementar Endpoints de Empresas

## Estado: done

## Objetivo

Implementar los endpoints de CRUD de empresas para:
1. Panel Admin (`/api/admin/empresas`)
2. Cliente con token acceso (`/api/empresas`)

---

## Endpoints Requeridos

### Panel Admin (`/api/admin/empresas`)

| # | Método | Endpoint | Descripción |
|---|--------|----------|-------------|
| 1 | `GET` | `/api/admin/empresas` | Listar empresas (paginado, búsqueda, filtro ativo) |
| 2 | `GET` | `/api/admin/empresas/{id}` | Obtener empresa por ID |
| 3 | `POST` | `/api/admin/empresas` | Crear empresa |
| 4 | `PUT` | `/api/admin_empresas/{id}` | Actualizar empresa |
| 5 | `DELETE` | `/api/admin/empresas/{id}` | Eliminar empresa (soft delete) |
| 6 | `POST` | `/api/admin/empresas/{id}/token` | Generar token de acceso |
| 7 | `POST` | `/api/admin/empresas/{id}/token/revoke` | Revocar token |

### Cliente (`/api/empresas`)

| # | Método | Endpoint | Descripción |
|---|--------|----------|-------------|
| 1 | `GET` | `/api/empresas` | Obtener su propia empresa |
| 2 | `PUT` | `/api/empresas` | Actualizar su empresa (RUC es readonly) |

---

## Notas de Diseño

- RUC es **solo lectura** para el cliente (no actualizable)
- El cliente solo puede ver/actualizar SU empresa
- Panel admin tiene acceso completo a todas las empresas

---

## AC

- [x] `/api/admin/empresas` CRUD funciona completamente
- [x] `/api/empresas` GET retorna solo la empresa del cliente
- [x] `/api/empresas` PUT no permite cambiar RUC
- [x] Token generation/revoke funciona desde admin
- [x] Handlers compilan correctamente

## Notas

Rutas implementadas:
- `/api/admin/empresas` (GET list, POST create)
- `/api/admin/empresas/{id}` (GET, PUT, DELETE)
- `/api/admin/empresas/{id}/token` (POST generate)
- `/api/admin/empresas/{id}/token/revoke` (POST revoke)
- `/api/empresas` (GET, PUT - solo propia)
