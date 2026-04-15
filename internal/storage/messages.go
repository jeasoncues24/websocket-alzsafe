package storage

// [QUÉ] Implementa la persistencia de mensajes en MariaDB usando el patrón Repository.
// [POR QUÉ] El patrón Repository abstrae el acceso a datos detrás de una interfaz.
// Esto permite: 1) testear handlers sin una DB real, 2) cambiar el motor de DB sin tocar handlers.

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"wsapi/internal/domain"
)

// MessagesRepository define todas las operaciones de persistencia sobre mensajes.
// [QUÉ] Es una interfaz, no una struct concreta.
// [POR QUÉ] Al ser interfaz, los tests pueden inyectar una implementación "falsa" (mock/stub)
// sin levantar MariaDB, lo que hace los tests rápidos y deterministas.
type MessagesRepository interface {
	// Create persiste un nuevo mensaje con estado 'pending'.
	Create(msg *domain.Message) error

	// UpdateEstado actualiza el estado de un mensaje y opcionalmente su error_reason.
	// [POR QUÉ] Separado de Create para ser llamado cuando el proveedor confirma/rechaza.
	UpdateEstado(referenceID string, estado domain.MessageState, errorReason string) error

	// GetByEmpresa retorna mensajes de una empresa, ordenados por timestamp DESC, con paginación.
	GetByEmpresa(rucEmpresa string, estado string, limit, offset int) ([]domain.Message, int, error)

	// GetByEmpresaAndDateRange filtra por empresa y rango de fechas, con paginación.
	// [POR QUÉ] Permite auditoría temporal: "dame todos los mensajes de enero de empresa X".
	GetByEmpresaAndDateRange(rucEmpresa string, start, end time.Time, estado string, limit, offset int) ([]domain.Message, int, error)
}

// mariaDBMessagesRepository implementa MessagesRepository contra MariaDB/MySQL.
type mariaDBMessagesRepository struct {
	db *sql.DB
}

// NewMessagesRepository crea un repositorio de mensajes conectado a la DB dada.
func NewMessagesRepository(db *sql.DB) MessagesRepository {
	return &mariaDBMessagesRepository{db: db}
}

// Create inserta un nuevo mensaje en la tabla messages.
// [QUÉ] Serializa los adjuntos a JSON para almacenarlos en la columna adjuntos_json (TEXT).
// [POR QUÉ] MariaDB no tiene tipo nativo de array; JSON en TEXT es simple y legible.
// [APRENDE] En Go, json.Marshal convierte cualquier struct/slice a []byte listo para string.
func (r *mariaDBMessagesRepository) Create(msg *domain.Message) error {
	// Serializar slice de adjuntos a JSON; si no hay adjuntos, guarda "null"
	adjuntosJSON, err := json.Marshal(msg.Adjuntos)
	if err != nil {
		return fmt.Errorf("error al serializar adjuntos: %w", err)
	}

	// [QUÉ] INSERT con parámetros posicionales ? (placeholders).
	// [POR QUÉ] Los placeholders previenen SQL Injection: el driver escapa los valores antes de enviar.
	// [APRENDE] NUNCA concatenes valores del usuario en queries SQL. Siempre usa ? o $N.
	_, err = r.db.Exec(`
		INSERT INTO messages
			(reference_id, ruc_empresa, destino, contenido, adjuntos_json, estado, timestamp_created)
		VALUES
			(?, ?, ?, ?, ?, ?, ?)
	`,
		msg.ReferenceID,
		msg.RUCEmpresa,
		msg.Destino,
		msg.Contenido,
		string(adjuntosJSON),
		string(msg.Estado),
		msg.TiempoEnvio,
	)

	return err
}

// UpdateEstado actualiza el estado y los timestamps correspondientes de un mensaje.
// [QUÉ] Usa lógica condicional en SQL: solo escribe timestamp_sent/confirmed si el estado lo amerita.
// [POR QUÉ] Mantiene integridad temporal: si marcamos 'sent', registramos cuándo ocurrió.
func (r *mariaDBMessagesRepository) UpdateEstado(referenceID string, estado domain.MessageState, errorReason string) error {
	now := time.Now()

	switch estado {
	case domain.MessageStateSent:
		_, err := r.db.Exec(`
			UPDATE messages SET estado = ?, timestamp_sent = ?, error_reason = NULL
			WHERE reference_id = ?
		`, string(estado), now, referenceID)
		return err

	case domain.MessageStateDelivered:
		_, err := r.db.Exec(`
			UPDATE messages SET estado = ?, timestamp_confirmed = ?
			WHERE reference_id = ?
		`, string(estado), now, referenceID)
		return err

	case domain.MessageStateFailed, domain.MessageStateRejected:
		var errReason *string
		if errorReason != "" {
			errReason = &errorReason
		}
		_, err := r.db.Exec(`
			UPDATE messages SET estado = ?, error_reason = ?
			WHERE reference_id = ?
		`, string(estado), errReason, referenceID)
		return err

	default:
		_, err := r.db.Exec(`
			UPDATE messages SET estado = ? WHERE reference_id = ?
		`, string(estado), referenceID)
		return err
	}
}

// GetByEmpresa retorna mensajes paginados de una empresa, ordenados por timestamp_created DESC.
// Retorna: lista de mensajes, total de registros (para calcular páginas), error.
// [APRENDE] Se hacen 2 queries: una para COUNT (total) y otra para los datos paginados.
func (r *mariaDBMessagesRepository) GetByEmpresa(rucEmpresa string, estado string, limit, offset int) ([]domain.Message, int, error) {
	// Query 1: contar total para paginación
	var total int
	var err error
	if estado != "" {
		err = r.db.QueryRow(
			`SELECT COUNT(*) FROM messages WHERE ruc_empresa = ? AND estado = ?`, rucEmpresa, estado,
		).Scan(&total)
	} else {
		err = r.db.QueryRow(
			`SELECT COUNT(*) FROM messages WHERE ruc_empresa = ?`, rucEmpresa,
		).Scan(&total)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("error contando mensajes: %w", err)
	}

	// Query 2: obtener página de datos
	var rows *sql.Rows
	if estado != "" {
		rows, err = r.db.Query(`
			SELECT id, reference_id, ruc_empresa, destino, contenido,
			       adjuntos_json, estado, error_reason,
			       timestamp_created, timestamp_sent, timestamp_confirmed
			FROM messages
			WHERE ruc_empresa = ? AND estado = ?
			ORDER BY timestamp_created DESC
			LIMIT ? OFFSET ?
		`, rucEmpresa, estado, limit, offset)
	} else {
		rows, err = r.db.Query(`
			SELECT id, reference_id, ruc_empresa, destino, contenido,
			       adjuntos_json, estado, error_reason,
			       timestamp_created, timestamp_sent, timestamp_confirmed
			FROM messages
			WHERE ruc_empresa = ?
			ORDER BY timestamp_created DESC
			LIMIT ? OFFSET ?
		`, rucEmpresa, limit, offset)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("error obteniendo mensajes: %w", err)
	}
	defer rows.Close()

	messages, err := scanMessages(rows)
	if err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// GetByEmpresaAndDateRange filtra mensajes por empresa y rango de fechas con paginación.
func (r *mariaDBMessagesRepository) GetByEmpresaAndDateRange(rucEmpresa string, start, end time.Time, estado string, limit, offset int) ([]domain.Message, int, error) {
	var total int
	var err error
	if estado != "" {
		err = r.db.QueryRow(
			`SELECT COUNT(*) FROM messages WHERE ruc_empresa = ? AND timestamp_created BETWEEN ? AND ? AND estado = ?`,
			rucEmpresa, start, end, estado,
		).Scan(&total)
	} else {
		err = r.db.QueryRow(
			`SELECT COUNT(*) FROM messages WHERE ruc_empresa = ? AND timestamp_created BETWEEN ? AND ?`,
			rucEmpresa, start, end,
		).Scan(&total)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("error contando mensajes por rango: %w", err)
	}

	var rows *sql.Rows
	if estado != "" {
		rows, err = r.db.Query(`
			SELECT id, reference_id, ruc_empresa, destino, contenido,
			       adjuntos_json, estado, error_reason,
			       timestamp_created, timestamp_sent, timestamp_confirmed
			FROM messages
			WHERE ruc_empresa = ? AND timestamp_created BETWEEN ? AND ? AND estado = ?
			ORDER BY timestamp_created DESC
			LIMIT ? OFFSET ?
		`, rucEmpresa, start, end, estado, limit, offset)
	} else {
		rows, err = r.db.Query(`
			SELECT id, reference_id, ruc_empresa, destino, contenido,
			       adjuntos_json, estado, error_reason,
			       timestamp_created, timestamp_sent, timestamp_confirmed
			FROM messages
			WHERE ruc_empresa = ? AND timestamp_created BETWEEN ? AND ?
			ORDER BY timestamp_created DESC
			LIMIT ? OFFSET ?
		`, rucEmpresa, start, end, limit, offset)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("error obteniendo mensajes por rango: %w", err)
	}
	defer rows.Close()

	messages, err := scanMessages(rows)
	if err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// scanMessages convierte las filas de un *sql.Rows en un slice de domain.Message.
// [QUÉ] Función de utilidad interna (minúscula = privada al paquete).
// [POR QUÉ] Reutilizar el escaneo evita duplicar lógica de conversión de tipos SQL → Go.
// [APRENDE] Los campos anulables en DB (NULL) deben escanearse en punteros (*string, *time.Time).
func scanMessages(rows *sql.Rows) ([]domain.Message, error) {
	var messages []domain.Message

	for rows.Next() {
		var msg domain.Message

		// [QUÉ] Campos anulables: se escanean en punteros para distinguir NULL de string vacío.
		var adjuntosJSON string
		var errorReason sql.NullString
		var timestampSent, timestampConfirmed sql.NullTime

		err := rows.Scan(
			&msg.ID,
			&msg.ReferenceID,
			&msg.RUCEmpresa,
			&msg.Destino,
			&msg.Contenido,
			&adjuntosJSON,
			&msg.Estado,
			&errorReason,
			&msg.TiempoEnvio,
			&timestampSent,
			&timestampConfirmed,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando fila de mensaje: %w", err)
		}

		// Deserializar adjuntos_json → []AttachmentInfo
		if adjuntosJSON != "" && adjuntosJSON != "null" {
			if err := json.Unmarshal([]byte(adjuntosJSON), &msg.Adjuntos); err != nil {
				return nil, fmt.Errorf("error deserializando adjuntos_json: %w", err)
			}
		}

		// Convertir sql.NullString / sql.NullTime a los tipos del dominio
		if errorReason.Valid {
			msg.ErrorReason = errorReason.String
		}
		if timestampSent.Valid {
			t := timestampSent.Time
			msg.TimestampSent = &t
		}
		if timestampConfirmed.Valid {
			t := timestampConfirmed.Time
			msg.TimestampConfirmed = &t
		}

		messages = append(messages, msg)
	}

	// [QUÉ] Verificar si el loop de rows tuvo algún error de red/cursor.
	// [POR QUÉ] rows.Err() puede retornar errores que no se detectan dentro del loop.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando filas de mensajes: %w", err)
	}

	// Retornar slice vacío en vez de nil para que el JSON serialice [] y no null
	if messages == nil {
		messages = []domain.Message{}
	}

	return messages, nil
}
