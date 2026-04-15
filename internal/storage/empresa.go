package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"wsapi/internal/domain"
)

type EmpresaStore struct {
	db *sql.DB
}

func NewEmpresaStore(db *sql.DB) *EmpresaStore {
	return &EmpresaStore{db: db}
}

// Create inserta una nueva empresa
func (s *EmpresaStore) Create(empresa *domain.Empresa) (int64, error) {
	query := `INSERT INTO empresas (ruc, nombre, nombre_comercial, telefono, direccion, activo) 
			  VALUES (?, ?, ?, ?, ?, ?)`

	result, err := s.db.Exec(query, empresa.RUC, empresa.Nombre, empresa.NombreComercial,
		empresa.Telefono, empresa.Direccion, empresa.Activo)
	if err != nil {
		return 0, fmt.Errorf("error al crear empresa: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error al obtener ID: %w", err)
	}

	empresa.ID = id
	empresa.CreatedAt = time.Now()
	empresa.UpdatedAt = time.Now()

	return id, nil
}

// GetByID obtiene una empresa por ID
func (s *EmpresaStore) GetByID(id int64) (*domain.Empresa, error) {
	query := `SELECT id, ruc, nombre, nombre_comercial, telefono, direccion, activo, created_at, updated_at 
			  FROM empresas WHERE id = ?`

	empresa := &domain.Empresa{}
	err := s.db.QueryRow(query, id).Scan(
		&empresa.ID, &empresa.RUC, &empresa.Nombre, &empresa.NombreComercial,
		&empresa.Telefono, &empresa.Direccion, &empresa.Activo,
		&empresa.CreatedAt, &empresa.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener empresa: %w", err)
	}

	return empresa, nil
}

// GetByRUC obtiene una empresa por RUC
func (s *EmpresaStore) GetByRUC(ruc string) (*domain.Empresa, error) {
	query := `SELECT id, ruc, nombre, nombre_comercial, telefono, direccion, activo, created_at, updated_at 
			  FROM empresas WHERE ruc = ?`

	empresa := &domain.Empresa{}
	err := s.db.QueryRow(query, ruc).Scan(
		&empresa.ID, &empresa.RUC, &empresa.Nombre, &empresa.NombreComercial,
		&empresa.Telefono, &empresa.Direccion, &empresa.Activo,
		&empresa.CreatedAt, &empresa.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener empresa por RUC: %w", err)
	}

	return empresa, nil
}

// GetAll obtiene todas las empresas con paginación y filtros
func (s *EmpresaStore) GetAll(page, limit int, search string, activo *bool) ([]domain.Empresa, int, error) {
	offset := (page - 1) * limit

	// Build WHERE clause
	conditions := []string{}
	args := []interface{}{}

	if search != "" {
		conditions = append(conditions, "(nombre LIKE ? OR ruc LIKE ?)")
		args = append(args, "%"+search+"%", "%"+search+"%")
	}
	if activo != nil {
		conditions = append(conditions, "activo = ?")
		args = append(args, *activo)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM empresas " + where
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error al contar empresas: %w", err)
	}

	// Get paginated results
	query := fmt.Sprintf(`SELECT id, ruc, nombre, nombre_comercial, telefono, direccion, activo, created_at, updated_at 
						  FROM empresas %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, where)

	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error al obtener empresas: %w", err)
	}
	defer rows.Close()

	var empresas []domain.Empresa
	for rows.Next() {
		var e domain.Empresa
		err := rows.Scan(&e.ID, &e.RUC, &e.Nombre, &e.NombreComercial,
			&e.Telefono, &e.Direccion, &e.Activo, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("error al escanear empresa: %w", err)
		}
		empresas = append(empresas, e)
	}

	return empresas, total, nil
}

// Update actualiza una empresa existente
func (s *EmpresaStore) Update(empresa *domain.Empresa) error {
	query := `UPDATE empresas SET nombre = ?, nombre_comercial = ?, telefono = ?, 
			  direccion = ?, activo = ?, updated_at = NOW() WHERE id = ?`

	_, err := s.db.Exec(query, empresa.Nombre, empresa.NombreComercial,
		empresa.Telefono, empresa.Direccion, empresa.Activo, empresa.ID)
	if err != nil {
		return fmt.Errorf("error al actualizar empresa: %w", err)
	}

	empresa.UpdatedAt = time.Now()
	return nil
}

// Delete realiza soft delete de una empresa
func (s *EmpresaStore) Delete(id int64) error {
	query := `UPDATE empresas SET activo = FALSE, updated_at = NOW() WHERE id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error al eliminar empresa: %w", err)
	}

	return nil
}
