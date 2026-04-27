package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"wsapi/internal/domain"
)

type RoleStore struct {
	db *sql.DB
}

var ErrRoleInUse = errors.New("role in use")

func NewRoleStore(db *sql.DB) *RoleStore {
	return &RoleStore{db: db}
}

func (s *RoleStore) Create(role *domain.Role) (int64, error) {
	permissionsJSON, err := json.Marshal(role.Permissions)
	if err != nil {
		permissionsJSON = []byte("[]")
	}
	query := `INSERT INTO roles (name, description, is_root, permissions) VALUES (?, ?, ?, ?)`

	result, err := s.db.Exec(query, role.Name, role.Description, role.IsRoot, string(permissionsJSON))
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
	query := `SELECT id, name, description, is_root, permissions,
		COALESCE((SELECT COUNT(*) FROM admin_users au WHERE au.role_id = roles.id), 0) AS usage_count,
		created_at, updated_at FROM roles WHERE id = ?`

	role := &domain.Role{}
	var permissionsJSON sql.NullString
	err := s.db.QueryRow(query, id).Scan(
		&role.ID, &role.Name, &role.Description, &role.IsRoot,
		&permissionsJSON, &role.UsageCount, &role.CreatedAt, &role.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener rol: %w", err)
	}
	if permissionsJSON.Valid && permissionsJSON.String != "" {
		_ = json.Unmarshal([]byte(permissionsJSON.String), &role.Permissions)
	}

	return role, nil
}

func (s *RoleStore) GetByName(name string) (*domain.Role, error) {
	query := `SELECT id, name, description, is_root, permissions,
		COALESCE((SELECT COUNT(*) FROM admin_users au WHERE au.role_id = roles.id), 0) AS usage_count,
		created_at, updated_at FROM roles WHERE name = ?`

	role := &domain.Role{}
	var permissionsJSON sql.NullString
	err := s.db.QueryRow(query, name).Scan(
		&role.ID, &role.Name, &role.Description, &role.IsRoot,
		&permissionsJSON, &role.UsageCount, &role.CreatedAt, &role.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener rol: %w", err)
	}
	if permissionsJSON.Valid && permissionsJSON.String != "" {
		_ = json.Unmarshal([]byte(permissionsJSON.String), &role.Permissions)
	}

	return role, nil
}

func (s *RoleStore) GetAll() ([]domain.Role, error) {
	query := `SELECT id, name, description, is_root, permissions,
		COALESCE((SELECT COUNT(*) FROM admin_users au WHERE au.role_id = roles.id), 0) AS usage_count,
		created_at, updated_at FROM roles ORDER BY id`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error al obtener roles: %w", err)
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var r domain.Role
		var permissionsJSON sql.NullString
		err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.IsRoot, &permissionsJSON, &r.UsageCount, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("error al escanear rol: %w", err)
		}
		if permissionsJSON.Valid && permissionsJSON.String != "" {
			_ = json.Unmarshal([]byte(permissionsJSON.String), &r.Permissions)
		}
		roles = append(roles, r)
	}

	return roles, nil
}

func (s *RoleStore) GetRootRole() (*domain.Role, error) {
	query := `SELECT id, name, description, is_root, permissions,
		COALESCE((SELECT COUNT(*) FROM admin_users au WHERE au.role_id = roles.id), 0) AS usage_count,
		created_at, updated_at FROM roles WHERE is_root = TRUE LIMIT 1`

	role := &domain.Role{}
	var permissionsJSON sql.NullString
	err := s.db.QueryRow(query).Scan(
		&role.ID, &role.Name, &role.Description, &role.IsRoot,
		&permissionsJSON, &role.UsageCount, &role.CreatedAt, &role.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener rol root: %w", err)
	}
	if permissionsJSON.Valid && permissionsJSON.String != "" {
		_ = json.Unmarshal([]byte(permissionsJSON.String), &role.Permissions)
	}

	return role, nil
}

func (s *RoleStore) Update(role *domain.Role) error {
	permissionsJSON, err := json.Marshal(role.Permissions)
	if err != nil {
		permissionsJSON = []byte("[]")
	}
	query := `UPDATE roles SET name = ?, description = ?, is_root = ?, permissions = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	_, err = s.db.Exec(query, role.Name, role.Description, role.IsRoot, string(permissionsJSON), role.ID)
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

// DeleteIfUnused elimina un rol solo si no está en uso.
func (s *RoleStore) DeleteIfUnused(id int64) error {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM admin_users WHERE role_id = ?`, id).Scan(&count); err != nil {
		return fmt.Errorf("error al validar uso del rol: %w", err)
	}
	if count > 0 {
		return ErrRoleInUse
	}
	return s.Delete(id)
}
