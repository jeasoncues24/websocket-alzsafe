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
	GetByEmpresa(empresaID int64, estado string, telefono string, limit, offset int) ([]domain.Message, int, error)

	// GetByEmpresaAndDateRange filtra por empresa y rango de fechas, con paginación.
	// [POR QUÉ] Permite auditoría temporal: "dame todos los mensajes de enero de empresa X".
	GetByEmpresaAndDateRange(empresaID int64, start, end time.Time, estado string, telefono string, limit, offset int) ([]domain.Message, int, error)

	// GetMessageMetricsByEmpresa retorna métricas de mensajes para una empresa específica
	GetMessageMetricsByEmpresa(empresaID int64) (*MessageMetrics, error)

	// GetAllMessageMetrics retorna métricas agregadas de todas las empresas
	GetAllMessageMetrics() (*MessageMetrics, error)

	// GetByReferenceID retrieve a single message by its reference_id
	GetByReferenceID(referenceID string) (*domain.Message, error)

	// UpdateContenido updates the contenido of a message (for edit feature)
	// Returns error if message is already sent/delivered
	UpdateContenido(referenceID string, contenido string) error

	// IncrementRetryCount increments the retry_count and updates last_attempt_at
	IncrementRetryCount(referenceID string) error
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
			(reference_id, empresa_id, telefono_id, destino, contenido, adjuntos_json, estado, timestamp_created)
		VALUES
			(?, ?, ?, ?, ?, ?, ?, ?)
	`,
		msg.ReferenceID,
		msg.EmpresaID,
		msg.TelefonoID,
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
func (r *mariaDBMessagesRepository) GetByEmpresa(empresaID int64, estado string, telefono string, limit, offset int) ([]domain.Message, int, error) {
	// Build WHERE clause dynamically
	whereClause := "empresa_id = ?"
	args := []interface{}{empresaID}

	if estado != "" {
		whereClause += " AND estado = ?"
		args = append(args, estado)
	}
	if telefono != "" {
		whereClause += " AND destino LIKE ?"
		args = append(args, "%"+telefono+"%")
	}

	// Query 1: contar total para paginación
	var total int
	countQuery := "SELECT COUNT(*) FROM messages WHERE " + whereClause
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error contando mensajes: %w", err)
	}

	// Query 2: obtener página de datos
	selectQuery := `
		SELECT id, reference_id, empresa_id, telefono_id, destino, contenido,
		       adjuntos_json, estado, error_reason,
		       timestamp_created, timestamp_sent, timestamp_confirmed,
		       retry_count, last_attempt_at
		FROM messages
		WHERE ` + whereClause + `
		ORDER BY timestamp_created DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)
	rows, err := r.db.Query(selectQuery, args...)
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
func (r *mariaDBMessagesRepository) GetByEmpresaAndDateRange(empresaID int64, start, end time.Time, estado string, telefono string, limit, offset int) ([]domain.Message, int, error) {
	whereClause := "empresa_id = ? AND timestamp_created BETWEEN ? AND ?"
	args := []interface{}{empresaID, start, end}

	if estado != "" {
		whereClause += " AND estado = ?"
		args = append(args, estado)
	}
	if telefono != "" {
		whereClause += " AND destino LIKE ?"
		args = append(args, "%"+telefono+"%")
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM messages WHERE " + whereClause
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error contando mensajes por rango: %w", err)
	}

	selectQuery := `
		SELECT id, reference_id, empresa_id, telefono_id, destino, contenido,
		       adjuntos_json, estado, error_reason,
		       timestamp_created, timestamp_sent, timestamp_confirmed,
		       retry_count, last_attempt_at
		FROM messages
		WHERE ` + whereClause + `
		ORDER BY timestamp_created DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)
	rows, err := r.db.Query(selectQuery, args...)
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

// MessageMetrics contiene las métricas calculadas para una empresa
type MessageMetrics struct {
	TotalMensajes        int64 `json:"total_mensajes"`
	MensajesHoy          int64 `json:"mensajes_hoy"`
	MensajesSemana       int64 `json:"mensajes_semana"`
	MensajesExitosos     int64 `json:"mensajes_exitosos"`
	MensajesFallidos     int64 `json:"mensajes_fallidos"`
	SesionesActivas      int   `json:"sesiones_activas"`
	BroadcastsEjecutados int64 `json:"broadcasts_ejecutados"`
}

// GetMessageMetricsByEmpresa retorna métricas de mensajes para una empresa específica
func (r *mariaDBMessagesRepository) GetMessageMetricsByEmpresa(empresaID int64) (*MessageMetrics, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -7)

	metrics := &MessageMetrics{}

	// Total mensajes
	err := r.db.QueryRow("SELECT COUNT(*) FROM messages WHERE empresa_id = ?", empresaID).Scan(&metrics.TotalMensajes)
	if err != nil {
		return nil, fmt.Errorf("error contando total mensajes: %w", err)
	}

	// Mensajes de hoy
	err = r.db.QueryRow("SELECT COUNT(*) FROM messages WHERE empresa_id = ? AND timestamp_created >= ?", empresaID, todayStart).Scan(&metrics.MensajesHoy)
	if err != nil {
		return nil, fmt.Errorf("error contando mensajes hoy: %w", err)
	}

	// Mensajes de la última semana
	err = r.db.QueryRow("SELECT COUNT(*) FROM messages WHERE empresa_id = ? AND timestamp_created >= ?", empresaID, weekStart).Scan(&metrics.MensajesSemana)
	if err != nil {
		return nil, fmt.Errorf("error contando mensajes semana: %w", err)
	}

	// Mensajes exitosos (sent, delivered)
	err = r.db.QueryRow("SELECT COUNT(*) FROM messages WHERE empresa_id = ? AND estado IN ('sent', 'delivered')", empresaID).Scan(&metrics.MensajesExitosos)
	if err != nil {
		return nil, fmt.Errorf("error contando mensajes exitosos: %w", err)
	}

	// Mensajes fallidos (failed, rejected)
	err = r.db.QueryRow("SELECT COUNT(*) FROM messages WHERE empresa_id = ? AND estado IN ('failed', 'rejected')", empresaID).Scan(&metrics.MensajesFallidos)
	if err != nil {
		return nil, fmt.Errorf("error contando mensajes fallidos: %w", err)
	}

	// Broadcasts ejecutados
	err = r.db.QueryRow("SELECT COUNT(*) FROM broadcasts WHERE empresa_id = ?", empresaID).Scan(&metrics.BroadcastsEjecutados)
	if err != nil {
		return nil, fmt.Errorf("error contando broadcasts: %w", err)
	}

	return metrics, nil
}

// GetAllMessageMetrics retorna métricas agregadas de todas las empresas
func (r *mariaDBMessagesRepository) GetAllMessageMetrics() (*MessageMetrics, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -7)

	metrics := &MessageMetrics{}

	// Total mensajes
	err := r.db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&metrics.TotalMensajes)
	if err != nil {
		return nil, fmt.Errorf("error contando total mensajes: %w", err)
	}

	// Mensajes de hoy
	err = r.db.QueryRow("SELECT COUNT(*) FROM messages WHERE timestamp_created >= ?", todayStart).Scan(&metrics.MensajesHoy)
	if err != nil {
		return nil, fmt.Errorf("error contando mensajes hoy: %w", err)
	}

	// Mensajes de la última semana
	err = r.db.QueryRow("SELECT COUNT(*) FROM messages WHERE timestamp_created >= ?", weekStart).Scan(&metrics.MensajesSemana)
	if err != nil {
		return nil, fmt.Errorf("error contando mensajes semana: %w", err)
	}

	// Mensajes exitosos
	err = r.db.QueryRow("SELECT COUNT(*) FROM messages WHERE estado IN ('sent', 'delivered')").Scan(&metrics.MensajesExitosos)
	if err != nil {
		return nil, fmt.Errorf("error contando mensajes exitosos: %w", err)
	}

	// Mensajes fallidos
	err = r.db.QueryRow("SELECT COUNT(*) FROM messages WHERE estado IN ('failed', 'rejected')").Scan(&metrics.MensajesFallidos)
	if err != nil {
		return nil, fmt.Errorf("error contando mensajes fallidos: %w", err)
	}

	// Broadcasts ejecutados
	err = r.db.QueryRow("SELECT COUNT(*) FROM broadcasts").Scan(&metrics.BroadcastsEjecutados)
	if err != nil {
		return nil, fmt.Errorf("error counting broadcasts: %w", err)
	}

	return metrics, nil
}

// GetByReferenceID retrieves a single message by its reference_id
func (r *mariaDBMessagesRepository) GetByReferenceID(referenceID string) (*domain.Message, error) {
	var msg domain.Message

	var adjuntosJSON string
	var errorReason sql.NullString
	var timestampSent, timestampConfirmed, lastAttemptAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, reference_id, empresa_id, telefono_id, destino, contenido,
		       adjuntos_json, estado, error_reason,
		       timestamp_created, timestamp_sent, timestamp_confirmed,
		       retry_count, last_attempt_at
		FROM messages
		WHERE reference_id = ?
	`, referenceID).Scan(
		&msg.ID,
		&msg.ReferenceID,
		&msg.EmpresaID,
		&msg.TelefonoID,
		&msg.Destino,
		&msg.Contenido,
		&adjuntosJSON,
		&msg.Estado,
		&errorReason,
		&msg.TiempoEnvio,
		&timestampSent,
		&timestampConfirmed,
		&msg.RetryCount,
		&lastAttemptAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching message by reference_id: %w", err)
	}

	if adjuntosJSON != "" && adjuntosJSON != "null" {
		if err := json.Unmarshal([]byte(adjuntosJSON), &msg.Adjuntos); err != nil {
			return nil, fmt.Errorf("error deserializando adjuntos_json: %w", err)
		}
	}

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
	if lastAttemptAt.Valid {
		t := lastAttemptAt.Time
		msg.LastAttemptAt = &t
	}

	return &msg, nil
}

// UpdateContenido updates the contenido of a message
// Returns error if message is already sent/delivered
func (r *mariaDBMessagesRepository) UpdateContenido(referenceID string, contenido string) error {
	result, err := r.db.Exec(`
		UPDATE messages
		SET contenido = ?, updated_at = NOW()
		WHERE reference_id = ? AND estado IN ('pending', 'failed', 'rejected')
	`, contenido, referenceID)
	if err != nil {
		return fmt.Errorf("error updating message contenido: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("message not found or already sent/delivered")
	}

	return nil
}

// IncrementRetryCount increments the retry_count and updates last_attempt_at
func (r *mariaDBMessagesRepository) IncrementRetryCount(referenceID string) error {
	_, err := r.db.Exec(`
		UPDATE messages
		SET retry_count = retry_count + 1,
		    last_attempt_at = NOW(),
		    estado = 'pending',
		    error_reason = NULL,
		    updated_at = NOW()
		WHERE reference_id = ?
	`, referenceID)
	if err != nil {
		return fmt.Errorf("error incrementing retry count: %w", err)
	}
	return nil
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
		var lastAttemptAt sql.NullTime

		err := rows.Scan(
			&msg.ID,
			&msg.ReferenceID,
			&msg.EmpresaID,
			&msg.TelefonoID,
			&msg.Destino,
			&msg.Contenido,
			&adjuntosJSON,
			&msg.Estado,
			&errorReason,
			&msg.TiempoEnvio,
			&timestampSent,
			&timestampConfirmed,
			&msg.RetryCount,
			&lastAttemptAt,
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
		if lastAttemptAt.Valid {
			t := lastAttemptAt.Time
			msg.LastAttemptAt = &t
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
