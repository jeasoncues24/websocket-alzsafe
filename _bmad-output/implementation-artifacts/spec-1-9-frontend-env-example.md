---
title: 'Story 1.9 — frontend .env.example alineado al proyecto'
type: 'chore'
created: '2026-05-04'
status: 'done'
baseline_commit: '31fb2e6af82f2702ade5201159269b6b10c38761'
context:
  - '{project-root}/_bmad-output/project-context.md'
  - '{project-root}/docs/bmad-project-rules.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** El frontend ya consume variables de entorno reales (`NEXT_PUBLIC_API_URL`, `NEXT_INTERNAL_API_URL`, `PORT`), pero el proyecto está inconsistente: existe `frontend/.env.local.example`, el README habla de `frontend/.env.example`, y la story 1.9 actual está enfocada en PM2 en vez de documentar únicamente el archivo de entorno que el frontend necesita.

**Approach:** Alinear el frontend para que tenga un único archivo plantilla versionado llamado `frontend/.env.example`, mantener `frontend/.env.local` como archivo local no commiteado, y reemplazar la story 1.9 actual por una nueva story enfocada solo en este entregable de frontend y su sincronización con sprint status.

## Boundaries & Constraints

**Always:** limitar el cambio al frontend y a los artefactos BMad afectados por la renumeración de la story; usar exclusivamente variables que el frontend ya consume hoy; mantener `frontend/.env.local` como archivo local no versionado; dejar instrucciones claras en español; mantener coherencia con `frontend/README.md`, con el comportamiento actual de `frontend/lib/api.ts` y `frontend/next.config.ts`, y con el sprint status.

**Ask First:** si durante la implementación aparece la necesidad de agregar variables nuevas no usadas hoy por el frontend, cambiar el flujo de arranque de Next.js, o conservar simultáneamente dos plantillas distintas (`.env.example` y `.env.local.example`).

**Never:** incluir backend, Docker, PM2 o cambios de despliegue fuera del frontend; inventar variables no respaldadas por el código actual; dejar dos nombres de plantilla activos que generen confusión; mantener la story 1.9 antigua de PM2 como si siguiera siendo el objetivo vigente.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Plantilla principal | Repo frontend sin `frontend/.env.example` consistente | Existe `frontend/.env.example` con las variables reales del proyecto y comentarios de uso | N/A |
| Guía de uso | Un desarrollador nuevo revisa el README del frontend | El README indica copiar `frontend/.env.example` hacia `frontend/.env.local` y explica cada variable necesaria | Si encuentra referencias al nombre viejo `.env.local.example`, deben corregirse |
| Artefactos BMad | La story 1.9 actual apunta a PM2 | La story 1.9 pasa a describir únicamente el `.env.example` del frontend y el sprint status refleja esa story | Si queda la clave vieja o un archivo story antiguo activo, deben eliminarse o renombrarse para evitar ambigüedad |

</frozen-after-approval>

## Code Map

- `frontend/.env.example` -- plantilla oficial versionada con las variables reales del frontend
- `frontend/README.md` -- documentación del frontend; ya referencia `.env.example` y debe quedar consistente
- `frontend/.gitignore` -- define qué archivos `.env*` se ignoran; puede requerir excepción para versionar `.env.example`
- `frontend/lib/api.ts` -- usa `NEXT_PUBLIC_API_URL`; valida qué variables son realmente necesarias
- `frontend/next.config.ts` -- usa `NEXT_INTERNAL_API_URL` o `NEXT_PUBLIC_API_URL`; confirma contrato de configuración
- `_bmad-output/implementation-artifacts/1-9-frontend-env-example.md` -- nueva story 1.9; debe cumplir la estructura mínima de stories exigida por las reglas BMad del proyecto
- `_bmad-output/implementation-artifacts/sprint-status.yaml` -- estado del sprint; al terminar la re-derivación debe quedar alineado con la fase real de la story

## Tasks & Acceptance

**Execution:**
- [x] `frontend/.env.example` -- crear la plantilla canónica versionada usando las variables reales del frontend (`NEXT_PUBLIC_API_URL`, `NEXT_INTERNAL_API_URL`, `PORT`) con comentarios claros y un ejemplo seguro para entorno local por defecto -- evita que el equipo adivine qué debe poner en `.env.local`
- [x] `frontend/.gitignore` -- permitir explícitamente versionar `frontend/.env.example` sin dejar de ignorar `frontend/.env.local` y otros `.env` locales -- asegura que la plantilla no desaparezca del repositorio
- [x] `frontend/README.md` -- confirmar y ajustar la guía para copiar `frontend/.env.example` a `frontend/.env.local`, explicando el propósito de cada variable y manteniendo el alcance solo en frontend -- elimina contradicciones de documentación
- [x] `_bmad-output/implementation-artifacts/1-9-frontend-env-example.md` -- crear la nueva story 1.9 enfocada en `.env.example` del frontend, incluyendo epic padre, alcance, fuera de alcance, pruebas requeridas, riesgos/edge cases y dependencias/bloqueos según reglas del proyecto -- reemplaza la story vieja de PM2 por la nueva prioridad pedida por el usuario
- [x] `_bmad-output/implementation-artifacts/1-9-frontend-next-start-pm2.md` -- eliminar o renombrar el archivo anterior para que no existan dos stories compitiendo por el identificador 1.9 -- evita ambigüedad operativa
- [x] `_bmad-output/implementation-artifacts/sprint-status.yaml` -- actualizar la clave de la story 1.9 y dejar el estado final alineado con la fase real de revisión del artefacto entregado -- mantiene trazabilidad del backlog actual

**Acceptance Criteria:**
- Given un desarrollador que entra por primera vez al frontend, when abre `frontend/.env.example`, then puede identificar sin ambigüedad qué variables debe copiar a `frontend/.env.local` y con qué propósito general se usan.
- Given la documentación del frontend, when sigue los pasos de configuración, then el flujo recomendado apunta a `frontend/.env.example` como plantilla única y no menciona `.env.local.example` como referencia activa.
- Given los artefactos BMad de implementación, when se revisa la story 1.9 y `sprint-status.yaml`, then ambos reflejan la nueva story del `.env.example` del frontend, no la story anterior de PM2, y el estado final coincide entre la story y el sprint status.
- Given la nueva story 1.9, when se valida contra `docs/bmad-project-rules.md`, then incluye la estructura mínima requerida para stories del proyecto.

## Spec Change Log

- 2026-05-04 — bad_spec
  - Triggering finding: la primera revisión detectó que la nueva story 1.9 no cumplía la estructura mínima exigida por `docs/bmad-project-rules.md` y que `sprint-status.yaml` quedó desalineado con el estado real de la story/spec.
  - Amendment: se reforzó la spec para exigir una story 1.9 completa como artefacto BMad (epic padre, alcance, fuera de alcance, pruebas, riesgos/edge cases y dependencias/bloqueos) y para exigir alineación final entre story/spec/sprint status. También se explicitó que `.env.example` debe usar un ejemplo seguro para entorno local por defecto.
  - Avoids known-bad state: artefactos parcialmente correctos en frontend pero story incompleta y estados inconsistentes en el sprint.
  - KEEP: mantener el alcance solo en frontend; conservar la decisión de usar `frontend/.env.example` como plantilla oficial; preservar la actualización de `frontend/README.md`, `frontend/.gitignore` y el reemplazo de la story 1.9 de PM2 por la nueva story de entorno.

## Design Notes

La implementación no debe cambiar el contrato funcional del frontend; solo debe volver explícito y coherente el contrato de configuración ya existente. La fuente de verdad para las variables permitidas es el propio código: `frontend/lib/api.ts` exige `NEXT_PUBLIC_API_URL`, mientras `frontend/next.config.ts` admite `NEXT_INTERNAL_API_URL` como override server-side y, si no existe, usa `NEXT_PUBLIC_API_URL`. `PORT` ya está contemplado por el flujo de arranque documentado del frontend. El ejemplo comentado para `NEXT_INTERNAL_API_URL` debe ser seguro y entendible para entorno local por defecto, evitando inducir una configuración que falle fuera de Docker.

## Verification

**Commands:**
- `test -f /home/fulanito/development/wsapi/frontend/.env.example` -- expected: el archivo existe en la raíz del frontend
- `rg -n "\.env\.local\.example|\.env\.example" /home/fulanito/development/wsapi/frontend /home/fulanito/development/wsapi/_bmad-output/implementation-artifacts` -- expected: las referencias activas del frontend y de la story 1.9 quedan alineadas al nombre final decidido
- `python3 - <<'PY'
from pathlib import Path
p = Path('/home/fulanito/development/wsapi/_bmad-output/implementation-artifacts/sprint-status.yaml')
text = p.read_text()
print('1-9 updated' if '1-9-frontend-env-example' in text else 'missing new story key')
PY` -- expected: el sprint status contiene la nueva clave de la story 1.9

## Suggested Review Order

**Contrato de entorno del frontend**

- Muestra la plantilla oficial y los valores ejemplo seguros para local.
  [`.env.example:1`](../../frontend/.env.example#L1)

- Explica cómo copiar la plantilla y cuándo definir cada variable.
  [`README.md:3`](../../frontend/README.md#L3)

**Versionado de la plantilla**

- Permite commitear solo la plantilla y seguir ignorando `.env.local`.
  [`.gitignore:38`](../../frontend/.gitignore#L38)

**Trazabilidad BMad**

- Reemplaza la story 1.9 anterior con la nueva definición completa.
  [`1-9-frontend-env-example.md:1`](1-9-frontend-env-example.md#L1)

- Sincroniza la clave y el estado final de la story en el sprint.
  [`sprint-status.yaml:44`](sprint-status.yaml#L44)
