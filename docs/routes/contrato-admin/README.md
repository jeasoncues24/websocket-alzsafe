# Contrato Admin — Referencia de API

Endpoints del panel de administración de wsapi. Permiten gestionar empresas, usuarios, teléfonos, API Keys y monitorear el sistema.

**Autenticación:** JWT de admin obtenido via `POST /api/auth/login`, enviado como `Authorization: Bearer <ADMIN_JWT>`.

Base URL: `https://tu-dominio.com`

---

## Índice

| Archivo | Endpoints cubiertos |
|---------|---------------------|
| [autenticacion.md](autenticacion.md) | `POST /api/auth/login`, `logout`, `refresh`, `GET /api/auth/me` |
| [empresas.md](empresas.md) | CRUD `/api/admin/empresas`, generación y revocación de JWT empresa |
| [telefonos.md](telefonos.md) | CRUD `/api/admin/telefonos`, conexión WS admin, webhooks por teléfono |
| [api-keys.md](api-keys.md) | CRUD, rotación, revocación y auditoría de API Keys |
| [usuarios.md](usuarios.md) | CRUD `/api/admin/users` y `/api/admin/usuario_admin`, módulos |
| [roles.md](roles.md) | CRUD `/api/admin/roles`, listado de módulos |
| [mensajes.md](mensajes.md) | `GET /api/admin/mensajes`, reintento admin |
| [sesiones.md](sesiones.md) | `GET/POST /api/admin/sesiones`, diagnóstico, QR-link |
| [metricas.md](metricas.md) | Dashboard métricas, difusiones admin, búsqueda de clientes, `/metrics`, `/health` |

---

## Roles del sistema

| Rol | Descripción |
|-----|-------------|
| `super_admin` | Acceso total, incluyendo generación/revocación de JWT de empresa |
| `admin` | Gestión general sin acceso a operaciones de token de empresa |
