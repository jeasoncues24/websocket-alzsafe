---
title: 'Dockerización y despliegue productivo de WSAPI'
type: 'feature'
created: '2026-04-20'
status: 'done'
baseline_commit: 'dd8d6f89bfad038d81796d1fe9b9ebaac03f64f1'
context:
  - '_bmad-output/project-context.md'
  - '_bmad-output/implementation-artifacts/sprint-status.yaml'
  - 'Makefile'
  - 'docker/go/Dockerfile'
  - 'frontend/Dockerfile'
  - 'docker-compose.yml'
  - 'docker/check-ports.sh'
  - 'docker/production.md'
  - '.dockerignore'
  - 'frontend/.dockerignore'
  - 'docker/nginx/default.conf'
  - '.gitignore'
  - '.env.example'
  - 'frontend/.env.example'
  - 'frontend/next.config.ts'
  - 'internal/config/config.go'
  - 'frontend/README.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** WSAPI todavía depende de un flujo local/manual para ejecutar backend y frontend, y el repositorio no protege bien los contextos de build ni los archivos que no deberían versionarse. Esto dificulta un despliegue productivo repetible.

**Approach:** Crear una base Docker productiva con Compose para backend y frontend, agregar un chequeo de puertos ocupados que solo sugiera cambios, endurecer `.dockerignore` y `.gitignore`, y documentar la instalación de producción con los comandos reales.

## Boundaries & Constraints

**Always:**
- Mantener `APP_PORT` en backend y `NEXT_PUBLIC_API_URL` en frontend como contratos vigentes.
- Exponer backend y frontend con puertos separados.
- No auto-modificar archivos `.env`; el script de puertos solo debe sugerir.
- No hardcodear `localhost`, IPs ni puertos en código fuente.
- Incluir higiene de Docker build context y Git tracking para evitar artefactos indebidos.

**Ask First:**
- Si el puerto sugerido ya está ocupado, confirmar el nuevo valor antes de aplicarlo manualmente en `.env`.
- Si luego se quiere empaquetar también la base de datos como contenedor, ese cambio se trata como alcance adicional.

**Never:**
- Mezclar esta tarea con refactors funcionales del backend o frontend.
- Introducir cambios de runtime que rompan el flujo local existente sin una guía clara.
- Dejar `docker compose` sin instrucciones de instalación y arranque.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| HAPPY_PATH | Puertos libres y `.env`/`.env.local` configurados | `docker compose` levanta backend y frontend con build productivo | N/A |
| PORT_IN_USE | Puerto backend o frontend ocupado | Script reporta el puerto ocupado y sugiere alternativas | No modifica archivos |
| DIRTY_CONTEXT | Checkout con artefactos locales | `.dockerignore` y `.gitignore` evitan subir/mandar basura al build | Build context limpio |
| MISSING_ENV | Variables de entorno requeridas ausentes | Docs explican cómo completar los valores antes de levantar | Fallo claro y accionable |

</frozen-after-approval>

## Code Map

- `docker/go/Dockerfile` -- base de imagen productiva del backend
- `frontend/Dockerfile` -- base de imagen productiva del frontend
- `docker-compose.yml` -- orquestación de servicios productivos
- `docker/check-ports.sh` -- detección de puertos ocupados y sugerencias
- `.dockerignore` -- exclusión de contextos de build no deseados
- `frontend/.dockerignore` -- exclusión de artefactos del build del frontend
- `.gitignore` -- exclusión de artefactos temporales, logs, envs y overrides locales
- `Makefile` -- build de producción y tareas de soporte
- `frontend/next.config.ts` -- wiring de URL del backend para rewrites
- `internal/config/config.go` -- fuente de verdad de `APP_PORT`
- `.env.example` -- variables backend para producción
- `frontend/.env.example` -- variables frontend para producción
- `docker/production.md` -- guía de instalación y operación en producción

## Tasks & Acceptance

**Execution:**
 - [x] `docker/go/Dockerfile` -- endurecer build backend de producción -- asegurar binario reproducible y runtime mínimo
 - [x] `frontend/Dockerfile` -- crear imagen productiva del frontend -- permitir build y ejecución en Compose
 - [x] `docker-compose.yml` -- orquestar backend y frontend con puertos separados -- levantar la app completa en producción
 - [x] `docker/check-ports.sh` -- detectar puertos ocupados y sugerir alternativas -- facilitar la selección manual del puerto correcto
 - [x] `.dockerignore` -- excluir build contexts, logs, envs y basura temporal -- reducir riesgo de artefactos indebidos en imágenes
 - [x] `frontend/.dockerignore` -- excluir dependencias y artefactos del build frontend -- proteger el contexto de build del frontend
 - [x] `.gitignore` -- ampliar exclusiones para deployment y runtime local -- evitar commits de archivos no deseados
 - [x] `Makefile` -- agregar build de producción y targets de soporte -- simplificar el flujo de release
 - [x] `docker/production.md` -- documentar instalación y arranque en producción -- dejar un procedimiento operable

**Acceptance Criteria:**
- Given un checkout limpio, when se ejecuta el build de producción, then backend y frontend se generan o levantan sin depender de comandos dev.
- Given puertos ocupados, when se ejecuta el script de puertos, then se reportan los conflictos y se sugieren alternativas sin modificar archivos.
- Given la configuración de producción, when se revisan `.dockerignore` y `.gitignore`, then no se incluyen envs, logs, o artefactos de build innecesarios.
- Given la documentación de producción, when una persona sigue los pasos, then puede desplegar y validar la aplicación sin improvisar.

## Design Notes

- Mantener el alcance en una sola entrega operativa: empaquetado, orquestación, higiene de archivos y guía de instalación.
- No acoplar el script de puertos a escritura automática de `.env`; la sugerencia manual evita cambios involuntarios.
- El frontend debe seguir usando `NEXT_PUBLIC_API_URL` para rewrites; la Dockerización no debe romper ese contrato.

## Verification

**Commands:**
- `make build` -- expected: build local consistente antes del empaquetado
- `docker compose config` -- expected: Compose valida sin errores de sintaxis
- `docker compose up --build` -- expected: backend y frontend levantan con la configuración productiva
- `bash docker/check-ports.sh` -- expected: detecta puertos ocupados y propone alternativas
