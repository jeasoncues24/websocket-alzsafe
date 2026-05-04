# Story 1.2: docker-compose build-only

Status: review

## Story

As a desarrollador de despliegue,
I want que el `docker-compose.yml` sea auto-suficiente como orquestador de build-only,
so that pueda ejecutar `docker compose run --rm backend-build` directamente sin necesitar el script wrapper ni condiciones previas en el host.

## Acceptance Criteria

1. El entrypoint del servicio `backend-build` crea `dist/` si no existe antes de copiar el binario, por lo que `docker compose run --rm backend-build` funciona aunque el directorio `dist/` no exista en el host.
2. El `docker-compose.yml` expone un build arg `BUILDKIT_INLINE_CACHE` con valor por defecto `1`, permitiendo que el build genere metadatos de caché en la imagen para uso en CI/CD.
3. El `docker-compose.yml` no contiene ni referencia ningún servicio de runtime activo: ningún servicio expone puertos del backend (`8080` u otros), no hay servicio de base de datos, no hay servicio que ejecute `wsapi` como proceso de larga vida.
4. El `docker-compose.yml` tiene comentarios inline que explican: propósito del archivo (build-only), motivo de `network: host` en build (acceso a internet durante `go mod download`), y advertencia de que el runtime del backend se gestiona con PM2 en el host.
5. El comando `docker compose build backend-build && docker compose run --rm backend-build` completa exitosamente sin requerir el script wrapper, dejando `dist/wsapi` en el host.

## Tasks / Subtasks

- [x] Hacer el entrypoint auto-suficiente en creación de dist/ (AC: 1, 5)
  - [x] Agregar `mkdir -p /dist` al inicio del entrypoint del servicio `backend-build`
  - [x] Verificar que la secuencia `mkdir -p /dist && cp /out/wsapi /dist/wsapi.tmp && mv /dist/wsapi.tmp /dist/wsapi` sigue siendo atómica y limpia

- [x] Agregar build arg BUILDKIT_INLINE_CACHE (AC: 2)
  - [x] Agregar `args: BUILDKIT_INLINE_CACHE: "1"` bajo `build:` del servicio `backend-build`
  - [x] Confirmar que el Dockerfile actual (`docker/go/Dockerfile`) recibe y aplica correctamente el arg (si no lo usa explícitamente, el arg sigue siendo válido como metadato de imagen)

- [x] Auditar y documentar el compose (AC: 3, 4)
  - [x] Verificar que no hay servicios con `ports:`, bases de datos ni procesos runtime de wsapi en el archivo
  - [x] Agregar/mejorar comentario de cabecera explicando que este compose es build-only
  - [x] Agregar comentario inline en `network: host` explicando su propósito (go mod download necesita red)
  - [x] Agregar comentario explicando que el runtime del backend usa PM2 en el host (ver `docs/deploy-backend.md`)

- [x] Validar flujo E2E sin script wrapper (AC: 5)
  - [x] Confirmar que `docker compose build backend-build` compila sin errores
  - [x] Confirmar que `docker compose run --rm backend-build` deja `dist/wsapi` en el host
  - [x] Confirmar que el resultado es el mismo con o sin el script `scripts/build-backend.sh`

## Dev Notes

### Estado actual del docker-compose.yml

El archivo actual (modificado en story 1-1) es:

```yaml
# docker-compose.yml
# Uso: docker compose run --rm backend-build
# Deja el binario compilado en ./dist/wsapi

services:
  backend-build:
    build:
      context: ./backend
      dockerfile: ../docker/go/Dockerfile
      network: host
    volumes:
      - ./dist:/dist
    entrypoint: ["sh", "-c", "cp /out/wsapi /dist/wsapi.tmp && mv /dist/wsapi.tmp /dist/wsapi && echo 'Binario listo en ./dist/wsapi'"]
    restart: "no"
```

**Problema actual:** Si `dist/` no existe en el host, el volumen `./dist:/dist` hace que Docker cree el directorio automáticamente, pero el entrypoint hace `cp` directamente a `/dist/wsapi.tmp` — si el volumen montado tiene problemas de permisos o el directorio fue eliminado, puede fallar. Agregar `mkdir -p /dist` al inicio del entrypoint lo hace explícito y robusto.

### Contexto heredado de story 1-1

- `docker/go/Dockerfile` es build-only: no tiene `EXPOSE` ni `CMD`. Solo tiene un stage builder que deja el binario en `/out/wsapi`.
- La exportación ya es atómica: `cp /out/wsapi /dist/wsapi.tmp && mv /dist/wsapi.tmp /dist/wsapi`.
- `backend/.dockerignore` excluye `.env` y `.env.*` del build context.
- `scripts/build-backend.sh` es el wrapper recomendado (valida `.env`, rutas absolutas, `--no-cache`).
- Runtime usa PM2 con `--cwd backend/` para que `godotenv.Load()` encuentre `backend/.env`.

### Sobre BUILDKIT_INLINE_CACHE

`BUILDKIT_INLINE_CACHE=1` le dice al builder que incluya metadatos de caché dentro de la imagen Docker generada. Esto permite usar `cache_from:` en entornos CI/CD para acelerar builds subsecuentes. No afecta builds locales normales. Referencia: [Docker BuildKit inline cache](https://docs.docker.com/build/cache/backends/).

**Cómo agregarlo en compose:**
```yaml
build:
  args:
    BUILDKIT_INLINE_CACHE: "1"
```

El Dockerfile no necesita declarar este arg explícitamente; es un arg especial reconocido por BuildKit.

### Entorno requerido

- Docker 29.4.1 + Docker Compose v5.1.3 (confirmado en el host). BuildKit está habilitado por defecto.
- El build context es `./backend`; el Dockerfile está en `../docker/go/Dockerfile`.
- `network: host` en build es necesario para que `go mod download` acceda a internet sin problemas de DNS/proxy en algunos entornos Linux. Es una configuración de build (no de runtime del contenedor).

### Restricciones técnicas

- No cambiar el nombre del servicio `backend-build` (es referenciado por `scripts/build-backend.sh`).
- No agregar servicios de runtime, ports, ni base de datos (fuera de alcance por PRD FR13, NFR4).
- No modificar `docker/go/Dockerfile` en esta story (scope de story 1-1, ya completo).
- No modificar `scripts/build-backend.sh` en esta story salvo que sea necesario por cambio en compose.
- Mantener la exportación atómica (`cp + mv`) introducida en story 1-1.

### Archivos a modificar

- `docker-compose.yml` — único archivo de implementación de esta story.

### Verificación manual

```bash
# Desde la raíz del repo, sin el script wrapper:
docker compose build backend-build
rm -rf dist/   # simular directorio inexistente
docker compose run --rm backend-build
ls -la dist/wsapi   # debe existir y ser ejecutable
```

### References

- [Source: docker-compose.yml] — estado actual post-story-1-1
- [Source: docker/go/Dockerfile] — Dockerfile build-only (sin EXPOSE ni CMD)
- [Source: backend/.dockerignore] — excluye .env del build context
- [Source: scripts/build-backend.sh] — wrapper que llama al compose
- [Source: docs/deploy-backend.md] — documentación operativa completa
- [Source: _bmad-output/planning-artifacts/prd.md#FR13, NFR4] — BD en host, no en compose
- [Source: _bmad-output/implementation-artifacts/1-1-dockerfile-optimizado.md] — story previa

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

### Completion Notes List

- `docker-compose.yml` actualizado con: `mkdir -p /dist` en entrypoint, `BUILDKIT_INLINE_CACHE: "1"` en build args, comentarios build-only, comentario `network: host`, entrypoint en formato multilinea legible.
- Auditado: cero servicios con `ports:`, sin DB, sin runtime wsapi.
- Validado: `docker compose config` parsea correctamente; `docker compose config --services` retorna `backend-build` únicamente.

### Change Log

- 2026-05-03: Implementación completa story 1-2. docker-compose.yml auto-suficiente con mkdir -p /dist, BUILDKIT_INLINE_CACHE, comentarios build-only.

### File List

- `docker-compose.yml`
- `_bmad-output/implementation-artifacts/1-2-docker-compose-build-only.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
