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
	query := `INSERT INTO admin_users (username, password_hash, email, role_id, activo, created_by, updated_by) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := s.db.Exec(query, user.Username, user.PasswordHash, user.Email,
		user.RoleID, user.Activo, user.CreatedBy, user.UpdatedBy)
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

// GetByID obtiene un usuario por ID (con JOIN a roles para IsRoot)
func (s *AdminUserStore) GetByID(id int64) (*domain.AdminUser, error) {
	query := `SELECT au.id, au.username, au.password_hash, au.email, au.role_id, r.name, au.activo, 
			  au.created_at, au.updated_at, au.last_login_at, au.created_by, au.updated_by,
			  COALESCE(r.is_root, FALSE) as is_root
			  FROM admin_users au
			  LEFT JOIN roles r ON au.role_id = r.id
			  WHERE au.id = ?`

	user := &domain.AdminUser{}
	var email sql.NullString
	var lastLogin sql.NullTime
	var roleID sql.NullInt64
	var roleName sql.NullString
	var createdBy sql.NullInt64
	var updatedBy sql.NullInt64
	var isRoot bool

	err := s.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &email,
		&roleID, &roleName, &user.Activo, &user.CreatedAt, &user.UpdatedAt, &lastLogin,
		&createdBy, &updatedBy, &isRoot,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener usuario: %w", err)
	}

	if email.Valid {
		user.Email = email.String
	}
	if roleID.Valid {
		user.RoleID = &roleID.Int64
	}
	if roleName.Valid {
		user.RoleName = domain.UserRole(roleName.String)
	}
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}
	if createdBy.Valid {
		user.CreatedBy = &createdBy.Int64
	}
	if updatedBy.Valid {
		user.UpdatedBy = &updatedBy.Int64
	}
	user.IsRoot = isRoot

	return user, nil
}

// GetByUsername obtiene un usuario por username (con JOIN a roles para IsRoot)
func (s *AdminUserStore) GetByUsername(username string) (*domain.AdminUser, error) {
	query := `SELECT au.id, au.username, au.password_hash, au.email, au.role_id, r.name, au.activo,
			  au.created_at, au.updated_at, au.last_login_at, au.created_by, au.updated_by,
			  COALESCE(r.is_root, FALSE) as is_root
			  FROM admin_users au
			  LEFT JOIN roles r ON au.role_id = r.id
			  WHERE au.username = ?`

	user := &domain.AdminUser{}
	var email sql.NullString
	var lastLogin sql.NullTime
	var roleID sql.NullInt64
	var roleName sql.NullString
	var createdBy sql.NullInt64
	var updatedBy sql.NullInt64
	var isRoot bool

	err := s.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &email,
		&roleID, &roleName, &user.Activo, &user.CreatedAt, &user.UpdatedAt, &lastLogin,
		&createdBy, &updatedBy, &isRoot,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener usuario: %w", err)
	}

	if email.Valid {
		user.Email = email.String
	}
	if roleID.Valid {
		user.RoleID = &roleID.Int64
	}
	if roleName.Valid {
		user.RoleName = domain.UserRole(roleName.String)
	}
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}
	if createdBy.Valid {
		user.CreatedBy = &createdBy.Int64
	}
	if updatedBy.Valid {
		user.UpdatedBy = &updatedBy.Int64
	}
	user.IsRoot = isRoot

	return user, nil
}

// GetAll obtiene todos los usuarios con paginación (con JOIN a roles para IsRoot)
func (s *AdminUserStore) GetAll(page, limit int) ([]domain.AdminUser, int, error) {
	offset := (page - 1) * limit

	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM admin_users").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error al contar usuarios: %w", err)
	}

	query := `SELECT au.id, au.username, au.password_hash, au.email, au.role_id, r.name, au.activo,
			  au.created_at, au.updated_at, au.last_login_at, au.created_by, au.updated_by,
			  COALESCE(r.is_root, FALSE) as is_root
			  FROM admin_users au
			  LEFT JOIN roles r ON au.role_id = r.id
			  ORDER BY au.created_at DESC LIMIT ? OFFSET ?`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error al obtener usuarios: %w", err)
	}
	defer rows.Close()

	var users []domain.AdminUser
	for rows.Next() {
		var u domain.AdminUser
		var email sql.NullString
		var lastLogin sql.NullTime
		var roleID sql.NullInt64
		var roleName sql.NullString
		var createdBy sql.NullInt64
		var updatedBy sql.NullInt64
		var isRoot bool

		err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &email,
			&roleID, &roleName, &u.Activo, &u.CreatedAt, &u.UpdatedAt, &lastLogin,
			&createdBy, &updatedBy, &isRoot)
		if err != nil {
			return nil, 0, fmt.Errorf("error al escanear usuario: %w", err)
		}

		if email.Valid {
			u.Email = email.String
		}
		if lastLogin.Valid {
			u.LastLogin = &lastLogin.Time
		}
		if roleID.Valid {
			u.RoleID = &roleID.Int64
		}
		if roleName.Valid {
			u.RoleName = domain.UserRole(roleName.String)
		}
		if createdBy.Valid {
			u.CreatedBy = &createdBy.Int64
		}
		if updatedBy.Valid {
			u.UpdatedBy = &updatedBy.Int64
		}
		u.IsRoot = isRoot

		u.PasswordHash = ""
		users = append(users, u)
	}

	return users, total, nil
}

// GetAllByEmpresa obtiene usuarios admin de una empresa (deprecated - empresa_id eliminado de admin_users)
// Se mantiene por compatibilidad pero no tiene efecto
func (s *AdminUserStore) GetAllByEmpresa(empresaID int64, page, limit int) ([]domain.AdminUser, int, error) {
	return s.GetAll(page, limit)
}

// Update actualiza un usuario existente (IsRoot se actualiza via cambio de rol)
func (s *AdminUserStore) Update(user *domain.AdminUser) error {
	query := `UPDATE admin_users SET email = ?, activo = ?, role_id = ?, updated_by = ?, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = ?`

	_, err := s.db.Exec(query, user.Email, user.Activo, user.RoleID, user.UpdatedBy, user.ID)
	if err != nil {
		return fmt.Errorf("error al actualizar usuario: %w", err)
	}

	user.UpdatedAt = time.Now()
	return nil
}

// UpdateProfile actualiza el perfil (username y email) de un usuario
func (s *AdminUserStore) UpdateProfile(id int64, username, email string) error {
	query := `UPDATE admin_users SET username = ?, email = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := s.db.Exec(query, username, email, id)
	if err != nil {
		return fmt.Errorf("error al actualizar perfil: %w", err)
	}
	return nil
}

// IsUsernameTaken verifica si el username ya está en uso por otro usuario
func (s *AdminUserStore) IsUsernameTaken(username string, excludeID int64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM admin_users WHERE username = ? AND id != ?`
	err := s.db.QueryRow(query, username, excludeID).Scan(&count)
	return count > 0, err
}

// IsEmailTaken verifica si el email ya está en uso por otro usuario
func (s *AdminUserStore) IsEmailTaken(email string, excludeID int64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM admin_users WHERE email = ? AND id != ?`
	err := s.db.QueryRow(query, email, excludeID).Scan(&count)
	return count > 0, err
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
		  (SELECT COUNT(*) FROM api_key_audit_events WHERE actor_user_id = ?)
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
