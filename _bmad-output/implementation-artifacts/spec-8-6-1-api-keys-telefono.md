# S-8.6.1 Fix: API Keys por Teléfono WhatsApp

## Estado

`done`.

## Motivo del fix intermedio

La versión anterior de `S-8.6` estaba planteada como JWT de empresa. Ese modelo sirve para identidad de tenant, pero no para un sistema tipo OpenAI/AWS con consumo por integración, métricas por clave y rotación segura.

## Decisión de arquitectura

- La empresa sigue siendo el tenant.
- El teléfono WhatsApp es la unidad de consumo y métricas.
- La API key no es el número de teléfono ni un JWT.
- La API key es un secreto opaco, mostrado solo una vez al crearla.
- El frontend puede mostrar el valor completo solo en el momento de creación/rotación y luego solo el último estado parcial o máscara.

## Objetivos del fix

1. Crear el modelo de datos real para API keys por teléfono.
2. Preparar la base de uso y auditoría por key.
3. Definir el mapa de handlers admin y público.
4. Dejar `S-8.6` pendiente hasta terminar este fix.

## Modelo de tablas

### `api_keys`

Campos:

- `id`
- `empresa_id`
- `telefono_id`
- `nombre`
- `key_prefix`
- `secret_hash`
- `scopes`
- `activo`
- `created_by_user_id`
- `last_used_at`
- `expires_at`
- `revoked_at`
- `rotated_from_id`
- `created_at`
- `updated_at`

### `api_key_usage_events`

Campos:

- `id`
- `api_key_id`
- `empresa_id`
- `telefono_id`
- `method`
- `endpoint`
- `status_code`
- `latency_ms`
- `request_units`
- `response_units`
- `request_id`
- `created_at`

### `api_key_usage_daily`

Campos:

- `day`
- `api_key_id`
- `empresa_id`
- `telefono_id`
- `request_count`
- `success_count`
- `error_count`
- `latency_avg_ms`
- `messages_sent`
- `broadcasts_sent`
- `bytes_in`
- `bytes_out`

### `api_key_audit_events`

Campos:

- `id`
- `api_key_id`
- `empresa_id`
- `telefono_id`
- `action`
- `actor_user_id`
- `metadata`
- `created_at`

## Mapa de handlers

### Admin

- `GET /api/admin/telefonos/{telefono_id}/api-keys` -> listar claves del teléfono
- `POST /api/admin/telefonos/{telefono_id}/api-keys` -> crear key y devolver el secreto una sola vez
- `GET /api/admin/api-keys/{api_key_id}` -> detalle de la key
- `POST /api/admin/api-keys/{api_key_id}/rotate` -> crear nueva key y revocar la anterior
- `POST /api/admin/api-keys/{api_key_id}/revoke` -> revocar key
- `GET /api/admin/api-keys/{api_key_id}/usage` -> métricas de uso
- `GET /api/admin/api-keys/{api_key_id}/audit` -> auditoría de eventos

### Público

- Header de autenticación: `X-API-Key`
- Base sugerida: `/api/v1`
- Middleware dedicado para validar key, scopes y vínculo con `telefono_id`

## Regla UX de seguridad

- La key se muestra completa solo una vez al crearla.
- Después de eso, solo se permite copiar desde el modal inicial o ver máscara parcial.
- El patrón debe parecerse a OpenAI / AWS / Gemini: secreto breve, visible una vez, almacenado con hash.

## Pendientes bloqueados

- UI nueva hasta que este fix quede estable.
- `S-8.6` sigue sin considerarse completado.
