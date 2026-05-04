# Story 1.6: Migraciones embebidas en el binario

Status: done

## Story

As a operador de despliegue,
I want que el binario `wsapi` lleve sus migraciones SQL embebidas en sí mismo,
so that pueda ejecutar `./wsapi migrate up` desde cualquier directorio sin depender de la estructura del proyecto fuente.

## Contexto y Problema

El binario compilado buscaba las migraciones en `file://internal/storage/migrations` — una ruta relativa al directorio de trabajo. Funcionaba cuando se ejecutaba desde la raíz del proyecto, pero fallaba al ejecutarse desde `dist/` en producción:

```
Error running migrations: failed to create migrator: failed to open source,
"file://internal/storage/migrations": open .: no such file or directory
```

La causa raíz: el binario dependía de artefactos externos (archivos SQL) que debían estar presentes en el servidor con una estructura de directorios específica. Esto contradice el modelo de despliegue del proyecto (binario único exportado a `dist/`).

## Solución adoptada: `embed.FS`

Las migraciones SQL se embeben en el binario en tiempo de compilación usando la directiva `//go:embed` de Go. El binario resultante es completamente autónomo — no requiere archivos externos para ejecutar migraciones.

**Por qué esta opción y no las alternativas:**
- `MIGRATIONS_PATH` env var: añade dependencia operacional; el operador debe gestionar la sincronización entre archivos y binario.
- Copiar carpeta junto al binario: frágil, el paso manual puede omitirse en deploys.
- `embed.FS`: binario + migraciones son atómicamente la misma versión. El compilador falla si los archivos SQL no existen. Cero dependencias externas en runtime.

## Acceptance Criteria

1. El binario `./wsapi migrate up` se ejecuta correctamente desde cualquier directorio sin archivos SQL externos presentes.
2. Las migraciones SQL (`*.sql`) quedan embebidas en el binario en tiempo de compilación.
3. Si los archivos SQL de migraciones no existen al compilar, el build falla con error claro (garantía del compilador).
4. La API del `MigrationRunner` no cambia — el resto del código no requiere modificaciones.

## Tasks / Subtasks

- [x] Crear `backend/internal/storage/migrations_embed.go` con directiva `//go:embed` (AC: 2, 3)
- [x] Modificar `backend/internal/storage/migration.go`: reemplazar driver `file` por `iofs`, eliminar campo `migrationsPath` (AC: 1, 4)
- [x] Verificar compilación limpia con `go build ./...` (AC: 1, 2, 3)

## Archivos modificados

| Archivo | Cambio |
|---------|--------|
| `backend/internal/storage/migrations_embed.go` | **Nuevo** — define `//go:embed migrations/*.sql` |
| `backend/internal/storage/migration.go` | Reemplaza driver `file` → `iofs`, elimina `migrationsPath` |

## Notas técnicas

- El paquete `iofs` es parte de `github.com/golang-migrate/migrate/v4` — no requiere dependencia nueva.
- `fs.Sub(migrationsFS, "migrations")` crea un sub-filesystem que apunta directamente al directorio embebido; `iofs.New(sub, ".")` lo registra como source driver.
- Los archivos `.up.sql` y `.down.sql` se incluyen automáticamente por el glob `migrations/*.sql`.
