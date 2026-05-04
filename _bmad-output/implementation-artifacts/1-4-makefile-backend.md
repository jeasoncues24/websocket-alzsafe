# Story 1.4: makefile backend

Status: review

## Story

As a desarrollador de despliegue,
I want un `Makefile` en la raíz del repositorio que centralice los comandos más usados del ciclo build/deploy/test del backend,
so that pueda operar el proyecto sin recordar rutas de scripts ni flags específicos.

## Acceptance Criteria

1. `make build` ejecuta `scripts/build-backend.sh` (valida `backend/.env`, compila con caché Docker).
2. `make build-clean` ejecuta `scripts/build-backend.sh --no-cache` (rebuild sin caché).
3. `make start` ejecuta `scripts/start-backend.sh` (inicia wsapi con PM2, idempotente).
4. `make restart` reinicia el proceso `wsapi` en PM2 si ya está registrado; imprime mensaje claro si no está registrado.
5. `make stop` detiene `wsapi` en PM2.
6. `make logs` muestra los logs de `wsapi` en PM2 en tiempo real.
7. `make test` ejecuta `cd backend && go test ./...` desde la raíz del repo.
8. `make` o `make help` muestra una lista de targets disponibles con descripción de una línea cada uno.
9. Todos los targets declarados como `.PHONY` para no colisionar con archivos del mismo nombre.

## Tasks / Subtasks

- [x] Crear `Makefile` en la raíz del repo con todos los targets (AC: 1–9)
  - [x] Declarar `.PHONY` con todos los targets
  - [x] Target `help` como objetivo por defecto (primer target del archivo)
  - [x] Target `build` → `./scripts/build-backend.sh`
  - [x] Target `build-clean` → `./scripts/build-backend.sh --no-cache`
  - [x] Target `start` → `./scripts/start-backend.sh`
  - [x] Target `restart` → `pm2 restart wsapi` con guardia si no está registrado
  - [x] Target `stop` → `pm2 stop wsapi`
  - [x] Target `logs` → `pm2 logs wsapi`
  - [x] Target `test` → `cd backend && go test ./...`

- [x] Validar el Makefile (AC: 1, 7, 8, 9)
  - [x] `make help` muestra todos los targets sin error
  - [x] `make test` compila y pasa los tests del backend sin regresiones
  - [x] `make --dry-run build` muestra el comando correcto sin ejecutarlo

## Dev Notes

### Archivos existentes que el Makefile debe invocar

| Script | Descripción |
|--------|-------------|
| `scripts/build-backend.sh` | Valida `backend/.env`, elimina binario previo, llama `docker compose build + run` |
| `scripts/build-backend.sh --no-cache` | Lo mismo pero con `--no-cache` al build de Docker |
| `scripts/start-backend.sh` | Inicia wsapi con PM2, `--cwd backend/`, idempotente |

**No duplicar** la lógica de los scripts en el Makefile. El Makefile es solo un dispatcher.

### Patrón `make help` recomendado

Usar comentarios `##` junto a cada target y extraerlos automáticamente. Patrón estándar para Makefiles autodocumentados:

```makefile
.DEFAULT_GOAL := help

help: ## Muestra esta ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
```

### Guardia para `make restart`

`pm2 restart wsapi` falla si el proceso no existe. Añadir guardia:

```makefile
restart: ## Reinicia wsapi en PM2 (requiere que ya esté registrado)
	@pm2 describe wsapi > /dev/null 2>&1 || \
	  (echo "ERROR: wsapi no está registrado en PM2. Ejecuta 'make start' primero." && exit 1)
	pm2 restart wsapi
```

### Convenciones del Makefile

- Primera línea real: `.DEFAULT_GOAL := help`
- Todos los targets en `.PHONY`
- Usar `@` en mensajes de error/ayuda para no duplicar el comando en la salida
- No usar tabs mezclados con espacios (Make requiere tabs para recetas)
- El Makefile invoca scripts con ruta relativa al repo: `./scripts/...`

### Restricciones

- No duplicar lógica de los scripts en el Makefile.
- No agregar targets fuera del alcance de esta story (frontend, DB, nginx — son stories posteriores).
- El Makefile vive en la raíz del repo, no en `backend/`.
- No modificar `scripts/build-backend.sh` ni `scripts/start-backend.sh`.

### Verificación de `make test`

```bash
make test
# Equivale a:
cd backend && go test ./...
```

El backend ya compila sin errores (`go build ./...` OK en story 1-1). `go test ./...` puede tener tests vacíos — es aceptable; lo que importa es que no falle.

### References

- [Source: scripts/build-backend.sh] — invocado por `make build` y `make build-clean`
- [Source: scripts/start-backend.sh] — invocado por `make start`
- [Source: docs/deploy-backend.md] — documenta ciclo completo; Makefile complementa la doc
- [Source: _bmad-output/project-context.md] — "Mantener comandos: `cd backend && go test ./...` y `cd backend && go build ./...`"
- [Source: _bmad-output/planning-artifacts/prd.md#NFR9] — deploy repetible sin conocimiento implícito

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

### Completion Notes List

- `Makefile` creado en raíz con 8 targets: `help`, `build`, `build-clean`, `start`, `restart`, `stop`, `logs`, `test`.
- `.DEFAULT_GOAL := help` — `make` sin argumentos muestra ayuda.
- `make help` usa `grep + awk` sobre comentarios `##` para autodocumentarse.
- `make restart` tiene guardia `pm2 describe wsapi` — falla con mensaje claro si no está registrado.
- `make test` ejecuta `cd backend && go test ./...` — 6 paquetes con tests pasan, 5 sin test files (OK).
- `make --dry-run build/build-clean` muestra comandos correctos sin ejecutar.
- Todos los targets en `.PHONY`.

### Change Log

- 2026-05-04: Story 1-4 implementada. Makefile creado con 8 targets de build/deploy/test.

### File List

- `Makefile`
- `_bmad-output/implementation-artifacts/1-4-makefile-backend.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
