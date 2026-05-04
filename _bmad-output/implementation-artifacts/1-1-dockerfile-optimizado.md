# Story 1.1: dockerfile optimizado

Status: review

<!-- Story recreada/ajustada después de discusión de alcance. Reemplaza el enfoque runtime en Docker por build-only/export en Docker + runtime con PM2 en host. -->

## Story

As a desarrollador de despliegue,
I want un flujo Docker que solo compile y exporte el binario `wsapi`, con validaciones operativas previas y arranque separado por PM2 en el host,
so that pueda generar builds reproducibles sin mezclar runtime, puertos ni dependencias de base de datos dentro de Docker.

## Acceptance Criteria

1. El `Dockerfile` usado por el flujo backend se limita a compilar y exportar el binario `wsapi`; no expone puertos, no define `CMD` de runtime y no asume ejecución de la aplicación dentro de Docker.
2. El flujo de exportación falla de forma clara si falta el archivo de entorno requerido para backend (`backend/.env` o la ruta/documento que se acuerde como oficial), antes de dejar un binario utilizable.
3. El binario compilado se exporta a `./dist/wsapi` mediante el flujo soportado del proyecto y el proceso limpia/evita artefactos viejos antes de publicar el nuevo binario.
4. Existe una forma documentada y ejecutable de forzar un rebuild limpio del binario cuando se necesite evitar caché previa (`--no-cache` o equivalente del flujo definido).
5. Existe un script idempotente para preparar ejecución con PM2 en el host: si PM2 no está disponible, lo instala; si la app `wsapi` ya existe en PM2, no crea un duplicado y termina con mensaje claro.
6. La ejecución final del binario queda fuera de Docker y usa PM2 en el host, con lectura de variables de entorno desde el entorno/archivo del host.
7. La solución deja explícito que la base de datos vive en el host y que este flujo no debe levantar ni exponer una base de datos o puerto backend desde Docker.

## Tasks / Subtasks

- [x] Ajustar el builder Docker del backend para dejarlo en modo build-only/export (AC: 1, 3, 4, 7)
  - [x] Eliminar responsabilidades de runtime del `docker/go/Dockerfile` (`EXPOSE`, `CMD`, stage runtime no necesaria si aplica)
  - [x] Mantener salida consistente del binario `wsapi` hacia una ubicación exportable por el flujo del proyecto
  - [x] Revisar el orden de capas para preservar caché útil de dependencias y recompilar correctamente cuando cambie el código
  - [x] Asegurar que exista una variante de build limpio documentada para invalidar caché cuando sea necesario

- [x] Ajustar el flujo de exportación para validar entorno requerido y limpiar artefactos previos (AC: 2, 3, 4)
  - [x] Definir el archivo de entorno oficial que debe existir antes del export (`backend/.env` salvo decisión distinta)
  - [x] Hacer que el flujo falle con mensaje claro si ese archivo no existe
  - [x] Limpiar `dist/wsapi` previo al copiado final para evitar confusión con binarios antiguos
  - [x] Confirmar que el flujo estándar sigue dejando el binario en `./dist/wsapi`

- [x] Incorporar operación host con PM2 sin mezclarla con Docker (AC: 5, 6, 7)
  - [x] Crear script idempotente para instalar/verificar PM2
  - [x] Crear script o configuración para registrar/arrancar `./dist/wsapi` con PM2 desde el host
  - [x] Cancelar alta duplicada si el proceso `wsapi` ya existe en PM2
  - [x] Validar que el runtime lea entorno del host y no dependa de runtime Docker

- [x] Documentar el flujo operativo para desarrollo/servidor (AC: 2, 4, 5, 6, 7)
  - [x] Documentar build normal vs build limpio sin caché
  - [x] Documentar precondición de `.env`
  - [x] Documentar arranque con PM2 y comportamiento idempotente
  - [x] Documentar que la base de datos está en el host y que Docker no publica el backend en este flujo

### Review Findings

- [x] [Review][Patch] El arranque con PM2 no carga el `backend/.env` oficial [scripts/start-backend.sh:31]
- [x] [Review][Patch] El build context sigue incluyendo `backend/.env` y puede filtrar secretos al stage builder [docker-compose.yml:8]
- [x] [Review][Patch] Si falta `backend/.env`, puede quedar un `dist/wsapi` viejo todavía utilizable [scripts/build-backend.sh:19]
- [x] [Review][Patch] La exportación del binario no es atómica y borra primero el artefacto previo [docker-compose.yml:13]
- [x] [Review][Patch] Los scripts dependen de ejecutarse desde la raíz del repo y no fijan `cwd`/rutas absolutas [scripts/build-backend.sh:8]
- [x] [Review][Patch] El binario `CGO_ENABLED=1` ahora se ejecuta en el host sin validar ni documentar dependencias nativas requeridas [docker/go/Dockerfile:23]
- [x] [Review][Patch] `sprint-status.yaml` quedó desalineado con los artefactos reales y lista stories inexistentes [ _bmad-output/implementation-artifacts/sprint-status.yaml:44 ]

## Dev Notes

### Contexto funcional acordado

- La decisión tomada en discusión fue separar responsabilidades:
  - Docker = compilar/exportar binario.
  - Host = ejecutar binario con PM2.
  - `.env` = validación previa del flujo operativo, no responsabilidad primaria del Dockerfile.
- Aunque el binario use variables de entorno en runtime, el usuario pidió que el flujo falle si falta `.env`. Eso debe resolverse en el flujo de build/export del host o wrapper operativo, no necesariamente dentro de la semántica del Dockerfile.

### Archivos actuales relevantes

- `docker/go/Dockerfile`
  - Hoy tiene stage builder + stage runtime.
  - Hoy copia `wsapi` a `/usr/local/bin/wsapi` en un runtime Debian.
  - Hoy incluye `EXPOSE 8080` y `CMD ["wsapi"]`, lo cual contradice el alcance acordado.
- `docker-compose.yml`
  - Hoy define `backend-build` con contexto `./backend` y usa `../docker/go/Dockerfile`.
  - Hoy el entrypoint copia `/usr/local/bin/wsapi` hacia `/dist/wsapi`.
  - Debe seguir siendo un flujo de exportación, no de runtime.
- `.dockerignore`
  - En raíz excluye `.env` y `.env.*`; revisar compatibilidad real con el contexto `./backend` si el flujo necesita solo validar presencia en host y no copiarlo al build context.
- `backend/main.go`
  - El binario usa `config.Load()` y falla si `APP_PORT` no existe al arrancar servidor.
- `backend/internal/config/config.go`
  - Usa `godotenv.Load()`; por diseño el runtime puede leer `.env` del host si se ejecuta desde el directorio apropiado o con entorno exportado.
- `backend/.env.copy`
  - Sirve como referencia de variables mínimas, pero no reemplaza la necesidad del archivo operativo real.

### Restricciones técnicas

- Mantener Go `1.25.0`.
- Mantener binario esperado: `wsapi`.
- No introducir nuevas librerías salvo necesidad clara.
- No cambiar el entrypoint real del backend (`backend/main.go`).
- No asumir base de datos dentro de Docker Compose.
- No mezclar este trabajo con despliegue completo del frontend.

### Recomendación de implementación

- Convertir `docker/go/Dockerfile` en builder-only o dejar un stage exportable sin runtime final.
- Dejar la validación de `.env` y la limpieza de `dist/wsapi` en script host o wrapper de build.
- Crear un script PM2 idempotente, por ejemplo bajo `scripts/`, que:
  - valide prerequisitos,
  - instale PM2 si hace falta,
  - no duplique la app si ya existe,
  - arranque o informe estado.
- Para evitar builds engañosos por caché:
  - mantener `COPY go.mod go.sum ./` antes de `go mod download`,
  - copiar código después,
  - documentar modo limpio con `docker compose build --no-cache backend-build` o equivalente.

### Testing Requirements

- Verificar que el flujo de build/export deja `dist/wsapi` actualizado.
- Verificar que el flujo falla cuando falta `.env`.
- Verificar que el script PM2 no crea duplicados si `wsapi` ya existe.
- Verificar que el Dockerfile final no contiene `EXPOSE` ni `CMD` de runtime.

### References

- `docker/go/Dockerfile`
- `docker-compose.yml`
- `.dockerignore`
- `backend/main.go`
- `backend/internal/config/config.go`
- `backend/.env.copy`
- `_bmad-output/project-context.md`
- `docs/bmad-project-rules.md`

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

- Story reabierta tras discusión de alcance sobre Dockerfile build-only, `.env` y PM2.

### Completion Notes List

- Se redefinió el alcance para separar build (Docker) y runtime (PM2 host).
- La status previa `review` ya no era válida porque cambió el objetivo de la story.
- Dockerfile reducido a un solo stage builder; eliminados `EXPOSE 8080`, `CMD ["wsapi"]` y el stage runtime Debian.
- `docker-compose.yml` actualizado: entrypoint copia desde `/out/wsapi` (builder) en lugar de `/usr/local/bin/wsapi` (runtime eliminado), y limpia binario previo con `rm -f`.
- `scripts/build-backend.sh`: valida `backend/.env`, crea `dist/`, soporta `--no-cache`.
- `scripts/start-backend.sh`: instala PM2 si falta, detecta proceso duplicado, arranca `./dist/wsapi`.
- `docs/deploy-backend.md`: documenta ciclo completo build + PM2 + BD en host.
- Backend Go sigue compilando sin cambios ni regresiones (`go build ./...` OK).
- [Review fixes 2026-05-03] PM2 ahora usa `--cwd backend/` para que godotenv encuentre `.env`.
- [Review fixes 2026-05-03] Creado `backend/.dockerignore` que excluye `.env` y `.env.*` del build context.
- [Review fixes 2026-05-03] `build-backend.sh` elimina binario previo antes de validar `.env` para no dejar artefactos viejos utilizables.
- [Review fixes 2026-05-03] Exportación atómica en docker-compose: `cp /out/wsapi /dist/wsapi.tmp && mv /dist/wsapi.tmp /dist/wsapi`.
- [Review fixes 2026-05-03] Scripts usan `REPO_ROOT=$(cd "$(dirname "$0")/.." && pwd)` para funcionar desde cualquier directorio.
- [Review fixes 2026-05-03] `docs/deploy-backend.md` documenta dependencia nativa `libsqlite3-0` requerida en el host.
- [Review fixes 2026-05-03] `sprint-status.yaml` corregido: stories 1-2 a 1-6 pasan a `backlog` (no tienen archivo de story).

### Change Log

- 2026-05-03: Implementación completa de story 1-1. Dockerfile build-only, scripts de build/PM2, documentación operativa.
- 2026-05-03: Review patches aplicados. Scripts con rutas absolutas, PM2 con --cwd, exportación atómica, backend/.dockerignore, docs de libsqlite3-0, sprint-status corregido.

### File List

- `docker/go/Dockerfile`
- `docker-compose.yml`
- `backend/.dockerignore`
- `scripts/build-backend.sh`
- `scripts/start-backend.sh`
- `docs/deploy-backend.md`
- `_bmad-output/implementation-artifacts/1-1-dockerfile-optimizado.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
