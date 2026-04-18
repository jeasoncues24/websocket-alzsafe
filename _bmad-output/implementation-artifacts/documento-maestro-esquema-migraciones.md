---
status: done
created: 2026-04-18
last_updated: 2026-04-18
---

# Documento Maestro del Esquema

## Criterio de Documentacion

Para cada columna se documenta:

- `nullable: true|false`
- `default: valor|none`

Nota: en SQL no existe `DEFAULT NOT NULL` como clausula unica; para evitar ambiguedad se documenta como `nullable: false` y `default: none` cuando el campo es obligatorio.

## Tablas

### `messages`

| Campo | Tipo | Nullable | Default | Notas |
|---|---|---:|---|---|
| id | BIGINT | false | none | PK auto increment |
| empresa_id | BIGINT | false | 0 | indice `idx_messages_empresa` |
| telefono_id | BIGINT | false | 0 | indice `idx_messages_telefono` |
| destino | VARCHAR(50) | false | none |  |
| contenido | TEXT | false | none |  |
| estado | VARCHAR(20) | false | pending | indice `idx_messages_estado` |
| reference_id | VARCHAR(100) | true | none | unique, indice `idx_messages_reference` |
| tiempo_envio | TIMESTAMP | true | none |  |
| created_at | TIMESTAMP | true | CURRENT_TIMESTAMP |  |
| updated_at | TIMESTAMP | true | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP |  |

### `broadcasts`

| Campo | Tipo | Nullable | Default | Notas |
|---|---|---:|---|---|
| id | BIGINT | false | none | PK auto increment |
| empresa_id | BIGINT | false | 0 | indice `idx_broadcasts_empresa` |
| telefono_id | BIGINT | false | 0 | indice `idx_broadcasts_telefono` |
| reference_id | VARCHAR(100) | false | none | unique, indice `idx_broadcasts_reference` |
| total | INT | false | 0 |  |
| status | VARCHAR(20) | false | pending | indice `idx_broadcasts_status` |
| created_at | TIMESTAMP | true | CURRENT_TIMESTAMP |  |
| updated_at | TIMESTAMP | true | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP |  |

### `broadcast_results`

| Campo | Tipo | Nullable | Default | Notas |
|---|---|---:|---|---|
| id | BIGINT | false | none | PK auto increment |
| broadcast_id | BIGINT | false | none | indice `idx_results_broadcast` |
| destino | VARCHAR(50) | false | none | indice `idx_results_destino` |
| status | VARCHAR(20) | false | pending |  |
| error_message | TEXT | true | none |  |
| sent_at | TIMESTAMP | true | none |  |
| delivered_at | TIMESTAMP | true | none |  |
| read_at | TIMESTAMP | true | none |  |
| created_at | TIMESTAMP | true | CURRENT_TIMESTAMP |  |

### `admin_users`

| Campo | Tipo | Nullable | Default | Notas |
|---|---|---:|---|---|
| id | BIGINT | false | none | PK auto increment |
| username | VARCHAR(50) | false | none | unique, indices `idx_admin_users_username` |
| password_hash | VARCHAR(255) | false | none |  |
| email | VARCHAR(100) | true | none |  |
| empresa_id | BIGINT | true | none | indice `idx_admin_users_empresa` |
| rol | VARCHAR(20) | false | operador | indice `idx_admin_users_rol` |
| role_id | BIGINT | true | none | indice `idx_admin_users_role` |
| is_root | BOOLEAN | false | false | agregado por migracion 011 |
| activo | BOOLEAN | false | true |  |
| created_at | TIMESTAMP | true | CURRENT_TIMESTAMP |  |
| updated_at | TIMESTAMP | true | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP |  |
| last_login_at | TIMESTAMP | true | none |  |

### `empresas`

| Campo | Tipo | Nullable | Default | Notas |
|---|---|---:|---|---|
| id | BIGINT | false | none | PK auto increment |
| ruc | VARCHAR(20) | false | none | unique, indices `idx_empresas_ruc` |
| nombre | VARCHAR(255) | false | none | indice `idx_empresas_nombre` |
| nombre_comercial | VARCHAR(255) | true | none |  |
| telefono | VARCHAR(30) | true | none | campo actual; en el nuevo contrato debe pasar a `telefono_contacto` |
| direccion | VARCHAR(500) | true | none |  |
| token_version | INT | false | 1 |  |
| permissions | JSON | true | none |  |
| activo | BOOLEAN | false | true | indice `idx_empresas_activo` |
| created_at | TIMESTAMP | true | CURRENT_TIMESTAMP |  |
| updated_at | TIMESTAMP | true | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP |  |

### `telefonos`

| Campo | Tipo | Nullable | Default | Notas |
|---|---|---:|---|---|
| id | BIGINT | false | none | PK auto increment |
| empresa_id | BIGINT | false | none | indice `idx_telefonos_empresa` |
| codigo_pais | VARCHAR(5) | false | +51 |  |
| numero | VARCHAR(20) | false | none | indice `idx_telefonos_numero` |
| numero_completo | VARCHAR(30) | false | none | unique, indice `idx_telefonos_completo` |
| status | VARCHAR(20) | false | disconnected | indice `idx_telefonos_status` |
| session_data | LONGBLOB | true | none |  |
| qr_string | TEXT | true | none |  |
| last_connected | TIMESTAMP | true | none |  |
| created_at | TIMESTAMP | true | CURRENT_TIMESTAMP |  |
| updated_at | TIMESTAMP | true | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP |  |

### `roles`

| Campo | Tipo | Nullable | Default | Notas |
|---|---|---:|---|---|
| id | BIGINT | false | none | PK auto increment |
| name | VARCHAR(50) | false | none | unique, indice `idx_roles_name` |
| description | VARCHAR(255) | true | none |  |
| permissions | JSON | true | none |  |
| created_at | TIMESTAMP | true | CURRENT_TIMESTAMP |  |
| updated_at | TIMESTAMP | true | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP |  |
| is_root | BOOLEAN | true | false | agregado por migracion 011 |

### `modules`

| Campo | Tipo | Nullable | Default | Notas |
|---|---|---:|---|---|
| id | BIGINT | false | none | PK auto increment |
| name | VARCHAR(50) | false | none | unique, indice `idx_modules_name` |
| description | VARCHAR(255) | true | none |  |
| created_at | TIMESTAMP | true | CURRENT_TIMESTAMP |  |
| updated_at | TIMESTAMP | true | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP |  |
| slug | VARCHAR(50) | true | none | unique; agregado por migracion 011 |

### `user_modules`

| Campo | Tipo | Nullable | Default | Notas |
|---|---|---:|---|---|
| id | BIGINT | false | none | PK auto increment |
| user_id | BIGINT | false | none | indice `idx_user_modules_user` |
| module_id | BIGINT | false | none | indice `idx_user_modules_module` |
| created_at | TIMESTAMP | true | CURRENT_TIMESTAMP |  |

Constraints:

- unique `uk_user_module` sobre `(user_id, module_id)`
- FK `fk_um_user` -> `admin_users(id)` ON DELETE CASCADE
- FK `fk_um_module` -> `modules(id)` ON DELETE CASCADE

### `token_blacklist`

| Campo | Tipo | Nullable | Default | Notas |
|---|---|---:|---|---|
| id | BIGINT | false | none | PK auto increment |
| jti | VARCHAR(100) | false | none | unique, indice `idx_blacklist_jti` |
| expires_at | TIMESTAMP | false | none | indice `idx_blacklist_expires` |
| created_at | TIMESTAMP | true | CURRENT_TIMESTAMP |  |

### `api_keys`

| Campo | Tipo | Nullable | Default | Notas |
|---|---|---:|---|---|
| id | BIGINT | false | none | PK auto increment |
| empresa_id | BIGINT | false | none | indice `idx_api_keys_empresa` |
| telefono_id | BIGINT | false | none | indice `idx_api_keys_telefono` |
| nombre | VARCHAR(120) | false | none |  |
| key_prefix | VARCHAR(12) | false | none | unique `uq_api_keys_key_prefix` |
| secret_hash | CHAR(64) | false | none | unique `uq_api_keys_secret_hash` |
| scopes | JSON | true | none |  |
| activo | BOOLEAN | false | true | indice `idx_api_keys_activo` |
| created_by_user_id | BIGINT | true | none |  |
| last_used_at | TIMESTAMP | true | none | indice `idx_api_keys_last_used` |
| expires_at | TIMESTAMP | true | none |  |
| revoked_at | TIMESTAMP | true | none |  |
| rotated_from_id | BIGINT | true | none |  |
| created_at | TIMESTAMP | true | CURRENT_TIMESTAMP |  |
| updated_at | TIMESTAMP | true | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP |  |

### `api_key_usage_events`

| Campo | Tipo | Nullable | Default | Notas |
|---|---|---:|---|---|
| id | BIGINT | false | none | PK auto increment |
| api_key_id | BIGINT | false | none | indice `idx_api_key_usage_key` |
| empresa_id | BIGINT | false | none | indice `idx_api_key_usage_empresa` |
| telefono_id | BIGINT | false | none | indice `idx_api_key_usage_telefono` |
| method | VARCHAR(10) | false | none |  |
| endpoint | VARCHAR(255) | false | none |  |
| status_code | INT | false | none |  |
| latency_ms | INT | false | none |  |
| request_units | INT | false | 1 |  |
| response_units | INT | false | 0 |  |
| request_id | VARCHAR(64) | true | none |  |
| created_at | TIMESTAMP | true | CURRENT_TIMESTAMP | indice `idx_api_key_usage_created_at` |

### `api_key_usage_daily`

| Campo | Tipo | Nullable | Default | Notas |
|---|---|---:|---|---|
| day | DATE | false | none | parte de PK compuesta |
| api_key_id | BIGINT | false | none | parte de PK compuesta |
| empresa_id | BIGINT | false | none | indice `idx_api_key_usage_daily_empresa` |
| telefono_id | BIGINT | false | none | indice `idx_api_key_usage_daily_telefono` |
| request_count | INT | false | 0 |  |
| success_count | INT | false | 0 |  |
| error_count | INT | false | 0 |  |
| latency_avg_ms | INT | false | 0 |  |
| messages_sent | INT | false | 0 |  |
| broadcasts_sent | INT | false | 0 |  |
| bytes_in | BIGINT | false | 0 |  |
| bytes_out | BIGINT | false | 0 |  |
| created_at | TIMESTAMP | true | CURRENT_TIMESTAMP |  |
| updated_at | TIMESTAMP | true | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP |  |

Constraints:

- PK compuesta `(day, api_key_id)`

### `api_key_audit_events`

| Campo | Tipo | Nullable | Default | Notas |
|---|---|---:|---|---|
| id | BIGINT | false | none | PK auto increment |
| api_key_id | BIGINT | false | none | indice `idx_api_key_audit_key` |
| empresa_id | BIGINT | false | none | indice `idx_api_key_audit_empresa` |
| telefono_id | BIGINT | false | none |  |
| action | VARCHAR(40) | false | none |  |
| actor_user_id | BIGINT | true | none |  |
| metadata | JSON | true | none |  |
| created_at | TIMESTAMP | true | CURRENT_TIMESTAMP | indice `idx_api_key_audit_created_at` |

### `schema_migrations`

Tabla gestionada por `golang-migrate` en runtime.

Campos esperados:

| Campo | Tipo | Nullable | Default | Notas |
|---|---|---:|---|---|
| version | BIGINT | false | none | version actual del migrador |
| dirty | BOOLEAN | false | false | estado de migracion fallida/incompleta |

## Observaciones Importantes

1. `empresas.telefono` es el nombre fisico actual de la columna, pero el contrato funcional debe documentarlo y exponerlo como `telefono_contacto` para evitar ambiguedad con los telefonos administrables.
2. Toda relacion sin FK fisica debe declararse como `logical relation` en el documento maestro, indicando tabla origen, tabla destino y campo de enlace.
3. Los defaults `0` en campos que representan relaciones (`empresa_id`, `telefono_id`, etc.) se consideran deuda tecnica del esquema y deben quedar marcados para eliminacion o reemplazo por una restriccion valida en el reset.
4. `api_keys` es la unidad de consumo por telefono. Todas las tablas de uso y auditoria deben seguir esa cardinalidad y no tratar la empresa como unidad primaria de consumo.

## Reglas del Reset

1. No borrar migraciones hasta que el documento maestro tenga estas observaciones resueltas y visibles por tabla/campo.
2. No introducir nuevos defaults `0` en columnas relacionales.
3. No asumir FK implícitas: si no existe constraint fisica, documentarlo como relacion logica.
4. Si el contrato de empresa usa `telefono_contacto`, el documento maestro debe reflejar la diferencia entre nombre fisico de columna y nombre funcional del campo.
5. Si una tabla depende de `telefono_id`, su comportamiento debe describirse desde el telefono, no desde la empresa.
