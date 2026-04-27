---
stepsCompleted:
  - step-01-init
  - step-02-discovery
  - step-02b-vision
  - step-02c-executive-summary
  - step-03-success
  - step-04-journeys
  - step-05-domain
  - step-06-innovation
  - step-07-project-type
  - step-08-scoping
  - step-09-functional
  - step-10-nonfunctional
  - step-11-polish
  - step-12-complete
inputDocuments:
  - "_bmad-output/project-context.md"
  - "docs/bmad-project-rules.md"
documentCounts:
  productBriefs: 0
  research: 0
  brainstorming: 0
  projectDocs: 2
workflowType: 'prd'
releaseMode: single-release
classification:
  projectType: "SaaS B2B / API backend / Web admin panel"
  domain: "General business messaging / WhatsApp operations"
  complexity: "medium"
  projectContext: "brownfield"
---

# Product Requirements Document - wsapi

**Author:** Fulanito
**Date:** 2026-04-26

## Executive Summary

`wsapi` es una plataforma brownfield B2B compuesta por un backend API, un panel web administrativo y una integración operativa con WhatsApp. El objetivo inmediato de este PRD es corregir y documentar el build/deploy con Docker Compose para que el sistema use correctamente sus variables de entorno y pueda desplegarse en servidor de forma reproducible, funcional y mantenible.

La visión del producto es convertir `wsapi` en una herramienta administrativa confiable para que el personal interno de soporte configure empresas, sesiones de WhatsApp, parámetros operativos, envío de mensajes, difusión y revisión de actividad sin depender del equipo de desarrollo ni intervenir directamente base de datos, código o configuración manual. El producto debe reducir fricción operativa, acelerar respuesta ante clientes y permitir explicación clara de problemas desde el propio sistema.

Los usuarios principales son administradores internos y personal de soporte técnico. El sistema sirve a clientes empresariales que consumen la API para envío de mensajes por teléfono/WhatsApp, mientras soporte gestiona configuración, monitoreo y operación desde el panel administrativo.

### What Makes This Special

El valor diferencial de `wsapi` es trasladar tareas operativas que hoy requieren intervención de desarrollo hacia un panel administrativo seguro, documentado y trazable. El producto no se limita a enviar mensajes; habilita autonomía operativa para soporte en configuración de clientes, revisión de mensajes enviados, administración de sesiones, diagnóstico de problemas y mantenimiento del servicio.

El insight central es que el problema profundo no es únicamente técnico, sino organizacional: soporte depende de desarrollo para resolver tareas recurrentes. El sistema debe convertir esas tareas en flujos administrables, reduciendo tiempos de respuesta, errores manuales y conocimiento tribal.

## Project Classification

- **Project Type:** SaaS B2B / API backend / Web admin panel
- **Domain:** General business messaging / WhatsApp operations
- **Complexity:** Medium
- **Project Context:** Brownfield

## Success Criteria

### User Success

El personal interno de soporte puede usar `wsapi` como herramienta operativa sin depender de desarrollo para tareas recurrentes de revisión y configuración básica. En el alcance actual, soporte debe poder acceder al panel administrativo existente, revisar empresas y sesiones, y validar que el sistema desplegado funciona con la configuración correcta.

El momento de éxito para soporte ocurre cuando, ante un problema operativo, puede revisar el sistema y distinguir si la causa es una configuración visible/corregible por soporte o si realmente requiere intervención de desarrollo. El sistema debe reducir llamadas innecesarias a desarrollo y convertir información técnica recurrente en información operativa consultable.

### Business Success

El despliegue de `wsapi` en servidor debe ser repetible y documentado, reduciendo fricción para poner el servicio en operación. El éxito de negocio inicial se mide por la disminución de dependencias hacia desarrollo para despliegue, diagnóstico básico y revisión de configuración.

A 3 meses, el producto será exitoso si soporte realiza más revisiones desde el panel administrativo antes de escalar a desarrollo, las llamadas a desarrollo por configuración o dudas operativas disminuyen, y los problemas escalados llegan con mejor explicación del síntoma, configuración observada y posible causa.

### Technical Success

El build/deploy con Docker Compose debe construir y ejecutar correctamente usando variables de entorno reales, sin fallar por errores como `APP_PORT` no definido u otras configuraciones obligatorias faltantes cuando el archivo de entorno esté correctamente provisto.

El sistema debe validar y consumir de forma consistente las variables necesarias para backend, frontend, base de datos, JWT/API keys, WhatsApp/session config y URL pública del API. La configuración debe ser funcional, segura, reproducible y documentada. También debe verificarse que archivos como `.dockerignore`, Dockerfile, compose y estructura de `.env` no excluyan ni rompan los archivos o variables requeridas durante build/runtime.

### Measurable Outcomes

- `docker compose build` o el flujo de build definido termina sin errores.
- Los contenedores requeridos arrancan correctamente usando variables de entorno configuradas.
- El backend no falla al iniciar por variables obligatorias ausentes cuando `.env` está correctamente definido.
- `APP_PORT` y demás variables críticas son leídas desde configuración de entorno y no dependen de valores hardcodeados incorrectos.
- El frontend queda configurado con `NEXT_PUBLIC_API_URL` de forma compatible con el build/deploy.
- Existe documentación clara para repetir el despliegue en servidor.
- Soporte puede acceder al panel administrativo desplegado y usar las áreas existentes de empresas y sesiones.
- Quedan identificadas las causas de configuración que sí puede resolver soporte y las que deben escalarse a desarrollo.

## Product Scope

### MVP - Minimum Viable Product

El MVP de este PRD se enfoca en corregir y documentar el proceso de build/deploy Docker Compose para que `wsapi` pueda subirse al servidor y ejecutarse correctamente con variables de entorno.

Incluye:

- Revisión de Dockerfile, Docker Compose, `.dockerignore` y archivos de configuración relacionados.
- Corrección del uso de variables de entorno en backend y frontend durante build/runtime.
- Verificación de variables críticas como `APP_PORT`, configuración de base de datos, JWT/API keys, WhatsApp/session config y `NEXT_PUBLIC_API_URL`.
- Documentación del flujo correcto de build y deploy.
- Validación de que el panel administrativo existente queda accesible para soporte.
- Validación de uso básico de áreas existentes de empresas y sesiones.

### Growth Features (Post-MVP)

Después del MVP, el producto debe evolucionar el panel administrativo para dar mayor autonomía a soporte y acercarse a un CRM profesional de WhatsApp.

Incluye potencialmente:

- Mejoras en visualización y diagnóstico de empresas.
- Mejoras en administración y revisión de sesiones WhatsApp.
- Revisión más clara de mensajes enviados.
- Diagnóstico operativo guiado para distinguir problemas de configuración, sesión, cliente, API o desarrollo.
- Métricas más útiles para soporte y operación.
- Configuraciones administrables desde UI que hoy requieren base de datos, código o intervención técnica.

### Vision (Future)

La visión futura es convertir `wsapi` en una plataforma administrativa profesional para operación WhatsApp multiempresa, con capacidades tipo CRM orientadas a soporte interno: configuración, monitoreo, diagnóstico, trazabilidad de mensajes y explicación de problemas desde un único panel.

Quedan fuera de alcance por ahora:

- Sistema de soporte/tickets.
- Pagos.
- Campañas avanzadas.
- Funcionalidades CRM completas no necesarias para resolver el despliegue inicial.

## User Journeys

### Journey 1: Soporte configura una empresa cliente para operación WhatsApp

María, integrante del equipo de soporte, recibe la solicitud de habilitar una nueva empresa cliente para usar el servicio de envío de mensajes por WhatsApp. Antes, esta tarea podía requerir pedir ayuda a desarrollo para crear registros directamente en base de datos o validar configuraciones internas. Ahora, María entra al panel administrativo y registra la empresa con los datos requeridos por el sistema: RUC, nombre legal, nombre comercial, teléfono de contacto, estado activo y permisos aplicables.

Después, María asocia uno o más teléfonos/sesiones WhatsApp a la empresa. El sistema le permite identificar el número completo, país, estado de conexión y datos necesarios para operar la sesión. Si corresponde, se genera o administra una API key asociada a la empresa y teléfono, con nombre, scopes, estado activo y vigencia. La configuración queda lista para que el cliente consuma la API de envío.

El momento de valor ocurre cuando soporte completa la configuración sin tocar base de datos ni llamar a desarrollo. La nueva realidad es que la operación inicial del cliente queda resuelta desde el panel, con datos visibles, auditables y corregibles.

Capacidades reveladas:
- Gestión de empresas.
- Gestión de teléfonos/sesiones por empresa.
- Gestión de API keys por empresa/teléfono.
- Visibilidad de estado activo/inactivo.
- Auditoría básica de creación/actualización.

### Journey 2: Soporte revisa por qué un mensaje no fue enviado

Carlos, soporte técnico, recibe un reporte de un cliente: “no se están enviando mensajes”. En lugar de escalar inmediatamente a desarrollo, entra al panel y busca la empresa cliente. Luego revisa los mensajes asociados a esa empresa y teléfono. Encuentra el mensaje reportado y observa su `estado`, `error_reason`, timestamps, número destino, reintentos y último intento.

Si el error indica una causa operativa entendible —por ejemplo sesión desconectada, configuración inválida, API key inactiva o un fallo que soporte fue capacitado para interpretar— Carlos puede explicar el problema al cliente y tomar la acción correspondiente. Si el error indica una causa técnica fuera del alcance de soporte, escala a desarrollo con evidencia concreta: empresa, teléfono, mensaje, estado, error y momento del fallo.

El clímax del journey ocurre cuando Carlos puede responder: “esto está bien, esto falló por esta razón, esto lo podemos resolver nosotros” o “esto sí requiere desarrollo, y aquí está la información precisa”. El sistema reduce llamadas innecesarias y mejora la calidad de los escalamientos.

Capacidades reveladas:
- Búsqueda y revisión de mensajes por empresa/teléfono.
- Visualización clara de estado y razón de error.
- Timestamps y trazabilidad de intentos.
- Reintento de mensajes cuando aplique.
- Mensajes de error comprensibles para soporte capacitado.

### Journey 3: Soporte opera difusión sin depender de desarrollo

Ana, soporte técnico, necesita revisar una difusión enviada para una empresa. Entra al panel, localiza la empresa y revisa las difusiones asociadas a un teléfono/sesión. Observa el `reference_id`, total de destinatarios y estado de la difusión. Si existen resultados o mensajes fallidos asociados, revisa la causa para entender si el problema fue de configuración, sesión, destinatario o un error técnico.

Cuando el sistema lo permite, Ana reenvía o reintenta mensajes fallidos siguiendo criterios operativos definidos. Si no puede resolverlo, escala con detalles concretos y no con una descripción genérica del problema.

El valor ocurre cuando difusión deja de ser una operación opaca. Soporte puede revisar estado, explicar fallos y ejecutar acciones permitidas sin manipular datos directamente.

Capacidades reveladas:
- Revisión de difusiones por empresa/teléfono.
- Visibilidad de estado y total de difusión.
- Relación entre difusión, mensajes/resultados y errores.
- Reintento o reenvío controlado cuando aplique.
- Separación entre acciones permitidas para soporte y acciones reservadas.

### Journey 4: Administración revisa métricas y estado general del servicio

Lucía, responsable administrativa, no necesita operar sesiones ni modificar empresas. Su objetivo es evaluar el comportamiento general del servicio. Entra al dashboard y revisa métricas: actividad de mensajes, uso por empresa, sesiones activas/inactivas, errores frecuentes, volumen operativo y señales de rendimiento.

A diferencia de soporte, Lucía no debe administrar roles, módulos, usuarios o configuraciones sensibles. Su journey se enfoca en visibilidad, evaluación y toma de decisiones. El dashboard le permite identificar si el servicio está creciendo, si hay clientes con problemas recurrentes o si soporte está resolviendo más casos desde el sistema.

El momento de valor ocurre cuando administración puede evaluar el servicio sin pedir reportes manuales a desarrollo o soporte.

Capacidades reveladas:
- Dashboard administrativo.
- Métricas de mensajes, sesiones, empresas y errores.
- Acceso de solo evaluación para administración.
- Separación de permisos entre soporte, administración y desarrollo.

### Journey 5: Desarrollo despliega `wsapi` en servidor usando Docker Compose

Diego, del equipo de desarrollo, necesita subir `wsapi` al servidor. Prepara el archivo de variables de entorno requerido y ejecuta el flujo documentado de Docker Compose. El build debe usar correctamente la configuración necesaria para backend y frontend, incluyendo `APP_PORT`, conexión de base de datos, JWT/API keys, configuración de WhatsApp/sesiones y `NEXT_PUBLIC_API_URL`.

Durante el despliegue, Diego valida que los contenedores construyen y arrancan sin errores como `APP_PORT no está definido`. Si el sistema falla, la documentación y la estructura de configuración deben permitir identificar si el problema está en `.env`, Dockerfile, docker-compose, `.dockerignore`, variables de build o variables runtime.

El valor ocurre cuando el deploy es repetible y no depende de memoria o pasos implícitos. Desarrollo sigue siendo responsable del despliegue, pero el proceso queda claro, funcional y mantenible.

Capacidades reveladas:
- Build Docker Compose reproducible.
- Variables de entorno documentadas.
- Separación clara entre build-time y runtime env.
- Validación de configuración crítica.
- Documentación de despliegue para servidor.

### Journey Requirements Summary

Los journeys revelan que `wsapi` necesita cubrir dos frentes conectados:

1. **Operación interna desde panel administrativo**
   - Gestión de empresas.
   - Gestión de teléfonos/sesiones WhatsApp.
   - Revisión de mensajes enviados.
   - Visualización de errores comprensibles para soporte.
   - Reintentos o reenvíos controlados cuando aplique.
   - Revisión de difusión.
   - Dashboard y métricas para administración.
   - Separación clara de permisos: soporte opera empresas/sesiones/mensajes; administración evalúa métricas; desarrollo mantiene despliegue y problemas técnicos reales.

2. **Deploy confiable**
   - Docker Compose debe construir y ejecutar el sistema usando variables de entorno correctas.
   - La configuración debe evitar errores de arranque por variables obligatorias faltantes cuando `.env` existe y está correctamente definido.
   - Debe revisarse que Dockerfile, compose y `.dockerignore` no bloqueen variables o archivos requeridos.
   - El proceso debe quedar documentado para que desarrollo pueda repetirlo sin depender de pasos implícitos.

El MVP de este PRD prioriza el deploy confiable y la validación de acceso al panel existente. Los journeys de soporte, administración y operación WhatsApp orientan el crecimiento posterior del panel hacia una herramienta profesional tipo CRM de WhatsApp.

## Domain-Specific Requirements

### Compliance & Regulatory

No se identifican requisitos regulatorios específicos de alto impacto como salud, pagos, gobierno o educación. El sistema debe manejarse como una plataforma B2B de mensajería operativa con datos de empresas, teléfonos, sesiones WhatsApp, API keys y registros de mensajes enviados.

El sistema no debe capturar ni procesar mensajes personales, externos o conversaciones completas ajenas al uso de la API. El alcance de datos debe limitarse a mensajes enviados mediante `wsapi`, sus estados, errores, timestamps, difusión y metadatos necesarios para operación y soporte.

### Technical Constraints

- Las API keys, JWT/secrets y credenciales deben manejarse exclusivamente mediante configuración segura de entorno o mecanismos equivalentes; no deben quedar hardcodeadas en código, imágenes Docker o documentación pública.
- El build/deploy Docker debe distinguir claramente variables de build y variables runtime, especialmente para frontend (`NEXT_PUBLIC_API_URL`) y backend (`APP_PORT`, DB, JWT/API keys, WhatsApp/session config).
- Las sesiones WhatsApp deben exponer al panel solo la información necesaria para operación: estado, número, empresa asociada, QR cuando aplique y último estado de conexión.
- Los mensajes enviados deben registrar estado, error entendible, timestamps e intentos para permitir diagnóstico por soporte.
- Las acciones sensibles deben quedar protegidas por roles/permisos; soporte no debe administrar usuarios, roles, módulos o permisos globales.
- Deben mantenerse logs/auditoría suficientes para saber quién creó o modificó empresas, teléfonos/sesiones, API keys u otros recursos operativos.

### Integration Requirements

- El backend debe mantener compatibilidad con clientes que consumen la API de envío por empresa/teléfono.
- Las API keys deben estar asociadas a empresa y teléfono/sesión para preservar trazabilidad multiempresa.
- El panel administrativo debe consumir endpoints internos/administrativos sin duplicar reglas críticas en frontend.
- La integración WhatsApp debe permitir diagnosticar estado de sesión, desconexiones, QR y fallos de envío de forma visible para soporte.
- El despliegue en servidor debe soportar conexión correcta con base de datos MySQL/MariaDB y persistencia/session storage WhatsApp según la configuración existente.

### Risk Mitigations

- **Riesgo:** Variables de entorno faltantes o mal aplicadas rompen el arranque del sistema.  
  **Mitigación:** documentar `.env`, validar variables críticas y corregir Docker Compose/Dockerfile para usarlas correctamente.

- **Riesgo:** Soporte escala problemas sin información suficiente.  
  **Mitigación:** mostrar estado, `error_reason`, timestamps, empresa, teléfono y contexto mínimo de diagnóstico en el panel.

- **Riesgo:** Soporte accede a funciones sensibles.  
  **Mitigación:** aplicar separación de permisos; soporte opera empresas/sesiones/mensajes, administración revisa métricas y desarrollo conserva tareas técnicas críticas.

- **Riesgo:** Captura de información no deseada de WhatsApp.  
  **Mitigación:** limitar almacenamiento y visualización a mensajes enviados mediante la API y metadatos necesarios para operación.

- **Riesgo:** Configuración manual en base de datos genera errores o conocimiento tribal.  
  **Mitigación:** migrar progresivamente tareas recurrentes al panel administrativo y documentar el flujo operativo.

## SaaS B2B / API Backend / Web Admin Panel Specific Requirements

### Project-Type Overview

`wsapi` es un sistema B2B multiempresa con backend API, panel administrativo y operación WhatsApp. Cada empresa mantiene separación lógica de datos mediante tablas y relaciones propias para empresas, teléfonos/sesiones, API keys, mensajes y difusiones. El panel administrativo es usado por usuarios internos, mientras que las empresas cliente consumen la API desde sus propios sistemas.

El alcance inmediato se centra en corregir el build Docker del backend Go para que genere y ejecute correctamente el binario `wsapi` usando variables de entorno runtime y conectándose a la base de datos existente en el host. El frontend/panel administrativo forma parte del producto, pero el deploy Docker inmediato prioriza backend.

### Technical Architecture Considerations

El backend debe conservar la separación multiempresa existente: empresas, teléfonos/sesiones WhatsApp, API keys, mensajes y difusiones deben asociarse de forma consistente para evitar mezcla de datos entre clientes.

La autenticación se divide en dos modelos:

- **JWT:** usado para el panel administrativo y usuarios internos.
- **API key:** usada por empresas cliente para consumir endpoints de API desde sus sistemas.

El sistema debe mantener permisos diferenciados para usuarios internos:

- **super admin/root:** acceso completo y gestión de configuración sensible.
- **soporte:** operación de empresas, sesiones, mensajes y difusión, sin gestión de usuarios, módulos ni roles.
- **administración:** visualización de métricas, dashboard y evaluación operativa, sin operación sensible.

El deploy Docker inmediato debe construir únicamente el backend Go. El contenedor/resultante debe ejecutar el binario `wsapi` y conectarse a una base de datos MySQL/MariaDB existente en el host, usando variables de entorno runtime. No debe asumir que la base de datos vive dentro del mismo Docker Compose.

### Tenant Model

Cada empresa debe operar como tenant lógico. La separación multiempresa se basa en `empresa_id` y relaciones con teléfonos, API keys, mensajes y difusiones.

Requisitos:

- Cada empresa tiene sus propios teléfonos/sesiones.
- Cada empresa/teléfono puede tener API keys asociadas.
- Los mensajes deben quedar asociados a empresa y teléfono.
- Las difusiones deben quedar asociadas a empresa y teléfono.
- El panel administrativo debe evitar mostrar o mezclar datos entre empresas de forma accidental.
- Los diagnósticos de soporte deben partir de empresa → teléfono/sesión → mensaje/difusión.

### Permission Model / RBAC

El sistema debe aplicar separación de responsabilidades por rol.

Requisitos:

- Soporte puede crear/revisar empresas, revisar sesiones, revisar mensajes, reenviar mensajes cuando aplique y revisar/operar difusión.
- Soporte no puede asignar usuarios, módulos, roles ni permisos globales.
- Administración puede revisar dashboard, métricas y estado general del servicio.
- Administración no debe modificar configuraciones sensibles ni operar permisos.
- Super admin/root conserva capacidades de administración total.
- Las acciones sensibles deben quedar protegidas en backend, no solo ocultas en frontend.

### Billing / Usage Metrics

El billing actual se limita a métricas internas de uso y evaluación. No hay pagos, planes ni límites por empresa en el MVP actual.

Requisitos:

- Registrar y exponer métricas útiles para evaluación interna.
- Permitir análisis de volumen de mensajes, uso por empresa, sesiones y errores.
- Diseñar el modelo sin bloquear futura incorporación de planes o límites por empresa.
- Dejar explícitamente fuera de alcance pagos y cobro automatizado.

### API Endpoint Scope

Los clientes empresa consumen endpoints relacionados con:

- Envío de mensaje individual.
- Envío de difusión.
- Consulta de estado de mensaje.
- Consulta o validación de sesión/teléfono cuando aplique.

Requisitos:

- La API debe autenticarse con API key para clientes empresa.
- Cada request debe resolverse contra empresa/teléfono autorizado.
- Los errores deben ser claros y trazables para permitir diagnóstico desde el panel.
- No se deben romper nombres de endpoints ni payloads existentes sin decisión explícita.

### Docker Backend Build Requirements

El alcance Docker inmediato es construir el backend Go y ejecutar el binario `wsapi`.

Requisitos:

- El build debe generar el binario backend esperado: `wsapi`.
- El runtime debe leer variables de entorno, incluyendo `APP_PORT` y configuración de base de datos.
- El backend debe conectarse a la base de datos MySQL/MariaDB existente en el host.
- Docker Compose/Dockerfile no deben depender de variables hardcodeadas que contradigan `.env`.
- `.dockerignore` no debe excluir archivos necesarios para compilar o ejecutar el backend.
- El proceso debe quedar documentado para que desarrollo pueda repetir el build/deploy.
- El contenedor no debe asumir que debe levantar frontend ni base de datos dentro del mismo flujo de build backend, salvo que se documente otro modo explícitamente.

### Integration Requirements

Para este alcance no se requieren integraciones externas adicionales como storage externo, colas, gateways de pago o servicios de terceros fuera de WhatsApp y MySQL/MariaDB.

Integraciones consideradas:

- WhatsApp mediante la integración existente del backend.
- MySQL/MariaDB existente en el host.
- Panel administrativo vía JWT.
- API de clientes vía API key.

### Implementation Considerations

La implementación debe respetar la arquitectura existente del proyecto:

- Backend Go bajo `backend/`, con entrypoint real en `backend/main.go`.
- Imports internos bajo `wsapi/internal/...`.
- Configuración centralizada en `backend/internal/config`.
- Rutas API/admin en `backend/internal/http`.
- Persistencia en `backend/internal/storage`.
- WhatsApp en `backend/internal/whatsapp`.

Antes de cambiar Docker, se deben revisar:

- `docker-compose.yml`
- archivos bajo `docker/`
- Dockerfile del backend si existe
- `.dockerignore`
- `backend/main.go`
- `backend/internal/config/config.go`
- documentación o scripts de build existentes

Validaciones mínimas:

- Build backend exitoso.
- Binario `wsapi` generado/ejecutado correctamente.
- Runtime sin error por `APP_PORT` faltante cuando `.env` está correctamente definido.
- Conexión a base de datos del host validada.
- Documentación del flujo de build/deploy actualizada.

## Project Scoping

### Strategy & Philosophy

**Approach:** Single-release enfocado en confiabilidad de deploy backend.

Este PRD define una entrega única y acotada cuyo objetivo es resolver el problema inmediato de build/deploy Docker del backend Go. La estrategia es corregir la base técnica necesaria para subir `wsapi` al servidor de forma reproducible, usando variables de entorno runtime y conectando contra la base de datos existente en el host.

La filosofía de alcance es lean y operacional: entregar el mínimo cambio que permita a desarrollo desplegar correctamente el backend, sin mezclar este trabajo con mejoras amplias del panel administrativo o funcionalidades CRM futuras.

**Resource Requirements:** desarrollo backend/DevOps con conocimiento de Go, Docker Compose, variables de entorno, configuración de base de datos y estructura actual del proyecto.

### Complete Feature Set

**Core User Journeys Supported:**

- Desarrollo despliega `wsapi` en servidor usando Docker Compose.
- Soporte accede al sistema desplegado y usa el panel existente para empresas y sesiones.
- Soporte puede revisar información operativa existente una vez el backend está correctamente en ejecución.
- Administración conserva acceso al dashboard/métricas existentes si ya están disponibles en el sistema.

**Must-Have Capabilities:**

- Build Docker del backend Go ejecutado correctamente.
- Generación o disponibilidad del binario esperado `wsapi`.
- Ejecución del backend sin error por `APP_PORT` faltante cuando `.env` está correctamente configurado.
- Lectura correcta de variables runtime necesarias para backend.
- Conexión desde el contenedor/backend hacia la base de datos MySQL/MariaDB existente en el host.
- Revisión y corrección de `docker-compose.yml`, Dockerfile/backend build y `.dockerignore` cuando afecten el build o runtime.
- Documentación del flujo correcto de build/deploy.
- Documentación de variables requeridas para ejecutar el backend.
- Validación de arranque del backend con configuración real.
- No romper endpoints, payloads ni comportamiento existente del backend.

**Nice-to-Have Capabilities:**

- Validación automatizada adicional del entorno antes de arrancar.
- Mensajes de error de configuración más claros para variables faltantes.
- Script auxiliar para build/deploy backend.
- Checklist operativo para deploy en servidor.
- Revisión ligera de acceso al panel administrativo después del deploy.
- Preparar el modelo de métricas para futuros planes/límites sin implementarlos en esta entrega.

### Out of Scope

Quedan fuera de esta single-release:

- Reconstrucción o rediseño del panel administrativo.
- Funcionalidades CRM completas.
- Sistema de soporte/tickets.
- Pagos.
- Planes/límites por empresa.
- Campañas avanzadas.
- Nuevos módulos de usuarios, roles o permisos.
- Frontend Docker deploy completo, salvo ajustes mínimos si afectan la validación del backend existente.
- Base de datos dentro de Docker Compose.
- Integraciones externas adicionales fuera de WhatsApp y MySQL/MariaDB existente.

### Risk Mitigation Strategy

**Technical Risks:**

- **Riesgo:** el backend sigue fallando por variables runtime no cargadas.  
  **Mitigación:** revisar configuración centralizada, `docker-compose.yml`, Dockerfile y `.env`; validar específicamente `APP_PORT` y DB.

- **Riesgo:** el build funciona pero el contenedor no se conecta a la base de datos del host.  
  **Mitigación:** documentar el host/DSN correcto para Docker y validar conectividad desde runtime.

- **Riesgo:** `.dockerignore` excluye archivos necesarios para compilar o ejecutar.  
  **Mitigación:** auditar `.dockerignore` contra archivos requeridos por backend y build.

- **Riesgo:** se mezclan cambios de deploy con mejoras de producto.  
  **Mitigación:** mantener esta entrega limitada a backend Docker build/runtime y documentación.

**Market / Operational Risks:**

- **Riesgo:** soporte espera mejoras de panel en esta entrega.  
  **Mitigación:** comunicar que esta release habilita deploy confiable; mejoras CRM/panel quedan para PRDs o epics posteriores.

- **Riesgo:** el deploy queda funcionando solo por conocimiento implícito de desarrollo.  
  **Mitigación:** documentar comandos, variables, archivos requeridos y validaciones.

**Resource Risks:**

- **Riesgo:** falta tiempo para revisar todo el sistema.  
  **Mitigación:** priorizar ruta crítica: build backend, env runtime, DB host, arranque y documentación.

- **Riesgo:** aparecen problemas no relacionados durante deploy.  
  **Mitigación:** registrar hallazgos fuera de alcance y no bloquear la entrega salvo que impidan el arranque backend.

## Functional Requirements

### Backend Docker Build & Runtime Configuration

- FR1: Desarrollo puede construir el backend Go mediante el flujo Docker definido para el proyecto.
- FR2: Desarrollo puede obtener o ejecutar el binario backend esperado `wsapi` como resultado del build.
- FR3: Desarrollo puede ejecutar el backend construido usando variables de entorno runtime provistas por archivo o entorno del servidor.
- FR4: El sistema puede leer `APP_PORT` desde variables de entorno durante el arranque backend.
- FR5: El sistema puede leer configuración de base de datos desde variables de entorno durante el arranque backend.
- FR6: El sistema puede iniciar sin errores de variable obligatoria faltante cuando el entorno requerido está correctamente configurado.
- FR7: El sistema puede reportar de forma clara qué variable obligatoria falta cuando la configuración requerida no existe.
- FR8: Desarrollo puede verificar que los archivos necesarios para compilar y ejecutar el backend no estén excluidos del contexto Docker.
- FR9: Desarrollo puede seguir documentación del proyecto para repetir el build/deploy backend en servidor.

### Database Host Connectivity

- FR10: El backend puede conectarse desde su runtime Docker a una base de datos MySQL/MariaDB existente fuera del compose.
- FR11: Desarrollo puede configurar el host, puerto, usuario, contraseña y nombre de base de datos necesarios para la conexión backend.
- FR12: El sistema puede fallar de forma diagnóstica cuando la conexión a base de datos no está disponible o está mal configurada.
- FR13: El deploy backend puede ejecutarse sin requerir que Docker Compose levante una base de datos propia.

### Administrative Access & Internal User Roles

- FR14: Usuarios internos pueden autenticarse en el panel administrativo mediante el modelo JWT existente.
- FR15: Usuarios con rol super admin/root pueden acceder a capacidades administrativas completas.
- FR16: Usuarios de soporte pueden operar empresas, sesiones, mensajes y difusión dentro de los permisos permitidos.
- FR17: Usuarios de soporte no pueden administrar usuarios, módulos, roles ni permisos globales.
- FR18: Usuarios de administración pueden revisar dashboard, métricas y estado general del servicio.
- FR19: Usuarios de administración no pueden ejecutar acciones operativas sensibles ni modificar permisos globales.
- FR20: El backend puede proteger acciones sensibles por permisos de servidor, no solo por visibilidad en frontend.

### Multiempresa / Tenant Operations

- FR21: Soporte puede crear y revisar empresas cliente.
- FR22: Soporte puede asociar teléfonos/sesiones WhatsApp a una empresa.
- FR23: Soporte puede revisar el estado de teléfonos/sesiones por empresa.
- FR24: El sistema puede mantener separación lógica de datos por empresa.
- FR25: El sistema puede asociar mensajes a empresa y teléfono/sesión.
- FR26: El sistema puede asociar difusiones a empresa y teléfono/sesión.
- FR27: El sistema puede evitar mezcla accidental de datos entre empresas en operaciones administrativas y API.

### Client API Access

- FR28: Empresas cliente pueden consumir la API usando API keys asociadas a empresa y teléfono/sesión.
- FR29: Empresas cliente pueden enviar mensajes individuales mediante endpoints existentes.
- FR30: Empresas cliente pueden solicitar envíos de difusión mediante endpoints existentes.
- FR31: Empresas cliente pueden consultar estado de mensajes cuando el endpoint exista o aplique.
- FR32: El sistema puede resolver cada request de cliente contra la empresa/teléfono autorizado.
- FR33: El sistema puede rechazar requests con API key inválida, inactiva, revocada o no autorizada.
- FR34: El sistema puede mantener compatibilidad con endpoints y payloads existentes salvo decisión explícita.

### Message & Broadcast Operations

- FR35: Soporte puede revisar mensajes enviados desde el sistema.
- FR36: Soporte puede ver estado, razón de error, timestamps e intentos de un mensaje.
- FR37: Soporte puede interpretar errores operativos documentados para decidir si resuelve o escala.
- FR38: Soporte puede reenviar o reintentar mensajes cuando la capacidad exista y el permiso lo permita.
- FR39: Soporte puede revisar difusiones asociadas a una empresa y teléfono/sesión.
- FR40: Soporte puede revisar estado y total de destinatarios de una difusión.
- FR41: El sistema puede limitar el registro operativo a mensajes enviados mediante `wsapi` y metadatos necesarios.

### Metrics & Operational Visibility

- FR42: Administración puede consultar métricas internas de uso del servicio.
- FR43: Administración puede revisar volumen de mensajes por empresa cuando la información esté disponible.
- FR44: Administración puede revisar estado general de sesiones y errores frecuentes.
- FR45: El sistema puede exponer información suficiente para explicar problemas operativos recurrentes.
- FR46: El sistema puede conservar base funcional para futuros planes o límites por empresa sin implementarlos en esta release.

### Auditability & Traceability

- FR47: El sistema puede registrar o conservar información de auditoría para creación y modificación de empresas.
- FR48: El sistema puede registrar o conservar información de auditoría para creación y modificación de teléfonos/sesiones.
- FR49: El sistema puede registrar o conservar información de auditoría para API keys.
- FR50: Soporte puede escalar problemas a desarrollo con contexto trazable: empresa, teléfono, mensaje/difusión, estado, error y momento del fallo.

### Scope Control

- FR51: El sistema mantiene fuera de esta release pagos, soporte/tickets, campañas avanzadas y funcionalidades CRM completas.
- FR52: El sistema mantiene fuera de esta release la base de datos dentro de Docker Compose.
- FR53: El sistema mantiene fuera de esta release un deploy Docker completo del frontend, salvo ajustes mínimos necesarios para validar operación existente.

## Non-Functional Requirements

### Deployment Reliability

- NFR1: El build Docker del backend debe completarse de forma reproducible usando los archivos versionados del repositorio y la configuración documentada.
- NFR2: El backend desplegado no debe fallar por `APP_PORT` indefinido cuando el entorno requerido está correctamente configurado.
- NFR3: El runtime backend debe poder reiniciarse usando la misma configuración sin requerir pasos manuales no documentados.
- NFR4: El flujo Docker backend no debe requerir levantar una base de datos dentro de Docker Compose para funcionar contra la base de datos del host.
- NFR5: El build no debe depender de archivos locales no versionados, excepto archivos de entorno explícitamente documentados.

### Configuration & Operability

- NFR6: Todas las variables obligatorias para ejecutar el backend deben estar documentadas con nombre, propósito y momento de uso.
- NFR7: La configuración debe distinguir variables runtime de cualquier variable build-time cuando aplique.
- NFR8: Los errores por configuración faltante o inválida deben identificar la variable o componente afectado siempre que sea posible.
- NFR9: El proceso de deploy debe poder ser repetido por desarrollo siguiendo documentación sin depender de conocimiento implícito.
- NFR10: La configuración no debe hardcodear valores de entorno que cambian entre desarrollo, servidor o producción.

### Security

- NFR11: Secrets, API keys, JWT secrets, credenciales de base de datos y configuración sensible no deben quedar hardcodeados en código, Dockerfile, imágenes ni documentación pública.
- NFR12: El sistema debe preservar separación entre autenticación administrativa JWT y autenticación cliente por API key.
- NFR13: Las acciones sensibles deben validarse en backend según permisos/rol.
- NFR14: El sistema no debe exponer secretos completos en logs, errores o respuestas HTTP.
- NFR15: El almacenamiento operativo debe limitarse a mensajes enviados mediante `wsapi` y metadatos requeridos para soporte.

### Data Isolation & Multi-Tenancy

- NFR16: Las consultas y operaciones administrativas/API deben preservar separación lógica por empresa.
- NFR17: Mensajes, difusiones, teléfonos/sesiones y API keys deben permanecer trazables a su empresa correspondiente.
- NFR18: El diagnóstico de soporte no debe mezclar datos entre empresas.
- NFR19: Cualquier futura mejora de métricas o billing debe respetar separación multiempresa desde el diseño.

### Observability & Diagnostics

- NFR20: Los fallos de arranque por configuración o base de datos deben producir mensajes suficientes para diagnóstico por desarrollo.
- NFR21: Los errores de envío de mensajes deben conservar información entendible para soporte capacitado.
- NFR22: El sistema debe conservar timestamps y estado suficiente para reconstruir el flujo de un mensaje o difusión.
- NFR23: Los problemas escalados a desarrollo deben poder incluir empresa, teléfono/sesión, mensaje/difusión, estado, error y momento del fallo.

### Compatibility & Regression Control

- NFR24: Los cambios de deploy no deben romper endpoints existentes de API cliente ni rutas administrativas existentes.
- NFR25: Los cambios no deben modificar payloads existentes salvo decisión explícita documentada.
- NFR26: El backend debe seguir usando el entrypoint real del proyecto y el binario esperado `wsapi`.
- NFR27: El flujo debe respetar la arquitectura existente del backend Go y sus paquetes internos.
- NFR28: El proyecto debe mantener compatibilidad con la base de datos MySQL/MariaDB existente.

### Documentation Quality

- NFR29: La documentación de deploy debe incluir comandos, archivos relevantes, variables requeridas y validaciones mínimas.
- NFR30: La documentación debe explicar que la base de datos está en el host y no dentro del compose para este alcance.
- NFR31: La documentación debe incluir una forma de verificar que el backend arrancó correctamente.
- NFR32: La documentación debe identificar problemas comunes de configuración, incluyendo `APP_PORT` faltante o conexión DB incorrecta.
