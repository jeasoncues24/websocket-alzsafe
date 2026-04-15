---
stepsCompleted:
  - validate-prerequisites
  - design-epics
  - create-stories
  - final-validation
inputDocuments:
  - _bmad-output/planning-artifacts/epics.md
status: "ready-for-dev"
---

# wsapi - Epic Breakdown: Empresa, Seguridad y Mensajería Enriquecida

## Overview

Este documento contiene los epics y stories para implementar:
- Gestión de empresas desde panel admin
- Seguridad con JWT para autenticación de usuarios
- Endpoints de mensajería enriquecidos con contexto (usuario, empresa, sesión)
- Endpoint /usuario/me para información del usuario actual

---

## Requirements Inventory

### Functional Requirements

- FR-11: Autenticación de usuarios admin con JWT
- FR-12: Middleware de protección para endpoints de empresas
- FR-13: Middleware de protección para endpoints de sesiones WhatsApp
- FR-14: Middleware de protección para endpoints de mensajes
- FR-15: CRUD de empresas desde panel admin
- FR-16: Asignación de sesión WhatsApp por empresa
- FR-17: Endpoint /usuario/me con información del usuario autenticado
- FR-18: Payload enriquecido en respuestas de mensajes (usuario_id, ruc_empresa, session_id)
- FR-19: Aislamiento de datos por empresa (cada empresa solo ve sus datos)
- FR-20: Rate limiting por empresa

### Non-Functional Requirements

- NFR-09: Tokens JWT con expiry configurable
- NFR-10: HTTPS obligatorio en producción
- NFR-11: Auditoría de acciones por usuario y empresa

---

## FR Coverage Map

| FR | Epic | Descripción |
|-----|------|-------------|
| FR-11 | Epic 7 | Autenticación JWT |
| FR-12 | Epic 7 | Middleware empresas |
| FR-13 | Epic 7 | Middleware sesiones |
| FR-14 | Epic 7 | Middleware mensajes |
| FR-15 | Epic 8 | CRUD empresas |
| FR-16 | Epic 8 | Asignación sesión |
| FR-17 | Epic 9 | Endpoint /usuario/me |
| FR-18 | Epic 9 | Payload enriquecido |
| FR-19 | Epic 9 | Aislamiento por empresa |
| FR-20 | Epic 9 | Filtros por empresa |

---

## Epic List

- Epic 7: Autenticación y Seguridad JWT
- Epic 8: Gestión de Empresas y Aprovisionamiento
- Epic 9: Mensajería Enriquecida con Contexto

---

# Epic 7: Autenticación y Seguridad JWT

Implementar sistema de autenticación robusto con JWT para proteger todos los endpoints del sistema.

## Story 7.1: Sistema de autenticación JWT con login y logout

**Given** un usuario administrador del sistema
**When** envía solicitud POST a /api/auth/login con username y password
**Then** el sistema valida credenciales contra tabla admin_users
**And** si son válidas, genera JWT con claims: usuario_id, username, rol, empresa_id, empresa_nombre
**And** retorna token con expiry configurable (default 24h)

**Given** un usuario autenticado
**When** envía solicitud POST a /api/auth/logout
**Then** el token se agrega a blacklist en Redis o DB
**And** retorna confirmación de logout exitoso

**Given** un JWT válido
**When** el token está por expirar (menos de 1 hora)
**Then** el endpoint /api/auth/refresh retorna nuevo token
**And** el token anterior se invalida

---

## Story 7.2: Schema de base de datos para usuarios y API keys

**Given** el sistema necesita almacenar usuarios admin
**When** se ejecutan migraciones
**Then** crea tabla `admin_users`:
- id (INT, PK, AUTO_INCREMENT)
- username (VARCHAR(50), UNIQUE, NOT NULL)
- password_hash (VARCHAR(255), NOT NULL)
- empresa_id (INT, FK, nullable para super_admin)
- rol (ENUM: 'super_admin', 'admin', 'operador')
- activo (BOOLEAN, DEFAULT TRUE)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)

**Given** el sistema necesita almacenar API keys
**When** se ejecutan migraciones
**Then** crea tabla `api_keys`:
- id (INT, PK, AUTO_INCREMENT)
- empresa_id (INT, FK, NOT NULL)
- key_hash (VARCHAR(255), NOT NULL)
- nombre (VARCHAR(100))
- activo (BOOLEAN, DEFAULT TRUE)
- created_at (TIMESTAMP)
- expires_at (TIMESTAMP, nullable)

**Given** el sistema necesita almacenar empresas
**When** se ejecutan migraciones
**Then** crea tabla `empresas`:
- id (INT, PK, AUTO_INCREMENT)
- ruc (VARCHAR(11), UNIQUE, NOT NULL)
- nombre (VARCHAR(255), NOT NULL)
- nombre_comercial (VARCHAR(255))
- telefono (VARCHAR(20))
- direccion (TEXT)
- activo (BOOLEAN, DEFAULT TRUE)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)

---

## Story 7.3: Middleware de protección para endpoints de empresas

**Given** una request a /api/companies
**When** no incluye header Authorization con JWT válido
**Then** retorna HTTP 401 con mensaje "Token requerido"
**And** no ejecuta la lógica del handler

**Given** un JWT válido con empresa_id
**When** accede a GET /api/companies/{id}
**Then** verifica que la empresa solicitada pertenezca al usuario
**And** si no pertenece, retorna HTTP 403 con mensaje "Acceso denegado a esta empresa"

**Given** un usuario con rol "super_admin"
**When** accede a cualquier endpoint de empresas
**Then** puede acceder a todas las empresas del sistema
**And** el middleware permite el acceso sin restricciones de empresa

---

## Story 7.4: Middleware de protección para endpoints de sesiones y mensajes

**Given** una request a /api/sessions/*, /api/message/*, o /api/broadcast/*
**When** no incluye JWT válido
**Then** retorna HTTP 401 con mensaje "Token requerido"

**Given** un JWT válido
**When** intenta acceder a recursos de otra empresa (distinta a su empresa_id en token)
**Then** retorna HTTP 403 Forbidden

**Given** un JWT válido con empresa_id null (super_admin)
**When** accede a endpoints de sesiones o mensajes
**Then** puede operar con todas las empresas
**And** debe incluir header X-Empresa-ID para especificar la empresa objetivo

---

## Story 7.5: Hash de passwords con bcrypt

**Given** un usuario admin configurado inicialmente
**When** se crea el usuario, su password se hashea con bcrypt (cost 12)
**And** se almacena solo el hash en la base de datos

**Given** un usuario intenta hacer login
**When** envía su password en texto plano
**Then** el sistema usa bcrypt.Compare para verificar contra el hash almacenado
**And** retorna error si no coincide

**Given** un usuario quiere cambiar su password
**When** envía POST /api/auth/password con password actual y nuevo
**Then** valida password actual
**And** si es correcta, hashea el nuevo y actualiza en la base de datos

---

# Epic 8: Gestión de Empresas y Aprovisionamiento

Permitir la gestión completa de empresas desde el panel administrativo.

## Story 8.1: CRUD de empresas desde API

**Given** un usuario admin autenticado
**When** envía POST /api/companies con datos: ruc, nombre, nombre_comercial, telefono
**Then** crea empresa en estado activo
**And** retorna empresa creada con ID y código HTTP 201

**Given** un usuario admin autenticado
**When** envía GET /api/companies sin parámetros
**Then** retorna lista de empresas accesibles para su usuario
**And** si es super_admin retorna todas las empresas
**And** si es admin/operador retorna solo su empresa

**Given** un usuario admin autenticado
**When** envía GET /api/companies?busqueda=term&estado=activo
**Then** filtra empresas por nombre/ruc que contengan "term"
**And** filtra por estado (activo/inactivo)

**Given** un usuario admin autenticado
**When** envía PUT /api/companies/{id} con datos a actualizar
**Then** actualiza solo los campos enviados
**And** valida que la empresa pertenezca a su empresa_id (o es super_admin)
**And** retorna empresa actualizada

**Given** un usuario admin autenticado
**When** envía DELETE /api/companies/{id}
**Then** si la empresa tiene sesiones WhatsApp activas, retorna error 409
**And** si no tiene sesiones, marca empresa como inactiva (soft delete)
**And** retorna HTTP 204

---

## Story 8.2: Asignación de sesión WhatsApp por empresa

**Given** una empresa creada en el sistema
**When** se solicita iniciar sesión WhatsApp para esa empresa
**Then** la sesión se asocia a esa empresa específico (ruc_empresa)
**And** el WebSocket usa el empresa_id del JWT para identificar la empresa

**Given** múltiples empresas en el sistema
**When** cada empresa inicia su propia sesión
**Then** el sistema aísla correctamente las sesiones por ruc_empresa
**And** no hay fuga de datos entre empresas

**Given** un super_admin
**When** inicia sesión WhatsApp
**Then** debe especificar empresa_id en el request
**And** la sesión se crea para esa empresa específica

---

## Story 8.3: Panel admin para gestión de empresas (Frontend)

**Given** un usuario autenticado en el panel
**When** accede a /companies
**Then** muestra lista de empresas con columnas: RUC, Nombre, Estado, Sesión WhatsApp
**And** incluye búsqueda por nombre o RUC
**And** incluye filtros por estado (activo/inactivo)

**Given** un usuario autenticado
**When** hace clic en "Nueva Empresa"
**Then** muestra modal/formulario con campos: RUC, Nombre, Teléfono, Dirección
**And** valida que RUC no exista previamente

**Given** un usuario autenticado
**When** hace clic en Editar empresa
**Then** muestra formulario con datos actuales
**And** permite modificar nombre, teléfono, dirección, estado

**Given** un usuario autenticado
**When** hace clic en Ver detalle de empresa
**Then** muestra panel con: datos de empresa, sesión WhatsApp actual, últimos mensajes

---

## Story 8.4: Migraciones para esquema de empresas y usuarios

**Given** el sistema necesita las tablas de seguridad
**When** se ejecutan migraciones
**Then** ejecuta en orden:
1. 005_create_empresas_table.up.sql
2. 006_create_admin_users_table.up.sql
3. 007_create_api_keys_table.up.sql
4. 005_create_empresas_table.down.sql (revertir)
5. 006_create_admin_users_table.down.sql (revertir)
6. 007_create_api_keys_table.down.sql (revertir)

**Given** migración hacia adelante
**When** se completa exitosamente
**Then** las tablas existen con los índices correspondientes
**And** foreign keys están configuradas correctamente

---

# Epic 9: Mensajería Enriquecida con Contexto

Enriquecer los endpoints de mensajes con información contextual del usuario, empresa y sesión.

## Story 9.1: Endpoint /usuario/me

**Given** un usuario autenticado con JWT válido
**When** envía GET /api/usuario/me
**Then** retorna información del usuario:
- id (del token)
- username (del token)
- rol (del token)
- empresa: { id, ruc, nombre } (del token)
- permisos: array de permisos basados en rol

**Response ejemplo:**
```json
{
  "id": 1,
  "username": "admin",
  "rol": "super_admin",
  "empresa": {
    "id": 1,
    "ruc": "20123456789",
    "nombre": "Empresa Demo"
  },
  "permisos": ["companies:read", "companies:write", "messages:read", "messages:write"]
}
```

**Given** un usuario sin empresa asociada (super_admin sin empresa asignada)
**When** invoca GET /api/usuario/me
**Then** retorna empresa: null
**And** puede operar con todas las empresas

---

## Story 9.2: Payload enriquecido en respuestas de mensajes

**Given** una respuesta de GET /api/message/send
**When** se retorna información del mensaje
**Then** incluye campos enriquecidos:
- usuario_id: ID del usuario que ejecutó la acción
- ruc_empresa: RUC de la empresa asociada
- empresa_nombre: Nombre de la empresa
- session_id: ID de la sesión WhatsApp (si existe)

**Given** una respuesta de GET /api/messages
**When** se retorna lista de mensajes
**Then** cada mensaje incluye los campos enriquecidos anteriores

**Given** una respuesta de POST /api/broadcast/send
**When** se retorna información del broadcast creado
**Then** incluye ruc_empresa y empresa_nombre en el broadcast

**Given** una respuesta de GET /api/broadcast/{id}/results
**When** se retornan resultados por destinatario
**Then** cada resultado incluye ruc_empresa del broadcast

---

## Story 9.3: Aislamiento de datos por empresa en queries

**Given** un usuario con empresa_id específica
**When** ejecuta cualquier query de mensajes
**Then** el sistema filtra automáticamente por empresa_id del token JWT
**And** no permite acceso a datos de otras empresas

**Given** un usuario con empresa_id específica
**When** ejecuta query de sesiones WhatsApp
**Then** filtra por su empresa_id
**And** solo ve su sesión de WhatsApp

**Given** un usuario con rol "super_admin" y empresa_id null
**When** ejecuta queries
**Then** puede acceder a datos de todas las empresas
**And** el payload incluye ruc_empresa para identificar origen

**Given** un super_admin
**When** quiere acceder a datos de empresa específica
**Then** puede usar header X-Empresa-ID para especificar la empresa
**And** las queries se ejecutan para esa empresa específica

---

## Story 9.4: Historial de mensajes por empresa con filtros

**Given** un usuario autenticado
**When** envía GET /api/messages?empresa_id=1&desde=2024-01-01&hasta=2024-01-31&telefono=+51
**Then** filtra mensajes por empresa (si no es super_admin, ignora empresa_id del filtro)
**And** filtra por rango de fechas
**And** filtra por teléfono destino (búsqueda parcial)
**And** retorna mensajes con paginación (page, limit)

**Given** un usuario autenticado
**When** envía GET /api/messages sin filtros
**Then** retorna todos los mensajes de su empresa (o todas si es super_admin)
**And** incluye paginación por defecto: page=1, limit=50

**Given** un usuario autenticado
**When** envía GET /api/messages?page=2&limit=20
**Then** retorna la segunda página con 20 resultados
**And** incluye metadatos: total, page, limit, total_pages

---

## Story 9.5: Dashboard con métricas por empresa

**Given** un usuario autenticado
**When** accede a GET /api/dashboard/metricas
**Then** retorna métricas de su empresa:
- total_mensajes: cantidad de mensajes enviados
- mensajes_hoy: mensajes de hoy
- mensajes_semana: mensajes de los últimos 7 días
- mensajes_exitosos: mensajes entregados exitosamente
- mensajes_fallidos: mensajes con error
- sesiones_activas: cantidad de sesiones WhatsApp activas
- broadcasts_ejecutados: broadcasts de los últimos 30 días

**Given** un super_admin
**When** accede a GET /api/dashboard/metricas
**Then** puede agregar parametro empresa_id para ver métricas de empresa específica
**And** sin parametro, retorna métricas agregadas de todas las empresas