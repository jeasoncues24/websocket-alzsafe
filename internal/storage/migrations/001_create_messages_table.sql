-- 001_create_messages_table.sql
-- Tabla principal de mensajes para trazabilidad y auditoría.
-- Registra cada intento de envío con su ciclo de vida completo: pending → sent → delivered/failed.

CREATE TABLE IF NOT EXISTS messages (
    -- PK autoincremental, referencia interna de la DB
    id                  BIGINT          AUTO_INCREMENT PRIMARY KEY,

    -- UUID generado en la aplicación, devuelto al cliente como referencia de trazabilidad
    reference_id        VARCHAR(36)     NOT NULL,

    -- Empresa que originó el mensaje (RUC normalizado)
    ruc_empresa         VARCHAR(20)     NOT NULL,

    -- Número destino en formato internacional (ej: 51999999999)
    destino             VARCHAR(20)     NOT NULL,

    -- Contenido textual del mensaje
    contenido           TEXT            NOT NULL,

    -- JSON con la lista de adjuntos (nombre, sha256, tamano_bytes). NULL si no hay adjuntos.
    adjuntos_json       TEXT            NULL,

    -- Estado del ciclo de vida del mensaje
    estado              ENUM('pending','sent','delivered','failed','rejected')
                        NOT NULL DEFAULT 'pending',

    -- Razón de fallo (solo se llena si estado = 'failed' o 'rejected')
    error_reason        VARCHAR(500)    NULL,

    -- Momento en que la API recibió y aceptó el mensaje (202 Accepted)
    timestamp_created   DATETIME(3)     NOT NULL,

    -- Momento en que el proveedor WhatsApp confirmó el envío
    timestamp_sent      DATETIME(3)     NULL,

    -- Momento en que el destinatario confirmó la entrega (doble check)
    timestamp_confirmed DATETIME(3)     NULL,

    -- Unicidad en reference_id para evitar duplicados
    UNIQUE KEY uk_reference_id (reference_id),

    -- Índice compuesto para las consultas más frecuentes: mensajes de una empresa por fecha
    INDEX idx_ruc_empresa_created (ruc_empresa, timestamp_created),

    -- Índice para filtrar por estado (ej: todos los 'pending' para reintentos)
    INDEX idx_estado (estado)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
