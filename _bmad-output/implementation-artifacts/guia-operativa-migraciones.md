---
status: done
created: 2026-04-18
last_updated: 2026-04-18
---

# Guía Operativa de Migraciones

## Stack

- Librería: `github.com/golang-migrate/migrate/v4`
- Driver DB: MySQL
- Source: file
- Ruta de migraciones: `internal/storage/migrations`

## Comandos

### Ver estado

```bash
go run . migrate status
```

Muestra:

- versión actual
- cantidad de migraciones aplicadas
- detalle de migraciones aplicadas

### Aplicar migraciones

```bash
go run . migrate up
```

Aplica todas las migraciones pendientes.

### Revertir última migración

```bash
go run . migrate down
```

Revierte un solo paso.

## Flujo Real

1. `main.go` recibe el subcomando `migrate`.
2. `runMigrateCommand()` crea la conexión MySQL.
3. `storage.NewMigrationRunner()` resuelve `internal/storage/migrations`.
4. `RunMigrations()` elimina la tabla legacy `schema_migrations` solo si tiene el esquema viejo.
5. `m.Up()` aplica lo pendiente.

## Convenciones

- Cada migración nueva debe tener par `*.up.sql` y `*.down.sql`.
- El nombre debe ser secuencial y descriptivo.
- `up` crea o altera.
- `down` revierte exactamente el cambio del `up`.

## Estado Dirty

Si `GetCurrentVersion()` detecta `dirty = true`, la base está en estado incompleto o fallido.

Acciones:

1. revisar la última migración ejecutada
2. validar el archivo SQL correspondiente
3. usar `status` antes de volver a correr `up`

## Referencias

- `main.go`
- `internal/storage/migration.go`
- `internal/storage/migrations/`
- `go.mod`
