package storage

// [QUÉ] Implementa la conexión al pool de base de datos MariaDB.
// [POR QUÉ] Centralizar la creación del *sql.DB garantiza que toda la aplicación
// comparta el mismo pool de conexiones (evita resource leaks y mejora rendimiento).

import (
	"database/sql"
	"fmt"

	// [QUÉ] Blank import registra el driver MySQL/MariaDB en database/sql.
	// [POR QUÉ] Go usa el patrón "registro de drivers": importar con _ ejecuta el
	// init() del paquete que llama sql.Register("mysql", ...).
	_ "github.com/go-sql-driver/mysql"

	"wsapi/internal/config"
)

// NewDB crea y retorna un pool de conexiones a MariaDB listo para usar.
// Retorna error si la cadena de conexión es inválida o si la DB no responde.
func NewDB(cfg *config.Config) (*sql.DB, error) {
	// [QUÉ] DSN (Data Source Name) con parámetros de seguridad y charset.
	// [POR QUÉ] parseTime=true convierte columnas DATETIME/TIMESTAMP directamente a time.Time en Go.
	// charset=utf8mb4 soporta emojis y caracteres especiales (importante para mensajes WhatsApp).
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	// [QUÉ] sql.Open NO establece conexión todavía; solo valida el driver y el DSN.
	// [POR QUÉ] La conexión real ocurre en el primer uso o en el Ping de abajo.
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error al abrir conexión a DB: %w", err)
	}

	// [QUÉ] Ping fuerza al driver a establecer una conexión real y verificar credenciales.
	// [POR QUÉ] Fail-fast en startup: mejor fallar aquí que en el primer request de producción.
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error al conectar con la DB (ping falló): %w", err)
	}

	return db, nil
}
