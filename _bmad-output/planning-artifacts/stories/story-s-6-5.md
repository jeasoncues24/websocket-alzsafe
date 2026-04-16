# Story S-6.5: Endpoints Empresa (v1/*)

## Epic
Epic 6: Sistema de Autenticación JWT por Empresa

## Prioridad
P0

## Estado
pending

## Overview

Implementar todos los endpoints que la empresa puede usar para gestionar sus números y enviar mensajes.

## Endpoints

| Método | Endpoint | Descripción | Permiso |
|--------|----------|-------------|--------|
| GET | `/v1/status` | Estado de la empresa y teléfonos | view |
| GET | `/v1/telefonos` | Listar teléfonos | view |
| GET | `/v1/telefonos/{id}` | Ver teléfono específico | view |
| POST | `/v1/telefonos` | Agregar nuevo teléfono | sessions |
| DELETE | `/v1/telefonos/{id}` | Eliminar teléfono | sessions |
| POST | `/v1/telefonos/{id}/connect` | Iniciar conexión QR | sessions |
| POST | `/v1/telefonos/{id}/disconnect` | Desconectar | sessions |
| POST | `/v1/message` | Enviar mensaje | send |
| GET | `/v1/messages` | Listar mensajes | view |
| POST | `/v1/broadcast` | Crear difusión | broadcast |
| GET | `/v1/broadcast/{id}` | Ver difusión | view |

## Ejemplo: Conectar Teléfono

```go
// POST /v1/telefonos/{id}/connect
type ConnectRequest struct {
    // No body needed - se genera QR automáticamente
}

type ConnectResponse struct {
    QRString  string `json:"qr_string"`  // Para convertir en QR
    ExpiresIn int    `json:"expires_in"` // Segundos (300)
}
```

## Ejemplo: Enviar Mensaje

```go
// POST /v1/message
type SendMessageRequest struct {
    To      string `json:"to"`      // +519999999999
    Message string `json:"message"`
    PhoneID string `json:"phone_id,omitempty"` // Opcional - usa primario si no especifica
}

type SendMessageResponse struct {
    ID         string `json:"id"`
    Status     string `json:"status"` // sent, queued, failed
    Timestamp  int64  `json:"timestamp"`
}
```

## Dependencias
- S-6.4 (middleware JWT)

## Estimated Effort
3 days