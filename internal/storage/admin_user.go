package storage

import (
	"database/sql"
	"fmt"
	"time"

	"wsapi/internal/domain"
)

type AdminUserStore struct {
	db *sql.DB
}

func NewAdminUserStore(db *sql.DB) *AdminUserStore {
	return &AdminUserStore{db: db}
}

// Create inserta un nuevo usuario admin
func (s *AdminUserStore) Create(user *domain.AdminUser) (int64, error) {
	query := `INSERT INTO admin_users (username, password_hash, email, empresa_id, rol, activo, role_id, is_root) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := s.db.Exec(query, user.Username, user.PasswordHash, user.Email,
		user.EmpresaID, user.Rol, user.Activo, user.RoleID, user.IsRoot)
	if err != nil {
		return 0, fmt.Errorf("error al crear usuario: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error al obtener ID: %w", err)
	}

	user.ID = id
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	return id, nil
}

// GetByID obtiene un usuario por ID
func (s *AdminUserStore) GetByID(id int64) (*domain.AdminUser, error) {
	query := `SELECT id, username, password_hash, email, empresa_id, rol, activo, created_at, updated_at, last_login_at, role_id, is_root 
			  FROM admin_users WHERE id = ?`

	user := &domain.AdminUser{}
	var email, role string
	var empresaID sql.NullInt64
	var lastLogin sql.NullTime
	var roleID sql.NullInt64

	err := s.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &email, &empresaID,
		&role, &user.Activo, &user.CreatedAt, &user.UpdatedAt, &lastLogin,
		&roleID, &user.IsRoot,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener usuario: %w", err)
	}

	user.Email = email
	if empresaID.Valid {
		user.EmpresaID = &empresaID.Int64
	}
	user.Rol = domain.UserRole(role)
	if roleID.Valid {
		user.RoleID = &roleID.Int64
	}
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return user, nil
}

// GetByUsername obtiene un usuario por username
func (s *AdminUserStore) GetByUsername(username string) (*domain.AdminUser, error) {
	query := `SELECT id, username, password_hash, email, empresa_id, rol, activo, created_at, updated_at, last_login_at, role_id, is_root 
			  FROM admin_users WHERE username = ?`

	user := &domain.AdminUser{}
	var email, role string
	var empresaID sql.NullInt64
	var lastLogin sql.NullTime
	var roleID sql.NullInt64

	err := s.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &email, &empresaID,
		&role, &user.Activo, &user.CreatedAt, &user.UpdatedAt, &lastLogin,
		&roleID, &user.IsRoot,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener usuario: %w", err)
	}

	user.Email = email
	if empresaID.Valid {
		user.EmpresaID = &empresaID.Int64
	}
	user.Rol = domain.UserRole(role)
	if roleID.Valid {
		user.RoleID = &roleID.Int64
	}
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return user, nil
}

// GetAll obtiene todos los usuarios con paginación
func (s *AdminUserStore) GetAll(page, limit int) ([]domain.AdminUser, int, error) {
	offset := (page - 1) * limit

	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM admin_users").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error al contar usuarios: %w", err)
	}

	query := `SELECT id, username, password_hash, email, empresa_id, rol, activo, created_at, updated_at, last_login_at, role_id, is_root 
			  FROM admin_users ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error al obtener usuarios: %w", err)
	}
	defer rows.Close()

	var users []domain.AdminUser
	for rows.Next() {
		var u domain.AdminUser
		var email, role string
		var empresaID sql.NullInt64
		var lastLogin sql.NullTime
		var roleID sql.NullInt64

		err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &email, &empresaID,
			&role, &u.Activo, &u.CreatedAt, &u.UpdatedAt, &lastLogin, &roleID, &u.IsRoot)
		if err != nil {
			return nil, 0, fmt.Errorf("error al escanear usuario: %w", err)
		}

		u.Email = email
		if empresaID.Valid {
			u.EmpresaID = &empresaID.Int64
		}
		u.Rol = domain.UserRole(role)
		if lastLogin.Valid {
			u.LastLogin = &lastLogin.Time
		}
		if roleID.Valid {
			u.RoleID = &roleID.Int64
		}

		// Don't expose password hash
		u.PasswordHash = ""
		users = append(users, u)
	}

	return users, total, nil
}

// GetAllByEmpresa obtiene usuarios admin de una empresa con paginación.
func (s *AdminUserStore) GetAllByEmpresa(empresaID int64, page, limit int) ([]domain.AdminUser, int, error) {
	offset := (page - 1) * limit

	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM admin_users WHERE empresa_id = ?", empresaID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error al contar usuarios: %w", err)
	}

	query := `SELECT id, username, password_hash, email, empresa_id, rol, activo, created_at, updated_at, last_login_at, role_id, is_root 
			  FROM admin_users WHERE empresa_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := s.db.Query(query, empresaID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error al obtener usuarios: %w", err)
	}
	defer rows.Close()

	var users []domain.AdminUser
	for rows.Next() {
		var u domain.AdminUser
		var email, role string
		var empresa sql.NullInt64
		var lastLogin sql.NullTime
		var roleID sql.NullInt64

		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &email, &empresa,
			&role, &u.Activo, &u.CreatedAt, &u.UpdatedAt, &lastLogin, &roleID, &u.IsRoot); err != nil {
			return nil, 0, fmt.Errorf("error al escanear usuario: %w", err)
		}

		u.Email = email
		if empresa.Valid {
			u.EmpresaID = &empresa.Int64
		}
		u.Rol = domain.UserRole(role)
		if lastLogin.Valid {
			u.LastLogin = &lastLogin.Time
		}
		if roleID.Valid {
			u.RoleID = &roleID.Int64
		}
		u.PasswordHash = ""
		users = append(users, u)
	}

	return users, total, nil
}

// Update actualiza un usuario existente
func (s *AdminUserStore) Update(user *domain.AdminUser) error {
	query := `UPDATE admin_users SET email = ?, empresa_id = ?, rol = ?, activo = ?, role_id = ?, is_root = ?, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = ?`

	_, err := s.db.Exec(query, user.Email, user.EmpresaID, user.Rol, user.Activo, user.RoleID, user.IsRoot, user.ID)
	if err != nil {
		return fmt.Errorf("error al actualizar usuario: %w", err)
	}

	user.UpdatedAt = time.Now()
	return nil
}

// UpdatePassword actualiza la contraseña de un usuario
func (s *AdminUserStore) UpdatePassword(id int64, newHash string) error {
	query := `UPDATE admin_users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	_, err := s.db.Exec(query, newHash, id)
	if err != nil {
		return fmt.Errorf("error al actualizar contraseña: %w", err)
	}

	return nil
}

// UpdateLastLogin actualiza el timestamp de último login
func (s *AdminUserStore) UpdateLastLogin(id int64) error {
	query := `UPDATE admin_users SET last_login_at = CURRENT_TIMESTAMP WHERE id = ?`

	_, err := s.db.Exec(query, id)
	return err
}

// Delete realiza soft delete de un usuario
func (s *AdminUserStore) Delete(id int64) error {
	_, err := s.DeleteWithPolicy(id)
	return err
}

// DeleteWithPolicy borra si no hay dependencias; si hay referencias, deshabilita.
func (s *AdminUserStore) DeleteWithPolicy(id int64) (string, error) {
	var refs int
	err := s.db.QueryRow(`
		SELECT
		  (SELECT COUNT(*) FROM user_modules WHERE user_id = ?) +
		  (SELECT COUNT(*) FROM api_keys WHERE created_by_user_id = ?) +
		  (SELECT COUNT(*) FROM api_key_events WHERE actor_user_id = ?)
	`, id, id, id).Scan(&refs)
	if err != nil {
		return "", fmt.Errorf("error al validar dependencias del usuario: %w", err)
	}

	if refs > 0 {
		query := `UPDATE admin_users SET activo = FALSE, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
		if _, err := s.db.Exec(query, id); err != nil {
			return "", fmt.Errorf("error al deshabilitar usuario: %w", err)
		}
		return "disabled", nil
	}

	query := `DELETE FROM admin_users WHERE id = ?`
	res, err := s.db.Exec(query, id)
	if err != nil {
		return "", fmt.Errorf("error al eliminar usuario: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return "", sql.ErrNoRows
	}
	return "deleted", nil
}
