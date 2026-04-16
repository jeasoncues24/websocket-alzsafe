# RESUMEN: Epic 6 - Sistema de Autenticación JWT por Empresa + Multi-Tenant

## 🔑 Idea Central

Sistema de dos autentificaciones paralelas:
1. **Admin** (existente, NO modificar) → `/api/admin/*`
2. **Empresa** (NUEVO) → `/v1/*`

Cada empresa tiene su propio token JWT (5 años). Las empresas pueden gestionar sus números de WhatsApp desde su propio sistema.

---

## 🏗️ Arquitectura Multi-Tenant

### Modelo de Datos

| Tabla: empresas | |
|---|---|
| id | **PK (INT)** |
| ruc | RUC único |
| nombre | Nombre |
| **token_version** | Para revocación |
| permissions | ["send", "broadcast", "sessions"] |
| created_at | TIMESTAMP |

| Tabla: telefonos (NUEVA) | |
|---|---|
| id | **PK (INT)** |
| empresa_id | **FK → empresas.id** |
| codigo_pais | +51 |
| numero | 999999999 |
| **numero_completo** | +519999999999 (COMPUTED) |
| status | active / qr_pending / disconnected |
| session_data | LONGBLOB |
| qr_string | TEXT |
| last_connected | TIMESTAMP |
| created_at | TIMESTAMP |

> **Una empresa puede tener N teléfonos (múltiples números)**

---

## 🗑️ ELIMINACIÓN COMPLETA DE ENDPOINTS LEGACY

Eliminar estos endpoints públicos (ya no existen):

```
/message           → ELIMINADO
/messages          → ELIMINADO
/broadcast         → ELIMINADO
/broadcast/{id}    → ELIMINADO
/metrics           → ELIMINADO
/companies        → ELIMINADO
/admin/*           → ELIMINADO (usar /api/admin/*)
/sessions          → ELIMINADO
```

También eliminar duplicados bajo `/api/*` que no respeten el nuevo esquema.

---

## 🔐 Autenticación - Dos Sistemas Paralelos

### 1. ADMIN (EXISTE - NO MODIFICAR)

| Sistema | Valor |
|--------|-------|
| Endpoints | `/api/admin/*` |
| Tipo | JWT sesión (short-lived) |
| Header | `Authorization: Bearer <JWT_ADMIN>` |
| Tablas | admin_users, roles, modules |
| Permisos | Full access |

### 2. API EMPRESA (NUEVO)

| Sistema | Valor |
|--------|-------|
| Endpoints | `/v1/*` |
| Tipo | JWT empresa (long-lived: 5 años) |
| Header | `Authorization: Bearer <JWT_EMPRESA>` |
| Permisos | Restringido por empresa_id |

### JWT API EMPRESA (nuevo formato simplificado)
```json
{
  "sub": "1",           // empresa_id
  "ver": 1,            // token_version
  "exp": 1769999999      // 5 años
}
```

---

## 🌐 NUEVA ESTRUCTURA ENDPOINTS `/v1/*`

### Solo públicos (sin token)
| Endpoint | Descripción |
|----------|------------|
| GET /health | Health check |

### Estado general
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/v1/status` | Estado del servicio (up/down, uptime, connections) |

### Teléfonos
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/v1/telefonos` | Listar teléfonos |
| GET | `/v1/telefonos/{id}` | Ver teléfono |

### Sesión WhatsApp
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | `/v1/telefonos/{id}/connect` | Solicitar conexión QR |
| POST | `/v1/telefonos/{id}/disconnect` | Desconectar |
| GET | `/v1/telefonos/{id}/status` | Ver estado (active/qr_pending/disconnected) |
| GET | `/v1/telefonos/{id}/qr` | Obtener QR string |

### Mensajes
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | `/v1/message` | Enviar mensaje |
| GET | `/v1/messages` | Listar mensajes |

### Difusiones
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | `/v1/broadcast` | Crear difusión |
| GET | `/v1/broadcast/{id}` | Ver difusión |

### WebSocket
| Endpoint | Descripción |
|----------|-------------|
| GET `/v1/ws` | WebSocket para tiempo real |

---

## 🌐 WEBSOCKET (`/v1/ws`)

### Endpoint
```
GET /v1/ws
```
Autenticación: JWT empresa (header o query param)

### Eventos WebSocket

```json
// QR disponible
{ "type": "qr", "phone_id": 1, "qr": "string" }

// Conectado
{ "type": "connected", "phone_id": 1 }

// Desconectado
{ "type": "disconnected", "phone_id": 1 }

// Estado de mensaje
{ "type": "message_status", "message_id": "abc", "status": "sent|delivered|read|failed" }

// Estado de servicio
{ "type": "service_status", "status": "up|down" }
```

---

## 🔄 FLUJO DE CONEXIÓN QR (COMPLETO)

### Paso 1: Solicitar conexión
```
POST /v1/telefonos/{id}/connect
```
Respuesta:
```json
{ "status": "qr_pending", "expires_in": 300 }
```

### Paso 2: Obtener QR
```
GET /v1/telefonos/{id}/qr
```
O vía WebSocket: evento `qr`

### Paso 3: Escaneo
Usuario escanea QR desde WhatsApp

### Paso 4: Activación
- status → `active`
- Guardar `session_data`
- Evento WS: `{ "type": "connected", "phone_id": 1 }`

---

## 📊 VALIDACIÓN DE ESTADO

### GET `/v1/telefonos/{id}/status`
```json
{
  "status": "active|qr_pending|disconnected",
  "last_connected": "timestamp"
}
```

### GET `/v1/status`
```json
{
  "service": "up|down",
  "uptime": "seconds",
  "connections": 10
}
```

---

## ⚠️ REGLAS CRÍTICAS

| Regla | Descripción |
|-------|-------------|
| JWT obligatorio | TODOS endpoints `/v1/*` requieren JWT empresa |
| Ownership | SIEMPRE validar `telefono_id` pertenece a empresa |
| Sin sesión activa | NO permitir operaciones sin status = active |
| QR expira | 300 segundos máximo |
| WS token | WebSocket cierra si token inválido |
| Duplicados | NO permitir endpoints duplicados |

```go
func validateV1Request(r *http.Request, claims *JWTClaims, telefonoID int64) error {
    // 1. JWT obligatorio
    if claims == nil {
        return ErrUnauthorized
    }
    
    // 2. Ownership
    telefono, _ := GetTelefono(telefonoID)
    if telefono.EmpresaID != claims.EmpresaID {
        return ErrForbidden
    }
    
    // 3. Sin sesión activa
    if telefono.Status != "active" && requiresActiveSession {
        return ErrSessionNotActive
    }
    
    return nil
}
```

---

## 📱 Migración Tablas Existentes

### messages
```sql
ALTER TABLE messages ADD COLUMN empresa_id INT NOT NULL;
ALTER TABLE messages ADD COLUMN telefono_id INT NOT NULL;
CREATE INDEX idx_messages_empresa ON messages(empresa_id);
CREATE INDEX idx_messages_telefono ON messages(telefono_id);
-- Eliminar ruc_empresa después de migración
```

### broadcasts
```sql
ALTER TABLE broadcasts ADD COLUMN empresa_id INT NOT NULL;
ALTER TABLE broadcasts ADD COLUMN telefono_id INT NOT NULL;
CREATE INDEX idx_broadcasts_empresa ON broadcasts(empresa_id);
-- Eliminar ruc_empresa después de migración
```

---

## 📋 NO Usar ruc_empresa

| Tabla | Campo Anterior | Campo Nuevo |
|-------|---------------|--------------|
| messages | ruc_empresa VARCHAR | empresa_id INT, telefono_id INT |
| broadcasts | ruc_empresa VARCHAR | empresa_id INT, telefono_id INT |

---

## ✅ VALIDACIÓN CRÍTICA (endpoints /v1/*)

```
1. Validar JWT empresa
2. empresa_id = claims.sub
3. obtener telefono por phone_id
4. validar telefono.empresa_id == empresa_id  ← OWNERSHIP
5. continuar
```

---

## 🔄 Endpoints Admin (`/api/*`) - SIN CAMBIOS

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET/POST | /api/empresas | CRUD empresas |
| GET/PUT/DELETE | /api/empresas/{id} | |
| GET/POST | /api/empresas/{id}/telefonos | CRUD teléfonos |
| POST | /api/empresas/{id}/token | Generar token |
| POST | /api/empresas/{id}/revoke | Revocar token |
| GET/POST | /api/admin/users | Usuarios admin |
| GET | /api/admin/roles | Roles |

---

## 🔄 Flujo de Revocación

```
1. Admin POST /api/empresas/1/revoke
2. empresa.token_version++ → 2
3. Cache invalidado
4. Token anterior → 401 "Token revoked"
```

---

## 🔄 Estados Sesión WhatsApp

| Estado | Descripción |
|--------|-------------|
| qr_pending | Esperando escaneo QR |
| active | Conectado y funcionando |
| disconnected | Sesión cayera o desconectada |

---

## ✅ Success Criteria

- [ ] TODOS los endpoints requieren JWT (excepto /health)
- [ ] Una empresa puede tener múltiples números
- [ ] Empresa puede autoconnectarse sin admin
- [ ] Revocación instantánea
- [ ] Isolation de datos garantizado (empresa_id owns telefono_id)
- [ ] NO se usa ruc_empresa como FK
- [ ] Admin JWT `/api/admin/*` funciona igual que antes
- [ ] Sistema empresa `/v1/*` funciona en paralelo
- [ ] WebSocket /v1/ws con eventos
- [ ] Flujo QR completo con expiración

---

## 📋 Stories (10)

| ID | Story | Prioridad |
|-----|-------|----------|
| S-6.1 | Schema DB: empresas + telefonos + migrate messages/broadcasts | P0 |
| S-6.2 | Modelo Empresa + Teléfono | P0 |
| S-6.3 | Migración messages/broadcasts (empresa_id, telefono_id) | P0 |
| S-6.4 | Generación JWT empresa | P0 |
| S-6.5 | Middleware auth JWT empresa + ownership | P0 |
| S-6.6 | Endpoints v1/* + elimina legacy | P0 |
| S-6.7 | WebSocket /v1/ws + eventos | P0 |
| S-6.8 | Flujo QR completo | P0 |
| S-6.9 | Tests unitarios + integración | P0 |
| S-6.10 | Documentación API | P1 |