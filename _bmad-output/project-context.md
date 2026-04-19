---
project_name: "wsapi"
user_name: "Fulanito"
date: "2026-04-17"
sections_completed:
  [
    "technology_stack",
    "language_rules",
    "framework_rules",
    "testing_rules",
    "quality_rules",
    "workflow_rules",
    "anti_patterns",
  ]
status: "complete"
rule_count: 30
optimized_for_llm: true
existing_patterns_found: 11
---

# Project Context for AI Agents

_Este archivo contiene reglas y patrones criticos que los agentes AI deben seguir al implementar codigo en este proyecto. Enfocarse en detalles no obvios para evitar errores de implementacion durante la migracion._

---

## Technology Stack & Versions

- Lenguaje principal: Go 1.22.12 ([go.mod](go.mod))
- Modulo: `wsapi`
- API HTTP: `net/http` (servidor basico en [main.go](main.go))
- Configuracion: variables de entorno via `os.Getenv` ([internal/config/config.go](internal/config/config.go))
- Integracion WhatsApp: `go.mau.fi/whatsmeow` (uso observado en [internal/whatsapp/manager.go](internal/whatsapp/manager.go))
- Persistencia/sesiones: estructura preparada en `internal/storage` y carpeta de runtime `sessions/`

## Critical Implementation Rules

### Language-Specific Rules (Go)

- Mantener la estructura por paquetes en [internal](internal): [internal/config](internal/config), [internal/http](internal/http), [internal/storage](internal/storage), [internal/whatsapp](internal/whatsapp).
- No introducir estado global mutable para sesiones/clients; usar managers con mutex (patron en [internal/whatsapp/manager.go](internal/whatsapp/manager.go)).
- Centralizar toda lectura de entorno en [internal/config/config.go](internal/config/config.go); no leer variables de entorno dispersas en handlers.
- Todo flujo de red o I/O debe manejar error de forma explicita; no ignorar errores retornados.
- Mantener nombres idiomaticos Go: paquetes lowercase, constructores NewX, metodos con receptor claro.

### Framework/Platform Rules (Migracion desde usqay)

- Fuente funcional legacy: rama usqay con Express + websocket + whatsapp-web.js.
- Casos de uso obligatorios a migrar:
  - Sesion por empresa (ruc_empresa) con eventos de estado, autenticacion, desconexion y QR.
  - Envio directo de mensajes con validacion de telefono/codigo postal y opcion de adjunto.
  - Difusion masiva con resultado por destinatario.
  - Persistencia de mensajes enviados y estados de usuario (activo/vinculado/servicio).
- Mantener compatibilidad funcional de eventos en tiempo real del legado:
  - init-session
  - qr-{ruc_empresa}
  - active-{ruc_empresa}
  - error-event
- En esta fase no depender de Docker para ejecutar ni validar comportamientos.

### Testing Rules

- Usar pruebas de tabla en Go para validaciones de negocio (telefono, codigo postal, construccion de chat id, parseo de payloads).
- Cubrir rutas felices y de error para cada caso de uso migrado: init-session, QR, desconexion, envio directo, difusion.
- Incluir pruebas de concurrencia para managers de sesiones/clientes cuando se agreguen operaciones de alta rotacion.
- Validar contractos de respuesta HTTP y eventos en tiempo real para no romper compatibilidad funcional del cliente actual.
- Mockear dependencias externas (WhatsApp y base de datos) en pruebas unitarias; evitar pruebas acopladas a servicios reales en CI.
- Si se agrega suite e2e, mantenerla separada de unitarias y ejecutable local sin Docker en esta etapa.

### Code Quality & Style Rules

- Mantener archivos y paquetes en lowercase, sin abreviaturas ambiguas.
- Evitar funciones largas en handlers; delegar logica de negocio a capas de servicio o dominio.
- Retornar errores enriquecidos y trazables; evitar mensajes genericos sin contexto operativo.
- Definir estructuras de request/response explicitas para endpoints en lugar de mapas dinamicos.
- No introducir dependencias nuevas sin justificar su necesidad para la migracion funcional.
- Mantener comentarios solo donde aclaren decisiones no obvias de concurrencia, reconexion o compatibilidad.
- No hardcodear `localhost`, IPs o puertos en codigo fuente. El backend debe leer valores de `.env` desde `internal/config/config.go`; el frontend debe leer su base URL desde `frontend/.env.local` y `frontend/.env.example`.

### Development Workflow Rules

- Priorizar migracion por capacidades: primero sesion/QR/estado, luego envio directo, luego difusion, luego endurecimiento tecnico.
- No mezclar cambios de infraestructura (Docker, despliegue, proxy) con PRs de migracion funcional.
- Validar cada capability migrada con evidencia minima: endpoint funcional, caso de error, y persistencia esperada.
- Mantener trazabilidad Node -> Go por caso de uso, citando el baseline usqay cuando aplique.
- Usar rama actual para implementacion Go sin reintroducir archivos del stack TypeScript eliminado.
- Si se agrega una variable de entorno nueva, documentarla en el `.env.example` correspondiente y en el README del proyecto afectado antes de usarla en codigo.

### Critical Don't-Miss Rules

- No migrar literalmente la estructura de carpetas Node; migrar comportamiento de negocio a diseño idiomatico Go.
- No mezclar en un mismo paso la migracion funcional con hardening de infraestructura.
- No eliminar cobertura de casos de desconexion/logout y limpieza de sesion.
- No perder la granularidad de respuesta por destinatario en difusion.
- No bloquear el proceso principal por operaciones de sesion de WhatsApp por usuario.
- No asumir que la ausencia actual de codigo en algunos archivos implica que la funcionalidad puede omitirse en la migracion.

### AI Working Rules

- Si falta contexto, preguntar antes de asumir o inventar rutas, estados, archivos o contratos.
- Fuente de verdad: `sprint-status.yaml` para estado del sprint, `project-context.md` para reglas de trabajo, `epic-*-context.md` para el epic activo.
- No crear alias de backend ni contratos en ingles salvo que el usuario lo pida explicitamente.
- Si se confirma un cambio de estado o alcance, actualizar el archivo correspondiente en la misma iteracion.
- Si una duda afecta al contrato o a la seguridad del flujo, detenerse y pedir confirmacion.
- Mantener los nombres de rutas, documentos y estados en espanol cuando el dominio del negocio ya esta estandarizado asi.

## Discovery Notes (Step-01)

- Proyecto en migracion de una base previa Node.js hacia Go.
- Prioridad actual: ejecucion local sin Docker.
- El repo contiene archivos de Docker ([docker-compose.yaml](docker-compose.yaml), `docker/`), pero quedan fuera de alcance en esta fase.
- Estado actual del arbol Go: varios paquetes inicializados con archivos placeholders vacios (`internal/http/handlers.go`, `internal/storage/mariadb.go`, `internal/storage/sessions.go`, `internal/whatsapp/client.go`, `internal/whatsapp/qr.go`).
- Convenciones detectadas:
  - Estructura por capas en `internal/` (`config`, `http`, `storage`, `whatsapp`)
  - Nombres de archivos en lowercase
  - Paquetes alineados al nombre de carpeta
  - Concurrencia protegida con `sync.RWMutex` para estado en memoria
  - Configuracion orientada a entorno (`APP_ENV`, `DB_*`)
  - Punto de entrada HTTP existente en [main.go](main.go)

## Estado del Proyecto (Sprint Tracking)

**IMPORTANTE:** Antes de iniciar cualquier trabajo, siempre consultar el sprint status.

- **Archivo de Tracking:** `_bmad-output/implementation-artifacts/sprint-status.yaml`
- **Ultima Actualizacion:** 2026-04-17T18:06:20Z

### Epics Completados:
| Epic | Estado | Stories |
|------|--------|---------|
| Epic 1: Sesiones WhatsApp | ✅ done | 4/4 |
| Epic 2: Mensajería Directa | ✅ done | 3/3 |
| Epic 3: Difusión Masiva | ✅ done | 3/3 |
| Epic 4: Infra | ✅ done | 3/3 |
| Epic 5: Panel Admin (Next.js) | ✅ done | 8/8 + 1 |
| Epic 6: Refactor shadcn/ui | ✅ done | 9/9 |
| Epic 7: Autenticación JWT | ✅ done | 4/4 |
| Epic 8: Gestión de Empresas | ✅ done | 8/8 |
| **Epic 8.5: Usuarios/Roles/Módulos** | ✅ done | 5/5 |
| **Epic 9: Mensajería Enriquecida** | ✅ done | 5/5 |

- **Epic 8 (empresa)**: completado en su alcance actual; `S-8.1` a `S-8.8` done, con `S-8.9` y `S-8.10` como futuros.

### Reglas de Tracking:
1. **Antes de iniciar cualquier tarea:** Consultar `sprint-status.yaml` para conocer el estado actual
2. **Al completar una story:** Actualizar su estado a "done" en el archivo YAML
3. **Al iniciar una story:** Cambiar estado de "ready-for-dev" a "in-progress"
4. **Al crear nuevos epics/stories:** Regenerar el sprint-status.yaml completo
5. **Al hacer planning:** Actualizar project-context.md con el estado actual

**Nota:** El agent debe会自动 detectar y sugerir el proximo paso basado en el sprint status.

## Legacy Baseline (usqay)

- Stack anterior: Express + TypeScript + websocket + mysql2 + whatsapp-web.js.
- Entrada principal legacy en [main.go](main.go) (rama actual Go minima) y equivalentes Node en usqay:
  - API Express con carga dinamica de rutas /api/v1/\*.
  - WebSocket para bootstrap de sesion WhatsApp por empresa.
  - Inicializacion de clientes activos al arrancar, basada en usuarios activos y vinculados.
- Ultimo commit en usqay: 201c85f; ajustes menores, no rediseno funcional.
- Objetivo actual: migracion de logica Node a Go para recuperar estabilidad operativa sin Docker por ahora.

---

## Usage Guidelines

Para agentes AI:

- Leer este archivo antes de implementar cualquier cambio.
- Seguir todas las reglas documentadas, priorizando compatibilidad funcional con el baseline usqay.
- En caso de duda, elegir la opcion mas conservadora para no romper sesiones ni eventos en tiempo real.
- Actualizar este archivo cuando aparezcan nuevos patrones estables en la implementacion Go.

Para humanos:

- Mantener este contexto conciso y centrado en reglas no obvias.
- Actualizar cuando cambie el stack, el flujo de sesion o la estrategia de pruebas.
- Revisar periodicamente y eliminar reglas que ya sean obvias o hayan quedado obsoletas.

Last Updated: 2026-04-17
