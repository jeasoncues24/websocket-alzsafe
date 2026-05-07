# Reglas BMad del Proyecto wsapi

Este documento define las reglas de trabajo para agentes BMad y cualquier colaborador que cree PRD, arquitectura, epics, stories, backlog, sprint status, código, QA o revisiones en este repositorio.

## Principios generales

- El proyecto se trabaja con flujo BMad disciplinado: descubrimiento → PRD → arquitectura → epics/stories → readiness → sprint planning → story → validación → desarrollo → review.
- Ningún agente debe saltar fases cuando existan artefactos requeridos faltantes.
- Toda decisión importante debe quedar trazable en artefactos Markdown o YAML bajo `_bmad-output/` o `docs/`.
- Las respuestas y documentos del proyecto deben estar en español salvo nombres técnicos, APIs o código.
- Los cambios deben respetar el código existente antes de introducir nuevos patrones.
- Las historias deben ser pequeñas, verificables y orientadas a valor.
- No se marca trabajo como terminado sin pruebas, build/lint aplicables o justificación explícita.

## Ubicación de artefactos

- Reglas del proyecto: `docs/bmad-project-rules.md`.
- Contexto técnico para IA: `_bmad-output/project-context.md`.
- Artefactos de planificación: `_bmad-output/planning-artifacts/`.
- Artefactos de implementación: `_bmad-output/implementation-artifacts/`.
- Sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`.
- Overrides de BMad: `_bmad/custom/`.

## Convención de nombres de archivos en `implementation-artifacts/`

Todos los archivos bajo `_bmad-output/implementation-artifacts/` siguen estas reglas sin excepción:

### Stories
```
{epic-n}-{story-n}-{nombre-slug}.md
```
Ejemplos: `2-1-eliminar-select-empresa-usuario-admin.md`, `2-2-5-limpiar-empresa-del-jwt-admin.md`

- El slug es kebab-case, sin prefijos `spec-` ni `story-`.
- La clave en `sprint-status.yaml` es exactamente el nombre del archivo sin la extensión `.md`.
- Las stories intermedias usan `{epic-n}-{story-n}-{decimal}` (ej. `2-2-5`). El filesystem puede ordenarlas distinto al orden lógico; el orden autoritativo es `sprint-status.yaml`.
- Nunca crear dos archivos para la misma story (uno de "tracking" y uno de "spec"). Un solo archivo por story.

### Epics
```
epic-{n}-{nombre-slug}.md
```
Ejemplo: `epic-2-mejoras-post-revision.md`

- El prefijo `epic-` distingue los archivos de resumen de epic de los archivos de story.
- La clave en `sprint-status.yaml` es `epic-{n}` (sin nombre slug).

### Documentos suplementarios
```
spec-{nombre-slug}.md
```
Ejemplos: `spec-readme-deploy-produccion.md`

- El prefijo `spec-` indica un documento de referencia o contexto que **no es una story** y **no se trackea en sprint-status**.
- No usar `spec-` para story files. Si un archivo tiene `spec-` y un ID de story (ej. `spec-1-9-*`), es un artefacto legacy que no debe replicarse.

### Resumen visual
```
implementation-artifacts/
├── 1-1-dockerfile-optimizado.md          ← story (epic 1, story 1)
├── 1-9-frontend-env-example.md           ← story (epic 1, story 9)
├── 2-1-eliminar-select-empresa-usuario-admin.md  ← story (epic 2, story 1)
├── 2-2-restaurar-empresas-desactivadas.md        ← story (epic 2, story 2)
├── 2-2-5-limpiar-empresa-del-jwt-admin.md        ← story intermedia (entre 2-2 y 2-3)
├── epic-2-mejoras-post-revision.md       ← epic overview
├── spec-readme-deploy-produccion.md      ← doc suplementario (no story)
└── sprint-status.yaml                    ← fuente de verdad de estados y orden
```

**Regla de oro:** si el archivo es una story, su nombre empieza con `{n}-{m}`. Si es un epic, empieza con `epic-`. Si es un doc de referencia, empieza con `spec-`. Nunca mezclar.

## Reglas de Epics

Cada epic debe incluir como mínimo:

1. Identificador estable: `Epic N` o `epic-N`.
2. Nombre claro orientado a resultado.
3. Objetivo de negocio o capacidad habilitada.
4. Alcance incluido.
5. Fuera de alcance.
6. Dependencias técnicas o funcionales.
7. Riesgos conocidos.
8. Criterios de éxito.
9. Lista de stories asociadas.
10. Condición de cierre del epic.

Reglas:

- Un epic debe entregar valor de negocio completo o habilitar explícitamente una capacidad indispensable.
- Evitar epics puramente técnicos si no están conectados a una capacidad del producto.
- Cada story debe pertenecer a exactamente un epic.
- Cada epic debe tener una retrospectiva opcional al cierre: `epic-N-retrospective`.
- Si cambia el alcance de un epic, actualizar PRD/arquitectura/backlog según impacto.

## Reglas de Stories

Cada story debe incluir como mínimo:

1. ID estable: `N.M` o slug equivalente `N-M-titulo`.
2. Estado actual.
3. Epic padre.
4. Contexto.
5. Objetivo.
6. Usuario/actor afectado si aplica.
7. Alcance.
8. Fuera de alcance.
9. Acceptance Criteria numerados: `AC1`, `AC2`, `AC3`, etc.
10. Tareas técnicas asociadas.
11. Archivos o áreas probablemente afectadas.
12. Pruebas requeridas.
13. Riesgos y edge cases.
14. Dependencias y bloqueos.
15. Notas de implementación.

Reglas:

- Los Acceptance Criteria deben ser observables y testeables.
- No usar criterios vagos como “funciona correctamente” sin detalle verificable.
- Una story no entra a desarrollo sin estado `ready-for-dev` o validación equivalente.
- La story debe ser suficientemente pequeña para implementarse y revisarse sin mezclar dominios no relacionados.
- Si durante desarrollo se descubre nuevo alcance, documentarlo y decidir si se agrega a la misma story o se crea otra.

## Estados oficiales

### Estados de epic

- `backlog`: epic definido pero no iniciado.
- `in-progress`: alguna story del epic está en progreso o revisión.
- `done`: todas las stories requeridas del epic están completadas.

### Estados de story

- `backlog`: story existe en epics/backlog, pero aún no está preparada para desarrollo.
- `ready-for-dev`: story file creado, completa y validada para desarrollo.
- `in-progress`: implementación activa.
- `review`: implementación terminada y pendiente de revisión.
- `changes-requested`: revisión encontró cambios obligatorios.
- `done`: aceptada, probada y cerrada.
- `blocked`: no puede avanzar por dependencia, decisión o incidente.
- `deferred`: pospuesta explícitamente.

### Estados de retrospectiva

- `optional`: disponible pero no requerida.
- `done`: retrospectiva ejecutada y registrada.

## Definition of Ready

Una story está lista para desarrollo solo si:

- Tiene epic padre y prioridad clara.
- Tiene Acceptance Criteria numerados y testeables.
- Tiene alcance y fuera de alcance explícitos.
- Tiene dependencias identificadas.
- Tiene impacto técnico razonablemente entendido.
- Tiene pruebas requeridas definidas.
- Tiene riesgos o edge cases documentados.
- No existen preguntas abiertas que bloqueen implementación.

## Definition of Done

Una story solo puede pasar a `done` si:

- Cumple todos los Acceptance Criteria.
- El código compila.
- Se ejecutaron pruebas aplicables o se documentó por qué no aplican.
- Se ejecutó lint/build aplicable cuando el área lo permita.
- No quedan errores conocidos sin registrar.
- Se actualizó el sprint status.
- La revisión de código no tiene hallazgos bloqueantes.
- Se documentaron cambios relevantes en artefactos si corresponde.

## Reglas de Backlog

- El backlog debe estar derivado de PRD, arquitectura y epics aprobados.
- No agregar stories sueltas sin epic padre.
- Las prioridades deben ser explícitas: alta, media o baja; o secuencia de sprint.
- Las dependencias deben registrarse en la story y reflejarse en el orden de sprint.
- Los bloqueos deben tener causa y condición de desbloqueo.
- Cualquier cambio grande durante implementación debe usar `bmad-correct-course`.

## Reglas de validación

Validaciones mínimas por fase:

1. PRD antes de cerrar planificación.
2. Arquitectura antes de cerrar solución técnica.
3. Epics/stories antes de sprint planning.
4. Implementation readiness antes de crear sprint final.
5. Story validation antes de desarrollo.
6. Code review antes de `done`.
7. QA/E2E cuando la story afecte flujos críticos de usuario o API.

## Reglas de Sprint Status

- El sprint status vive en `_bmad-output/implementation-artifacts/sprint-status.yaml`.
- Debe reflejar todos los epics y stories existentes en los archivos de epics.
- No debe contener stories inexistentes.
- Los estados existentes más avanzados no deben degradarse sin decisión explícita.
- Después de crear, validar, desarrollar o revisar una story, se debe sincronizar el sprint status.
- Si no hay epics aún, el sprint status puede existir como base inicial vacía, pero debe regenerarse con `bmad-sprint-planning` cuando existan epics.

## Reglas técnicas del proyecto

- Backend: Go module `wsapi`, Go `1.25.0`.
- Frontend: Next.js `16.2.3`, React `19.2.4`, TypeScript strict.
- API frontend requiere `NEXT_PUBLIC_API_URL` en `frontend/.env.local`.
- El frontend usa alias `@/*`.
- El backend usa paquetes internos bajo `backend/internal/` por dominio técnico: `config`, `http`, `storage`, `domain`, `auth`, `whatsapp`, `metrics`.
- Mantener nombres y respuestas API existentes en español cuando ya existan así.
- No introducir frameworks nuevos sin justificación en arquitectura.
- Preferir cambios pequeños, trazables y testeables.

## Comandos de verificación recomendados

Frontend:

```bash
cd frontend && npm run lint
cd frontend && npm run build
```

Backend:

```bash
cd backend && go test ./...
cd backend && go build ./...
```

Docker backend build:

```bash
docker compose run --rm backend-build
```
