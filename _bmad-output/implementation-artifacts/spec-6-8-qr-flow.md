# S-6.8: Flujo QR completo

## Story

As a empresa user,
I want to connect/disconnect my phone numbers via QR code,
so that I can manage WhatsApp sessions.

## Endpoints

| Method | Endpoint | Descripción |
|--------|----------|-------------|
| POST | /v1/telefonos/{id}/connect | Generar QR para conexión |
| GET | /v1/telefonos/{id}/qr | Obtener QR existente |
| GET | /v1/telefonos/{id}/status | Ver estado del teléfono |
| POST | /v1/telefonos/{id}/disconnect | Desconectar sesión |

## Estado

- qr_pending → Esperando escaneo
- active → Conectado
- disconnected → Desconectado

## Implementación

- Requieren ownership validation (telefono_id属于empresa)
- connect genera QR via WhatsApp
- Guardar session_data al conectar
- Emitir eventos WS: qr, connected, disconnected