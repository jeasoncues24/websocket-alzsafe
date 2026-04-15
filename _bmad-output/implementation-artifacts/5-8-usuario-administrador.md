# Story 5-8: Usuario Administrador

Status: ready-for-dev

## Story

As a operations manager,
I want tener un usuario administrador para acceder al panel,
So that pueda probar el sistema completo.

## Acceptance Criteria

1. **Given** la base de datos, **When** corre migración, **Then** crea tabla `admin_users`.

2. **Given** el usuario admin, **When** existe, **Then** puede hacer login con credenciales por defecto.

3. **Given** el login, **When** es exitoso, **Then** redirige al dashboard.

## Tasks

- [x] Crear migración 004 para tabla admin_users
- [x] Crear usuario admin por defecto
- [x] Actualizar sprint status

## Dev Notes

- Username: `admin`
- Password: `admin123`
- Tabla: `admin_users`
- Roles: admin, operator, viewer