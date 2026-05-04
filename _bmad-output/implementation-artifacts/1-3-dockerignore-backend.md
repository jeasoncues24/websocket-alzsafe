# Story 1.3: dockerignore backend

Status: review

## Story

As a desarrollador de despliegue,
I want que `backend/.dockerignore` excluya todos los archivos innecesarios y sensibles del build context de Docker,
so that el build sea seguro (sin secretos en capas), eficiente (context mínimo) y correcto (todos los archivos requeridos para compilar siguen incluidos).

## Acceptance Criteria

1. `backend/.dockerignore` excluye todos los archivos de entorno y secretos: `.env`, `.env.*`, `.env.copy` — ninguno puede aparecer en el build context.
2. `backend/.dockerignore` excluye archivos innecesarios para compilación: `docs/`, `*.md`, `*.test`, `coverage.txt`, `*.log`, `.gitignore`.
3. Los archivos requeridos para compilar siguen disponibles en el build context: `go.mod`, `go.sum`, `main.go`, `internal/`, `cmd/` — ninguno está excluido.
4. `backend/.dockerignore` tiene comentarios que agrupan las exclusiones por categoría (secretos/entorno, documentación, artefactos de test, metadata de herramientas).
5. `docker compose config` valida el compose sin errores tras el cambio, y el Dockerfile sigue usando el mismo `COPY go.mod go.sum ./` + `COPY . .` sin necesitar ajustes.

## Tasks / Subtasks

- [x] Auditar archivos en `backend/` que no son necesarios para compilación (AC: 1, 2, 3)
  - [x] Identificar archivos de entorno/secretos: `.env`, `.env.*`, `.env.copy`
  - [x] Identificar documentación: `docs/`, `*.md`
  - [x] Identificar artefactos de test: `*.test`, `coverage.txt`
  - [x] Identificar metadata de herramientas: `.gitignore`, `*.log`, `*.swp`, `*.swo`
  - [x] Confirmar que `go.mod`, `go.sum`, `main.go`, `internal/`, `cmd/` NO deben excluirse

- [x] Actualizar `backend/.dockerignore` con exclusiones completas y comentadas (AC: 1, 2, 4)
  - [x] Sección: secretos/entorno
  - [x] Sección: documentación
  - [x] Sección: artefactos de test y cobertura
  - [x] Sección: metadata de herramientas (git, IDE, temporales)

- [x] Validar que el build context es correcto tras el cambio (AC: 3, 5)
  - [x] Verificar con `docker compose config` que el compose sigue siendo válido
  - [x] Confirmar que el Dockerfile (`COPY go.mod go.sum ./` + `COPY . .`) no requiere cambios

## Dev Notes

### Estado actual de `backend/.dockerignore`

Creado en story 1-1 (review patch #2). Contenido actual:

```
# Excluye archivos de entorno del build context para no filtrar secretos en capas de imagen.
.env
.env.*
```

Solo cubre secretos de entorno. Faltan exclusiones para documentación, test artifacts y metadata.

### Contexto del Dockerfile (NO modificar en esta story)

```dockerfile
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN ... go build -o /out/wsapi .
```

El `COPY . .` envía TODO el build context al daemon. El `.dockerignore` controla qué llega. Es crítico que `go.mod`, `go.sum`, `main.go`, `internal/` y `cmd/` NO estén excluidos.

### Archivos en `backend/` — análisis de necesidad

| Archivo/Dir | Necesario para build | Acción |
|-------------|---------------------|--------|
| `go.mod`, `go.sum` | ✅ Sí | Incluir (no excluir) |
| `main.go` | ✅ Sí | Incluir (no excluir) |
| `internal/` | ✅ Sí | Incluir (no excluir) |
| `cmd/` | ✅ Sí | Incluir (no excluir) |
| `.env` | ❌ Secreto | Excluir (ya hecho) |
| `.env.*` | ❌ Secreto | Excluir (ya hecho) |
| `.env.copy` | ❌ Referencia | Excluir |
| `docs/` | ❌ Documentación | Excluir |
| `*.md` | ❌ Documentación | Excluir |
| `*.test` | ❌ Binarios test | Excluir |
| `coverage.txt` | ❌ Artefacto test | Excluir |
| `*.log` | ❌ Logs | Excluir |
| `.gitignore` | ❌ Metadata git | Excluir |
| `*.swp`, `*.swo` | ❌ Temporales IDE | Excluir |

### Relación con `.dockerignore` en raíz del repo

El `.dockerignore` de la raíz **no aplica** al build context `./backend`. Docker usa el `.dockerignore` ubicado en la raíz del build context (o junto al Dockerfile, con precedencia al del context). Dado que:
- `context: ./backend` en `docker-compose.yml`
- `backend/.dockerignore` existe → Docker lo usa para este build

Son archivos independientes con propósitos distintos:
- `.dockerignore` (raíz): aplica si el context fuera la raíz del repo (no es el caso actual)
- `backend/.dockerignore`: aplica al build del backend

### Restricciones técnicas

- No modificar `docker/go/Dockerfile` (scope de story 1-1).
- No modificar `docker-compose.yml` (scope de story 1-2).
- No modificar el `.dockerignore` de la raíz (diferente propósito).
- El Dockerfile usa `COPY . .` — el `.dockerignore` es la única barrera; debe ser exhaustivo.

### Verificación manual esperada

```bash
# Verificar que el compose sigue válido:
docker compose config

# Opcional — ver qué archivos llegarían al build context (requiere BuildKit):
# docker build --no-cache --progress=plain -f docker/go/Dockerfile ./backend 2>&1 | grep "transferring"
```

### References

- [Source: backend/.dockerignore] — estado actual (solo .env)
- [Source: docker/go/Dockerfile] — usa COPY go.mod go.sum + COPY . .
- [Source: docker-compose.yml] — build context = ./backend
- [Source: _bmad-output/planning-artifacts/prd.md#FR8] — verificar que archivos requeridos no estén excluidos
- [Source: _bmad-output/planning-artifacts/prd.md#NFR11] — secretos no deben quedar en imágenes
- [Source: _bmad-output/implementation-artifacts/1-1-dockerfile-optimizado.md] — story 1-1 creó backend/.dockerignore inicial
- [Source: _bmad-output/implementation-artifacts/1-2-docker-compose-build-only.md] — story 1-2: contexto build

## Dev Agent Record

### Agent Model Used

gpt-5

### Debug Log References

- `./scripts/test-backend-dockerignore.sh` (falló antes del cambio por secciones/reglas faltantes; pasó después de actualizar `backend/.dockerignore`).
- `docker compose config` OK.
- `cd backend && go build ./...` OK.
- `cd backend && go test ./...` OK.

### Completion Notes List

- Se amplió `backend/.dockerignore` con exclusiones agrupadas por categoría: secretos/entorno, documentación, artefactos de test/cobertura y metadata/temporales.
- Se excluyó explícitamente `.env.copy` además de `.env` y `.env.*` para evitar que referencias de entorno entren al build context.
- Se confirmó que `go.mod`, `go.sum`, `main.go`, `internal/` y `cmd/` siguen disponibles para el `COPY . .` del Dockerfile.
- Se agregó una validación automatizada en `scripts/test-backend-dockerignore.sh` para proteger las reglas mínimas de `backend/.dockerignore`.
- Se verificó que `docker compose config` sigue válido y que el Dockerfile mantiene `COPY go.mod go.sum ./` + `COPY . .` sin cambios.

### File List

- `backend/.dockerignore`
- `scripts/test-backend-dockerignore.sh`

## Change Log

- 2026-05-03: Story 1.3 implementada; `backend/.dockerignore` endurecido y validación automatizada agregada.
