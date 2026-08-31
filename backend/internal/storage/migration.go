package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/MMaZX/goforge/database/mariadb"
	"github.com/MMaZX/goforge/migration"
)

// MigrationRunner ejecuta las migraciones SQL embebidas usando el motor de
// goforge (github.com/MMaZX/goforge) contra MySQL/MariaDB. Sustituye al runner
// casero basado en golang-migrate.
type MigrationRunner struct {
	dsn string
}

// NewMigrationRunner crea un runner que se conecta con el DSN indicado
// (formato go-sql-driver/mysql, p. ej. "user:pass@tcp(host:3306)/db?parseTime=true").
func NewMigrationRunner(dsn string) *MigrationRunner {
	return &MigrationRunner{dsn: dsn}
}

// migrationsSource devuelve el sub-FS con los archivos NNNNNN_*.{up,down}.sql.
func migrationsSource() (fs.FS, error) {
	sub, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir el sub-FS de migraciones: %w", err)
	}
	return sub, nil
}

// open construye el provider de MariaDB y el motor de goforge a partir de las
// migraciones embebidas. El caller debe invocar provider.Close().
func (r *MigrationRunner) open(ctx context.Context) (*migration.Engine, *mariadb.Provider, error) {
	src, err := migrationsSource()
	if err != nil {
		return nil, nil, err
	}
	entries, err := migration.Load(src, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("no se pudieron cargar las migraciones: %w", err)
	}

	provider, err := mariadb.Open(ctx, r.dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("no se pudo conectar a la base de datos: %w", err)
	}

	engine, err := migration.NewEngine(provider.DB(), provider, entries)
	if err != nil {
		provider.Close()
		return nil, nil, fmt.Errorf("no se pudo crear el motor de migraciones: %w", err)
	}
	return engine, provider, nil
}

// RunMigrations aplica todas las migraciones pendientes en orden ascendente.
// Que no haya pendientes no se considera un error.
func (r *MigrationRunner) RunMigrations(ctx context.Context) error {
	engine, provider, err := r.open(ctx)
	if err != nil {
		return err
	}
	defer provider.Close()

	if _, err := engine.Up(ctx, 0); err != nil {
		if errors.Is(err, migration.ErrNoMigrations) {
			return nil
		}
		return fmt.Errorf("fallo al aplicar migraciones: %w", err)
	}
	return nil
}

// Rollback revierte migraciones aplicadas. steps == 0 revierte el último batch;
// steps > 0 revierte esas últimas N migraciones.
func (r *MigrationRunner) Rollback(ctx context.Context, steps int) error {
	engine, provider, err := r.open(ctx)
	if err != nil {
		return err
	}
	defer provider.Close()

	if _, err := engine.Rollback(ctx, steps); err != nil {
		if errors.Is(err, migration.ErrNoMigrations) {
			return nil
		}
		return fmt.Errorf("fallo al revertir migración: %w", err)
	}
	return nil
}

// Status devuelve el estado (aplicada / pendiente / dirty) de cada migración
// conocida, ordenado por versión.
func (r *MigrationRunner) Status(ctx context.Context) ([]migration.StatusEntry, error) {
	engine, provider, err := r.open(ctx)
	if err != nil {
		return nil, err
	}
	defer provider.Close()

	return engine.Status(ctx)
}

// Adopt siembra la tabla de historial de goforge (goforge_migrations) a partir
// de la tabla legacy schema_migrations que dejó golang-migrate: marca como
// aplicadas, con su checksum correcto, todas las migraciones cuya versión sea
// menor o igual a la versión registrada por golang-migrate. Es idempotente y no
// modifica ni elimina la tabla legacy.
func (r *MigrationRunner) Adopt(ctx context.Context) ([]uint64, error) {
	src, err := migrationsSource()
	if err != nil {
		return nil, err
	}
	entries, err := migration.Load(src, nil)
	if err != nil {
		return nil, fmt.Errorf("no se pudieron cargar las migraciones: %w", err)
	}

	provider, err := mariadb.Open(ctx, r.dsn)
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar a la base de datos: %w", err)
	}
	defer provider.Close()

	db := provider.DB()
	hist := provider.History()

	legacyVersion, hasLegacy, err := legacySchemaMigrationsVersion(ctx, db)
	if err != nil {
		return nil, err
	}
	if !hasLegacy {
		return nil, errors.New("no se encontró la tabla schema_migrations (golang-migrate); no hay historial que adoptar")
	}

	if err := hist.EnsureTable(ctx, db); err != nil {
		return nil, fmt.Errorf("no se pudo crear goforge_migrations: %w", err)
	}
	existing, err := hist.List(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer goforge_migrations: %w", err)
	}
	already := make(map[uint64]struct{}, len(existing))
	for _, rec := range existing {
		already[rec.Version] = struct{}{}
	}

	var adopted []uint64
	for _, e := range entries {
		v := e.Version()
		if v > legacyVersion {
			continue
		}
		if _, ok := already[v]; ok {
			continue
		}
		if err := hist.Begin(ctx, db, v, e.Name(), e.Checksum, 1); err != nil {
			return adopted, fmt.Errorf("no se pudo registrar la migración %d: %w", v, err)
		}
		if err := hist.Complete(ctx, db, v, 0); err != nil {
			return adopted, fmt.Errorf("no se pudo marcar como aplicada la migración %d: %w", v, err)
		}
		adopted = append(adopted, v)
	}
	return adopted, nil
}

// legacySchemaMigrationsVersion lee la versión actual de la tabla
// schema_migrations dejada por golang-migrate. Devuelve (0, false, nil) si la
// tabla no existe.
func legacySchemaMigrationsVersion(ctx context.Context, db migration.DB) (uint64, bool, error) {
	var version uint64
	var dirty bool
	err := db.QueryRowContext(ctx,
		"SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1",
	).Scan(&version, &dirty)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		if isTableMissing(err) {
			return 0, false, nil
		}
		if isUnknownColumn(err) {
			// Tabla legacy sin la columna dirty (esquema propio anterior a
			// golang-migrate): tomamos solo la versión máxima.
			var maxVersion sql.NullInt64
			if err2 := db.QueryRowContext(ctx,
				"SELECT MAX(version) FROM schema_migrations",
			).Scan(&maxVersion); err2 != nil {
				if isTableMissing(err2) {
					return 0, false, nil
				}
				return 0, false, fmt.Errorf("no se pudo leer schema_migrations: %w", err2)
			}
			if !maxVersion.Valid {
				return 0, false, nil
			}
			return uint64(maxVersion.Int64), true, nil
		}
		return 0, false, fmt.Errorf("no se pudo leer schema_migrations: %w", err)
	}
	if dirty {
		return version, true, fmt.Errorf(
			"schema_migrations legacy está en estado dirty (versión %d): resuélvelo antes de adoptar", version)
	}
	return version, true, nil
}

func isTableMissing(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "doesn't exist") ||
		strings.Contains(msg, "Unknown table") ||
		strings.Contains(msg, "no such table")
}

func isUnknownColumn(err error) bool {
	return strings.Contains(err.Error(), "Unknown column")
}
