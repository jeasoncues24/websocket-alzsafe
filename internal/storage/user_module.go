package storage

import (
	"database/sql"
	"fmt"

	"wsapi/internal/domain"
)

type UserModuleStore struct {
	db *sql.DB
}

func NewUserModuleStore(db *sql.DB) *UserModuleStore {
	return &UserModuleStore{db: db}
}

func (s *UserModuleStore) Create(userID, moduleID int64) error {
	query := `INSERT INTO user_modules (user_id, module_id) VALUES (?, ?)`

	_, err := s.db.Exec(query, userID, moduleID)
	if err != nil {
		return fmt.Errorf("error al crear usuario-módulo: %w", err)
	}

	return nil
}

func (s *UserModuleStore) GetByUserID(userID int64) ([]domain.Module, error) {
	query := `
		SELECT m.id, m.name, m.description, m.slug, m.created_at 
		FROM modules m 
		JOIN user_modules um ON m.id = um.module_id 
		WHERE um.user_id = ?`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener módulos de usuario: %w", err)
	}
	defer rows.Close()

	var modules []domain.Module
	for rows.Next() {
		var m domain.Module
		err := rows.Scan(&m.ID, &m.Name, &m.Description, &m.Slug, &m.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error al escanear módulo: %w", err)
		}
		modules = append(modules, m)
	}

	return modules, nil
}

func (s *UserModuleStore) GetByModuleID(moduleID int64) ([]domain.AdminUser, error) {
	query := `
		SELECT u.id, u.username, u.password_hash, u.email, u.empresa_id, u.role, u.is_active, u.created_at, u.updated_at, u.last_login_at
		FROM admin_users u 
		JOIN user_modules um ON u.id = um.user_id 
		WHERE um.module_id = ?`

	rows, err := s.db.Query(query, moduleID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener usuarios de módulo: %w", err)
	}
	defer rows.Close()

	var users []domain.AdminUser
	for rows.Next() {
		var u domain.AdminUser
		var email, role string
		var empresaID sql.NullInt64
		var lastLogin sql.NullTime

		err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &email, &empresaID,
			&role, &u.Activo, &u.CreatedAt, &u.UpdatedAt, &lastLogin)
		if err != nil {
			return nil, fmt.Errorf("error al escanear usuario: %w", err)
		}

		u.Email = email
		if empresaID.Valid {
			u.EmpresaID = &empresaID.Int64
		}
		u.Rol = domain.UserRole(role)
		if lastLogin.Valid {
			u.LastLogin = &lastLogin.Time
		}

		users = append(users, u)
	}

	return users, nil
}

func (s *UserModuleStore) Delete(userID, moduleID int64) error {
	query := `DELETE FROM user_modules WHERE user_id = ? AND module_id = ?`

	_, err := s.db.Exec(query, userID, moduleID)
	if err != nil {
		return fmt.Errorf("error al eliminar usuario-módulo: %w", err)
	}

	return nil
}

func (s *UserModuleStore) DeleteByUserID(userID int64) error {
	query := `DELETE FROM user_modules WHERE user_id = ?`

	_, err := s.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("error al eliminar módulos de usuario: %w", err)
	}

	return nil
}

func (s *UserModuleStore) AssignModules(userID int64, moduleIDs []int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error al iniciar transacción: %w", err)
	}
	defer tx.Rollback()

	// Delete existing
	_, err = tx.Exec(`DELETE FROM user_modules WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("error al eliminar módulos existentes: %w", err)
	}

	// Insert new
	for _, moduleID := range moduleIDs {
		_, err = tx.Exec(`INSERT INTO user_modules (user_id, module_id) VALUES (?, ?)`, userID, moduleID)
		if err != nil {
			return fmt.Errorf("error al asignar módulo: %w", err)
		}
	}

	return tx.Commit()
}

func (s *UserModuleStore) HasModuleAccess(userID int64, moduleSlug string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM user_modules um
		JOIN modules m ON um.module_id = m.id
		WHERE um.user_id = ? AND m.slug = ?`

	var count int
	err := s.db.QueryRow(query, userID, moduleSlug).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error al verificar acceso: %w", err)
	}

	return count > 0, nil
}
