package storage

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

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
	keyHash := HashKey(rawKey)

	query := `INSERT INTO api_keys (empresa_id, key_hash, nombre, activo, expires_at) 
			  VALUES (?, ?, ?, ?, ?)`

	result, err := s.db.Exec(query, apiKey.EmpresaID, keyHash, apiKey.Nombre, apiKey.Activo, apiKey.ExpiresAt)
	if err != nil {
		return "", fmt.Errorf("error al crear API key: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("error al obtener ID: %w", err)
	}

	apiKey.ID = id
	apiKey.KeyHash = keyHash
	apiKey.CreatedAt = time.Now()

	return rawKey, nil
}

// GetByID obtiene una API key por ID (sin el hash)
func (s *ApiKeyStore) GetByID(id int64) (*domain.ApiKey, error) {
	query := `SELECT id, empresa_id, key_hash, nombre, activo, created_at, expires_at 
			  FROM api_keys WHERE id = ?`

	key := &domain.ApiKey{}
	var expiresAt sql.NullTime

	err := s.db.QueryRow(query, id).Scan(&key.ID, &key.EmpresaID, &key.KeyHash,
		&key.Nombre, &key.Activo, &key.CreatedAt, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener API key: %w", err)
	}

	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Time
	}

	// Don't expose hash
	key.KeyHash = ""

	return key, nil
}

// Validate verifica si una API key es válida
func (s *ApiKeyStore) Validate(rawKey string) (*domain.ApiKey, error) {
	keyHash := HashKey(rawKey)

	query := `SELECT id, empresa_id, key_hash, nombre, activo, created_at, expires_at 
			  FROM api_keys WHERE key_hash = ? AND activo = TRUE`

	key := &domain.ApiKey{}
	var expiresAt sql.NullTime

	err := s.db.QueryRow(query, keyHash).Scan(&key.ID, &key.EmpresaID, &key.KeyHash,
		&key.Nombre, &key.Activo, &key.CreatedAt, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al validar API key: %w", err)
	}

	if expiresAt.Valid && time.Now().After(expiresAt.Time) {
		return nil, nil // Expired
	}

	key.KeyHash = "" // Don't expose
	return key, nil
}

// GetByEmpresaID obtiene todas las API keys de una empresa
func (s *ApiKeyStore) GetByEmpresaID(empresaID int64) ([]domain.ApiKey, error) {
	query := `SELECT id, empresa_id, key_hash, nombre, activo, created_at, expires_at 
			  FROM api_keys WHERE empresa_id = ? ORDER BY created_at DESC`

	rows, err := s.db.Query(query, empresaID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener API keys: %w", err)
	}
	defer rows.Close()

	var keys []domain.ApiKey
	for rows.Next() {
		var k domain.ApiKey
		var expiresAt sql.NullTime

		err := rows.Scan(&k.ID, &k.EmpresaID, &k.KeyHash, &k.Nombre, &k.Activo, &k.CreatedAt, &expiresAt)
		if err != nil {
			return nil, fmt.Errorf("error al escanear API key: %w", err)
		}

		if expiresAt.Valid {
			k.ExpiresAt = &expiresAt.Time
		}

		k.KeyHash = "" // Don't expose
		keys = append(keys, k)
	}

	return keys, nil
}

// Delete elimina (desactiva) una API key
func (s *ApiKeyStore) Delete(id int64) error {
	query := `UPDATE api_keys SET activo = FALSE WHERE id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error al eliminar API key: %w", err)
	}

	return nil
}

// Revoke desactiva una API key específica
func (s *ApiKeyStore) Revoke(id int64) error {
	query := `UPDATE api_keys SET activo = FALSE WHERE id = ?`

	_, err := s.db.Exec(query, id)
	return err
}
