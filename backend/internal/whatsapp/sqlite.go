package whatsapp

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.mau.fi/whatsmeow/store/sqlstore"
	_ "modernc.org/sqlite"
)

// sqliteDBPath calcula la ruta del archivo SQLite de un accountID, aplicando el
// mismo baseDir por defecto y saneo de nombre que usa openSQLiteContainer.
func sqliteDBPath(baseDir, accountID string) string {
	if strings.TrimSpace(baseDir) == "" {
		baseDir = "sessions/whatsappmeow"
	}
	return filepath.Join(baseDir, sanitizeSQLiteFilename(accountID)+".db")
}

func openSQLiteContainer(ctx context.Context, baseDir, accountID string) (*sqlstore.Container, error) {
	if strings.TrimSpace(baseDir) == "" {
		baseDir = "sessions/whatsappmeow"
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("no se pudo crear el directorio sqlite: %w", err)
	}

	path := sqliteDBPath(baseDir, accountID)
	container, err := openSQLiteContainerAtPath(ctx, path, accountID)
	if err == nil {
		return container, nil
	}
	if !isWhatsmeowUpgradeConflictError(err) {
		return nil, err
	}

	if container != nil {
		_ = container.Close()
	}
	if err := removeSQLiteArtifacts(path); err != nil {
		return nil, fmt.Errorf("no se pudo reiniciar sqlite store: %w", err)
	}

	container, err = openSQLiteContainerAtPath(ctx, path, accountID)
	if err != nil {
		return nil, fmt.Errorf("no se pudo recrear sqlite store: %w", err)
	}
	return container, nil
}

func openSQLiteContainerAtPath(ctx context.Context, path string, accountID string) (*sqlstore.Container, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", filepath.ToSlash(path))

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	dbLog := NewWhatsAppDBLogger(accountID)
	container := sqlstore.NewWithDB(db, "sqlite3", dbLog)
	if err := container.Upgrade(ctx); err != nil {
		_ = container.Close()
		return nil, fmt.Errorf("no se pudo actualizar sqlite store: %w", err)
	}

	return container, nil
}

func removeSQLiteArtifacts(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Remove(path + "-wal"); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Remove(path + "-shm"); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func sanitizeSQLiteFilename(accountID string) string {
	accountID = NormalizeAccountID(accountID)
	if accountID == "" {
		return "default"
	}

	var b strings.Builder
	for _, r := range accountID {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
			continue
		}
		b.WriteRune('_')
	}

	result := strings.Trim(b.String(), "_")
	if result == "" {
		return "default"
	}
	return result
}

func isWhatsmeowUpgradeConflictError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "already exists") && strings.Contains(msg, "whatsmeow_")
}
