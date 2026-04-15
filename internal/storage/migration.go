package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Migration struct {
	Version     int
	Description string
	AppliedAt   string
	Checksum    string
}

type MigrationRunner struct {
	migrationsPath string
}

func NewMigrationRunner() *MigrationRunner {
	return &MigrationRunner{
		migrationsPath: "internal/storage/migrations",
	}
}

func (r *MigrationRunner) RunMigrations(db *sql.DB) error {
	if err := r.ensureSchemaMigrationsTable(db); err != nil {
		return err
	}

	currentVersion, err := r.GetCurrentVersion(db)
	if err != nil {
		return err
	}

	migrations, err := r.getMigrationFiles()
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if m.Version <= currentVersion {
			continue
		}

		if err := r.runUp(db, m); err != nil {
			return fmt.Errorf("migration %d failed: %w", m.Version, err)
		}

		if err := r.recordMigration(db, m); err != nil {
			return fmt.Errorf("failed to record migration %d: %w", m.Version, err)
		}
	}

	return nil
}

func (r *MigrationRunner) Rollback(db *sql.DB) error {
	currentVersion, err := r.GetCurrentVersion(db)
	if err != nil {
		return err
	}

	if currentVersion == 0 {
		return fmt.Errorf("no migrations to rollback")
	}

	migrations, err := r.getMigrationFiles()
	if err != nil {
		return err
	}

	var lastMigration Migration
	for _, m := range migrations {
		if m.Version == currentVersion {
			lastMigration = m
			break
		}
	}

	if lastMigration.Version == 0 {
		return fmt.Errorf("migration file for version %d not found", currentVersion)
	}

	if err := r.runDown(db, lastMigration); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	if err := r.removeMigrationRecord(db, currentVersion); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	return nil
}

func (r *MigrationRunner) GetCurrentVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	return version, err
}

func (r *MigrationRunner) GetAppliedMigrations(db *sql.DB) ([]Migration, error) {
	rows, err := db.Query("SELECT version, description, applied_at, checksum FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []Migration
	for rows.Next() {
		var m Migration
		if err := rows.Scan(&m.Version, &m.Description, &m.AppliedAt, &m.Checksum); err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}

	return migrations, rows.Err()
}

func (r *MigrationRunner) ensureSchemaMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INT PRIMARY KEY,
			description VARCHAR(255),
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			checksum VARCHAR(64)
		)
	`
	_, err := db.Exec(query)
	return err
}

func (r *MigrationRunner) getMigrationFiles() ([]Migration, error) {
	entries, err := os.ReadDir(r.migrationsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Migration{}, nil
		}
		return nil, err
	}

	var migrations []Migration
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			m, err := r.parseMigrationFile(entry.Name())
			if err != nil {
				continue
			}
			if m.Description != "" {
				migrations = append(migrations, m)
			}
		}
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func (r *MigrationRunner) parseMigrationFile(filename string) (Migration, error) {
	name := filename
	// Remove .up.sql or .down.sql extension
	if strings.HasSuffix(name, ".up.sql") {
		name = name[:len(name)-7] // remove ".up.sql"
	} else if strings.HasSuffix(name, ".down.sql") {
		name = name[:len(name)-9] // remove ".down.sql"
	}

	// Now split by first underscore to get version and description
	parts := strings.SplitN(name, "_", 2)
	if len(parts) != 2 {
		return Migration{}, fmt.Errorf("invalid migration filename: %s", filename)
	}
	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return Migration{}, err
	}
	description := parts[1]

	return Migration{
		Version:     version,
		Description: description,
	}, nil
}

func (r *MigrationRunner) runUp(db *sql.DB, m Migration) error {
	// m.Description already has the full name without extension
	filepath := filepath.Join(r.migrationsPath, fmt.Sprintf("%03d_%s.up.sql", m.Version, m.Description))
	content, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	_, err = db.Exec(string(content))
	return err
}

func (r *MigrationRunner) runDown(db *sql.DB, m Migration) error {
	filename := fmt.Sprintf("%03d_%s.down.sql", m.Version, m.Description)
	filepath := filepath.Join(r.migrationsPath, filename)
	content, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	_, err = db.Exec(string(content))
	return err
}

func (r *MigrationRunner) recordMigration(db *sql.DB, m Migration) error {
	checksum := fmt.Sprintf("%x", time.Now().Unix())
	_, err := db.Exec(
		"INSERT INTO schema_migrations (version, description, checksum) VALUES (?, ?, ?)",
		m.Version, m.Description, checksum,
	)
	return err
}

func (r *MigrationRunner) removeMigrationRecord(db *sql.DB, version int) error {
	_, err := db.Exec("DELETE FROM schema_migrations WHERE version = ?", version)
	return err
}

func (r *MigrationRunner) ListMigrations() ([]Migration, error) {
	migrations, err := r.getMigrationFiles()
	if err != nil {
		return nil, err
	}
	return migrations, nil
}
