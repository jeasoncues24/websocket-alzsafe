package storage

import (
	"database/sql"
	"fmt"
	"time"
)

type TokenBlacklistStore struct {
	db *sql.DB
}

func NewTokenBlacklistStore(db *sql.DB) *TokenBlacklistStore {
	return &TokenBlacklistStore{db: db}
}

// Add adds a token to the blacklist
func (s *TokenBlacklistStore) Add(jti string, expiresAt time.Time) error {
	query := `INSERT INTO token_blacklist (jti, expires_at) VALUES (?, ?) ON DUPLICATE KEY UPDATE expires_at = ?`
	_, err := s.db.Exec(query, jti, expiresAt, expiresAt)
	return err
}

// IsBlacklisted checks if a token is blacklisted
func (s *TokenBlacklistStore) IsBlacklisted(jti string) (bool, error) {
	query := `SELECT 1 FROM token_blacklist WHERE jti = ? AND expires_at > NOW() LIMIT 1`
	var exists int
	err := s.db.QueryRow(query, jti).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("error checking blacklist: %w", err)
	}
	return true, nil
}

// Cleanup removes expired tokens from blacklist
func (s *TokenBlacklistStore) Cleanup() error {
	query := `DELETE FROM token_blacklist WHERE expires_at < NOW()`
	_, err := s.db.Exec(query)
	return err
}
