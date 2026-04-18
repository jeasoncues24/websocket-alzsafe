# S-8.1: Testing CRUD Empresas (Endpoints Backend)

## Objetivo

Validar que todos los endpoints de gestión de empresas funcionen correctamente.

## Endpoints Panel Admin (`/api/admin/empresas`)

| # | Método | Endpoint | Descripción |
|---|--------|----------|-------------|
| 1 | `GET` | `/api/admin/empresas` | Listar empresas (paginado, búsqueda, filtro activo) |
| 2 | `GET` | `/api/admin/empresas/{id}` | Obtener empresa por ID |
| 3 | `POST` | `/api/admin/empresas` | Crear empresa |
| 4 | `PUT` | `/api/admin/empresas/{id}` | Actualizar empresa |
| 5 | `DELETE` | `/api/admin/empresas/{id}` | Eliminar empresa (soft delete) |
| 6 | `POST` | `/api/admin/empresas/{id}/token` | Generar token de acceso |
| 7 | `POST` | `/api/admin/empresas/{id}/token/revoke` | Revocar token |

## Endpoints Cliente (`/api/empresas`)

| # | Método | Endpoint | Descripción |
|---|--------|----------|-------------|
| 1 | `GET` | `/api/empresas` | Obtener su propia empresa |
| 2 | `PUT` | `/api/empresas` | Actualizar su empresa (RUC es readonly) |

## Requisitos Previos

- Backend corriendo en `http://localhost:8080`
- Usuario admin autenticado (cookie de sesión para admin)
- Token de empresa (para routes de cliente)

## Plan de Testing

### Panel Admin - GET /api/admin/empresas

Casos:
- [ ] Sin parámetros → retorna lista paginada
- [ ] Con `?page=2&limit=5` → paginado correcto
- [ ] Con `?busqueda=empresa` → búsqueda por nombre/RUC
- [ ] Con `?estado=activo` → filtro solo activas
- [ ] Con `?estado=inactivo` → filtro solo inactivas

### Panel Admin - POST /api/admin/empresas

Request:
```json
{
  "ruc": "20100000001",
  "nombre": "Empresa Test SPA",
  "nombre_comercial": "TestCorp",
  "telefono": "+51999999999",
  "direccion": "Av. Test 123"
}
```

Casos:
- [ ] Crear empresa válida → 201 + datos creados
- [ ] RUC duplicado → 409
- [ ] RUC vacío → 400
- [ ] Nombre vacío → 400

### Panel Admin - POST /api/admin/empresas/{id}/token

Casos:
- [ ] Generar token para empresa → retorna token JWT
- [ ] Empresa inexistente → 404
- [ ] Empresa inactiva → 409

### Panel Admin - POST /api/admin/empresas/{id}/token/revoke

Casos:
- [ ] Revocar token → token_version incrementa
- [ ] Siguiente request con token anterior → 401

### Cliente - GET /api/empresas

Casos:
- [ ] Con token válido → retorna empresa propia
- [ ] Sin token → 401

### Cliente - PUT /api/empresas

Casos:
- [ ] Actualizar nombre → OK
- [ ] Intentar cambiar RUC → ignorado (RUC es readonly)

## Checklist

- [ ] Todos los endpoints responden correctamente
- [ ] Manejo de errores apropiado (400, 404, 409)
- [ ] Validación de datos de entrada funciona
- [ ] Tokens se generan y revokan correctamente
- [ ] Cliente no puede cambiar RUC