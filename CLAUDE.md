# wsapi - Guía de Comandos y Estilos (CLAUDE.md)

Este archivo sirve como referencia de comandos y convenciones del proyecto para agentes de IA (Antigravity/Claude).

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
