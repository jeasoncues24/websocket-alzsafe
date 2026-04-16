package storage

import (
	"database/sql"
	"fmt"
	"time"

	"wsapi/internal/domain"
)

type RoleStore struct {
	db *sql.DB
}

func NewRoleStore(db *sql.DB) *RoleStore {
	return &RoleStore{db: db}
}

func (s *RoleStore) Create(role *domain.Role) (int64, error) {
	query := `INSERT INTO roles (name, description, is_root) VALUES (?, ?, ?)`

	result, err := s.db.Exec(query, role.Name, role.Description, role.IsRoot)
	if err != nil {
		return 0, fmt.Errorf("error al crear rol: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error al obtener ID: %w", err)
	}

	role.ID = id
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	return id, nil
}

func (s *RoleStore) GetByID(id int64) (*domain.Role, error) {
	query := `SELECT id, name, description, is_root, created_at, updated_at FROM roles WHERE id = ?`

	role := &domain.Role{}
	err := s.db.QueryRow(query, id).Scan(
		&role.ID, &role.Name, &role.Description, &role.IsRoot,
		&role.CreatedAt, &role.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener rol: %w", err)
	}

	return role, nil
}

func (s *RoleStore) GetByName(name string) (*domain.Role, error) {
	query := `SELECT id, name, description, is_root, created_at, updated_at FROM roles WHERE name = ?`

	role := &domain.Role{}
	err := s.db.QueryRow(query, name).Scan(
		&role.ID, &role.Name, &role.Description, &role.IsRoot,
		&role.CreatedAt, &role.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener rol: %w", err)
	}

	return role, nil
}

func (s *RoleStore) GetAll() ([]domain.Role, error) {
	query := `SELECT id, name, description, is_root, created_at, updated_at FROM roles ORDER BY id`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error al obtener roles: %w", err)
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var r domain.Role
		err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.IsRoot, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("error al escanear rol: %w", err)
		}
		roles = append(roles, r)
	}

	return roles, nil
}

func (s *RoleStore) GetRootRole() (*domain.Role, error) {
	query := `SELECT id, name, description, is_root, created_at, updated_at FROM roles WHERE is_root = TRUE LIMIT 1`

	role := &domain.Role{}
	err := s.db.QueryRow(query).Scan(
		&role.ID, &role.Name, &role.Description, &role.IsRoot,
		&role.CreatedAt, &role.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener rol root: %w", err)
	}

	return role, nil
}

func (s *RoleStore) Update(role *domain.Role) error {
	query := `UPDATE roles SET name = ?, description = ?, updated_at = NOW() WHERE id = ?`

	_, err := s.db.Exec(query, role.Name, role.Description, role.ID)
	if err != nil {
		return fmt.Errorf("error al actualizar rol: %w", err)
	}

	role.UpdatedAt = time.Now()
	return nil
}

func (s *RoleStore) Delete(id int64) error {
	query := `DELETE FROM roles WHERE id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error al eliminar rol: %w", err)
	}

	return nil
}
