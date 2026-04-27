package storage

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"wsapi/internal/auth"
	"wsapi/internal/domain"
)

type ApiKeyStore struct {
	db *sql.DB
}

func NewApiKeyStore(db *sql.DB) *ApiKeyStore {
	return &ApiKeyStore{db: db}
}

// HashKey genera el hash SHA-256 de una API key
func HashKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

// Create genera y persiste una nueva API key
func (s *ApiKeyStore) Create(apiKey *domain.ApiKey, rawKey string) (string, error) {
	secretHash := HashKey(rawKey)
	scopesJSON, err := json.Marshal(apiKey.Scopes)
	if err != nil {
		scopesJSON = []byte("[]")
	}

	query := `INSERT INTO api_keys (
		empresa_id, telefono_id, nombre, key_prefix, secret_hash, scopes,
		activo, created_by_user_id, expires_at, rotated_from_id
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := s.db.Exec(
		query,
		apiKey.EmpresaID,
		apiKey.TelefonoID,
		apiKey.Nombre,
		apiKey.KeyPrefix,
		secretHash,
		string(scopesJSON),
		apiKey.Activo,
		apiKey.CreatedByUserID,
		apiKey.ExpiresAt,
		apiKey.RotatedFromID,
	)
	if err != nil {
		return "", fmt.Errorf("error al crear API key: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("error al obtener ID: %w", err)
	}

	apiKey.ID = id
	apiKey.SecretHash = secretHash
	apiKey.CreatedAt = time.Now()
	apiKey.UpdatedAt = time.Now()

	return rawKey, nil
}

// Rotate crea una nueva key y revoca la anterior dentro de una transacción.
func (s *ApiKeyStore) Rotate(oldKey *domain.ApiKey, newKey *domain.ApiKey, rawKey string) (string, error) {
	secretHash := HashKey(rawKey)
	scopesJSON, err := json.Marshal(newKey.Scopes)
	if err != nil {
		scopesJSON = []byte("[]")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return "", fmt.Errorf("error al iniciar rotación: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	insertQuery := `INSERT INTO api_keys (
		empresa_id, telefono_id, nombre, key_prefix, secret_hash, scopes,
		activo, created_by_user_id, expires_at, rotated_from_id
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := tx.Exec(
		insertQuery,
		newKey.EmpresaID,
		newKey.TelefonoID,
		newKey.Nombre,
		newKey.KeyPrefix,
		secretHash,
		string(scopesJSON),
		newKey.Activo,
		newKey.CreatedByUserID,
		newKey.ExpiresAt,
		oldKey.ID,
	)
	if err != nil {
		return "", fmt.Errorf("error al crear key rotada: %w", err)
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("error al obtener ID de key rotada: %w", err)
	}

	if _, err := tx.Exec(`UPDATE api_keys SET activo = FALSE, revoked_at = NOW(), updated_at = NOW() WHERE id = ?`, oldKey.ID); err != nil {
		return "", fmt.Errorf("error al revocar key anterior: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("error al confirmar rotación: %w", err)
	}

	newKey.ID = newID
	newKey.SecretHash = secretHash
	newKey.CreatedAt = time.Now()
	newKey.UpdatedAt = time.Now()
	return rawKey, nil
}

// GetByID obtiene una API key por ID (sin el hash)
func (s *ApiKeyStore) GetByID(id int64) (*domain.ApiKey, error) {
	query := `SELECT id, empresa_id, telefono_id, nombre, key_prefix, secret_hash, scopes, activo, created_by_user_id, created_at, updated_at, last_used_at, expires_at, revoked_at, rotated_from_id
			  FROM api_keys WHERE id = ?`

	key := &domain.ApiKey{}
	var scopesJSON sql.NullString
	var createdBy sql.NullInt64
	var expiresAt sql.NullTime
	var lastUsedAt sql.NullTime
	var revokedAt sql.NullTime
	var rotatedFrom sql.NullInt64

	err := s.db.QueryRow(query, id).Scan(
		&key.ID, &key.EmpresaID, &key.TelefonoID, &key.Nombre, &key.KeyPrefix, &key.SecretHash,
		&scopesJSON, &key.Activo, &createdBy, &key.CreatedAt, &key.UpdatedAt,
		&lastUsedAt, &expiresAt, &revokedAt, &rotatedFrom,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener API key: %w", err)
	}
	if scopesJSON.Valid {
		_ = json.Unmarshal([]byte(scopesJSON.String), &key.Scopes)
	}
	if createdBy.Valid {
		v := createdBy.Int64
		key.CreatedByUserID = &v
	}

	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		key.LastUsedAt = &lastUsedAt.Time
	}
	if revokedAt.Valid {
		key.RevokedAt = &revokedAt.Time
	}
	if rotatedFrom.Valid {
		v := rotatedFrom.Int64
		key.RotatedFromID = &v
	}

	// Don't expose hash
	key.SecretHash = ""

	return key, nil
}

// Validate verifica si una API key es válida
func (s *ApiKeyStore) Validate(rawKey string) (*domain.ApiKey, error) {
	prefix, ok := auth.ParseAPIKey(rawKey)
	if !ok {
		return nil, nil
	}

	query := `SELECT id, empresa_id, telefono_id, nombre, key_prefix, secret_hash, scopes, activo, created_by_user_id, created_at, updated_at, last_used_at, expires_at, revoked_at, rotated_from_id
			  FROM api_keys WHERE key_prefix = ? AND activo = TRUE`

	key := &domain.ApiKey{}
	var scopesJSON sql.NullString
	var createdBy sql.NullInt64
	var expiresAt sql.NullTime
	var lastUsedAt sql.NullTime
	var revokedAt sql.NullTime
	var rotatedFrom sql.NullInt64

	err := s.db.QueryRow(query, prefix).Scan(
		&key.ID, &key.EmpresaID, &key.TelefonoID, &key.Nombre, &key.KeyPrefix, &key.SecretHash,
		&scopesJSON, &key.Activo, &createdBy, &key.CreatedAt, &key.UpdatedAt,
		&lastUsedAt, &expiresAt, &revokedAt, &rotatedFrom,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al validar API key: %w", err)
	}
	if key.SecretHash != HashKey(rawKey) {
		return nil, nil
	}
	if scopesJSON.Valid {
		_ = json.Unmarshal([]byte(scopesJSON.String), &key.Scopes)
	}
	if createdBy.Valid {
		v := createdBy.Int64
		key.CreatedByUserID = &v
	}

	if expiresAt.Valid && time.Now().After(expiresAt.Time) {
		return nil, nil // Expired
	}
	if lastUsedAt.Valid {
		key.LastUsedAt = &lastUsedAt.Time
	}
	if revokedAt.Valid {
		key.RevokedAt = &revokedAt.Time
	}
	if rotatedFrom.Valid {
		v := rotatedFrom.Int64
		key.RotatedFromID = &v
	}

	key.SecretHash = "" // Don't expose
	return key, nil
}

// GetByEmpresaID obtiene todas las API keys de una empresa
func (s *ApiKeyStore) GetByEmpresaID(empresaID int64) ([]domain.ApiKey, error) {
	query := `SELECT id, empresa_id, telefono_id, nombre, key_prefix, secret_hash, scopes, activo, created_by_user_id, created_at, updated_at, last_used_at, expires_at, revoked_at, rotated_from_id
			  FROM api_keys WHERE empresa_id = ? ORDER BY created_at DESC`

	rows, err := s.db.Query(query, empresaID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener API keys: %w", err)
	}
	defer rows.Close()

	var keys []domain.ApiKey
	for rows.Next() {
		var k domain.ApiKey
		var scopesJSON sql.NullString
		var createdBy sql.NullInt64
		var expiresAt sql.NullTime
		var lastUsedAt sql.NullTime
		var revokedAt sql.NullTime
		var rotatedFrom sql.NullInt64

		err := rows.Scan(
			&k.ID, &k.EmpresaID, &k.TelefonoID, &k.Nombre, &k.KeyPrefix, &k.SecretHash,
			&scopesJSON, &k.Activo, &createdBy, &k.CreatedAt, &k.UpdatedAt,
			&lastUsedAt, &expiresAt, &revokedAt, &rotatedFrom,
		)
		if err != nil {
			return nil, fmt.Errorf("error al escanear API key: %w", err)
		}
		if scopesJSON.Valid {
			_ = json.Unmarshal([]byte(scopesJSON.String), &k.Scopes)
		}
		if createdBy.Valid {
			v := createdBy.Int64
			k.CreatedByUserID = &v
		}

		if expiresAt.Valid {
			k.ExpiresAt = &expiresAt.Time
		}
		if lastUsedAt.Valid {
			k.LastUsedAt = &lastUsedAt.Time
		}
		if revokedAt.Valid {
			k.RevokedAt = &revokedAt.Time
		}
		if rotatedFrom.Valid {
			v := rotatedFrom.Int64
			k.RotatedFromID = &v
		}

		k.SecretHash = "" // Don't expose
		keys = append(keys, k)
	}

	return keys, nil
}

// GetByTelefonoID obtiene todas las API keys de un teléfono.
func (s *ApiKeyStore) GetByTelefonoID(telefonoID int64) ([]domain.ApiKey, error) {
	query := `SELECT id, empresa_id, telefono_id, nombre, key_prefix, secret_hash, scopes, activo, created_by_user_id, created_at, updated_at, last_used_at, expires_at, revoked_at, rotated_from_id
			  FROM api_keys WHERE telefono_id = ? ORDER BY created_at DESC`

	rows, err := s.db.Query(query, telefonoID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener API keys por telefono: %w", err)
	}
	defer rows.Close()

	var keys []domain.ApiKey
	for rows.Next() {
		var k domain.ApiKey
		var scopesJSON sql.NullString
		var createdBy sql.NullInt64
		var expiresAt sql.NullTime
		var lastUsedAt sql.NullTime
		var revokedAt sql.NullTime
		var rotatedFrom sql.NullInt64

		err := rows.Scan(
			&k.ID, &k.EmpresaID, &k.TelefonoID, &k.Nombre, &k.KeyPrefix, &k.SecretHash,
			&scopesJSON, &k.Activo, &createdBy, &k.CreatedAt, &k.UpdatedAt,
			&lastUsedAt, &expiresAt, &revokedAt, &rotatedFrom,
		)
		if err != nil {
			return nil, fmt.Errorf("error al escanear API key por telefono: %w", err)
		}
		if scopesJSON.Valid {
			_ = json.Unmarshal([]byte(scopesJSON.String), &k.Scopes)
		}
		if createdBy.Valid {
			v := createdBy.Int64
			k.CreatedByUserID = &v
		}
		if expiresAt.Valid {
			k.ExpiresAt = &expiresAt.Time
		}
		if lastUsedAt.Valid {
			k.LastUsedAt = &lastUsedAt.Time
		}
		if revokedAt.Valid {
			k.RevokedAt = &revokedAt.Time
		}
		if rotatedFrom.Valid {
			v := rotatedFrom.Int64
			k.RotatedFromID = &v
		}
		k.SecretHash = ""
		keys = append(keys, k)
	}

	return keys, nil
}

// Delete elimina (desactiva) una API key
func (s *ApiKeyStore) Delete(id int64) error {
	query := `UPDATE api_keys SET activo = FALSE, revoked_at = NOW(), updated_at = NOW() WHERE id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error al eliminar API key: %w", err)
	}

	return nil
}

// RevokeByTelefonoID desactiva todas las API keys asociadas a un teléfono.
func (s *ApiKeyStore) RevokeByTelefonoID(telefonoID int64) error {
	_, err := s.db.Exec(`UPDATE api_keys SET activo = FALSE, revoked_at = NOW(), updated_at = NOW() WHERE telefono_id = ?`, telefonoID)
	if err != nil {
		return fmt.Errorf("error al revocar API keys por telefono: %w", err)
	}
	return nil
}

// Revoke desactiva una API key específica
func (s *ApiKeyStore) Revoke(id int64) error {
	query := `UPDATE api_keys SET activo = FALSE, revoked_at = NOW(), updated_at = NOW() WHERE id = ?`

	_, err := s.db.Exec(query, id)
	return err
}

// TouchLastUsed actualiza last_used_at al validar o consumir una key.
func (s *ApiKeyStore) TouchLastUsed(id int64) error {
	_, err := s.db.Exec(`UPDATE api_keys SET last_used_at = NOW(), updated_at = NOW() WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("error al actualizar last_used_at: %w", err)
	}
	return nil
}
