---
stepsCompleted:
  - validate-prerequisites
  - design-epics
  - create-stories
  - final-validation
inputDocuments:
  - _bmad-output/implementation-artifacts/epic-8-context.md
  - _bmad-output/planning-artifacts/api-key-ux-redesign.md
  - _bmad-output/implementation-artifacts/sprint-status.yaml
status: "ready-for-dev"
---

# wsapi - Epic Breakdown: Migraciones y Telefonos por Empresa

## Overview

Este documento define el bloque intermedio de trabajo que el equipo quiere resolver ahora:

1. Crear un documento maestro de migraciones antes de borrar o rehacer nada.
2. Estandarizar la gestion de telefonos por empresa desde el panel administrativo.
3. Separar el telefono de contacto de la empresa del telefono WhatsApp administrable.

## Requirements Inventory

### Functional Requirements

FR1: Documentar el esquema actual como fuente de verdad antes de eliminar migraciones.
FR2: Definir para cada tabla y columna si admite `NULL`, `NOT NULL` y cual es su `default` explicito.
FR3: Documentar indices, llaves primarias, llaves foraneas, uniques y relaciones criticas del esquema.
FR4: Documentar el uso correcto de la libreria de migraciones y su forma de ejecucion.
FR5: Definir un plan seguro para borrar y recrear migraciones sin duplicados ni cambios inconsistentes.
FR6: Crear CRUD de telefonos por empresa desde el panel administrativo.
FR7: Integrar la creacion, rotacion y revocacion de API keys/tokens dentro del flujo de telefonos.
FR8: Invalidar las API keys asociadas cuando se elimina un telefono.
FR9: Renombrar el campo de empresa `telefono` a `telefono_contacto` para evitar ambiguedad.
FR10: Ajustar frontend y contratos admin para consumir los endpoints de telefonos y no permitir editar telefonos dentro del PUT de empresas.

### NonFunctional Requirements

NFR1: El documento maestro debe ser claro, consistente y apto para operar como fuente unica de verdad.
NFR2: La documentacion de migraciones debe ser reproducible y no depender de conocimiento oral.
NFR3: Los indices documentados deben apoyar consultas reales por empresa, telefono, key prefix y fechas.
NFR4: El acceso a gestion de telefonos y tokens debe quedar restringido al panel administrativo.
NFR5: Los secretos de API key deben mostrarse una sola vez y almacenarse con hash.
NFR6: La recreacion de migraciones debe ser deterministica y verificable desde cero.

### Additional Requirements

- El repositorio ya usa `github.com/golang-migrate/migrate/v4` con driver MySQL y source file.
- El runner actual apunta a `internal/storage/migrations` y soporta `status`, `up` y `down` desde `go run . migrate`.
- El runner actual limpia una tabla legacy `schema_migrations` antes de aplicar migraciones nuevas.
- Antes de borrar migraciones viejas, debe existir un documento maestro con estructura de tablas aprobada.
- El documento de migraciones debe incluir por tabla una seccion de campos, tipos, nullability, default, indices y relaciones.
- En telefonos, el contrato admin ya tiene vistas relacionadas con empresas, telefonos y API keys; el trabajo nuevo debe ampliar ese flujo sin romperlo.

### UX Design Requirements

UX-DR1: La vista de API keys por telefono debe ser centrada en el numero WhatsApp, no en la empresa.
UX-DR2: La key o secreto debe mostrarse una sola vez al crear o rotar, con accion clara para copiar.
UX-DR3: La UI debe mostrar estados operativos del telefono y acciones visibles para crear, rotar y revocar keys.

### FR Coverage Map

FR1: Epic 1 - Story 1.1 documenta el esquema actual.
FR2: Epic 1 - Story 1.1 fija nullability y defaults por columna.
FR3: Epic 1 - Story 1.1 inventaria indices y relaciones; Story 1.2 documenta consultas criticas.
FR4: Epic 1 - Story 1.2 explica libreria y comandos de migracion.
FR5: Epic 1 - Story 1.3 define el plan de reset y recreacion.
FR6: Epic 2 - Story 2.2 crea el CRUD de telefonos por empresa.
FR7: Epic 2 - Story 2.3 integra el ciclo de vida de keys por telefono.
FR8: Epic 2 - Story 2.3 invalida keys al borrar el telefono.
FR9: Epic 2 - Story 2.1 renombra el contrato de contacto a `telefono_contacto`.
FR10: Epic 2 - Story 2.1 y Story 2.4 actualizan backend y frontend admin.

## Epic List

### Epic 1: Documento Maestro y Reset Controlado de Migraciones
Antes de borrar cualquier migracion, crear una fuente de verdad del esquema, del uso de la libreria de migraciones y del plan de recreacion limpia para evitar duplicados, indices inconsistentes y cambios improvisados.

**FRs covered:** FR1, FR2, FR3, FR4, FR5

### Epic 2: Gestion de Telefonos por Empresa
Permitir administrar telefonos WhatsApp por empresa desde el panel administrativo, separar el telefono de contacto del telefono administrable y conectar ese flujo con la creacion e invalidacion de API keys.

**FRs covered:** FR6, FR7, FR8, FR9, FR10

## Epic 1: Documento Maestro y Reset Controlado de Migraciones

Crear una documentacion canonica del esquema y del proceso de migraciones antes de eliminar o rehacer archivos viejos.

### Story 1.1: Documento maestro del esquema actual

Como equipo tecnico,
quiero un documento maestro del esquema actual,
para que la estructura de tablas quede fijada antes de borrar migraciones.

**Acceptance Criteria:**

**Given** la base de datos actual y el codigo existente
**When** se crea el documento maestro
**Then** cada tabla relevante queda listada con sus columnas, tipos, relaciones y proposito
**And** cada columna indica explicitamente si es nullable o not null
**And** cada columna indica su default de forma explicita
**And** el documento incluye al menos `empresas`, `telefonos`, `api_keys`, `api_key_usage_events`, `api_key_usage_daily`, `api_key_audit_events` y `schema_migrations`

**Given** una consulta critica del sistema
**When** se revisa el documento
**Then** se puede identificar que indice soporta esa consulta
**And** la documentacion no deja indices o relaciones importantes sin mencionar

**Given** el documento maestro no fue aprobado
**When** alguien intenta borrar migraciones
**Then** el borrado no debe considerarse valido

### Story 1.2: Guia operativa de la libreria de migraciones

Como desarrollador,
quiero una guia clara de la libreria de migraciones,
para que sepas como ejecutar y mantener las migraciones correctamente.

**Acceptance Criteria:**

**Given** la implementacion actual de migraciones
**When** alguien lee la guia operativa
**Then** entiende que se usa `golang-migrate/migrate/v4` con MySQL y source file
**And** entiende que la ruta de migraciones es `internal/storage/migrations`
**And** entiende como ejecutar `go run . migrate status`, `go run . migrate up` y `go run . migrate down`
**And** entiende como funciona la limpieza de la tabla legacy `schema_migrations`

**Given** se crea una migracion nueva
**When** se consulta la guia
**Then** queda claro el formato, orden y convencion de nombres a seguir
**And** queda claro como verificar estado sucio o fallas de ejecucion

### Story 1.3: Plan de recreacion limpia de migraciones

Como desarrollador,
quiero un plan de recreacion limpia de migraciones,
para que el reset del esquema sea controlado y repetible.

**Acceptance Criteria:**

**Given** el documento maestro ya fue aprobado
**When** se redacta el plan de recreacion
**Then** el plan define los pasos previos al borrado, incluyendo respaldo y validacion
**And** el plan indica que no se elimina ninguna migracion antes de tener la documentacion canonica lista
**And** el plan define como reconstruir el esquema desde cero en el orden correcto

**Given** las nuevas migraciones ya fueron recreadas
**When** se validan
**Then** no hay indices duplicados
**And** no hay diferencias accidentales de esquema respecto al documento maestro
**And** el flujo de up/down sigue siendo reproducible

## Epic 2: Gestion de Telefonos por Empresa

Administrar telefonos de WhatsApp por empresa desde el panel administrativo y conectar ese flujo con las API keys del telefono.

### Story 2.1: Renombrar el telefono de contacto de empresa

Como administrador,
quiero que el contrato de empresa use `telefono_contacto`,
para que no se confunda con el telefono WhatsApp administrable.

**Acceptance Criteria:**

**Given** el contrato de empresa para `POST` y `PUT`
**When** se envian los datos de contacto
**Then** el campo persistido y expuesto es `telefono_contacto`
**And** el contrato ya no usa `telefono` para el contacto de la empresa

**Given** el formulario de empresa en frontend
**When** se edita o crea una empresa
**Then** la etiqueta y el campo visible hacen referencia a `telefono_contacto`
**And** no sugieren que ese numero sea un telefono WhatsApp administrable

**Given** un `PUT` de empresa
**When** se intenta modificar telefonos de la empresa desde ese payload
**Then** el backend no acepta editar telefonos por ese camino

### Story 2.2: CRUD backend de telefonos por empresa

Como administrador del panel,
quiero crear, listar, editar y eliminar telefonos por empresa,
para que cada numero se gestione desde un flujo administrativo unico.

**Acceptance Criteria:**

**Given** un usuario sin permisos administrativos
**When** intenta acceder a los endpoints de telefonos
**Then** la operacion es rechazada con el codigo de autenticacion/autorizacion correspondiente

**Given** una empresa valida
**When** se crea un telefono
**Then** el telefono queda asociado a esa empresa
**And** el listado de telefonos devuelve solo los telefonos de esa empresa
**And** el numero se valida para evitar duplicados o inconsistencias basicas

**Given** un telefono existente
**When** se edita o elimina
**Then** el backend verifica ownership por empresa
**And** no permite modificar telefonos de otra empresa

**Given** la ruta admin de empresas
**When** el panel consulta telefonos
**Then** los endpoints estan disponibles solo bajo el contexto administrativo

### Story 2.3: Ciclo de vida de API keys por telefono

Como administrador,
quiero que las API keys vivan ligadas al telefono,
para que la integracion se invalide automaticamente cuando el numero se elimina.

**Acceptance Criteria:**

**Given** un telefono administrado desde panel admin
**When** se crea una API key para ese telefono
**Then** la key se guarda asociada a `telefono_id` y `empresa_id`
**And** el secreto se muestra una sola vez
**And** el secreto no vuelve a exponerse en respuestas posteriores

**Given** una API key existente
**When** se rota o revoca
**Then** la version anterior queda invalida
**And** el nuevo secreto sigue la misma regla de exposicion unica

**Given** un telefono con keys activas
**When** el telefono se elimina
**Then** todas sus API keys quedan invalidadas o revocadas
**And** ya no deben validar contra el middleware de API keys

**Given** las tablas de uso y auditoria
**When** se registran eventos
**Then** la informacion queda asociada al telefono y a la empresa correctamente

### Story 2.4: Frontend admin para telefonos y API keys

Como administrador del panel,
quiero consumir los endpoints de telefonos desde el frontend,
para que pueda gestionar telefonos y keys sin salir del panel administrativo.

**Acceptance Criteria:**

**Given** la vista de empresas en frontend
**When** abro el detalle o la seccion de telefonos
**Then** la UI consume los endpoints admin de telefonos
**And** permite ver la lista por empresa

**Given** un telefono seleccionado
**When** entro al detalle de API keys
**Then** la pantalla permite crear, rotar y revocar keys segun el flujo definido
**And** el modal de secreto solo aparece al crear o rotar

**Given** el formulario de empresa
**When** se muestra el campo de contacto
**Then** usa `telefono_contacto`
**And** no mezcla ese dato con la gestion de telefonos WhatsApp administrables

**Given** una accion de borrar telefono
**When** el backend invalida sus keys
**Then** el frontend refleja el estado correcto sin dejar acciones ambiguas o rotas
