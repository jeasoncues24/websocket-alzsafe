# wsapi - Guía de Comandos y Estilos (CLAUDE.md)

Este archivo sirve como referencia de comandos y convenciones del proyecto para agentes de IA (Antigravity/Claude).

## Regla obligatoria: preguntas antes de cualquier plan o artifact (ASK_QUESTION_USER)

* **Gate obligatorio:** Antes de generar **cualquier plan** (modo plan / `ExitPlanMode`) o **cualquier artifact**, el agente DEBE presentar primero al usuario un bloque de preguntas de aclaración mediante `ASK_QUESTION_USER` (la herramienta `AskUserQuestion`) y esperar las respuestas.
* **Sin preguntas respondidas → sin plan ni artifact:** No se puede construir, mostrar ni ejecutar un plan o un artifact si no hay un ciclo de preguntas terminado. Si crees que no hay nada que preguntar, igual debes plantear al menos las decisiones abiertas (alcance, ambigüedades reales, criterios de aceptación, restricciones técnicas) y confirmarlas.
* **Cobertura de las preguntas:** alcance exacto, interpretaciones alternativas válidas, archivos/módulos candidatos, criterios de aceptación, restricciones (portabilidad MySQL/SQLite, `wsapi/internal/...`, seguridad de campos sensibles) y cualquier requisito subjetivo sin criterio claro.
* **Única excepción:** que el usuario diga explícitamente "sin preguntas" / "procede directo". En ese caso se registra esa instrucción y se avanza.

## Comandos de Verificación Comunes

### Backend (Go)
* **Ejecutar Pruebas:** `cd backend && go test ./...`
* **Compilar Proyecto:** `cd backend && go build ./...`

### Frontend (Next.js)
* **Ejecutar Linter:** `cd frontend && npm run lint`
* **Compilar Proyecto:** `cd frontend && npm run build` (Requiere desactivar sandbox o habilitar red para fuentes de Google)

### Docker & Entorno
* **Compilar Backend en Contenedor:** `docker compose run --rm backend-build`
* **Levantar Entorno Completo:** `docker compose up -d`

## Convenciones de Código y Estilo

### General
* **Idioma:** Mantener textos orientados al usuario final, mensajes de error y documentación del proyecto en **Español** (salvo código de Go/React e imports técnicos).
* **Flujo de Trabajo BMad:** Seguir rigurosamente la secuencia de planificación, arquitectura, historias y revisión de código (`bmad-code-review`).

### Base de Datos & SQL (MySQL / MariaDB)
* **Criterio Obligatorio:** Cualquier modificación SQL (migraciones, CREATE, ALTER, JOINs complejos o índices) debe pasar por el análisis de la skill `/sql-optimization`.
* **Portabilidad:** Evitar cláusulas específicas como `FOR UPDATE` si se realizan lecturas en hilos de prueba unitaria que corren bajo SQLite en memoria.
* **Solución Concurrencia:** Utilizar sentencias `UPDATE` universales con condicionales `CASE WHEN` para mantener la atomicidad transaccional de forma segura y portable entre MySQL y SQLite.

### Backend (Go)
* **Imports:** Rutas relativas comenzando siempre por `wsapi/internal/...`.
* **Estructura:** Respetar la separación entre dominio (`internal/domain`), persistencia (`internal/storage`) y capa HTTP (`internal/http`).
* **Seguridad:** Proteger campos sensibles (como secretos de webhook) etiquetándolos con `json:"-"` en las entidades de dominio para evitar fugas.

### Frontend (Next.js / shadcn/ui)
* **Sistema de diseño obligatorio:** Todo trabajo de UI parte de `DESIGN.md` y `PRODUCT.md` (raíz). Respetar las reglas *No-Rainbow*, *Grey-State*, *Warning-Light* y *Flat-At-Rest*: el color sale solo de tokens de `frontend/app/globals.css`, sin paletas por vista ni degradados/glassmorphism.
* **Skill `impeccable` — invocar las skills necesarias:** Al ejecutar `/impeccable` (o encarar cualquier tarea de diseño o refinamiento de interfaz) el agente DEBE, además, invocar las skills que dejan el resultado funcional:
  * **`shadcn`** — siempre que se agregue, modifique, depure o componga un componente de UI. El proyecto usa shadcn/ui (`frontend/components.json`, estilo `base-nova`, iconos `lucide`, alias `@/components/ui`). Resolver componentes con la skill, no copiarlos a mano.
  * **`migrate-radix-to-base`** — solo cuando se pida explícitamente migrar primitivas de Radix UI a Base UI.
* **Primitivas:** Vestir siempre sobre `frontend/components/ui/*` con `cn()` de `@/lib/utils`; no introducir una segunda librería de componentes ni una segunda familia tipográfica (Plus Jakarta Sans es la única).
* **Verificar:** `cd frontend && npm run lint` y `npm run build` antes de cerrar un cambio de UI.
