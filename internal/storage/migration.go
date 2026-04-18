package storage

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type MigrationRunner struct {
	migrationsPath string
}

func NewMigrationRunner() *MigrationRunner {
	return &MigrationRunner{
		migrationsPath: "internal/storage/migrations",
	}
}

func (r *MigrationRunner) RunMigrations(db *sql.DB) error {
	if err := r.dropLegacySchemaMigrationsTable(db); err != nil {
		return err
	}

	m, err := r.newMigrator(db)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

func (r *MigrationRunner) Rollback(db *sql.DB) error {
	m, err := r.newMigrator(db)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Steps(-1); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return fmt.Errorf("rollback failed: %w", err)
	}

	return nil
}

func (r *MigrationRunner) GetCurrentVersion(db *sql.DB) (int, error) {
	m, err := r.newMigrator(db)
	if err != nil {
		return 0, fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion || err == migrate.ErrNoChange {
			return 0, nil
		}
		return 0, err
	}

	if dirty {
		return int(version), fmt.Errorf("database is in a dirty state")
	}

	return int(version), nil
}

func (r *MigrationRunner) newMigrator(db *sql.DB) (*migrate.Migrate, error) {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create mysql driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", r.migrationsPath),
		"mysql",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return m, nil
}

// dropLegacySchemaMigrationsTable drops the custom schema_migrations table that
// pre-dated golang-migrate. It only drops the table when it has the legacy schema
// (identified by the presence of the 'description' column). Once golang-migrate
// takes over it creates its own schema_migrations with just version+dirty, so
// subsequent calls are safe no-ops.
func (r *MigrationRunner) dropLegacySchemaMigrationsTable(db *sql.DB) error {
	var colName string
	err := db.QueryRow(
		"SELECT column_name FROM information_schema.columns " +
			"WHERE table_schema = DATABASE() AND table_name = 'schema_migrations' " +
			"AND column_name = 'description' LIMIT 1",
	).Scan(&colName)
	if err == sql.ErrNoRows {
		// Table doesn't exist or already has the golang-migrate schema — nothing to do.
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to inspect schema_migrations: %w", err)
	}
	_, err = db.Exec("DROP TABLE schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to drop legacy schema_migrations table: %w", err)
	}
	return nil
}

type Migration struct {
	Version     int
	Description string
	AppliedAt   string
	Checksum    string
}

func (r *MigrationRunner) GetAppliedMigrations(db *sql.DB) ([]Migration, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations WHERE dirty = 0 ORDER BY version")
	if err != nil {
		// Treat "table doesn't exist" as no migrations applied.
		if strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "no such table") {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query schema_migrations: %w", err)
	}
	defer rows.Close()

	var migrations []Migration
	for rows.Next() {
		var m Migration
		if err := rows.Scan(&m.Version); err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}
	return migrations, rows.Err()
}
