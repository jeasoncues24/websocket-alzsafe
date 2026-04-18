package storage

import (
	"database/sql"
	"fmt"
	"time"

	"wsapi/internal/domain"
)

type TelefonoStore struct {
	db *sql.DB
}

func NewTelefonoStore(db *sql.DB) *TelefonoStore {
	return &TelefonoStore{db: db}
}

// Create inserta un nuevo teléfono
func (s *TelefonoStore) Create(t *domain.Telefono) (int64, error) {
	if t.Status == "" {
		t.Status = domain.TelefonoStatusDisconnected
	}
	if t.NumeroCompleto == "" {
		t.NumeroCompleto = t.CodigoPais + t.Numero
	}

	query := `INSERT INTO telefonos (empresa_id, codigo_pais, numero, numero_completo, status) VALUES (?, ?, ?, ?, ?)`

	result, err := s.db.Exec(query, t.EmpresaID, t.CodigoPais, t.Numero, t.NumeroCompleto, t.Status)
	if err != nil {
		return 0, fmt.Errorf("error al crear telefono: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error al obtener ID telefono: %w", err)
	}

	t.ID = id
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	return id, nil
}

// Update actualiza los datos base de un teléfono.
func (s *TelefonoStore) Update(t *domain.Telefono) error {
	if t.Status == "" {
		t.Status = domain.TelefonoStatusDisconnected
	}
	if t.NumeroCompleto == "" {
		t.NumeroCompleto = t.CodigoPais + t.Numero
	}

	_, err := s.db.Exec(`
		UPDATE telefonos
		SET codigo_pais = ?, numero = ?, numero_completo = ?, status = ?, updated_at = NOW()
		WHERE id = ?
	`, t.CodigoPais, t.Numero, t.NumeroCompleto, t.Status, t.ID)
	if err != nil {
		return fmt.Errorf("error al actualizar telefono: %w", err)
	}

	t.UpdatedAt = time.Now()
	return nil
}

// GetByID obtiene un teléfono por ID
func (s *TelefonoStore) GetByID(id int64) (*domain.Telefono, error) {
	query := `SELECT id, empresa_id, codigo_pais, numero, numero_completo, status, session_data, qr_string, last_connected, created_at, updated_at
			  FROM telefonos WHERE id = ?`

	return s.scanOne(s.db.QueryRow(query, id))
}

// GetByEmpresa lista todos los teléfonos de una empresa
func (s *TelefonoStore) GetByEmpresa(empresaID int64) ([]domain.Telefono, error) {
	query := `SELECT id, empresa_id, codigo_pais, numero, numero_completo, status, session_data, qr_string, last_connected, created_at, updated_at
			  FROM telefonos WHERE empresa_id = ? ORDER BY created_at ASC`

	rows, err := s.db.Query(query, empresaID)
	if err != nil {
		return nil, fmt.Errorf("error al listar telefonos: %w", err)
	}
	defer rows.Close()

	var telefonos []domain.Telefono
	for rows.Next() {
		t, err := s.scanRow(rows)
		if err != nil {
			return nil, err
		}
		telefonos = append(telefonos, *t)
	}
	return telefonos, nil
}

// GetByNumeroCompleto busca un teléfono por numero_completo (ej. "51999888777")
func (s *TelefonoStore) GetByNumeroCompleto(numeroCompleto string) (*domain.Telefono, error) {
	query := `SELECT id, empresa_id, codigo_pais, numero, numero_completo, status, session_data, qr_string, last_connected, created_at, updated_at
			  FROM telefonos WHERE numero_completo = ?`

	return s.scanOne(s.db.QueryRow(query, numeroCompleto))
}

// UpdateStatus actualiza el estado de un teléfono
func (s *TelefonoStore) UpdateStatus(id int64, status domain.TelefonoStatus) error {
	_, err := s.db.Exec(`UPDATE telefonos SET status = ?, updated_at = NOW() WHERE id = ?`, status, id)
	if err != nil {
		return fmt.Errorf("error al actualizar status de telefono: %w", err)
	}
	return nil
}

// UpdateSessionData guarda los datos de sesión serializada de whatsmeow
func (s *TelefonoStore) UpdateSessionData(id int64, sessionData []byte) error {
	_, err := s.db.Exec(`UPDATE telefonos SET session_data = ?, updated_at = NOW() WHERE id = ?`, sessionData, id)
	if err != nil {
		return fmt.Errorf("error al actualizar session_data: %w", err)
	}
	return nil
}

// UpdateQRString guarda el QR string para escanear
func (s *TelefonoStore) UpdateQRString(id int64, qrString string) error {
	_, err := s.db.Exec(`UPDATE telefonos SET qr_string = ?, status = ?, updated_at = NOW() WHERE id = ?`,
		qrString, domain.TelefonoStatusQRPending, id)
	if err != nil {
		return fmt.Errorf("error al actualizar qr_string: %w", err)
	}
	return nil
}

// SetConnected marca el teléfono como activo y actualiza last_connected
func (s *TelefonoStore) SetConnected(id int64) error {
	_, err := s.db.Exec(`UPDATE telefonos SET status = ?, qr_string = NULL, last_connected = NOW(), updated_at = NOW() WHERE id = ?`,
		domain.TelefonoStatusActive, id)
	if err != nil {
		return fmt.Errorf("error al marcar telefono como conectado: %w", err)
	}
	return nil
}

// SetDisconnected marca el teléfono como desconectado
func (s *TelefonoStore) SetDisconnected(id int64) error {
	_, err := s.db.Exec(`UPDATE telefonos SET status = ?, qr_string = NULL, updated_at = NOW() WHERE id = ?`,
		domain.TelefonoStatusDisconnected, id)
	if err != nil {
		return fmt.Errorf("error al marcar telefono como desconectado: %w", err)
	}
	return nil
}

// Delete elimina un teléfono (hard delete — no tiene soft delete por diseño)
func (s *TelefonoStore) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM telefonos WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("error al eliminar telefono: %w", err)
	}
	return nil
}

// BelongsToEmpresa verifica que el telefono_id pertenece a la empresa dada (para ownership check)
func (s *TelefonoStore) BelongsToEmpresa(telefonoID, empresaID int64) (bool, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM telefonos WHERE id = ? AND empresa_id = ?`, telefonoID, empresaID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error al verificar ownership: %w", err)
	}
	return count > 0, nil
}

// --- helpers ---

func (s *TelefonoStore) scanOne(row *sql.Row) (*domain.Telefono, error) {
	t := &domain.Telefono{}
	var qrString sql.NullString
	var lastConnected sql.NullTime

	err := row.Scan(
		&t.ID, &t.EmpresaID, &t.CodigoPais, &t.Numero, &t.NumeroCompleto,
		&t.Status, &t.SessionData, &qrString, &lastConnected,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al escanear telefono: %w", err)
	}

	if qrString.Valid {
		t.QRString = qrString.String
	}
	if lastConnected.Valid {
		t.LastConnected = &lastConnected.Time
	}
	return t, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func (s *TelefonoStore) scanRow(row scannable) (*domain.Telefono, error) {
	t := &domain.Telefono{}
	var qrString sql.NullString
	var lastConnected sql.NullTime

	err := row.Scan(
		&t.ID, &t.EmpresaID, &t.CodigoPais, &t.Numero, &t.NumeroCompleto,
		&t.Status, &t.SessionData, &qrString, &lastConnected,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error al escanear telefono: %w", err)
	}

	if qrString.Valid {
		t.QRString = qrString.String
	}
	if lastConnected.Valid {
		t.LastConnected = &lastConnected.Time
	}
	return t, nil
}
