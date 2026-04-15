---
stepsCompleted:
  - validate-prerequisites
  - design-epics
  - create-stories
  - final-validation
inputDocuments:
  - _bmad-output/planning-artifacts/prd.md
  - _bmad-output/project-context.md
status: "complete"
---

# wsapi - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for wsapi, decomposing the requirements from the PRD and Project Context into implementable stories.

## Requirements Inventory

### Functional Requirements

- FR-01 inicio de sesion WhatsApp por empresa
- FR-02 emision de QR por empresa
- FR-03 emision de estados por empresa
- FR-04 envio directo con validaciones
- FR-05 adjunto opcional con validaciones
- FR-06 difusion masiva con resultado individual
- FR-07 persistencia de mensajes
- FR-08 manejo concurrente seguro de clientes por empresa
- FR-09 manejo de logout/desconexion aislado
- FR-10 activacion/desactivacion de empresas

### NonFunctional Requirements

- NFR-01 confiabilidad por aislamiento multiempresa
- NFR-02 escalabilidad por particion logica por ruc_empresa
- NFR-03 rendimiento con concurrencia controlada
- NFR-04 observabilidad con logs y metricas por empresa
- NFR-05 seguridad de secretos
- NFR-06 mantenibilidad con contratos versionables
- NFR-07 integridad de datos con migraciones seguras
- NFR-08 testabilidad automatizada

### Additional Requirements

- Mantener compatibilidad funcional con baseline usqay (API, WebSocket, WhatsApp).
- No depender de Docker en esta fase.
- Preparar el sistema para niveles de produccion.

### UX Design Requirements

No aplica (servicio backend/API-first en esta fase).

### FR Coverage Map

- Epic 1: FR-01, FR-02, FR-03, FR-08, FR-09, NFR-01, NFR-03
- Epic 2: FR-04, FR-05, FR-07, NFR-06, NFR-08
- Epic 3: FR-06, FR-07, NFR-03, NFR-08
- Epic 4: FR-10, NFR-02, NFR-04, NFR-05, NFR-07

## Epic List

- Epic 1: Nucleo de Sesion Multiempresa y WebSocket de Estado
- Epic 2: API de Mensajeria Directa con Persistencia y Adjuntos
- Epic 3: Motor de Difusion Masiva Resiliente y Escalable
- Epic 4: Plataforma de Produccion (DB Migrations, Seguridad, Observabilidad)
- Epic 6: Frontend shadcn Refactor

## Epic 1: Nucleo de Sesion Multiempresa y WebSocket de Estado

Implementar el ciclo de vida completo de sesion WhatsApp por empresa con eventos en tiempo real compatibles y aislamiento concurrente.

### Story 1.1: Registro de sesiones multiempresa en memoria segura

As a backend platform,
I want administrar clientes por ruc_empresa con control concurrente,
So that evitemos colisiones y estados corruptos entre empresas.

**Acceptance Criteria:**

**Given** un servicio iniciado
**When** se registra o consulta un cliente por ruc_empresa
**Then** las operaciones son thread-safe
**And** no existe fuga de referencias entre empresas.

### Story 1.2: Inicio de sesion por WebSocket con evento init-session

As a operador de empresa,
I want iniciar sesion enviando init-session con mi ruc_empresa,
So that pueda establecer conectividad de WhatsApp para mi empresa.

**Acceptance Criteria:**

**Given** un websocket conectado y una empresa habilitada
**When** se recibe init-session
**Then** se dispara la inicializacion del cliente de esa empresa
**And** se retorna error-event claro si la solicitud es invalida o no autorizada.

### Story 1.3: Emision de QR y confirmacion de estado activo por empresa

As a operador de empresa,
I want recibir qr-ruc_empresa y active-ruc_empresa,
So that pueda autenticar y confirmar que mi sesion quedo lista.

**Acceptance Criteria:**

**Given** una sesion en inicializacion
**When** el proveedor emite QR y luego autenticacion/ready
**Then** el backend publica eventos por canal de empresa
**And** persiste el estado de servicio de la empresa.

### Story 1.4: Manejo de desconexion y logout con limpieza aislada

As a administrador de plataforma,
I want manejar desconexiones por empresa sin impacto global,
So that el sistema mantenga continuidad para otras empresas activas.

**Acceptance Criteria:**

**Given** una empresa desconectada o logout
**When** ocurre el evento de desconexion
**Then** se limpia su sesion y se notifica active-ruc_empresa con isActive false
**And** no se interrumpe el servicio de otras empresas.

## Epic 2: API de Mensajeria Directa con Persistencia y Adjuntos

Migrar y robustecer la API de envio directo con validaciones de entrada, adjuntos opcionales y persistencia confiable.

### Story 2.1: Endpoint de envio directo con validacion de payload

As a operador de empresa,
I want enviar mensaje directo con validacion estricta,
So that evite errores por datos incompletos o mal formados.

**Acceptance Criteria:**

**Given** una solicitud de envio directo
**When** faltan campos o formato invalido de destino
**Then** la API responde 4xx con detalle de validacion
**And** no intenta enviar al proveedor.

### Story 2.2: Soporte de adjuntos en envio directo con politicas de seguridad

As a operador de empresa,
I want adjuntar archivos permitidos en mensajes directos,
So that pueda enviar evidencia o documentos a mis destinatarios.

**Acceptance Criteria:**

**Given** un adjunto en la solicitud
**When** el tipo o tamano no cumple politica
**Then** la API rechaza con error explicito
**And** solo archivos permitidos son procesados y enviados.

### Story 2.3: Persistencia de mensajes directos y trazabilidad minima

As a administrador de plataforma,
I want registrar cada envio directo,
So that tenga trazabilidad y auditoria operativa.

**Acceptance Criteria:**

**Given** un envio directo exitoso
**When** el proveedor confirma el envio
**Then** se persiste registro con ruc_empresa, destino, timestamp y estado
**And** se retorna respuesta consistente al cliente API.

## Epic 3: Motor de Difusion Masiva Resiliente y Escalable

Implementar difusion multi-destino con control de concurrencia, reintentos y resultados por destinatario.

### Story 3.1: Endpoint de difusion con validacion de lista_difusion

As a operador de empresa,
I want enviar una lista de destinos y mensajes,
So that pueda ejecutar campanas desde una sola solicitud.

**Acceptance Criteria:**

**Given** una solicitud de difusion
**When** lista_difusion no es JSON valido o no es array
**Then** la API responde error de validacion
**And** no inicia procesamiento parcial.

### Story 3.2: Procesamiento por lotes con worker pool y limites por empresa

As a plataforma backend,
I want procesar difusion con concurrencia controlada,
So that mantengamos rendimiento estable sin saturar el servicio.

**Acceptance Criteria:**

**Given** una difusion con multiples destinatarios
**When** inicia el procesamiento
**Then** se ejecuta con limite configurable de workers
**And** se evita bloqueo del servidor principal.

### Story 3.3: Resultado granular por destinatario y persistencia parcial

As a operador de empresa,
I want ver estado enviado/error por cada destinatario,
So that pueda tomar acciones correctivas puntuales.

**Acceptance Criteria:**

**Given** una difusion en ejecucion
**When** algunos destinos fallan y otros completan
**Then** la respuesta final incluye matriz de resultados por destinatario
**And** los envios exitosos se persisten aunque existan fallas parciales.

## Epic 4: Plataforma de Produccion (DB Migrations, Seguridad, Observabilidad)

Preparar el sistema para operacion productiva con migraciones seguras, seguridad de configuracion y observabilidad operacional.

### Story 4.1: Framework de migraciones DB versionadas (up/down)

As a administrador de plataforma,
I want versionar cambios de esquema,
So that podamos evolucionar base de datos con rollback seguro.

**Acceptance Criteria:**

**Given** una nueva version de esquema
**When** se aplica migracion en staging
**Then** existe script up/down idempotente
**And** la aplicacion mantiene compatibilidad con datos existentes.

### Story 4.2: Endurecimiento de esquema e indices por carga operativa

As a administrador de plataforma,
I want optimizar tablas e indices clave,
So that consultas por empresa y periodo escalen correctamente.

**Acceptance Criteria:**

**Given** consultas por ruc_empresa, estado y fecha
**When** se analizan planes de ejecucion
**Then** existen indices efectivos para lecturas operativas
**And** constraints previenen inconsistencias frecuentes.

### Story 4.3: Observabilidad y seguridad baseline para produccion

As a equipo de operacion,
I want logs estructurados, metricas y manejo de secretos,
So that podamos detectar incidentes y operar con seguridad.

**Acceptance Criteria:**

**Given** el servicio ejecutandose
**When** ocurren eventos de sesion o envio
**Then** se generan logs estructurados con correlacion por empresa
**And** metricas y configuracion sensible cumplen politicas de seguridad.

## Epic 6: Frontend shadcn Refactor

Adopcion completa de shadcn/ui para el frontend. Reemplazar CSS manual con componentes shadcn, corregir theme light/dark.

### Stories

- Story 6-1: Corregir Theme Light/Dark y globals.css
- Story 6-2: Refactorizar Login con shadcn completo
- Story 6-3: Refactorizar Sidebar con shadcn
- Story 6-4: Refactorizar Dashboard con shadcn Cards y Tabs
- Story 6-5: Refactorizar Companies con shadcn Table y Select
- Story 6-6: Refactorizar Messages con shadcn Table y Tabs
- Story 6-7: Refactorizar Sessions con shadcn Cards y Dialog
- Story 6-8: Refactorizar Broadcasts con shadcn Table y Sheet
- Story 6-9: Refactorizar Settings con shadcn Tabs
